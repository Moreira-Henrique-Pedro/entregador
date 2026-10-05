package services

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*entities.User, error)
}

type UserRegistry interface {
	CreateUser(ctx context.Context, user *entities.User, password string) (*entities.User, error)
}
