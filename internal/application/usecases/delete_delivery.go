package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeleteDelivery struct {
	deliveryRepository out.DeliveryRepository
	scheduler          out.NotificationScheduler
}

func NewDeleteDelivery(deliveryRepository out.DeliveryRepository, scheduler out.NotificationScheduler) *DeleteDelivery {
	return &DeleteDelivery{
		deliveryRepository: deliveryRepository,
		scheduler:          scheduler,
	}
}

// Execute marks the delivery as picked up and schedules the pickup notification.
// It is idempotent: retrying after a failed schedule only schedules it again.
func (uc *DeleteDelivery) Execute(ctx context.Context, deliveryID string) error {
	if deliveryID == "" {
		return fmt.Errorf("%w: delivery_id is required", domain.ErrInvalidDelivery)
	}

	logger := logger.GetLoggerFromContext(ctx).With("delivery_id", deliveryID)

	delivery, err := uc.deliveryRepository.FindByDeliveryID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to find delivery: deliveryID=%s: %w", deliveryID, err)
	}

	if delivery.Status == domain.DeliveryStatusPending {
		err := uc.deliveryRepository.MarkAsDeleted(ctx, delivery.DeliveryID)
		if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
			return fmt.Errorf("failed to delete delivery: deliveryID=%s: %w", delivery.DeliveryID, err)
		}
		logger.Info("Delivery deleted")
	} else {
		logger.Info("Delivery already deleted")
	}

	// Runs even for an already deleted delivery, so a retry after a failed schedule still notifies.
	if !delivery.PickupNotifiedAt.IsZero() {
		return nil
	}
	if err := uc.scheduler.Schedule(ctx, delivery.DeliveryID, domain.NotificationTypeDeliveryPickedUp); err != nil {
		return fmt.Errorf("failed to schedule pickup notification: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	return nil
}
