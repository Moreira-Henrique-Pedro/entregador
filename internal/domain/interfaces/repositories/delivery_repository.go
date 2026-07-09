package interfaces

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type DeliveryRepositoryPort interface {
	Create(ctx context.Context, delivery *entities.Delivery) error
	GetByID(ctx context.Context, id string) (*entities.Delivery, error)
	Update(ctx context.Context, delivery *entities.Delivery) error
	DeleteByDeliveryID(ctx context.Context, deliveryID string) error
}
