package interfaces

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type ResidentRepositoryPort interface {
	Insert(ctx context.Context, resident *entities.Resident) error
}
