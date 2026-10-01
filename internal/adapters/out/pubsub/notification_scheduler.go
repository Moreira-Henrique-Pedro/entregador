package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	pkgPubsub "github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
)

// NotificationScheduler publishes a NotifyDelivery message to a Pub/Sub topic, which pushes it
// back to the API (see adapters/in/pubsub).
type NotificationScheduler struct {
	publisher *gcppubsub.Publisher
}

func NewNotificationScheduler(client *gcppubsub.Client, topic string) *NotificationScheduler {
	return &NotificationScheduler{publisher: client.Publisher(topic)}
}

func (s *NotificationScheduler) Schedule(ctx context.Context, deliveryID string, notificationType domain.NotificationType) error {
	data, err := json.Marshal(messages.NewNotifyDelivery(deliveryID, notificationType))
	if err != nil {
		return fmt.Errorf("marshal NotifyDelivery: %w", err)
	}

	result := s.publisher.Publish(ctx, &gcppubsub.Message{
		Data: data,
		Attributes: map[string]string{
			pkgPubsub.EventTypeHeader: messages.NotifyDeliveryType,
			pkgPubsub.KeyHeader:       deliveryID,
			pkgPubsub.SourceHeader:    config.AppName,
		},
	})

	// Waits for the server ack: the request only succeeds once the notification is queued.
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("publish NotifyDelivery: %w", err)
	}
	return nil
}

// Stop flushes the pending messages.
func (s *NotificationScheduler) Stop() {
	s.publisher.Stop()
}
