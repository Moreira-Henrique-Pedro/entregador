package transporters

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestCreateResidentTransporter(t *testing.T) {
	handle := func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
		return NewCreateResidentTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.CreateResident{Name: "Ana", Apartment: "202", Phone: "11999999999"})
	}

	t.Run("publishes the ProcessCreateResident command on the internal topic", func(t *testing.T) {
		msg := publishMessage(t, context.Background(), handle)
		assertHeaders(t, msg, commands.ProcessCreateResidentCommandType, "Ana")

		command := commandOf[commands.ProcessCreateResidentCommand](t, msg)
		assert.NotEmpty(t, command.CommandID)
		command.CommandID = ""
		assert.Equal(t, commands.ProcessCreateResidentCommand{Name: "Ana", Apartment: "202", Phone: "11999999999"}, *command)
	})

	t.Run("command id", func(t *testing.T) {
		assertCommandID(t, handle, func(c *commands.ProcessCreateResidentCommand) string { return c.CommandID })
	})

	t.Run("publish error is returned", func(t *testing.T) {
		assertPublishError(t, handle)
	})
}
