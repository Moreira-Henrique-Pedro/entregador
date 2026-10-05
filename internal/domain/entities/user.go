package entities

import (
	"fmt"
	"slices"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/validation"
)

const minPasswordLength = 8

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleDoorman Role = "doorman"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleDoorman:
		return true
	}
	return false
}

type User struct {
	ID    string
	Email string
	Name  string
	Role  Role
}

func (u *User) HasAnyRole(roles ...Role) bool {
	return len(roles) == 0 || slices.Contains(roles, u.Role)
}

func (u *User) ValidateForCreate(password string) error {
	return validation.First(
		validation.Required(ErrInvalidUser, "email", u.Email),
		validation.Required(ErrInvalidUser, "name", u.Name),
		u.validateRole(),
		validatePassword(password),
	)
}

func (u *User) validateRole() error {
	if !u.Role.IsValid() {
		return fmt.Errorf("%w: invalid role %q", ErrInvalidUser, u.Role)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("%w: password must have at least %d characters", ErrInvalidUser, minPasswordLength)
	}
	return nil
}
