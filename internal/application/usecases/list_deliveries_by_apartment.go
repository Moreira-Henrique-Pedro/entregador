package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListDeliveriesByApartment struct {
	deliveryRepository out.DeliveryRepository
}

func NewListDeliveriesByApartment(deliveryRepository out.DeliveryRepository) *ListDeliveriesByApartment {
	return &ListDeliveriesByApartment{
		deliveryRepository: deliveryRepository,
	}
}

func (uc *ListDeliveriesByApartment) Execute(ctx context.Context, apartment string, status *domain.DeliveryStatus) ([]*domain.Delivery, error) {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Getting deliveries by apartment", "apartment", apartment)

	if apartment == "" {
		return nil, errors.New("apartment is required")
	}
	if status != nil && !status.IsValid() {
		return nil, fmt.Errorf("invalid delivery status: %s", *status)
	}

	deliveries, err := uc.deliveryRepository.FindByApartment(ctx, apartment, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find deliveries by apartment: apartment=%s: %w", apartment, err)
	}

	return deliveries, nil
}
