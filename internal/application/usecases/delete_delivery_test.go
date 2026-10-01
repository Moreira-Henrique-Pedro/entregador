package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	outmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

func TestDeleteDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	pending := &domain.Delivery{DeliveryID: "d1", Status: domain.DeliveryStatusPending}

	tests := []struct {
		name           string
		delivery       *domain.Delivery
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
			delivery: &domain.Delivery{DeliveryID: "d1", Status: domain.DeliveryStatusDeleted},
			schedule: true,
		},
		{
			name:     "already deleted and notified delivery does nothing",
			delivery: &domain.Delivery{DeliveryID: "d1", Status: domain.DeliveryStatusDeleted, PickupNotifiedAt: time.Now()},
		},
		{
			name:    "unknown delivery returns not found",
			findErr: domain.ErrEntityNotFound,
			wantErr: domain.ErrEntityNotFound,
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
			markDeletedErr: domain.ErrEntityNotFound,
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
			deliveries := outmocks.NewDeliveryRepository(t)
			scheduler := outmocks.NewNotificationScheduler(t)
			deliveries.EXPECT().FindByDeliveryID(mock.Anything, "d1").Return(tt.delivery, tt.findErr).Once()
			if tt.markDeleted {
				deliveries.EXPECT().MarkAsDeleted(mock.Anything, "d1").Return(tt.markDeletedErr).Once()
			}
			if tt.schedule {
				scheduler.EXPECT().Schedule(mock.Anything, "d1", domain.NotificationTypeDeliveryPickedUp).Return(tt.scheduleErr).Once()
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
	err := NewDeleteDelivery(outmocks.NewDeliveryRepository(t), outmocks.NewNotificationScheduler(t)).Execute(context.Background(), "")

	require.ErrorIs(t, err, domain.ErrInvalidDelivery)
}
