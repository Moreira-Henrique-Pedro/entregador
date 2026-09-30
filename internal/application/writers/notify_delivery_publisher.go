package writers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/google/uuid"
)

// publishNotifyDelivery publishes the internal command that notifies the residents.
// The command id is derived from the delivery and the notification type, so retries
// publish the same command.
func publishNotifyDelivery(
	ctx context.Context,
	publisher pubsub.MessagePublisher[any],
	topic string,
	deliveryID string,
	notificationType notifier.NotificationType,
) error {
	command := &commands.ProcessNotifyDeliveryCommand{
		CommandID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(string(notificationType)+"/"+deliveryID)).String(),
		DeliveryID:       deliveryID,
		NotificationType: notificationType,
	}

	headers := pubsub.NewHeaders(commands.ProcessNotifyDeliveryCommandType, deliveryID)
	message := pubsub.NewMessage[any](ctx, headers, command)
	return publisher.Publish(ctx, topic, message)
}
