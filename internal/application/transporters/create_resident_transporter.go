package transporters

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/google/uuid"
)

type CreateResidentTransporter struct {
	publisher     pubsub.MessagePublisher[any]
	internalTopic string
	sourceTopic   string
}

func NewCreateResidentTransporter(
	publisher pubsub.MessagePublisher[any],
	internalTopic,
	sourceTopic string,
) *CreateResidentTransporter {
	return &CreateResidentTransporter{
		publisher:     publisher,
		internalTopic: internalTopic,
		sourceTopic:   sourceTopic,
	}
}

func (t *CreateResidentTransporter) Handle(ctx context.Context, event *events.CreateResident) error {
	logger := logger.GetLoggerFromContext(ctx)

	logger.Info("Publishing CreateResident event to topic %s", t.internalTopic)

	command := t.buildInternalCommand(event)

	if err := t.publishCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessCreateResident: commandID=%s: %w", command.CommandID, err)
	}

	return nil
}

func (t *CreateResidentTransporter) buildInternalCommand(event *events.CreateResident) *commands.ProcessCreateResidentCommand {
	return &commands.ProcessCreateResidentCommand{
		CommandID: uuid.New().String(),
		Name:      event.Name,
		Apartment: event.Apartment,
		Phone:     event.Phone,
	}
}

func (t *CreateResidentTransporter) publishCommand(ctx context.Context, command *commands.ProcessCreateResidentCommand) error {
	headers := pubsub.NewHeaders(
		commands.ProcessCreateResidentCommandType,
		command.Name,
	)
	headers.Source = t.sourceTopic

	message := pubsub.NewMessage[any](ctx, headers, command)
	return t.publisher.Publish(ctx, t.internalTopic, message)
}
