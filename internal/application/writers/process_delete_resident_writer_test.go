package writers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessDeleteResident(t *testing.T) {
	failure := errors.New("mongo down")
	now := time.Now()

	tests := []struct {
		name             string
		residentID       string
		findErr          error
		deleteErr        error
		ensurePrimaryErr error
		wantErr          error
		wantDeleted      bool
		wantPrimaries    []string
		wantNewPrimary   string
	}{
		{name: "secondary resident is deleted without touching the primary", residentID: "bia", wantDeleted: true},
		{
			name:           "deleting the primary promotes the oldest remaining resident",
			residentID:     "ana",
			wantDeleted:    true,
			wantPrimaries:  []string{"101"},
			wantNewPrimary: "bia",
		},
		{
			name:             "promotion error is returned",
			residentID:       "ana",
			ensurePrimaryErr: failure,
			wantErr:          failure,
			wantDeleted:      true,
		},
		{name: "other resident is never deleted", residentID: entities.OtherResidentID("101")},
		{name: "unknown resident is ignored", residentID: "ghost"},
		{name: "find error is returned", residentID: "ana", findErr: failure, wantErr: failure},
		{name: "resident deleted concurrently is ignored", residentID: "ana", deleteErr: entities.ErrEntityNotFound},
		{name: "delete error is returned", residentID: "ana", deleteErr: failure, wantErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(
				&entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypePrimary, CreatedAt: now.Add(-3 * time.Hour)},
				&entities.Resident{ResidentID: "bia", Apartment: "101", Type: entities.ResidentTypeSecondary, CreatedAt: now.Add(-2 * time.Hour)},
				&entities.Resident{ResidentID: "caio", Apartment: "101", Type: entities.ResidentTypeSecondary, CreatedAt: now.Add(-time.Hour)},
				entities.NewOtherResident("101"),
			)
			residents.findErr = tt.findErr
			residents.deleteErr = tt.deleteErr
			residents.ensurePrimaryErr = tt.ensurePrimaryErr

			err := NewProcessDeleteResident(residents).Handle(context.Background(), &commands.ProcessDeleteResidentCommand{ResidentID: tt.residentID})

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got := len(residents.deletedIDs) == 1; got != tt.wantDeleted {
				t.Errorf("deleted = %v, want deleted %v", residents.deletedIDs, tt.wantDeleted)
			}
			if !equalStrings(residents.ensuredPrimaries, tt.wantPrimaries) {
				t.Errorf("ensured primaries = %v, want %v", residents.ensuredPrimaries, tt.wantPrimaries)
			}
			if tt.wantNewPrimary != "" {
				if got := residents.residents[tt.wantNewPrimary].Type; got != entities.ResidentTypePrimary {
					t.Errorf("%s type = %q, want primary", tt.wantNewPrimary, got)
				}
			}
		})
	}
}
