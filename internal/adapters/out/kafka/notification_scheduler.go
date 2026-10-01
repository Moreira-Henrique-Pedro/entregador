package kafka

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
)

// NotificationScheduler publishes a NotifyDelivery message, consumed by the worker.
type NotificationScheduler struct {
	publisher pubsub.MessagePublisher[any]
	topic     string
}

func NewNotificationScheduler(publisher pubsub.MessagePublisher[any], topic string) *NotificationScheduler {
	return &NotificationScheduler{
		publisher: publisher,
		topic:     topic,
	}
}

func (s *NotificationScheduler) Schedule(ctx context.Context, deliveryID string, notificationType domain.NotificationType) error {
	command := messages.NewNotifyDelivery(deliveryID, notificationType)

	headers := pubsub.NewHeaders(messages.NotifyDeliveryType, deliveryID)
	message := pubsub.NewMessage[any](ctx, headers, command)
	return s.publisher.Publish(ctx, s.topic, message)
}
