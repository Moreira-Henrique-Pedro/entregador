package writers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

// defaultPackageLabel fills the package type in the message when the delivery has none,
// since WhatsApp templates do not accept empty variables.
const defaultPackageLabel = "encomenda"

type ProcessNotifyDelivery struct {
	deliveryRepository interfaces.DeliveryRepositoryPort
	residentRepository interfaces.ResidentRepositoryPort
	notifier           notifier.NotifierPort
}

func NewProcessNotifyDelivery(
	deliveryRepository interfaces.DeliveryRepositoryPort,
	residentRepository interfaces.ResidentRepositoryPort,
	notifier notifier.NotifierPort,
) *ProcessNotifyDelivery {
	return &ProcessNotifyDelivery{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
		notifier:           notifier,
	}
}

func (w *ProcessNotifyDelivery) Handle(ctx context.Context, command *commands.ProcessNotifyDeliveryCommand) error {
	logger := logger.GetLoggerFromContext(ctx).With(
		"delivery_id", command.DeliveryID,
		"notification_type", string(command.NotificationType),
	)
	logger.Info("Processing ProcessNotifyDelivery command", "command_id", command.CommandID)

	if !command.NotificationType.IsValid() {
		logger.Warn("Discarding ProcessNotifyDelivery command with invalid notification type")
		return nil
	}

	delivery, err := w.deliveryRepository.FindByDeliveryID(ctx, command.DeliveryID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Delivery not found for notification")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find delivery: deliveryID=%s: %w", command.DeliveryID, err)
	}

	if reason, skip := shouldSkipNotification(delivery, command.NotificationType); skip {
		logger.Info("Skipping delivery notification", "reason", reason)
		return nil
	}

	recipients, err := w.resolveRecipients(ctx, delivery)
	if err != nil {
		return fmt.Errorf("failed to resolve notification recipients: deliveryID=%s: %w", delivery.DeliveryID, err)
	}
	if len(recipients) == 0 {
		logger.Warn("No resident with phone to notify", "apartment", delivery.Apartment, "resident_id", delivery.ResidentID)
	}

	sent := 0
	for _, resident := range recipients {
		err := w.notifier.Send(ctx, buildNotification(command.NotificationType, resident, delivery))
		if errors.Is(err, notifier.ErrInvalidRecipient) {
			// A bad phone must not block the other residents nor retry forever.
			logger.Warn("Resident phone rejected by the provider", "resident_id", resident.ResidentID, "error", err.Error())
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to notify resident: residentID=%s: %w", resident.ResidentID, err)
		}
		sent++
	}

	if err := w.markAsNotified(ctx, delivery.DeliveryID, command.NotificationType); err != nil {
		return fmt.Errorf("failed to mark delivery as notified: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	logger.Info("Delivery notified", "recipients", len(recipients), "sent", sent)

	return nil
}

func shouldSkipNotification(delivery *entities.Delivery, notificationType notifier.NotificationType) (string, bool) {
	switch notificationType {
	case notifier.NotificationTypeDeliveryArrived:
		if !delivery.ArrivalNotifiedAt.IsZero() {
			return "arrival already notified", true
		}
		if delivery.Status != entities.DeliveryStatusPending {
			return "delivery already picked up", true
		}
	case notifier.NotificationTypeDeliveryPickedUp:
		if !delivery.PickupNotifiedAt.IsZero() {
			return "pickup already notified", true
		}
		if delivery.Status != entities.DeliveryStatusDeleted {
			return "delivery not picked up yet", true
		}
	}
	return "", false
}

func (w *ProcessNotifyDelivery) markAsNotified(ctx context.Context, deliveryID string, notificationType notifier.NotificationType) error {
	if notificationType == notifier.NotificationTypeDeliveryPickedUp {
		return w.deliveryRepository.MarkPickupAsNotified(ctx, deliveryID)
	}
	return w.deliveryRepository.MarkArrivalAsNotified(ctx, deliveryID)
}

// resolveRecipients returns the delivery resident, or the apartment's primary
// resident when the delivery belongs to the "other" resident, keeping only
// those with a phone.
func (w *ProcessNotifyDelivery) resolveRecipients(ctx context.Context, delivery *entities.Delivery) ([]*entities.Resident, error) {
	var candidates []*entities.Resident

	resident, err := w.residentRepository.FindByResidentID(ctx, delivery.ResidentID)
	switch {
	case err == nil && !resident.IsOther():
		candidates = []*entities.Resident{resident}
	case err == nil, errors.Is(err, entities.ErrEntityNotFound):
		primary, err := w.findPrimary(ctx, delivery.Apartment)
		if err != nil {
			return nil, err
		}
		if primary != nil {
			candidates = []*entities.Resident{primary}
		}
	default:
		return nil, err
	}

	recipients := make([]*entities.Resident, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.IsOther() && candidate.Phone != "" {
			recipients = append(recipients, candidate)
		}
	}
	return recipients, nil
}

// findPrimary returns the apartment's primary resident, promoting one first when
// the apartment has none (residents stored before primary existed, or a promotion
// interrupted by a failure). It returns nil when the apartment has no resident.
func (w *ProcessNotifyDelivery) findPrimary(ctx context.Context, apartment string) (*entities.Resident, error) {
	primary, err := w.findPrimaryInApartment(ctx, apartment)
	if err != nil || primary != nil {
		return primary, err
	}

	if err := w.residentRepository.EnsurePrimaryResident(ctx, apartment); err != nil {
		return nil, err
	}
	return w.findPrimaryInApartment(ctx, apartment)
}

func (w *ProcessNotifyDelivery) findPrimaryInApartment(ctx context.Context, apartment string) (*entities.Resident, error) {
	residents, err := w.residentRepository.FindByApartment(ctx, apartment)
	if err != nil {
		return nil, err
	}
	for _, resident := range residents {
		if resident.IsPrimary() {
			return resident, nil
		}
	}
	return nil, nil
}

// buildNotification builds the message for the resident. Template variables, in order:
// {{1}} resident name, {{2}} apartment, {{3}} package type.
func buildNotification(notificationType notifier.NotificationType, resident *entities.Resident, delivery *entities.Delivery) notifier.Notification {
	packageLabel := delivery.PackageType
	if packageLabel == "" {
		packageLabel = defaultPackageLabel
	}

	return notifier.Notification{
		Type:      notificationType,
		Phone:     resident.Phone,
		Body:      buildMessageBody(notificationType, resident, delivery),
		Variables: []string{resident.Name, delivery.Apartment, packageLabel},
	}
}

func buildMessageBody(notificationType notifier.NotificationType, resident *entities.Resident, delivery *entities.Delivery) string {
	packageSuffix := ""
	if delivery.PackageType != "" {
		packageSuffix = fmt.Sprintf(" (%s)", delivery.PackageType)
	}

	if notificationType == notifier.NotificationTypeDeliveryPickedUp {
		return fmt.Sprintf("Olá, %s! A entrega%s do apartamento %s foi retirada na portaria.", resident.Name, packageSuffix, delivery.Apartment)
	}

	message := fmt.Sprintf("Olá, %s! Chegou uma entrega para o apartamento %s%s. Retire na portaria.", resident.Name, delivery.Apartment, packageSuffix)
	if delivery.Urgency != "" {
		message += fmt.Sprintf(" Urgência: %s.", delivery.Urgency)
	}
	return message
}
