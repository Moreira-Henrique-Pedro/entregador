package interfaces

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type ResidentRepositoryPort interface {
	Insert(ctx context.Context, resident *entities.Resident) error
	EnsureOtherResident(ctx context.Context, apartment string) error
	EnsurePrimaryResident(ctx context.Context, apartment string) error
	Update(ctx context.Context, resident *entities.Resident) error
	DeleteByResidentID(ctx context.Context, residentID string) error
	FindByResidentID(ctx context.Context, residentID string) (*entities.Resident, error)
	FindByApartment(ctx context.Context, apartment string) ([]*entities.Resident, error)
	FindByPhone(ctx context.Context, phone string) ([]*entities.Resident, error)
}
