package kafka

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	"github.com/google/uuid"
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
	command := &messages.NotifyDelivery{
		// Deterministic: scheduling the same notification twice yields the same command.
		CommandID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(string(notificationType)+"/"+deliveryID)).String(),
		DeliveryID:       deliveryID,
		NotificationType: notificationType,
	}

	headers := pubsub.NewHeaders(messages.NotifyDeliveryType, deliveryID)
	message := pubsub.NewMessage[any](ctx, headers, command)
	return s.publisher.Publish(ctx, s.topic, message)
}
