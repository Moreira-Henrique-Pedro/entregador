package readers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type ResidentsReaderPort interface {
	Handle(ctx context.Context, value string) ([]*entities.Resident, error)
}

type DeliveriesReaderPort interface {
	Handle(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error)
}
