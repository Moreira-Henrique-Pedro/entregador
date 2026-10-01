package writers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
)

func TestProcessDeleteResident(t *testing.T) {
	failure := errors.New("mongo down")
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypePrimary}
	bia := &entities.Resident{ResidentID: "bia", Apartment: "101", Type: entities.ResidentTypeSecondary}
	other := entities.NewOtherResident("101")

	tests := []struct {
		name       string
		residentID string
		current    *entities.Resident
		findErr    error
		delete     bool
		deleteErr  error
		promote    bool
		promoteErr error
		wantErr    error
	}{
		{name: "secondary resident is deleted without touching the primary", residentID: "bia", current: bia, delete: true},
		{name: "deleting the primary promotes the oldest remaining resident", residentID: "ana", current: ana, delete: true, promote: true},
		{
			name:       "promotion error is returned",
			residentID: "ana",
			current:    ana,
			delete:     true,
			promote:    true,
			promoteErr: failure,
			wantErr:    failure,
		},
		{name: "other resident is never deleted", residentID: other.ResidentID, current: other},
		{name: "unknown resident is ignored", residentID: "ghost", findErr: entities.ErrEntityNotFound},
		{name: "find error is returned", residentID: "ana", findErr: failure, wantErr: failure},
		{name: "resident deleted concurrently is ignored", residentID: "ana", current: ana, delete: true, deleteErr: entities.ErrEntityNotFound},
		{name: "delete error is returned", residentID: "ana", current: ana, delete: true, deleteErr: failure, wantErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepositoryPort(t)
			residents.EXPECT().FindByResidentID(mock.Anything, tt.residentID).Return(tt.current, tt.findErr).Once()
			if tt.delete {
				deleted := residents.EXPECT().DeleteByResidentID(mock.Anything, tt.residentID).Return(tt.deleteErr).Once()
				if tt.promote {
					residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(tt.promoteErr).Once().NotBefore(deleted)
				}
			}

			err := NewProcessDeleteResident(residents).Handle(context.Background(), &commands.ProcessDeleteResidentCommand{ResidentID: tt.residentID})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
