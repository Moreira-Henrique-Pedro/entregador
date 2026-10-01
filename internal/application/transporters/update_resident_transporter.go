package transporters

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type UpdateResidentTransporter struct {
	publisher     pubsub.MessagePublisher[any]
	internalTopic string
	sourceTopic   string
}

func NewUpdateResidentTransporter(
	publisher pubsub.MessagePublisher[any],
	internalTopic,
	sourceTopic string,
) *UpdateResidentTransporter {
	return &UpdateResidentTransporter{
		publisher:     publisher,
		internalTopic: internalTopic,
		sourceTopic:   sourceTopic,
	}
}

func (t *UpdateResidentTransporter) Handle(ctx context.Context, event *events.UpdateResident) error {
	logger := logger.GetLoggerFromContext(ctx)

	logger.Info("Publishing UpdateResident event", "internal_topic", t.internalTopic)

	command := t.buildInternalCommand(ctx, event)

	if err := t.publishCommand(ctx, command); err != nil {
		return fmt.Errorf("failed to publish internal command ProcessUpdateResident: commandID=%s: %w", command.CommandID, err)
	}

	return nil
}

func (t *UpdateResidentTransporter) buildInternalCommand(ctx context.Context, event *events.UpdateResident) *commands.ProcessUpdateResidentCommand {
	return &commands.ProcessUpdateResidentCommand{
		CommandID:  newCommandID(ctx),
		ResidentID: event.ResidentID,
		Name:       event.Name,
		Apartment:  event.Apartment,
		Phone:      event.Phone,
	}
}

func (t *UpdateResidentTransporter) publishCommand(ctx context.Context, command *commands.ProcessUpdateResidentCommand) error {
	headers := pubsub.NewHeaders(
		commands.ProcessUpdateResidentCommandType,
		command.ResidentID,
	)
	headers.Source = t.sourceTopic

	message := pubsub.NewMessage[any](ctx, headers, command)
	return t.publisher.Publish(ctx, t.internalTopic, message)
}
