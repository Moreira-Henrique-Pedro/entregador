package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteResident(t *testing.T) {
	failure := errors.New("mongo down")
	ana := &entities.Resident{ResidentID: "ana", Apartment: "101", Type: entities.ResidentTypePrimary}
	bia := &entities.Resident{ResidentID: "bia", Apartment: "101", Type: entities.ResidentTypeSecondary}
	other := entities.NewOtherResident("101")

	tests := []struct {
		name       string
		residentID string
		find       bool
		current    *entities.Resident
		findErr    error
		delete     bool
		deleteErr  error
		promote    bool
		promoteErr error
		wantErr    error
	}{
		{name: "secondary resident is deleted without touching the primary", residentID: "bia", find: true, current: bia, delete: true},
		{name: "deleting the primary promotes the oldest remaining resident", residentID: "ana", find: true, current: ana, delete: true, promote: true},
		{
			name:       "promotion error is returned",
			residentID: "ana",
			find:       true,
			current:    ana,
			delete:     true,
			promote:    true,
			promoteErr: failure,
			wantErr:    failure,
		},
		{name: "resident id is required", wantErr: entities.ErrInvalidResident},
		{name: "other resident is never deleted", residentID: other.ResidentID, find: true, current: other, wantErr: entities.ErrOtherResidentReadOnly},
		{name: "unknown resident is not found", residentID: "ghost", find: true, findErr: entities.ErrEntityNotFound, wantErr: entities.ErrEntityNotFound},
		{name: "find error is returned", residentID: "ana", find: true, findErr: failure, wantErr: failure},
		{
			name:       "resident deleted concurrently is not found",
			residentID: "ana",
			find:       true,
			current:    ana,
			delete:     true,
			deleteErr:  entities.ErrEntityNotFound,
			wantErr:    entities.ErrEntityNotFound,
		},
		{name: "delete error is returned", residentID: "ana", find: true, current: ana, delete: true, deleteErr: failure, wantErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepository(t)
			if tt.find {
				residents.EXPECT().FindByResidentID(mock.Anything, tt.residentID).Return(tt.current, tt.findErr).Once()
			}
			if tt.delete {
				deleted := residents.EXPECT().DeleteByResidentID(mock.Anything, tt.residentID).Return(tt.deleteErr).Once()
				if tt.promote {
					residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(tt.promoteErr).Once().NotBefore(deleted)
				}
			}

			err := NewDeleteResident(residents).Execute(context.Background(), tt.residentID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
