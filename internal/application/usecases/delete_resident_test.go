package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	outmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

func TestDeleteResident(t *testing.T) {
	failure := errors.New("mongo down")
	ana := &domain.Resident{ResidentID: "ana", Apartment: "101", Type: domain.ResidentTypePrimary}
	bia := &domain.Resident{ResidentID: "bia", Apartment: "101", Type: domain.ResidentTypeSecondary}
	other := domain.NewOtherResident("101")

	tests := []struct {
		name       string
		residentID string
		find       bool
		current    *domain.Resident
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
		{name: "resident id is required", wantErr: domain.ErrInvalidResident},
		{name: "other resident is never deleted", residentID: other.ResidentID, find: true, current: other, wantErr: domain.ErrOtherResidentReadOnly},
		{name: "unknown resident is not found", residentID: "ghost", find: true, findErr: domain.ErrEntityNotFound, wantErr: domain.ErrEntityNotFound},
		{name: "find error is returned", residentID: "ana", find: true, findErr: failure, wantErr: failure},
		{
			name:       "resident deleted concurrently is not found",
			residentID: "ana",
			find:       true,
			current:    ana,
			delete:     true,
			deleteErr:  domain.ErrEntityNotFound,
			wantErr:    domain.ErrEntityNotFound,
		},
		{name: "delete error is returned", residentID: "ana", find: true, current: ana, delete: true, deleteErr: failure, wantErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := outmocks.NewResidentRepository(t)
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
