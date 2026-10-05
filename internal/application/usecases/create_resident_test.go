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

func TestCreateResident(t *testing.T) {
	failure := errors.New("mongo down")
	input := &entities.Resident{Name: "Ana", Apartment: "101", Phone: "11999999999"}

	inserted := &entities.Resident{
		ID: "r1", ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: entities.ResidentTypeSecondary, Status: entities.ResidentStatusCreated,
	}
	promoted := *inserted
	promoted.Type = entities.ResidentTypePrimary

	tests := []struct {
		name          string
		input         *entities.Resident
		insert        bool
		insertErr     error
		ensureOther   bool
		otherErr      error
		ensurePrimary bool
		primaryErr    error
		find          bool
		findErr       error
		want          *entities.Resident
		wantErr       error
	}{
		{
			name:          "resident is inserted as secondary, the apartment primary is ensured and the stored resident is returned",
			input:         input,
			insert:        true,
			ensureOther:   true,
			ensurePrimary: true,
			find:          true,
			want:          &promoted,
		},
		{name: "name is required", input: &entities.Resident{Apartment: "101", Phone: "1"}, wantErr: entities.ErrInvalidResident},
		{name: "apartment is required", input: &entities.Resident{Name: "Ana", Phone: "1"}, wantErr: entities.ErrInvalidResident},
		{name: "phone is required", input: &entities.Resident{Name: "Ana", Apartment: "101"}, wantErr: entities.ErrInvalidResident},
		{name: "insert error is returned and nothing else runs", input: input, insert: true, insertErr: failure, wantErr: failure},
		{name: "ensure other error is returned", input: input, insert: true, ensureOther: true, otherErr: failure, wantErr: failure},
		{
			name:          "ensure primary error is returned",
			input:         input,
			insert:        true,
			ensureOther:   true,
			ensurePrimary: true,
			primaryErr:    failure,
			wantErr:       failure,
		},
		{
			name:          "find error is returned",
			input:         input,
			insert:        true,
			ensureOther:   true,
			ensurePrimary: true,
			find:          true,
			findErr:       failure,
			wantErr:       failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepository(t)
			var calls []*mock.Call
			if tt.insert {
				calls = append(calls, residents.EXPECT().Insert(mock.Anything, inserted).Return(tt.insertErr).Once())
			}
			if tt.ensureOther {
				calls = append(calls, residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(tt.otherErr).Once())
			}
			if tt.ensurePrimary {
				calls = append(calls, residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(tt.primaryErr).Once())
			}
			if tt.find {
				calls = append(calls, residents.EXPECT().FindByResidentID(mock.Anything, "r1").Return(tt.want, tt.findErr).Once())
			}
			mock.InOrder(calls...)

			useCase := NewCreateResident(residents)
			useCase.newID = func() string { return "r1" }
			got, err := useCase.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
