package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

// WhatsApp templates do not accept empty variables.
const defaultPackageLabel = "encomenda"

type NotifyDelivery struct {
	deliveryRepository out.DeliveryRepository
	residentRepository out.ResidentRepository
	notifier           out.Notifier
}

func NewNotifyDelivery(
	deliveryRepository out.DeliveryRepository,
	residentRepository out.ResidentRepository,
	notifier out.Notifier,
) *NotifyDelivery {
	return &NotifyDelivery{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
		notifier:           notifier,
	}
}

func (uc *NotifyDelivery) Execute(ctx context.Context, deliveryID string, notificationType domain.NotificationType) error {
	logger := logger.GetLoggerFromContext(ctx).With(
		"delivery_id", deliveryID,
		"notification_type", string(notificationType),
	)

	if !notificationType.IsValid() {
		logger.Warn("Discarding notification with invalid type")
		return nil
	}

	delivery, err := uc.deliveryRepository.FindByDeliveryID(ctx, deliveryID)
	if errors.Is(err, domain.ErrEntityNotFound) {
		logger.Warn("Delivery not found for notification")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find delivery: deliveryID=%s: %w", deliveryID, err)
	}

	if reason, skip := shouldSkipNotification(delivery, notificationType); skip {
		logger.Info("Skipping delivery notification", "reason", reason)
		return nil
	}

	recipients, err := uc.resolveRecipients(ctx, delivery)
	if err != nil {
		return fmt.Errorf("failed to resolve notification recipients: deliveryID=%s: %w", delivery.DeliveryID, err)
	}
	if len(recipients) == 0 {
		logger.Warn("No resident with phone to notify", "apartment", delivery.Apartment, "resident_id", delivery.ResidentID)
	}

	sent := 0
	for _, resident := range recipients {
		err := uc.notifier.Send(ctx, buildNotification(notificationType, resident, delivery))
		if errors.Is(err, out.ErrInvalidRecipient) {

			logger.Warn("Resident phone rejected by the provider", "resident_id", resident.ResidentID, "error", err.Error())
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to notify resident: residentID=%s: %w", resident.ResidentID, err)
		}
		sent++
	}

	if err := uc.markAsNotified(ctx, delivery.DeliveryID, notificationType); err != nil {
		return fmt.Errorf("failed to mark delivery as notified: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	logger.Info("Delivery notified", "recipients", len(recipients), "sent", sent)

	return nil
}

func shouldSkipNotification(delivery *domain.Delivery, notificationType domain.NotificationType) (string, bool) {
	switch notificationType {
	case domain.NotificationTypeDeliveryArrived:
		if !delivery.ArrivalNotifiedAt.IsZero() {
			return "arrival already notified", true
		}
		if delivery.Status != domain.DeliveryStatusPending {
			return "delivery already picked up", true
		}
	case domain.NotificationTypeDeliveryPickedUp:
		if !delivery.PickupNotifiedAt.IsZero() {
			return "pickup already notified", true
		}
		if delivery.Status != domain.DeliveryStatusDeleted {
			return "delivery not picked up yet", true
		}
	}
	return "", false
}

func (uc *NotifyDelivery) markAsNotified(ctx context.Context, deliveryID string, notificationType domain.NotificationType) error {
	if notificationType == domain.NotificationTypeDeliveryPickedUp {
		return uc.deliveryRepository.MarkPickupAsNotified(ctx, deliveryID)
	}
	return uc.deliveryRepository.MarkArrivalAsNotified(ctx, deliveryID)
}

func (uc *NotifyDelivery) resolveRecipients(ctx context.Context, delivery *domain.Delivery) ([]*domain.Resident, error) {
	var candidates []*domain.Resident

	resident, err := uc.residentRepository.FindByResidentID(ctx, delivery.ResidentID)
	switch {
	case err == nil && !resident.IsOther():
		candidates = []*domain.Resident{resident}
	case err == nil, errors.Is(err, domain.ErrEntityNotFound):
		primary, err := uc.findPrimary(ctx, delivery.Apartment)
		if err != nil {
			return nil, err
		}
		if primary != nil {
			candidates = []*domain.Resident{primary}
		}
	default:
		return nil, err
	}

	recipients := make([]*domain.Resident, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.IsOther() && candidate.Phone != "" {
			recipients = append(recipients, candidate)
		}
	}
	return recipients, nil
}

func (uc *NotifyDelivery) findPrimary(ctx context.Context, apartment string) (*domain.Resident, error) {
	primary, err := uc.findPrimaryInApartment(ctx, apartment)
	if err != nil || primary != nil {
		return primary, err
	}

	if err := uc.residentRepository.EnsurePrimaryResident(ctx, apartment); err != nil {
		return nil, err
	}
	return uc.findPrimaryInApartment(ctx, apartment)
}

func (uc *NotifyDelivery) findPrimaryInApartment(ctx context.Context, apartment string) (*domain.Resident, error) {
	residents, err := uc.residentRepository.FindByApartment(ctx, apartment)
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

// Template variables: {{1}} resident name, {{2}} apartment, {{3}} package type.
func buildNotification(notificationType domain.NotificationType, resident *domain.Resident, delivery *domain.Delivery) out.Notification {
	packageLabel := delivery.PackageType
	if packageLabel == "" {
		packageLabel = defaultPackageLabel
	}

	return out.Notification{
		Type:      notificationType,
		Phone:     resident.Phone,
		Body:      buildMessageBody(notificationType, resident, delivery),
		Variables: []string{resident.Name, delivery.Apartment, packageLabel},
	}
}

func buildMessageBody(notificationType domain.NotificationType, resident *domain.Resident, delivery *domain.Delivery) string {
	packageSuffix := ""
	if delivery.PackageType != "" {
		packageSuffix = fmt.Sprintf(" (%s)", delivery.PackageType)
	}

	if notificationType == domain.NotificationTypeDeliveryPickedUp {
		return fmt.Sprintf("Olá, %s! A entrega%s do apartamento %s foi retirada na portaria.", resident.Name, packageSuffix, delivery.Apartment)
	}

	message := fmt.Sprintf("Olá, %s! Chegou uma entrega para o apartamento %s%s. Retire na portaria.", resident.Name, delivery.Apartment, packageSuffix)
	if delivery.Urgency != "" {
		message += fmt.Sprintf(" Urgência: %s.", delivery.Urgency)
	}
	return message
}
