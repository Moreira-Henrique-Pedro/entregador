package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateResident(t *testing.T) {
	failure := errors.New("mongo down")
	ana := &entities.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Type: entities.ResidentTypePrimary}
	bia := &entities.Resident{ResidentID: "bia", Name: "Bia", Apartment: "101", Type: entities.ResidentTypeSecondary}
	other := entities.NewOtherResident("101")
	stored := &entities.Resident{ResidentID: "ana", Name: "stored"}

	tests := []struct {
		name          string
		input         entities.Resident
		current       *entities.Resident
		findErr       error
		wantUpdate    *entities.Resident
		updateErr     error
		wantOther     string
		otherErr      error
		wantPrimaries []string
		primaryErr    error
		reread        bool
		rereadErr     error
		wantErr       error
	}{
		{
			name:       "updates only the informed fields and returns the stored resident",
			input:      entities.Resident{ResidentID: "ana", Name: "Ana Maria"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "Ana Maria"},
			reread:     true,
		},
		{
			name:          "primary moving out arrives as secondary and the old apartment promotes the next resident",
			input:         entities.Resident{ResidentID: "ana", Apartment: "202"},
			current:       ana,
			wantUpdate:    &entities.Resident{ResidentID: "ana", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantOther:     "202",
			wantPrimaries: []string{"101", "202"},
			reread:        true,
		},
		{
			name:          "secondary moving to an apartment without residents becomes its primary",
			input:         entities.Resident{ResidentID: "bia", Apartment: "303"},
			current:       bia,
			wantUpdate:    &entities.Resident{ResidentID: "bia", Apartment: "303"},
			wantOther:     "303",
			wantPrimaries: []string{"101", "303"},
			reread:        true,
		},
		{
			name:       "same apartment does not ensure other nor primary",
			input:      entities.Resident{ResidentID: "ana", Apartment: "101"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Apartment: "101"},
			reread:     true,
		},
		{name: "resident id is required", input: entities.Resident{Name: "x"}, wantErr: entities.ErrInvalidResident},
		{name: "at least one field is required", input: entities.Resident{ResidentID: "ana"}, wantErr: entities.ErrInvalidResident},
		{
			name:    "unknown resident is not found",
			input:   entities.Resident{ResidentID: "ghost", Name: "x"},
			findErr: entities.ErrEntityNotFound,
			wantErr: entities.ErrEntityNotFound,
		},
		{
			name:    "other resident is never updated",
			input:   entities.Resident{ResidentID: other.ResidentID, Name: "x"},
			current: other,
			wantErr: entities.ErrOtherResidentReadOnly,
		},
		{name: "find error is returned", input: entities.Resident{ResidentID: "ana", Name: "x"}, findErr: failure, wantErr: failure},
		{
			name:       "resident deleted concurrently is not found",
			input:      entities.Resident{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "x"},
			updateErr:  entities.ErrEntityNotFound,
			wantErr:    entities.ErrEntityNotFound,
		},
		{
			name:       "update error is returned",
			input:      entities.Resident{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "x"},
			updateErr:  failure,
			wantErr:    failure,
		},
		{
			name:       "ensure other error is returned",
			input:      entities.Resident{ResidentID: "ana", Apartment: "202"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantOther:  "202",
			otherErr:   failure,
			wantErr:    failure,
		},
		{
			name:          "ensure primary error is returned",
			input:         entities.Resident{ResidentID: "ana", Apartment: "202"},
			current:       ana,
			wantUpdate:    &entities.Resident{ResidentID: "ana", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantOther:     "202",
			wantPrimaries: []string{"101"},
			primaryErr:    failure,
			wantErr:       failure,
		},
		{
			name:       "re-read error is returned",
			input:      entities.Resident{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "x"},
			reread:     true,
			rereadErr:  failure,
			wantErr:    failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepository(t)
			var calls []*mock.Call
			if tt.current != nil || tt.findErr != nil {
				calls = append(calls, residents.EXPECT().FindByResidentID(mock.Anything, tt.input.ResidentID).Return(tt.current, tt.findErr).Once())
			}
			if tt.wantUpdate != nil {
				calls = append(calls, residents.EXPECT().Update(mock.Anything, tt.wantUpdate).Return(tt.updateErr).Once())
			}
			if tt.wantOther != "" {
				calls = append(calls, residents.EXPECT().EnsureOtherResident(mock.Anything, tt.wantOther).Return(tt.otherErr).Once())
			}
			for i, apartment := range tt.wantPrimaries {
				var err error
				if i == 0 {
					err = tt.primaryErr
				}
				calls = append(calls, residents.EXPECT().EnsurePrimaryResident(mock.Anything, apartment).Return(err).Once())
			}
			if tt.reread {
				calls = append(calls, residents.EXPECT().FindByResidentID(mock.Anything, tt.input.ResidentID).Return(stored, tt.rereadErr).Once())
			}
			mock.InOrder(calls...)

			got, err := NewUpdateResident(residents).Execute(context.Background(), &tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, stored, got)
		})
	}
}
