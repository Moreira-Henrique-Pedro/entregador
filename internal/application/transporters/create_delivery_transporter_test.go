package transporters

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestCreateDeliveryTransporter(t *testing.T) {
	handle := func(ctx context.Context, pub pubsub.MessagePublisher[any]) error {
		return NewCreateDeliveryTransporter(pub, testInternalTopic, testSourceTopic).Handle(ctx, &events.CreateDelivery{Apartment: "101", ResidentID: "r1", PackageType: "box", Urgency: "high"})
	}

	t.Run("publishes the ProcessCreateDelivery command on the internal topic", func(t *testing.T) {
		msg := publishMessage(t, context.Background(), handle)
		assertHeaders(t, msg, commands.ProcessCreateDeliveryCommandType, "101")

		command := commandOf[commands.ProcessCreateDeliveryCommand](t, msg)
		assert.NotEmpty(t, command.CommandID)
		command.CommandID = ""
		assert.Equal(t, commands.ProcessCreateDeliveryCommand{Apartment: "101", ResidentID: "r1", PackageType: "box", Urgency: "high"}, *command)
	})

	t.Run("command id", func(t *testing.T) {
		assertCommandID(t, handle, func(c *commands.ProcessCreateDeliveryCommand) string { return c.CommandID })
	})

	t.Run("publish error is returned", func(t *testing.T) {
		assertPublishError(t, handle)
	})
}
