package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessCreateDelivery_ResolvesRecipient(t *testing.T) {
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypeResident}

	tests := []struct {
		name          string
		residentID    string
		wantResident  string
		wantEnsuredOK bool
	}{
		{name: "resident of the apartment", residentID: "ana", wantResident: "ana"},
		{name: "no resident informed goes to other", residentID: "", wantResident: entities.OtherResidentID("101"), wantEnsuredOK: true},
		{name: "unknown resident goes to other", residentID: "ghost", wantResident: entities.OtherResidentID("101"), wantEnsuredOK: true},
		{name: "resident from another apartment goes to other", residentID: "bob", wantResident: entities.OtherResidentID("101"), wantEnsuredOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bob := &entities.Resident{ResidentID: "bob", Apartment: "202", Type: entities.ResidentTypeResident}
			residents := newFakeResidentRepository(ana, bob)
			deliveries := &fakeDeliveryRepository{}
			writer := NewProcessCreateDelivery(deliveries, residents)

			err := writer.Handle(context.Background(), &commands.ProcessCreateDeliveryCommand{
				CommandID:  "cmd-1",
				Apartment:  "101",
				ResidentID: tt.residentID,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(deliveries.inserted) != 1 {
				t.Fatalf("inserted %d deliveries, want 1", len(deliveries.inserted))
			}
			delivery := deliveries.inserted[0]
			if delivery.ResidentID != tt.wantResident {
				t.Errorf("resident = %q, want %q", delivery.ResidentID, tt.wantResident)
			}
			if delivery.Status != entities.DeliveryStatusPending {
				t.Errorf("status = %q, want pending", delivery.Status)
			}
			if delivery.DeliveryID != "cmd-1" || delivery.ID != "cmd-1" {
				t.Errorf("ids = %q/%q, want cmd-1", delivery.ID, delivery.DeliveryID)
			}
			if tt.wantEnsuredOK != (len(residents.ensuredOthers) == 1 && residents.ensuredOthers[0] == "101") {
				t.Errorf("ensured others = %v", residents.ensuredOthers)
			}
		})
	}
}

func TestProcessCreateDelivery_DiscardsWithoutApartment(t *testing.T) {
	deliveries := &fakeDeliveryRepository{}
	writer := NewProcessCreateDelivery(deliveries, newFakeResidentRepository())

	if err := writer.Handle(context.Background(), &commands.ProcessCreateDeliveryCommand{CommandID: "cmd-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deliveries.inserted) != 0 {
		t.Errorf("inserted %d deliveries, want 0", len(deliveries.inserted))
	}
}

func TestProcessCreateDelivery_FailsOnRepositoryError(t *testing.T) {
	residents := newFakeResidentRepository()
	residents.findErr = errors.New("mongo down")
	deliveries := &fakeDeliveryRepository{}
	writer := NewProcessCreateDelivery(deliveries, residents)

	err := writer.Handle(context.Background(), &commands.ProcessCreateDeliveryCommand{CommandID: "cmd-1", Apartment: "101", ResidentID: "ana"})
	if err == nil {
		t.Fatal("expected error so the message is retried")
	}
	if len(deliveries.inserted) != 0 {
		t.Errorf("inserted %d deliveries, want 0", len(deliveries.inserted))
	}
}
