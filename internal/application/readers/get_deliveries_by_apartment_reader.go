package readers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type GetDeliveriesByApartment struct {
	deliveryRepository interfaces.DeliveryRepositoryPort
}

func NewGetDeliveriesByApartment(deliveryRepository interfaces.DeliveryRepositoryPort) *GetDeliveriesByApartment {
	return &GetDeliveriesByApartment{
		deliveryRepository: deliveryRepository,
	}
}

func (r *GetDeliveriesByApartment) Handle(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Getting deliveries by apartment", "apartment", apartment)

	if apartment == "" {
		return nil, errors.New("apartment is required")
	}
	if status != nil && !status.IsValid() {
		return nil, fmt.Errorf("invalid delivery status: %s", *status)
	}

	deliveries, err := r.deliveryRepository.FindByApartment(ctx, apartment, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find deliveries by apartment: apartment=%s: %w", apartment, err)
	}

	return deliveries, nil
}
