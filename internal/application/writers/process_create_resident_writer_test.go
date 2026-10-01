package writers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessCreateResident(t *testing.T) {
	failure := errors.New("mongo down")
	command := &commands.ProcessCreateResidentCommand{CommandID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999"}
	existingPrimary := func() *entities.Resident {
		return &entities.Resident{ResidentID: "bia", Apartment: "101", Type: entities.ResidentTypePrimary, CreatedAt: time.Now().Add(-time.Hour)}
	}

	tests := []struct {
		name             string
		existing         []*entities.Resident
		insertErr        error
		ensureOtherErr   error
		ensurePrimaryErr error
		wantErr          error
		wantInserted     bool
		wantEnsured      []string
		wantPrimaries    []string
		wantStoredType   entities.ResidentType
	}{
		{
			name:           "first resident of the apartment becomes primary",
			wantInserted:   true,
			wantEnsured:    []string{"101"},
			wantPrimaries:  []string{"101"},
			wantStoredType: entities.ResidentTypePrimary,
		},
		{
			name:           "resident of an apartment with primary stays secondary",
			existing:       []*entities.Resident{existingPrimary(), entities.NewOtherResident("101")},
			wantInserted:   true,
			wantEnsured:    []string{"101"},
			wantPrimaries:  []string{"101"},
			wantStoredType: entities.ResidentTypeSecondary,
		},
		{name: "insert error is returned and nothing else runs", insertErr: failure, wantErr: failure},
		{name: "ensure other error is returned", ensureOtherErr: failure, wantErr: failure, wantInserted: true},
		{
			name:             "ensure primary error is returned",
			ensurePrimaryErr: failure,
			wantErr:          failure,
			wantInserted:     true,
			wantEnsured:      []string{"101"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(tt.existing...)
			residents.insertErr = tt.insertErr
			residents.ensureOtherErr = tt.ensureOtherErr
			residents.ensurePrimaryErr = tt.ensurePrimaryErr

			err := NewProcessCreateResident(residents).Handle(context.Background(), command)

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got := len(residents.inserted) == 1; got != tt.wantInserted {
				t.Fatalf("inserted = %v, want inserted %v", residents.inserted, tt.wantInserted)
			}
			if tt.wantInserted {
				want := entities.Resident{ID: "r1", ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999", Type: entities.ResidentTypeSecondary, Status: entities.ResidentStatusCreated}
				if *residents.inserted[0] != want {
					t.Errorf("inserted = %+v, want %+v", *residents.inserted[0], want)
				}
			}
			if !equalStrings(residents.ensuredOthers, tt.wantEnsured) {
				t.Errorf("ensured others = %v, want %v", residents.ensuredOthers, tt.wantEnsured)
			}
			if !equalStrings(residents.ensuredPrimaries, tt.wantPrimaries) {
				t.Errorf("ensured primaries = %v, want %v", residents.ensuredPrimaries, tt.wantPrimaries)
			}
			if tt.wantStoredType != "" {
				if got := residents.residents["r1"].Type; got != tt.wantStoredType {
					t.Errorf("stored type = %q, want %q", got, tt.wantStoredType)
				}
			}
		})
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
