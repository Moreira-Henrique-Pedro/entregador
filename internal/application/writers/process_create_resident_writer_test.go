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

func TestProcessCreateResident(t *testing.T) {
	failure := errors.New("mongo down")
	command := &commands.ProcessCreateResidentCommand{CommandID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999"}

	inserted := &entities.Resident{
		ID: "r1", ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: entities.ResidentTypeSecondary, Status: entities.ResidentStatusCreated,
	}

	tests := []struct {
		name          string
		insertErr     error
		ensureOther   bool
		otherErr      error
		ensurePrimary bool
		primaryErr    error
		wantErr       error
	}{
		{
			name:          "resident is inserted as secondary and the apartment primary is ensured",
			ensureOther:   true,
			ensurePrimary: true,
		},
		{name: "insert error is returned and nothing else runs", insertErr: failure, wantErr: failure},
		{name: "ensure other error is returned", ensureOther: true, otherErr: failure, wantErr: failure},
		{
			name:          "ensure primary error is returned",
			ensureOther:   true,
			ensurePrimary: true,
			primaryErr:    failure,
			wantErr:       failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepositoryPort(t)
			insert := residents.EXPECT().Insert(mock.Anything, inserted).Return(tt.insertErr).Once()
			if tt.ensureOther {
				other := residents.EXPECT().EnsureOtherResident(mock.Anything, "101").Return(tt.otherErr).Once().NotBefore(insert)
				if tt.ensurePrimary {
					residents.EXPECT().EnsurePrimaryResident(mock.Anything, "101").Return(tt.primaryErr).Once().NotBefore(other)
				}
			}

			err := NewProcessCreateResident(residents).Handle(context.Background(), command)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
