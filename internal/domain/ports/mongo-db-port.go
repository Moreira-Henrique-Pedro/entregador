package ports

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type ResidentRepositoryPort interface {
	Create(ctx context.Context, resident *entities.Resident) error
	GetByApartment(ctx context.Context, apartamento string) (*entities.Resident, error)
	Update(ctx context.Context, resident *entities.Resident) error
	Delete(ctx context.Context, apartamento string) error
}
