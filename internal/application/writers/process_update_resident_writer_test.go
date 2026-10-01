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

func TestProcessUpdateResident(t *testing.T) {
	failure := errors.New("mongo down")
	ana := &entities.Resident{ResidentID: "ana", Name: "Ana", Apartment: "101", Type: entities.ResidentTypePrimary}
	bia := &entities.Resident{ResidentID: "bia", Name: "Bia", Apartment: "101", Type: entities.ResidentTypeSecondary}
	other := entities.NewOtherResident("101")

	tests := []struct {
		name          string
		command       commands.ProcessUpdateResidentCommand
		current       *entities.Resident
		findErr       error
		wantUpdate    *entities.Resident
		updateErr     error
		wantOther     string
		otherErr      error
		wantPrimaries []string
		primaryErr    error
		wantErr       error
	}{
		{
			name:       "updates only the informed fields",
			command:    commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "Ana Maria"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "Ana Maria"},
		},
		{
			name:          "primary moving out arrives as secondary and the old apartment promotes the next resident",
			command:       commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			current:       ana,
			wantUpdate:    &entities.Resident{ResidentID: "ana", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantOther:     "202",
			wantPrimaries: []string{"101", "202"},
		},
		{
			name:          "secondary moving to an apartment without residents becomes its primary",
			command:       commands.ProcessUpdateResidentCommand{ResidentID: "bia", Apartment: "303"},
			current:       bia,
			wantUpdate:    &entities.Resident{ResidentID: "bia", Apartment: "303"},
			wantOther:     "303",
			wantPrimaries: []string{"101", "303"},
		},
		{
			name:       "same apartment does not ensure other nor primary",
			command:    commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "101"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Apartment: "101"},
		},
		{
			name:    "command without resident id is discarded",
			command: commands.ProcessUpdateResidentCommand{Name: "x"},
		},
		{
			name:    "unknown resident is ignored",
			command: commands.ProcessUpdateResidentCommand{ResidentID: "ghost", Name: "x"},
			findErr: entities.ErrEntityNotFound,
		},
		{
			name:    "other resident is never updated",
			command: commands.ProcessUpdateResidentCommand{ResidentID: other.ResidentID, Name: "x"},
			current: other,
		},
		{
			name:    "find error is returned",
			command: commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "x"},
			findErr: failure,
			wantErr: failure,
		},
		{
			name:       "resident deleted concurrently is ignored",
			command:    commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "x"},
			updateErr:  entities.ErrEntityNotFound,
		},
		{
			name:       "update error is returned",
			command:    commands.ProcessUpdateResidentCommand{ResidentID: "ana", Name: "x"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Name: "x"},
			updateErr:  failure,
			wantErr:    failure,
		},
		{
			name:       "ensure other error is returned",
			command:    commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			current:    ana,
			wantUpdate: &entities.Resident{ResidentID: "ana", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantOther:  "202",
			otherErr:   failure,
			wantErr:    failure,
		},
		{
			name:          "ensure primary error is returned",
			command:       commands.ProcessUpdateResidentCommand{ResidentID: "ana", Apartment: "202"},
			current:       ana,
			wantUpdate:    &entities.Resident{ResidentID: "ana", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantOther:     "202",
			wantPrimaries: []string{"101"},
			primaryErr:    failure,
			wantErr:       failure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			residents := repomocks.NewResidentRepositoryPort(t)
			if tt.command.ResidentID != "" {
				residents.EXPECT().FindByResidentID(mock.Anything, tt.command.ResidentID).Return(tt.current, tt.findErr).Once()
			}
			var calls []*mock.Call
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
			mock.InOrder(calls...)

			err := NewProcessUpdateResident(residents).Handle(context.Background(), &tt.command)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
