package writers

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessCreateResident_EnsuresOther(t *testing.T) {
	residents := newFakeResidentRepository()
	writer := NewProcessCreateResident(residents)

	err := writer.Handle(context.Background(), &commands.ProcessCreateResidentCommand{CommandID: "r1", Name: "Ana", Apartment: "101"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(residents.insertedResult) != 1 || residents.insertedResult[0].Type != entities.ResidentTypeResident {
		t.Errorf("inserted = %+v, want one resident of type resident", residents.insertedResult)
	}
	if len(residents.ensuredOthers) != 1 || residents.ensuredOthers[0] != "101" {
		t.Errorf("ensured others = %v, want [101]", residents.ensuredOthers)
	}
}

func TestProcessUpdateResident(t *testing.T) {
	other := entities.NewOtherResident("101")
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypeResident}

	t.Run("other resident is not updated", func(t *testing.T) {
		residents := newFakeResidentRepository(other)
		err := NewProcessUpdateResident(residents).Handle(context.Background(), &commands.ProcessUpdateResidentCommand{ResidentID: other.ResidentID, Name: "x"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(residents.updated) != 0 {
			t.Errorf("updated = %v, want none", residents.updated)
		}
	})

	t.Run("moving apartment ensures other in the new one", func(t *testing.T) {
		residents := newFakeResidentRepository(ana)
		err := NewProcessUpdateResident(residents).Handle(context.Background(), &commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(residents.updated) != 1 {
			t.Errorf("updated %d residents, want 1", len(residents.updated))
		}
		if len(residents.ensuredOthers) != 1 || residents.ensuredOthers[0] != "202" {
			t.Errorf("ensured others = %v, want [202]", residents.ensuredOthers)
		}
	})

	t.Run("same apartment does not ensure other", func(t *testing.T) {
		residents := newFakeResidentRepository(ana)
		err := NewProcessUpdateResident(residents).Handle(context.Background(), &commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "Ana Maria"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(residents.ensuredOthers) != 0 {
			t.Errorf("ensured others = %v, want none", residents.ensuredOthers)
		}
	})
}

func TestProcessDeleteResident(t *testing.T) {
	other := entities.NewOtherResident("101")
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypeResident}

	tests := []struct {
		name        string
		residentID  string
		wantDeleted bool
	}{
		{name: "regular resident is deleted", residentID: "ana", wantDeleted: true},
		{name: "other resident is not deleted", residentID: other.ResidentID},
		{name: "unknown resident is ignored", residentID: "ghost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(other, ana)
			err := NewProcessDeleteResident(residents).Handle(context.Background(), &commands.ProcessDeleteResidentCommand{ResidentID: tt.residentID})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := len(residents.deletedIDs) == 1; got != tt.wantDeleted {
				t.Errorf("deleted = %v, want %v", residents.deletedIDs, tt.wantDeleted)
			}
		})
	}
}
