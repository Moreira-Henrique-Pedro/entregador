package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
)

const internalTopic = "delivery-internal.commands"

func TestProcessCreateDelivery(t *testing.T) {
	failure := errors.New("mongo down")
	otherID := entities.OtherResidentID("101")

	tests := []struct {
		name           string
		command        commands.ProcessCreateDeliveryCommand
		findErr        error
		ensureOtherErr error
		insertErr      error
		publishErr     error
		wantErr        error
		wantResident   string
		wantEnsured    []string
		wantPublished  bool
	}{
		{
			name:          "resident of the apartment receives the delivery",
			command:       commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"},
			wantResident:  "ana",
			wantPublished: true,
		},
		{
			name:          "no resident informed goes to the apartment other",
			command:       commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101"},
			wantResident:  otherID,
			wantEnsured:   []string{"101"},
			wantPublished: true,
		},
		{
			name:          "unknown resident goes to the apartment other",
			command:       commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ghost"},
			wantResident:  otherID,
			wantEnsured:   []string{"101"},
			wantPublished: true,
		},
		{
			name:          "resident from another apartment goes to the apartment other",
			command:       commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "bob"},
			wantResident:  otherID,
			wantEnsured:   []string{"101"},
			wantPublished: true,
		},
		{
			name:    "command without apartment is discarded",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", ResidentID: "ana"},
		},
		{
			name:    "resident lookup error is returned",
			command: commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			findErr: failure,
			wantErr: failure,
		},
		{
			name:           "ensure other error is returned",
			command:        commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101"},
			ensureOtherErr: failure,
			wantErr:        failure,
		},
		{
			name:      "insert error is returned",
			command:   commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			insertErr: failure,
			wantErr:   failure,
		},
		{
			name:         "notify publish error is returned so the command is retried",
			command:      commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101", ResidentID: "ana"},
			publishErr:   failure,
			wantErr:      failure,
			wantResident: "ana",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(
				&entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypeResident},
				&entities.Resident{ResidentID: "bob", Apartment: "202", Type: entities.ResidentTypeResident},
			)
			residents.findErr = tt.findErr
			residents.ensureOtherErr = tt.ensureOtherErr
			deliveries := newFakeDeliveryRepository()
			deliveries.insertErr = tt.insertErr
			publisher := &fakePublisher{err: tt.publishErr}

			err := NewProcessCreateDelivery(deliveries, residents, publisher, internalTopic).Handle(context.Background(), &tt.command)

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if !equalStrings(residents.ensuredOthers, tt.wantEnsured) {
				t.Errorf("ensured others = %v, want %v", residents.ensuredOthers, tt.wantEnsured)
			}

			if tt.wantResident == "" {
				if len(deliveries.inserted) != 0 {
					t.Errorf("inserted = %v, want none", deliveries.inserted)
				}
			} else {
				if len(deliveries.inserted) != 1 {
					t.Fatalf("inserted %d deliveries, want 1", len(deliveries.inserted))
				}
				want := entities.Delivery{
					ID:          tt.command.CommandID,
					DeliveryID:  tt.command.CommandID,
					Apartment:   tt.command.Apartment,
					ResidentID:  tt.wantResident,
					PackageType: tt.command.PackageType,
					Urgency:     tt.command.Urgency,
					Status:      entities.DeliveryStatusPending,
				}
				if *deliveries.inserted[0] != want {
					t.Errorf("inserted = %+v, want %+v", *deliveries.inserted[0], want)
				}
			}

			if got := len(publisher.messages) == 1; got != tt.wantPublished {
				t.Fatalf("published = %d messages, want published %v", len(publisher.messages), tt.wantPublished)
			}
			if tt.wantPublished {
				assertNotifyCommand(t, publisher, "d1", notifier.NotificationTypeDeliveryArrived)
			}
		})
	}
}

func TestProcessCreateDelivery_RetryPublishesSameNotifyCommand(t *testing.T) {
	publisher := &fakePublisher{}
	writer := NewProcessCreateDelivery(newFakeDeliveryRepository(), newFakeResidentRepository(), publisher, internalTopic)
	command := &commands.ProcessCreateDeliveryCommand{CommandID: "d1", Apartment: "101"}

	for i := 0; i < 2; i++ {
		if err := writer.Handle(context.Background(), command); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	first := publisher.messages[0].Payload.Data.(*commands.ProcessNotifyDeliveryCommand)
	second := publisher.messages[1].Payload.Data.(*commands.ProcessNotifyDeliveryCommand)
	if first.CommandID != second.CommandID {
		t.Errorf("command ids = %q and %q, want the same", first.CommandID, second.CommandID)
	}
}

func assertNotifyCommand(t *testing.T, publisher *fakePublisher, deliveryID string, notificationType notifier.NotificationType) {
	t.Helper()

	if publisher.topics[0] != internalTopic {
		t.Errorf("topic = %q, want %q", publisher.topics[0], internalTopic)
	}
	message := publisher.messages[0]
	if message.Headers.EventType != commands.ProcessNotifyDeliveryCommandType || message.Headers.Key != deliveryID {
		t.Errorf("headers = %+v, want %s keyed by %s", message.Headers, commands.ProcessNotifyDeliveryCommandType, deliveryID)
	}
	command, ok := message.Payload.Data.(*commands.ProcessNotifyDeliveryCommand)
	if !ok {
		t.Fatalf("payload = %T, want *commands.ProcessNotifyDeliveryCommand", message.Payload.Data)
	}
	if command.DeliveryID != deliveryID || command.NotificationType != notificationType || command.CommandID == "" {
		t.Errorf("command = %+v, want delivery %s and type %s", command, deliveryID, notificationType)
	}
}
