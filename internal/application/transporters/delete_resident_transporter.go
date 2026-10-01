package transporters

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeleteResidentTransporter struct {
	publisher     pubsub.MessagePublisher[any]
	internalTopic string
	sourceTopic   string
}

func NewDeleteResidentTransporter(
	publisher pubsub.MessagePublisher[any],
	internalTopic,
	sourceTopic string,
) *DeleteResidentTransporter {
	return &DeleteResidentTransporter{
		publisher:     publisher,
		internalTopic: internalTopic,
		sourceTopic:   sourceTopic,
	}
}

func (t *DeleteResidentTransporter) Handle(ctx context.Context, event *events.DeleteResident) error {
	logger := logger.GetLoggerFromContext(ctx)

	logger.Info("Publishing DeleteResident event", "internal_topic", t.internalTopic)

	command := t.buildInternalCommand(ctx, event)

	if err := t.publishCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessDeleteResident: commandID=%s: %w", command.CommandID, err)
	}

	return nil
}

func (t *DeleteResidentTransporter) buildInternalCommand(ctx context.Context, event *events.DeleteResident) *commands.ProcessDeleteResidentCommand {
	return &commands.ProcessDeleteResidentCommand{
		CommandID:  newCommandID(ctx),
		ResidentID: event.ResidentID,
	}
}

func (t *DeleteResidentTransporter) publishCommand(ctx context.Context, command *commands.ProcessDeleteResidentCommand) error {
	headers := pubsub.NewHeaders(
		commands.ProcessDeleteResidentCommandType,
		command.ResidentID,
	)
	headers.Source = t.sourceTopic

	message := pubsub.NewMessage[any](ctx, headers, command)
	return t.publisher.Publish(ctx, t.internalTopic, message)
}
