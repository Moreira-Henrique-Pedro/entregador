// package ports contem interfaces que definem os métodos para interagir com diferentes serviços e repositórios.
package ports

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

// ResidentRepositoryPort é a interface que define os métodos para o repositório de residentes
type ResidentRepositoryPort interface {
	Create(ctx context.Context, resident *entities.Resident) error
	GetByApartment(ctx context.Context, apartamento string) (*entities.Resident, error)
	Update(ctx context.Context, resident *entities.Resident) error
	Delete(ctx context.Context, apartamento string) error
}
