package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type NotificationScheduler struct {
	publisher *gcppubsub.Publisher
}

func NewNotificationScheduler(client *gcppubsub.Client, topic string) *NotificationScheduler {
	return &NotificationScheduler{publisher: client.Publisher(topic)}
}

func (s *NotificationScheduler) Schedule(ctx context.Context, deliveryID string, notificationType entities.NotificationType) error {
	data, err := json.Marshal(commands.NewNotifyDeliveryCommand(deliveryID, notificationType))
	if err != nil {
		return fmt.Errorf("marshal NotifyDelivery: %w", err)
	}

	result := s.publisher.Publish(ctx, &gcppubsub.Message{
		Data: data,
		Attributes: map[string]string{
			commands.AttributeEventType: commands.NotifyDeliveryCommandType,
			commands.AttributeKey:       deliveryID,
			commands.AttributeSource:    config.AppName,
		},
	})

	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("publish NotifyDelivery: %w", err)
	}
	return nil
}

func (s *NotificationScheduler) Stop() {
	s.publisher.Stop()
}
