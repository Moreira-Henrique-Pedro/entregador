package transporters

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestDeleteResidentTransporter(t *testing.T) {
	handle := func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
		return NewDeleteResidentTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.DeleteResident{ResidentID: "r1"})
	}

	t.Run("publishes the ProcessDeleteResident command on the internal topic", func(t *testing.T) {
		msg := publishMessage(t, context.Background(), handle)
		assertHeaders(t, msg, commands.ProcessDeleteResidentCommandType, "r1")

		command := commandOf[commands.ProcessDeleteResidentCommand](t, msg)
		assert.NotEmpty(t, command.CommandID)
		command.CommandID = ""
		assert.Equal(t, commands.ProcessDeleteResidentCommand{ResidentID: "r1"}, *command)
	})

	t.Run("command id", func(t *testing.T) {
		assertCommandID(t, handle, func(c *commands.ProcessDeleteResidentCommand) string { return c.CommandID })
	})

	t.Run("publish error is returned", func(t *testing.T) {
		assertPublishError(t, handle)
	})
}
