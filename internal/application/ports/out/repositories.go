// Package out holds the driven ports: what the use cases need from the outside world.
package out

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

type ResidentRepository interface {
	Insert(ctx context.Context, resident *domain.Resident) error
	EnsureOtherResident(ctx context.Context, apartment string) error
	EnsurePrimaryResident(ctx context.Context, apartment string) error
	Update(ctx context.Context, resident *domain.Resident) error
	DeleteByResidentID(ctx context.Context, residentID string) error
	FindByResidentID(ctx context.Context, residentID string) (*domain.Resident, error)
	FindByApartment(ctx context.Context, apartment string) ([]*domain.Resident, error)
	FindByPhone(ctx context.Context, phone string) ([]*domain.Resident, error)
}

type DeliveryRepository interface {
	Insert(ctx context.Context, delivery *domain.Delivery) error
	FindByDeliveryID(ctx context.Context, deliveryID string) (*domain.Delivery, error)
	FindByApartment(ctx context.Context, apartment string, status *domain.DeliveryStatus) ([]*domain.Delivery, error)
	MarkAsDeleted(ctx context.Context, deliveryID string) error
	MarkArrivalAsNotified(ctx context.Context, deliveryID string) error
	MarkPickupAsNotified(ctx context.Context, deliveryID string) error
}
