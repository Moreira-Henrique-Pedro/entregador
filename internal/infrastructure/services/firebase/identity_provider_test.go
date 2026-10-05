package firebase

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type fakeAuthClient struct {
	token       *auth.Token
	verifyErr   error
	createErr   error
	claimsErr   error
	deleteErr   error
	gotClaims   map[string]interface{}
	deletedUIDs []string
}

func (f *fakeAuthClient) VerifyIDToken(context.Context, string) (*auth.Token, error) {
	return f.token, f.verifyErr
}

func (f *fakeAuthClient) CreateUser(context.Context, *auth.UserToCreate) (*auth.UserRecord, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &auth.UserRecord{UserInfo: &auth.UserInfo{UID: "u1", Email: "ana@condominio.com", DisplayName: "Ana"}}, nil
}

func (f *fakeAuthClient) SetCustomUserClaims(_ context.Context, _ string, claims map[string]interface{}) error {
	f.gotClaims = claims
	return f.claimsErr
}

func (f *fakeAuthClient) DeleteUser(_ context.Context, uid string) error {
	f.deletedUIDs = append(f.deletedUIDs, uid)
	return f.deleteErr
}

func TestVerifyToken(t *testing.T) {
	client := &fakeAuthClient{token: &auth.Token{UID: "u1", Claims: map[string]interface{}{
		"email": "ana@condominio.com", "name": "Ana", "role": "admin",
	}}}

	user, err := NewIdentityProvider(client).VerifyToken(context.Background(), "token")

	require.NoError(t, err)
	assert.Equal(t, &entities.User{ID: "u1", Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleAdmin}, user)
}

func TestVerifyToken_WithoutRole(t *testing.T) {
	client := &fakeAuthClient{token: &auth.Token{UID: "u1", Claims: map[string]interface{}{}}}

	user, err := NewIdentityProvider(client).VerifyToken(context.Background(), "token")

	require.NoError(t, err)
	assert.Equal(t, entities.Role(""), user.Role)
}

func TestVerifyToken_Invalid(t *testing.T) {
	_, err := NewIdentityProvider(&fakeAuthClient{verifyErr: errors.New("expired")}).VerifyToken(context.Background(), "token")

	assert.ErrorIs(t, err, entities.ErrUnauthenticated)
}

func TestCreateUser(t *testing.T) {
	client := &fakeAuthClient{}
	input := &entities.User{Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleDoorman}

	user, err := NewIdentityProvider(client).CreateUser(context.Background(), input, "12345678")

	require.NoError(t, err)
	assert.Equal(t, &entities.User{ID: "u1", Email: "ana@condominio.com", Name: "Ana", Role: entities.RoleDoorman}, user)
	assert.Equal(t, map[string]interface{}{"role": "doorman"}, client.gotClaims)
	assert.Empty(t, client.deletedUIDs)
}

func TestCreateUser_CreateError(t *testing.T) {
	client := &fakeAuthClient{createErr: errors.New("firebase down")}

	_, err := NewIdentityProvider(client).CreateUser(context.Background(), &entities.User{Role: entities.RoleAdmin}, "12345678")

	require.Error(t, err)
	assert.Nil(t, client.gotClaims)
}

func TestCreateUser_RoleErrorDeletesTheUser(t *testing.T) {
	tests := []struct {
		name      string
		deleteErr error
	}{
		{name: "user deleted"},
		{name: "delete also fails", deleteErr: errors.New("delete failed")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeAuthClient{claimsErr: errors.New("claims failed"), deleteErr: tt.deleteErr}

			_, err := NewIdentityProvider(client).CreateUser(context.Background(), &entities.User{Role: entities.RoleAdmin}, "12345678")

			require.Error(t, err)
			assert.Equal(t, []string{"u1"}, client.deletedUIDs)
		})
	}
}
