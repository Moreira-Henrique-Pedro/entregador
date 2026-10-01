package writers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessUpdateResident(t *testing.T) {
	failure := errors.New("mongo down")
	now := time.Now()

	tests := []struct {
		name             string
		command          commands.ProcessUpdateResidentCommand
		findErr          error
		updateErr        error
		ensureOtherErr   error
		ensurePrimaryErr error
		wantErr          error
		wantUpdated      bool
		wantUpdatedType  entities.ResidentType
		wantEnsured      []string
		wantPrimaries    []string
		// wantStoredTypes checks the type of residents after the command, by id.
		wantStoredTypes map[string]entities.ResidentType
	}{
		{
			name:        "updates only the informed fields",
			command:     commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "Ana Maria"},
			wantUpdated: true,
		},
		{
			name:            "primary moving out arrives as secondary and the old apartment promotes the next resident",
			command:         commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			wantUpdated:     true,
			wantUpdatedType: entities.ResidentTypeSecondary,
			wantEnsured:     []string{"202"},
			wantPrimaries:   []string{"101", "202"},
			wantStoredTypes: map[string]entities.ResidentType{
				"ana":   entities.ResidentTypeSecondary,
				"bia":   entities.ResidentTypePrimary,
				"carla": entities.ResidentTypePrimary,
			},
		},
		{
			name:          "secondary moving to an apartment without residents becomes its primary",
			command:       commands.ProcessUpdateResidentCommand{ResidentID: "bia", Apartment: "303"},
			wantUpdated:   true,
			wantEnsured:   []string{"303"},
			wantPrimaries: []string{"101", "303"},
			wantStoredTypes: map[string]entities.ResidentType{
				"ana": entities.ResidentTypePrimary,
				"bia": entities.ResidentTypePrimary,
			},
		},
		{
			name:        "same apartment does not ensure other nor primary",
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
			name:            "ensure other error is returned",
			command:         commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			ensureOtherErr:  failure,
			wantErr:         failure,
			wantUpdated:     true,
			wantUpdatedType: entities.ResidentTypeSecondary,
		},
		{
			name:             "ensure primary error is returned",
			command:          commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			ensurePrimaryErr: failure,
			wantErr:          failure,
			wantUpdated:      true,
			wantUpdatedType:  entities.ResidentTypeSecondary,
			wantEnsured:      []string{"202"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(
				&entities.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Type: entities.ResidentTypePrimary, CreatedAt: now.Add(-2 * time.Hour)},
				&entities.Resident{ResidentID: "bia", Name: "Bia", Apartment: "101", Type: entities.ResidentTypeSecondary, CreatedAt: now.Add(-time.Hour)},
				&entities.Resident{ResidentID: "carla", Name: "Carla", Apartment: "202", Type: entities.ResidentTypePrimary, CreatedAt: now},
				entities.NewOtherResident("101"),
			)
			residents.findErr = tt.findErr
			residents.updateErr = tt.updateErr
			residents.ensureOtherErr = tt.ensureOtherErr
			residents.ensurePrimaryErr = tt.ensurePrimaryErr

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
				if updated.Type != tt.wantUpdatedType {
					t.Errorf("updated type = %q, want %q", updated.Type, tt.wantUpdatedType)
				}
			}
			if !equalStrings(residents.ensuredOthers, tt.wantEnsured) {
				t.Errorf("ensured others = %v, want %v", residents.ensuredOthers, tt.wantEnsured)
			}
			if !equalStrings(residents.ensuredPrimaries, tt.wantPrimaries) {
				t.Errorf("ensured primaries = %v, want %v", residents.ensuredPrimaries, tt.wantPrimaries)
			}
			for id, want := range tt.wantStoredTypes {
				if got := residents.residents[id].Type; got != want {
					t.Errorf("%s type = %q, want %q", id, got, want)
				}
			}
		})
	}
}
