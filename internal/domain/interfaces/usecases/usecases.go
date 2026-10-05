package usecases

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type CreateResident interface {
	Execute(ctx context.Context, resident *entities.Resident) (*entities.Resident, error)
}

type UpdateResident interface {
	Execute(ctx context.Context, resident *entities.Resident) (*entities.Resident, error)
}

type DeleteResident interface {
	Execute(ctx context.Context, residentID string) error
}

type ListResidents interface {
	Execute(ctx context.Context, value string) ([]*entities.Resident, error)
}

type RegisterDelivery interface {
	Execute(ctx context.Context, delivery *entities.Delivery) (*entities.Delivery, error)
}

type DeleteDelivery interface {
	Execute(ctx context.Context, deliveryID string) error
}

type ListDeliveries interface {
	Execute(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error)
}

type NotifyDelivery interface {
	Execute(ctx context.Context, deliveryID string, notificationType entities.NotificationType) error
}

type CreateUser interface {
	Execute(ctx context.Context, user *entities.User, password string) (*entities.User, error)
}
