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

	logger.Info("Publishing DeleteResident event to topic %s", t.internalTopic)

	command := t.buildInternalCommand(event)

	if err := t.publishCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessDeleteResident: commandID=%s: %w", command.CommandID, err)
	}

	return nil
}

func (t *DeleteResidentTransporter) buildInternalCommand(event *events.DeleteResident) *commands.ProcessDeleteResidentCommand {
	return &commands.ProcessDeleteResidentCommand{
		CommandID:  uuid.New().String(),
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
