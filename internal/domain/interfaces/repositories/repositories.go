package repositories

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type ResidentRepository interface {
	Insert(ctx context.Context, resident *entities.Resident) error
	EnsureOtherResident(ctx context.Context, apartment string) error
	EnsurePrimaryResident(ctx context.Context, apartment string) error
	Update(ctx context.Context, resident *entities.Resident) error
	DeleteByResidentID(ctx context.Context, residentID string) error
	FindByResidentID(ctx context.Context, residentID string) (*entities.Resident, error)
	FindByApartment(ctx context.Context, apartment string) ([]*entities.Resident, error)
	FindByPhone(ctx context.Context, phone string) ([]*entities.Resident, error)
}

type DeliveryRepository interface {
	Insert(ctx context.Context, delivery *entities.Delivery) error
	FindByDeliveryID(ctx context.Context, deliveryID string) (*entities.Delivery, error)
	FindByApartment(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error)
	MarkAsDeleted(ctx context.Context, deliveryID string) error
	MarkArrivalAsNotified(ctx context.Context, deliveryID string) error
	MarkPickupAsNotified(ctx context.Context, deliveryID string) error
}
