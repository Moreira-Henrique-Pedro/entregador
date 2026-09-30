package writers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
)

func TestProcessDeleteDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	pending := func() *entities.Delivery {
		return &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusPending}
	}

	tests := []struct {
		name           string
		delivery       *entities.Delivery
		findErr        error
		markDeletedErr error
		publishErr     error
		wantErr        error
		wantDeleted    bool
		wantPublished  bool
	}{
		{
			name:          "pending delivery is deleted and the pickup is notified",
			delivery:      pending(),
			wantDeleted:   true,
			wantPublished: true,
		},
		{
			name:          "already deleted delivery without pickup notification publishes it again",
			delivery:      &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusDeleted},
			wantPublished: true,
		},
		{
			name:     "already deleted and notified delivery does nothing",
			delivery: &entities.Delivery{DeliveryID: "d1", Status: entities.DeliveryStatusDeleted, PickupNotifiedAt: time.Now()},
		},
		{
			name: "unknown delivery is ignored",
		},
		{
			name:     "find error is returned",
			delivery: pending(),
			findErr:  failure,
			wantErr:  failure,
		},
		{
			name:           "delivery deleted concurrently still notifies the pickup",
			delivery:       pending(),
			markDeletedErr: entities.ErrEntityNotFound,
			wantPublished:  true,
		},
		{
			name:           "delete error is returned without notifying",
			delivery:       pending(),
			markDeletedErr: failure,
			wantErr:        failure,
		},
		{
			name:        "notify publish error is returned so the command is retried",
			delivery:    pending(),
			publishErr:  failure,
			wantErr:     failure,
			wantDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := newFakeDeliveryRepository()
			if tt.delivery != nil {
				deliveries = newFakeDeliveryRepository(tt.delivery)
			}
			deliveries.findErr = tt.findErr
			deliveries.markDeletedErr = tt.markDeletedErr
			publisher := &fakePublisher{err: tt.publishErr}

			err := NewProcessDeleteDelivery(deliveries, publisher, internalTopic).Handle(context.Background(), &commands.ProcessDeleteDeliveryCommand{DeliveryID: "d1"})

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got := len(deliveries.deletedIDs) == 1; got != tt.wantDeleted {
				t.Errorf("deleted = %v, want deleted %v", deliveries.deletedIDs, tt.wantDeleted)
			}
			if got := len(publisher.messages) == 1; got != tt.wantPublished {
				t.Fatalf("published = %d messages, want published %v", len(publisher.messages), tt.wantPublished)
			}
			if tt.wantPublished {
				assertNotifyCommand(t, publisher, "d1", notifier.NotificationTypeDeliveryPickedUp)
			}
		})
	}
}
