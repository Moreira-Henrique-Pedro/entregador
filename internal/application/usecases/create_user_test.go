package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	servicemocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	input := &entities.User{Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleDoorman}
	created := &entities.User{ID: "u1", Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleDoorman}

	tests := []struct {
		name        string
		input       *entities.User
		password    string
		expect      bool
		registryOut *entities.User
		registryErr error
		wantErr     error
	}{
		{name: "creates the user", input: input, password: "12345678", expect: true, registryOut: created},
		{name: "invalid user is not sent to the registry", input: &entities.User{Name: "Ana"}, password: "12345678", wantErr: entities.ErrInvalidUser},
		{name: "already exists", input: input, password: "12345678", expect: true, registryErr: entities.ErrUserAlreadyExists, wantErr: entities.ErrUserAlreadyExists},
		{name: "registry error", input: input, password: "12345678", expect: true, registryErr: errors.New("firebase down"), wantErr: errAny},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := servicemocks.NewUserRegistry(t)
			if tt.expect {
				registry.EXPECT().CreateUser(mock.Anything, tt.input, tt.password).Return(tt.registryOut, tt.registryErr).Once()
			}

			got, err := NewCreateUser(registry).Execute(context.Background(), tt.input, tt.password)

			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.wantErr != errAny {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, created, got)
		})
	}
}
