package transporters

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestUpdateResidentTransporter(t *testing.T) {
	handle := func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
		return NewUpdateResidentTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.UpdateResident{ResidentID: "r1", Name: "Ana", Apartment: "202", Phone: "11999999999"})
	}

	t.Run("publishes the ProcessUpdateResident command on the internal topic", func(t *testing.T) {
		msg := publishMessage(t, context.Background(), handle)
		assertHeaders(t, msg, commands.ProcessUpdateResidentCommandType, "r1")

		command := commandOf[commands.ProcessUpdateResidentCommand](t, msg)
		assert.NotEmpty(t, command.CommandID)
		command.CommandID = ""
		assert.Equal(t, commands.ProcessUpdateResidentCommand{ResidentID: "r1", Name: "Ana", Apartment: "202", Phone: "11999999999"}, *command)
	})

	t.Run("command id", func(t *testing.T) {
		assertCommandID(t, handle, func(c *commands.ProcessUpdateResidentCommand) string { return c.CommandID })
	})

	t.Run("publish error is returned", func(t *testing.T) {
		assertPublishError(t, handle)
	})
}
