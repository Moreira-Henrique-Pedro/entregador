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
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ProcessDeleteDelivery struct {
	deliveryRepository interfaces.DeliveryRepositoryPort
	publisher          pubsub.MessagePublisher[any]
	internalTopic      string
}

func NewProcessDeleteDelivery(
	deliveryRepository interfaces.DeliveryRepositoryPort,
	publisher pubsub.MessagePublisher[any],
	internalTopic string,
) *ProcessDeleteDelivery {
	return &ProcessDeleteDelivery{
		deliveryRepository: deliveryRepository,
		publisher:          publisher,
		internalTopic:      internalTopic,
	}
}

func (w *ProcessDeleteDelivery) Handle(ctx context.Context, command *commands.ProcessDeleteDeliveryCommand) error {
	logger := logger.GetLoggerFromContext(ctx).With("delivery_id", command.DeliveryID)
	logger.Info("Processing ProcessDeleteDelivery command", "command_id", command.CommandID)

	delivery, err := w.deliveryRepository.FindByDeliveryID(ctx, command.DeliveryID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Delivery not found for delete")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find delivery: deliveryID=%s: %w", command.DeliveryID, err)
	}

	if delivery.Status == entities.DeliveryStatusPending {
		err := w.deliveryRepository.MarkAsDeleted(ctx, delivery.DeliveryID)
		// Not found here means it was deleted concurrently, which is the state we want.
		if err != nil && !errors.Is(err, entities.ErrEntityNotFound) {
			return fmt.Errorf("failed to delete delivery: deliveryID=%s: %w", delivery.DeliveryID, err)
		}
		logger.Info("Delivery deleted")
	} else {
		logger.Info("Delivery already deleted")
	}

	// Also runs for an already deleted delivery, so a retry after a failed publish
	// still notifies the pickup; the notify writer ignores duplicates.
	if !delivery.PickupNotifiedAt.IsZero() {
		return nil
	}
	if err := publishNotifyDelivery(ctx, w.publisher, w.internalTopic, delivery.DeliveryID, notifier.NotificationTypeDeliveryPickedUp); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessNotifyDelivery: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	return nil
}
