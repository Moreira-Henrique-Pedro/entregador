package interfaces

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type DeliveryRepositoryPort interface {
	Insert(ctx context.Context, delivery *entities.Delivery) error
	FindByDeliveryID(ctx context.Context, deliveryID string) (*entities.Delivery, error)
	// FindByApartment returns every delivery of the apartment; a nil status means any status.
	FindByApartment(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error)
	// MarkAsDeleted sets a pending delivery as deleted (picked up).
	MarkAsDeleted(ctx context.Context, deliveryID string) error
	MarkArrivalAsNotified(ctx context.Context, deliveryID string) error
	MarkPickupAsNotified(ctx context.Context, deliveryID string) error
}
