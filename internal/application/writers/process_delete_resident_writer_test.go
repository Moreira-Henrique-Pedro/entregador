package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessDeleteResident(t *testing.T) {
	failure := errors.New("mongo down")

	tests := []struct {
		name        string
		residentID  string
		findErr     error
		deleteErr   error
		wantErr     error
		wantDeleted bool
	}{
		{name: "regular resident is deleted", residentID: "ana", wantDeleted: true},
		{name: "other resident is never deleted", residentID: entities.OtherResidentID("101")},
		{name: "unknown resident is ignored", residentID: "ghost"},
		{name: "find error is returned", residentID: "ana", findErr: failure, wantErr: failure},
		{name: "resident deleted concurrently is ignored", residentID: "ana", deleteErr: entities.ErrEntityNotFound},
		{name: "delete error is returned", residentID: "ana", deleteErr: failure, wantErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository(
				&entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypeResident},
				entities.NewOtherResident("101"),
			)
			residents.findErr = tt.findErr
			residents.deleteErr = tt.deleteErr

			err := NewProcessDeleteResident(residents).Handle(context.Background(), &commands.ProcessDeleteResidentCommand{ResidentID: tt.residentID})

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got := len(residents.deletedIDs) == 1; got != tt.wantDeleted {
				t.Errorf("deleted = %v, want deleted %v", residents.deletedIDs, tt.wantDeleted)
			}
		})
	}
}
