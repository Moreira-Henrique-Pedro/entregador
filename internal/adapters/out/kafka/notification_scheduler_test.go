package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	pubsubmocks "github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub/mocks"
)

const topic = "delivery-internal.commands"

func TestNotificationScheduler_Schedule(t *testing.T) {
	failure := errors.New("kafka down")

	tests := []struct {
		name             string
		notificationType domain.NotificationType
		publishErr       error
	}{
		{name: "arrival", notificationType: domain.NotificationTypeDeliveryArrived},
		{name: "pickup", notificationType: domain.NotificationTypeDeliveryPickedUp},
		{name: "publish error is returned", notificationType: domain.NotificationTypeDeliveryArrived, publishErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := pubsubmocks.NewMessagePublisher[any](t)
			expectNotifyDelivery(t, publisher, "d1", tt.notificationType, tt.publishErr)

			err := NewNotificationScheduler(publisher, topic).Schedule(context.Background(), "d1", tt.notificationType)

			if tt.publishErr != nil {
				require.ErrorIs(t, err, tt.publishErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNotificationScheduler_SameNotificationSameCommandID(t *testing.T) {
	publisher := pubsubmocks.NewMessagePublisher[any](t)
	first := expectNotifyDelivery(t, publisher, "d1", domain.NotificationTypeDeliveryArrived, nil)
	second := expectNotifyDelivery(t, publisher, "d1", domain.NotificationTypeDeliveryArrived, nil)
	pickup := expectNotifyDelivery(t, publisher, "d1", domain.NotificationTypeDeliveryPickedUp, nil)

	scheduler := NewNotificationScheduler(publisher, topic)
	require.NoError(t, scheduler.Schedule(context.Background(), "d1", domain.NotificationTypeDeliveryArrived))
	require.NoError(t, scheduler.Schedule(context.Background(), "d1", domain.NotificationTypeDeliveryArrived))
	require.NoError(t, scheduler.Schedule(context.Background(), "d1", domain.NotificationTypeDeliveryPickedUp))

	require.NotEmpty(t, first.CommandID)
	assert.Equal(t, first.CommandID, second.CommandID)
	assert.NotEqual(t, first.CommandID, pickup.CommandID)
}

func expectNotifyDelivery(
	t *testing.T,
	publisher *pubsubmocks.MessagePublisher[any],
	deliveryID string,
	notificationType domain.NotificationType,
	publishErr error,
) *messages.NotifyDelivery {
	t.Helper()

	published := &messages.NotifyDelivery{}
	publisher.EXPECT().Publish(mock.Anything, topic, mock.Anything).
		Run(func(_ context.Context, _ string, msgs ...*pubsub.Message[any]) {
			require.Len(t, msgs, 1)
			message := msgs[0]
			assert.Equal(t, messages.NotifyDeliveryType, message.Headers.EventType)
			assert.Equal(t, deliveryID, message.Headers.Key)

			command, ok := message.Payload.Data.(*messages.NotifyDelivery)
			require.True(t, ok, "payload = %T, want *messages.NotifyDelivery", message.Payload.Data)
			assert.Equal(t, deliveryID, command.DeliveryID)
			assert.Equal(t, notificationType, command.NotificationType)
			*published = *command
		}).
		Return(publishErr).Once()
	return published
}
