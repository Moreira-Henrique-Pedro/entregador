package ports

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type DeliveryRepositoryPort interface {
	CreateDelivery(ctx context.Context, delivery *entities.Delivery) (*entities.Delivery, error)
	DeleteDeliveryByID(ctx context.Context, id string) error
	GetDeliveryByID(ctx context.Context, id string) (*entities.Delivery, error)
	UpdateDelivery(ctx context.Context, delivery *entities.Delivery) error
}
