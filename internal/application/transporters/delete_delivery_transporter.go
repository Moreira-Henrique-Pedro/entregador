package transporters

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeleteDeliveryTransporter struct {
	publisher     pubsub.MessagePublisher[any]
	internalTopic string
	sourceTopic   string
}

func NewDeleteDeliveryTransporter(
	publisher pubsub.MessagePublisher[any],
	internalTopic,
	sourceTopic string,
) *DeleteDeliveryTransporter {
	return &DeleteDeliveryTransporter{
		publisher:     publisher,
		internalTopic: internalTopic,
		sourceTopic:   sourceTopic,
	}
}

func (t *DeleteDeliveryTransporter) Handle(ctx context.Context, event *events.DeleteDelivery) error {
	logger := logger.GetLoggerFromContext(ctx)

	logger.Info("Publishing DeleteDelivery event to topic %s", t.internalTopic)

	command := t.buildInternalCommand(ctx, event)

	if err := t.publishCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessDeleteDelivery: commandID=%s: %w", command.CommandID, err)
	}

	return nil
}

func (t *DeleteDeliveryTransporter) buildInternalCommand(ctx context.Context, event *events.DeleteDelivery) *commands.ProcessDeleteDeliveryCommand {
	return &commands.ProcessDeleteDeliveryCommand{
		CommandID:  newCommandID(ctx),
		DeliveryID: event.DeliveryID,
	}
}

func (t *DeleteDeliveryTransporter) publishCommand(ctx context.Context, command *commands.ProcessDeleteDeliveryCommand) error {
	headers := pubsub.NewHeaders(
		commands.ProcessDeleteDeliveryCommandType,
		command.DeliveryID,
	)
	headers.Source = t.sourceTopic

	message := pubsub.NewMessage[any](ctx, headers, command)
	return t.publisher.Publish(ctx, t.internalTopic, message)
}
