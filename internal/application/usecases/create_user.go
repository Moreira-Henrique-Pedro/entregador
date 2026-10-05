package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type CreateUser struct {
	userRegistry services.UserRegistry
}

func NewCreateUser(userRegistry services.UserRegistry) *CreateUser {
	return &CreateUser{
		userRegistry: userRegistry,
	}
}

func (uc *CreateUser) Execute(ctx context.Context, input *entities.User, password string) (*entities.User, error) {
	if err := input.ValidateForCreate(password); err != nil {
		return nil, err
	}

	user, err := uc.userRegistry.CreateUser(ctx, input, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: email=%s: %w", input.Email, err)
	}

	logger.GetLoggerFromContext(ctx).Info("User created", "user_id", user.ID, "role", string(user.Role))

	return user, nil
}
