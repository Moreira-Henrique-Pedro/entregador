package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
	servicemocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	pending := &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusPending}

	tests := []struct {
		name           string
		delivery       *entities.Delivery
		findErr        error
		markDeleted    bool
		markDeletedErr error
		schedule       bool
		scheduleErr    error
		wantErr        error
	}{
		{
			name:        "pending delivery is deleted and the pickup is scheduled",
			delivery:    pending,
			markDeleted: true,
			schedule:    true,
		},
		{
			name:     "already deleted delivery without pickup notification schedules it again",
			delivery: &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusDeleted},
			schedule: true,
		},
		{
			name:     "already deleted and notified delivery does nothing",
			delivery: &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusDeleted, PickupNotifiedAt: time.Now()},
		},
		{
			name:    "unknown delivery returns not found",
			findErr: entities.ErrEntityNotFound,
			wantErr: entities.ErrEntityNotFound,
		},
		{
			name:    "find error is returned",
			findErr: failure,
			wantErr: failure,
		},
		{
			name:           "delivery deleted concurrently still schedules the pickup",
			delivery:       pending,
			markDeleted:    true,
			markDeletedErr: entities.ErrEntityNotFound,
			schedule:       true,
		},
		{
			name:           "delete error is returned without scheduling",
			delivery:       pending,
			markDeleted:    true,
			markDeletedErr: failure,
			wantErr:        failure,
		},
		{
			name:        "schedule error is returned so the client can retry",
			delivery:    pending,
			markDeleted: true,
			schedule:    true,
			scheduleErr: failure,
			wantErr:     failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := repomocks.NewDeliveryRepository(t)
			scheduler := servicemocks.NewNotificationScheduler(t)
			deliveries.EXPECT().FindByDeliveryID(mock.Anything, "d1").Return(tt.delivery, tt.findErr).Once()
			if tt.markDeleted {
				deliveries.EXPECT().MarkAsDeleted(mock.Anything, "d1").Return(tt.markDeletedErr).Once()
			}
			if tt.schedule {
				scheduler.EXPECT().Schedule(mock.Anything, "d1", entities.NotificationTypeDeliveryPickedUp).Return(tt.scheduleErr).Once()
			}

			err := NewDeleteDelivery(deliveries, scheduler).Execute(context.Background(), "d1")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteDelivery_RequiresID(t *testing.T) {
	err := NewDeleteDelivery(repomocks.NewDeliveryRepository(t), servicemocks.NewNotificationScheduler(t)).Execute(context.Background(), "")

	require.ErrorIs(t, err, entities.ErrInvalidDelivery)
}
