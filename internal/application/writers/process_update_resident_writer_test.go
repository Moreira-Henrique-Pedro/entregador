package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessUpdateResident(t *testing.T) {
	failure := errors.New("mongo down")

	tests := []struct {
		name           string
		command        commands.ProcessUpdateResidentCommand
		findErr        error
		updateErr      error
		ensureOtherErr error
		wantErr        error
		wantUpdated    bool
		wantEnsured    []string
	}{
		{
			name:        "updates only the informed fields",
			command:     commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "Ana Maria"},
			wantUpdated: true,
		},
		{
			name:        "moving to another apartment ensures its other",
			command:     commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			wantUpdated: true,
			wantEnsured: []string{"202"},
		},
		{
			name:        "same apartment does not ensure other",
			command:     commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "101"},
			wantUpdated: true,
		},
		{
			name:    "command without resident id is discarded",
			command: commands.ProcessUpdateResidentCommand{Name: "x"},
		},
		{
			name:    "unknown resident is ignored",
			command: commands.ProcessUpdateResidentCommand{ResidentID: "ghost", Name: "x"},
		},
		{
			name:    "other resident is never updated",
			command: commands.ProcessUpdateResidentCommand{ResidentID: entities.OtherResidentID("101"), Name: "x"},
		},
		{
			name:    "find error is returned",
			command: commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "x"},
			findErr: failure,
			wantErr: failure,
		},
		{
			name:      "resident deleted concurrently is ignored",
			command:   commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "x"},
			updateErr: entities.ErrEntityNotFound,
		},
		{
			name:      "update error is returned",
			command:   commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "x"},
			updateErr: failure,
			wantErr:   failure,
		},
		{
			name:           "ensure other error is returned",
			command:        commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			ensureOtherErr: failure,
			wantErr:        failure,
			wantUpdated:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(
				&entities.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Type: entities.ResidentTypeResident},
				entities.NewOtherResident("101"),
			)
			residents.findErr = tt.findErr
			residents.updateErr = tt.updateErr
			residents.ensureOtherErr = tt.ensureOtherErr

			err := NewProcessUpdateResident(residents).Handle(context.Background(), &tt.command)

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got := len(residents.updated) == 1; got != tt.wantUpdated {
				t.Fatalf("updated = %v, want updated %v", residents.updated, tt.wantUpdated)
			}
			if tt.wantUpdated {
				updated := residents.updated[0]
				if updated.ResidentID != tt.command.ResidentID || updated.Name != tt.command.Name ||
					updated.Apartment != tt.command.Apartment || updated.Phone != tt.command.Phone {
					t.Errorf("updated = %+v, want fields from %+v", updated, tt.command)
				}
			}
			if !equalStrings(residents.ensuredOthers, tt.wantEnsured) {
				t.Errorf("ensured others = %v, want %v", residents.ensuredOthers, tt.wantEnsured)
			}
		})
	}
}
