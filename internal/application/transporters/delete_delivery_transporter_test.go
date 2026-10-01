package transporters

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestDeleteDeliveryTransporter(t *testing.T) {
	handle := func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
		return NewDeleteDeliveryTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.DeleteDelivery{DeliveryID: "d1"})
	}

	t.Run("publishes the ProcessDeleteDelivery command on the internal topic", func(t *testing.T) {
		msg := publishMessage(t, context.Background(), handle)
		assertHeaders(t, msg, commands.ProcessDeleteDeliveryCommandType, "d1")

		command := commandOf[commands.ProcessDeleteDeliveryCommand](t, msg)
		assert.NotEmpty(t, command.CommandID)
		command.CommandID = ""
		assert.Equal(t, commands.ProcessDeleteDeliveryCommand{DeliveryID: "d1"}, *command)
	})

	t.Run("command id", func(t *testing.T) {
		assertCommandID(t, handle, func(c *commands.ProcessDeleteDeliveryCommand) string { return c.CommandID })
	})

	t.Run("publish error is returned", func(t *testing.T) {
		assertPublishError(t, handle)
	})
}
