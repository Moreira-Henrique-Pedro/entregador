package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestProcessCreateResident(t *testing.T) {
	failure := errors.New("mongo down")
	command := &commands.ProcessCreateResidentCommand{CommandID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999"}

	tests := []struct {
		name           string
		insertErr      error
		ensureOtherErr error
		wantErr        error
		wantInserted   bool
		wantEnsured    []string
	}{
		{name: "creates the resident and ensures the apartment other", wantInserted: true, wantEnsured: []string{"101"}},
		{name: "insert error is returned and other is not ensured", insertErr: failure, wantErr: failure},
		{name: "ensure other error is returned", ensureOtherErr: failure, wantErr: failure, wantInserted: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := newFakeResidentRepository()
			residents.insertErr = tt.insertErr
			residents.ensureOtherErr = tt.ensureOtherErr

			err := NewProcessCreateResident(residents).Handle(context.Background(), command)

			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got := len(residents.inserted) == 1; got != tt.wantInserted {
				t.Fatalf("inserted = %v, want inserted %v", residents.inserted, tt.wantInserted)
			}
			if tt.wantInserted {
				want := entities.Resident{ID: "r1", ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999", Type: entities.ResidentTypeResident}
				if *residents.inserted[0] != want {
					t.Errorf("inserted = %+v, want %+v", *residents.inserted[0], want)
				}
			}
			if !equalStrings(residents.ensuredOthers, tt.wantEnsured) {
				t.Errorf("ensured others = %v, want %v", residents.ensuredOthers, tt.wantEnsured)
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
