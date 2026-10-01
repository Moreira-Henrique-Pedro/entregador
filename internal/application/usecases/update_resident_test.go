package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	outmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

func TestUpdateResident(t *testing.T) {
	failure := errors.New("mongo down")
	ana := &domain.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Type: domain.ResidentTypePrimary}
	bia := &domain.Resident{ResidentID: "bia", Name: "Bia", Apartment: "101", Type: domain.ResidentTypeSecondary}
	other := domain.NewOtherResident("101")
	stored := &domain.Resident{ResidentID: "ana", Name: "stored"}

	tests := []struct {
		name          string
		input         domain.Resident
		current       *domain.Resident
		findErr       error
		wantUpdate    *domain.Resident
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
			input:      domain.Resident{ResidentID: "ana", Name: "Ana Maria"},
			current:    ana,
			wantUpdate: &domain.Resident{ResidentID: "ana", Name: "Ana Maria"},
			reread:     true,
		},
		{
			name:          "primary moving out arrives as secondary and the old apartment promotes the next resident",
			input:         domain.Resident{ResidentID: "ana", Apartment: "202"},
			current:       ana,
			wantUpdate:    &domain.Resident{ResidentID: "ana", Apartment: "202", Type: domain.ResidentTypeSecondary},
			wantOther:     "202",
			wantPrimaries: []string{"101", "202"},
			reread:        true,
		},
		{
			name:          "secondary moving to an apartment without residents becomes its primary",
			input:         domain.Resident{ResidentID: "bia", Apartment: "303"},
			current:       bia,
			wantUpdate:    &domain.Resident{ResidentID: "bia", Apartment: "303"},
			wantOther:     "303",
			wantPrimaries: []string{"101", "303"},
			reread:        true,
		},
		{
			name:       "same apartment does not ensure other nor primary",
			input:      domain.Resident{ResidentID: "ana", Apartment: "101"},
			current:    ana,
			wantUpdate: &domain.Resident{ResidentID: "ana", Apartment: "101"},
			reread:     true,
		},
		{name: "resident id is required", input: domain.Resident{Name: "x"}, wantErr: domain.ErrInvalidResident},
		{name: "at least one field is required", input: domain.Resident{ResidentID: "ana"}, wantErr: domain.ErrInvalidResident},
		{
			name:    "unknown resident is not found",
			input:   domain.Resident{ResidentID: "ghost", Name: "x"},
			findErr: domain.ErrEntityNotFound,
			wantErr: domain.ErrEntityNotFound,
		},
		{
			name:    "other resident is never updated",
			input:   domain.Resident{ResidentID: other.ResidentID, Name: "x"},
			current: other,
			wantErr: domain.ErrOtherResidentReadOnly,
		},
		{name: "find error is returned", input: domain.Resident{ResidentID: "ana", Name: "x"}, findErr: failure, wantErr: failure},
		{
			name:       "resident deleted concurrently is not found",
			input:      domain.Resident{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &domain.Resident{ResidentID: "ana", Name: "x"},
			updateErr:  domain.ErrEntityNotFound,
			wantErr:    domain.ErrEntityNotFound,
		},
		{
			name:       "update error is returned",
			input:      domain.Resident{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &domain.Resident{ResidentID: "ana", Name: "x"},
			updateErr:  failure,
			wantErr:    failure,
		},
		{
			name:       "ensure other error is returned",
			input:      domain.Resident{ResidentID: "ana", Apartment: "202"},
			current:    ana,
			wantUpdate: &domain.Resident{ResidentID: "ana", Apartment: "202", Type: domain.ResidentTypeSecondary},
			wantOther:  "202",
			otherErr:   failure,
			wantErr:    failure,
		},
		{
			name:          "ensure primary error is returned",
			input:         domain.Resident{ResidentID: "ana", Apartment: "202"},
			current:       ana,
			wantUpdate:    &domain.Resident{ResidentID: "ana", Apartment: "202", Type: domain.ResidentTypeSecondary},
			wantOther:     "202",
			wantPrimaries: []string{"101"},
			primaryErr:    failure,
			wantErr:       failure,
		},
		{
			name:       "re-read error is returned",
			input:      domain.Resident{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &domain.Resident{ResidentID: "ana", Name: "x"},
			reread:     true,
			rereadErr:  failure,
			wantErr:    failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := outmocks.NewResidentRepository(t)
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
