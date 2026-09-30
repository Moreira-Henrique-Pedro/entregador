package writers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ProcessDeleteDelivery struct {
	deliveryRepository interfaces.DeliveryRepositoryPort
}

func NewProcessDeleteDelivery(deliveryRepository interfaces.DeliveryRepositoryPort) *ProcessDeleteDelivery {
	return &ProcessDeleteDelivery{
		deliveryRepository: deliveryRepository,
	}
}

func (w *ProcessDeleteDelivery) Handle(ctx context.Context, command *commands.ProcessDeleteDeliveryCommand) error {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Processing ProcessDeleteDelivery command: commandID=%s", command.CommandID)

	err := w.deliveryRepository.MarkAsDeleted(ctx, command.DeliveryID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Pending delivery not found for delete: DeliveryID=%s", command.DeliveryID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete delivery: deliveryID=%s: %w", command.DeliveryID, err)
	}

	logger.Info("Delivery deleted: DeliveryID=%s", command.DeliveryID)

	return nil
}
