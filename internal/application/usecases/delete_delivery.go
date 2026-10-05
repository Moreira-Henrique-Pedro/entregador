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

type DeleteDelivery struct {
	deliveryRepository repositories.DeliveryRepository
	scheduler          services.NotificationScheduler
}

func NewDeleteDelivery(deliveryRepository repositories.DeliveryRepository, scheduler services.NotificationScheduler) *DeleteDelivery {
	return &DeleteDelivery{
		deliveryRepository: deliveryRepository,
		scheduler:          scheduler,
	}
}

func (uc *DeleteDelivery) Execute(ctx context.Context, deliveryID string) error {
	if err := entities.ValidateDeliveryID(deliveryID); err != nil {
		return err
	}

	delivery, err := uc.deliveryRepository.FindByDeliveryID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to find delivery: deliveryID=%s: %w", deliveryID, err)
	}

	if err := uc.markAsPickedUp(ctx, delivery); err != nil {
		return err
	}

	return uc.schedulePickupNotification(ctx, delivery)
}

func (uc *DeleteDelivery) markAsPickedUp(ctx context.Context, delivery *entities.Delivery) error {
	log := logger.GetLoggerFromContext(ctx).With("delivery_id", delivery.DeliveryID)

	if !delivery.IsPending() {
		log.Info("Delivery already deleted")
		return nil
	}

	err := uc.deliveryRepository.MarkAsDeleted(ctx, delivery.DeliveryID)
	if err != nil && !errors.Is(err, entities.ErrEntityNotFound) {
		return fmt.Errorf("failed to delete delivery: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	log.Info("Delivery deleted")
	return nil
}

func (uc *DeleteDelivery) schedulePickupNotification(ctx context.Context, delivery *entities.Delivery) error {
	if delivery.IsPickupNotified() {
		return nil
	}
	if err := uc.scheduler.Schedule(ctx, delivery.DeliveryID, entities.NotificationTypeDeliveryPickedUp); err != nil {
		return fmt.Errorf("failed to schedule pickup notification: deliveryID=%s: %w", delivery.DeliveryID, err)
	}
	return nil
}
