package writers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	pubsubmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub/mocks"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
)

func TestProcessDeleteDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	pending := &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusPending}

	tests := []struct {
		name           string
		delivery       *entities.Delivery
		findErr        error
		markDeleted    bool
		markDeletedErr error
		publish        bool
		publishErr     error
		wantErr        error
	}{
		{
			name:        "pending delivery is deleted and the pickup is notified",
			delivery:    pending,
			markDeleted: true,
			publish:     true,
		},
		{
			name:     "already deleted delivery without pickup notification publishes it again",
			delivery: &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusDeleted},
			publish:  true,
		},
		{
			name:     "already deleted and notified delivery does nothing",
			delivery: &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusDeleted, PickupNotifiedAt: time.Now()},
		},
		{
			name:    "unknown delivery is ignored",
			findErr: entities.ErrEntityNotFound,
		},
		{
			name:    "find error is returned",
			findErr: failure,
			wantErr: failure,
		},
		{
			name:           "delivery deleted concurrently still notifies the pickup",
			delivery:       pending,
			markDeleted:    true,
			markDeletedErr: entities.ErrEntityNotFound,
			publish:        true,
		},
		{
			name:           "delete error is returned without notifying",
			delivery:       pending,
			markDeleted:    true,
			markDeletedErr: failure,
			wantErr:        failure,
		},
		{
			name:        "notify publish error is returned so the command is retried",
			delivery:    pending,
			markDeleted: true,
			publish:     true,
			publishErr:  failure,
			wantErr:     failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := repomocks.NewDeliveryRepositoryPort(t)
			publisher := pubsubmocks.NewMessagePublisher[any](t)
			deliveries.EXPECT().FindByDeliveryID(mock.Anything, "d1").Return(tt.delivery, tt.findErr).Once()
			if tt.markDeleted {
				deliveries.EXPECT().MarkAsDeleted(mock.Anything, "d1").Return(tt.markDeletedErr).Once()
			}
			if tt.publish {
				expectNotifyCommand(t, publisher, "d1", notifier.NotificationTypeDeliveryPickedUp, tt.publishErr)
			}

			err := NewProcessDeleteDelivery(deliveries, publisher, internalTopic).Handle(context.Background(), &commands.ProcessDeleteDeliveryCommand{DeliveryID: "d1"})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
