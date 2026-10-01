// Package in holds the driving ports: the use cases the adapters (HTTP, Kafka) call.
package in

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

type CreateResident interface {
	Execute(ctx context.Context, resident *domain.Resident) (*domain.Resident, error)
}

type UpdateResident interface {
	Execute(ctx context.Context, resident *domain.Resident) (*domain.Resident, error)
}

type DeleteResident interface {
	Execute(ctx context.Context, residentID string) error
}

type ListResidents interface {
	Execute(ctx context.Context, value string) ([]*domain.Resident, error)
}

type RegisterDeliveryInput struct {
	Apartment   string
	ResidentID  string
	PackageType string
	Urgency     string
}

type RegisterDelivery interface {
	Execute(ctx context.Context, input RegisterDeliveryInput) (*domain.Delivery, error)
}

type DeleteDelivery interface {
	Execute(ctx context.Context, deliveryID string) error
}

type ListDeliveries interface {
	Execute(ctx context.Context, apartment string, status *domain.DeliveryStatus) ([]*domain.Delivery, error)
}

type NotifyDelivery interface {
	Execute(ctx context.Context, deliveryID string, notificationType domain.NotificationType) error
}
