package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type NotifyDelivery struct {
	deliveryRepository repositories.DeliveryRepository
	residentRepository repositories.ResidentRepository
	notifier           services.Notifier
}

func NewNotifyDelivery(
	deliveryRepository repositories.DeliveryRepository,
	residentRepository repositories.ResidentRepository,
	notifier services.Notifier,
) *NotifyDelivery {
	return &NotifyDelivery{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
		notifier:           notifier,
	}
}

func (uc *NotifyDelivery) Execute(ctx context.Context, deliveryID string, notificationType entities.NotificationType) error {
	log := logger.GetLoggerFromContext(ctx).With("delivery_id", deliveryID, "notification_type", string(notificationType))
	ctx = log.AddToContext(ctx, log)

	delivery, err := uc.findDeliveryToNotify(ctx, deliveryID, notificationType)
	if err != nil || delivery == nil {
		return err
	}

	recipient, err := uc.resolveRecipient(ctx, delivery)
	if err != nil {
		return fmt.Errorf("failed to resolve notification recipient: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	if err := uc.send(ctx, notificationType, recipient, delivery); err != nil {
		return err
	}

	return uc.markAsNotified(ctx, delivery.DeliveryID, notificationType)
}

func (uc *NotifyDelivery) findDeliveryToNotify(ctx context.Context, deliveryID string, notificationType entities.NotificationType) (*entities.Delivery, error) {
	log := logger.GetLoggerFromContext(ctx)

	if !notificationType.IsValid() {
		log.Warn("Discarding notification with invalid type")
		return nil, nil
	}

	delivery, err := uc.deliveryRepository.FindByDeliveryID(ctx, deliveryID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		log.Warn("Delivery not found for notification")
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find delivery: deliveryID=%s: %w", deliveryID, err)
	}

	if reason := delivery.NotificationSkipReason(notificationType); reason != "" {
		log.Info("Skipping delivery notification", "reason", reason)
		return nil, nil
	}
	return delivery, nil
}

func (uc *NotifyDelivery) resolveRecipient(ctx context.Context, delivery *entities.Delivery) (*entities.Resident, error) {
	resident, err := uc.residentRepository.FindByResidentID(ctx, delivery.ResidentID)
	if err != nil && !errors.Is(err, entities.ErrEntityNotFound) {
		return nil, err
	}
	if err != nil || resident.IsOther() {
		return uc.findPrimary(ctx, delivery.Apartment)
	}
	return resident, nil
}

func (uc *NotifyDelivery) findPrimary(ctx context.Context, apartment string) (*entities.Resident, error) {
	primary, err := uc.findPrimaryInApartment(ctx, apartment)
	if err != nil || primary != nil {
		return primary, err
	}

	if err := ensurePrimaryResident(ctx, uc.residentRepository, apartment); err != nil {
		return nil, err
	}
	return uc.findPrimaryInApartment(ctx, apartment)
}

func (uc *NotifyDelivery) findPrimaryInApartment(ctx context.Context, apartment string) (*entities.Resident, error) {
	residents, err := findApartmentResidents(ctx, uc.residentRepository, apartment)
	if err != nil {
		return nil, err
	}
	return entities.FindPrimary(residents), nil
}

func (uc *NotifyDelivery) send(ctx context.Context, notificationType entities.NotificationType, recipient *entities.Resident, delivery *entities.Delivery) error {
	log := logger.GetLoggerFromContext(ctx)

	if recipient == nil || !recipient.CanBeNotified() {
		log.Warn("No resident with phone to notify", "apartment", delivery.Apartment, "resident_id", delivery.ResidentID)
		return nil
	}

	err := uc.notifier.Send(ctx, entities.NewDeliveryNotification(notificationType, recipient, delivery))
	if errors.Is(err, entities.ErrInvalidRecipient) {
		log.Warn("Resident phone rejected by the provider", "resident_id", recipient.ResidentID, "error", err.Error())
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to notify resident: residentID=%s: %w", recipient.ResidentID, err)
	}

	log.Info("Delivery notified", "resident_id", recipient.ResidentID)
	return nil
}

func (uc *NotifyDelivery) markAsNotified(ctx context.Context, deliveryID string, notificationType entities.NotificationType) error {
	mark := uc.deliveryRepository.MarkArrivalAsNotified
	if notificationType == entities.NotificationTypeDeliveryPickedUp {
		mark = uc.deliveryRepository.MarkPickupAsNotified
	}
	if err := mark(ctx, deliveryID); err != nil {
		return fmt.Errorf("failed to mark delivery as notified: deliveryID=%s: %w", deliveryID, err)
	}
	return nil
}
