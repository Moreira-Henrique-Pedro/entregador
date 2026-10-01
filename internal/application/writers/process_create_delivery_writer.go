package writers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ProcessCreateDelivery struct {
	deliveryRepository interfaces.DeliveryRepositoryPort
	residentRepository interfaces.ResidentRepositoryPort
	publisher          pubsub.MessagePublisher[any]
	internalTopic      string
}

func NewProcessCreateDelivery(
	deliveryRepository interfaces.DeliveryRepositoryPort,
	residentRepository interfaces.ResidentRepositoryPort,
	publisher pubsub.MessagePublisher[any],
	internalTopic string,
) *ProcessCreateDelivery {
	return &ProcessCreateDelivery{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
		publisher:          publisher,
		internalTopic:      internalTopic,
	}
}

func (w *ProcessCreateDelivery) Handle(ctx context.Context, command *commands.ProcessCreateDeliveryCommand) error {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Processing ProcessCreateDelivery command", "command_id", command.CommandID)

	if command.Apartment == "" {
		logger.Warn("Discarding ProcessCreateDelivery command without apartment", "command_id", command.CommandID)
		return nil
	}

	hasResident, err := w.apartmentHasResident(ctx, command.Apartment)
	if err != nil {
		return fmt.Errorf("failed to find apartment residents: apartment=%s: %w", command.Apartment, err)
	}
	if !hasResident {
		logger.Warn("Rejecting delivery for apartment without residents", "command_id", command.CommandID, "apartment", command.Apartment)
		// Retrying will not register a resident, so the command goes straight to the DLQ.
		return pkgEvents.Permanent(fmt.Errorf("apartment %s: %w", command.Apartment, entities.ErrNoResidentInApartment))
	}

	residentID, err := w.resolveRecipient(ctx, command)
	if err != nil {
		return fmt.Errorf("failed to resolve delivery recipient: apartment=%s: %w", command.Apartment, err)
	}

	delivery := w.buildDeliveryEntity(command, residentID)
	if err := w.deliveryRepository.Insert(ctx, delivery); err != nil {
		return fmt.Errorf("failed to insert delivery: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	logger.Info("Delivery created", "delivery_id", delivery.DeliveryID, "apartment", delivery.Apartment, "resident_id", delivery.ResidentID)

	// On failure the command is retried: the insert is deduplicated by id and the
	// notification is published again, so the resident is still notified.
	if err := publishNotifyDelivery(ctx, w.publisher, w.internalTopic, delivery.DeliveryID, notifier.NotificationTypeDeliveryArrived); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessNotifyDelivery: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	return nil
}

// apartmentHasResident reports whether the apartment has a registered resident
// besides its "other" resident.
func (w *ProcessCreateDelivery) apartmentHasResident(ctx context.Context, apartment string) (bool, error) {
	residents, err := w.residentRepository.FindByApartment(ctx, apartment)
	if err != nil {
		return false, err
	}
	for _, resident := range residents {
		if !resident.IsOther() {
			return true, nil
		}
	}
	return false, nil
}

// resolveRecipient returns the informed resident when it lives in the apartment,
// otherwise the apartment's "other" resident, creating it if needed.
func (w *ProcessCreateDelivery) resolveRecipient(ctx context.Context, command *commands.ProcessCreateDeliveryCommand) (string, error) {
	logger := logger.GetLoggerFromContext(ctx)

	if command.ResidentID != "" {
		resident, err := w.residentRepository.FindByResidentID(ctx, command.ResidentID)
		switch {
		case err == nil && resident.Apartment == command.Apartment:
			return resident.ResidentID, nil
		case err == nil:
			logger.Warn("Resident does not live in the delivery apartment, using other", "resident_id", command.ResidentID, "apartment", command.Apartment)
		case errors.Is(err, entities.ErrEntityNotFound):
			logger.Warn("Resident not found, using other", "resident_id", command.ResidentID, "apartment", command.Apartment)
		default:
			return "", err
		}
	}

	if err := w.residentRepository.EnsureOtherResident(ctx, command.Apartment); err != nil {
		return "", err
	}
	return entities.OtherResidentID(command.Apartment), nil
}

func (w *ProcessCreateDelivery) buildDeliveryEntity(command *commands.ProcessCreateDeliveryCommand, residentID string) *entities.Delivery {
	return &entities.Delivery{
		ID:          command.CommandID,
		DeliveryID:  command.CommandID,
		Apartment:   command.Apartment,
		ResidentID:  residentID,
		PackageType: command.PackageType,
		Urgency:     command.Urgency,
		Status:      entities.DeliveryStatusPending,
	}
}
