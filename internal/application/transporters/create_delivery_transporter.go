package transporters

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type CreateDeliveryTransporter struct {
	publisher     pubsub.MessagePublisher[any]
	internalTopic string
	sourceTopic   string
}

func NewCreateDeliveryTransporter(
	publisher pubsub.MessagePublisher[any],
	internalTopic,
	sourceTopic string,
) *CreateDeliveryTransporter {
	return &CreateDeliveryTransporter{
		publisher:     publisher,
		internalTopic: internalTopic,
		sourceTopic:   sourceTopic,
	}
}

func (t *CreateDeliveryTransporter) Handle(ctx context.Context, event *events.CreateDelivery) error {
	logger := logger.GetLoggerFromContext(ctx)

	logger.Info("Publishing CreateDelivery event to topic %s", t.internalTopic)

	command := t.buildInternalCommand(ctx, event)

	if err := t.publishCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessCreateDelivery: commandID=%s: %w", command.CommandID, err)
	}

	return nil
}

func (t *CreateDeliveryTransporter) buildInternalCommand(ctx context.Context, event *events.CreateDelivery) *commands.ProcessCreateDeliveryCommand {
	return &commands.ProcessCreateDeliveryCommand{
		CommandID:   newCommandID(ctx),
		Apartment:   event.Apartment,
		ResidentID:  event.ResidentID,
		PackageType: event.PackageType,
		Urgency:     event.Urgency,
	}
}

func (t *CreateDeliveryTransporter) publishCommand(ctx context.Context, command *commands.ProcessCreateDeliveryCommand) error {
	headers := pubsub.NewHeaders(
		commands.ProcessCreateDeliveryCommandType,
		command.Apartment,
	)
	headers.Source = t.sourceTopic

	message := pubsub.NewMessage[any](ctx, headers, command)
	return t.publisher.Publish(ctx, t.internalTopic, message)
}
