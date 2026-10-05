package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleIsValid(t *testing.T) {
	assert.True(t, RoleAdmin.IsValid())
	assert.True(t, RoleDoorman.IsValid())
	assert.False(t, Role("resident").IsValid())
	assert.False(t, Role("").IsValid())
}

func TestUserHasAnyRole(t *testing.T) {
	doorman := &User{Role: RoleDoorman}

	assert.True(t, doorman.HasAnyRole())
	assert.True(t, doorman.HasAnyRole(RoleAdmin, RoleDoorman))
	assert.False(t, doorman.HasAnyRole(RoleAdmin))
}

func TestUserValidateForCreate(t *testing.T) {
	valid := User{Email: "ana@condominio.com", Name: "Ana", Role: RoleDoorman}

	tests := []struct {
		name     string
		mutate   func(u *User)
		password string
		wantErr  string
	}{
		{name: "valid", password: "12345678"},
		{name: "missing email", mutate: func(u *User) { u.Email = "" }, password: "12345678", wantErr: "invalid user: email is required"},
		{name: "missing name", mutate: func(u *User) { u.Name = "" }, password: "12345678", wantErr: "invalid user: name is required"},
		{name: "invalid role", mutate: func(u *User) { u.Role = "resident" }, password: "12345678", wantErr: `invalid user: invalid role "resident"`},
		{name: "short password", password: "1234567", wantErr: "invalid user: password must have at least 8 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := valid
			if tt.mutate != nil {
				tt.mutate(&user)
			}

			err := user.ValidateForCreate(tt.password)

			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
			assert.ErrorIs(t, err, ErrInvalidUser)
		})
	}
}
