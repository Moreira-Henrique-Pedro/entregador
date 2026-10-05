package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListDeliveriesByApartment struct {
	deliveryRepository repositories.DeliveryRepository
}

func NewListDeliveriesByApartment(deliveryRepository repositories.DeliveryRepository) *ListDeliveriesByApartment {
	return &ListDeliveriesByApartment{
		deliveryRepository: deliveryRepository,
	}
}

func (uc *ListDeliveriesByApartment) Execute(ctx context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	if err := entities.ValidateDeliveryFilter(apartment, status); err != nil {
		return nil, err
	}

	logger.GetLoggerFromContext(ctx).Info("Getting deliveries by apartment", "apartment", apartment)

	deliveries, err := uc.deliveryRepository.FindByApartment(ctx, apartment, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find deliveries by apartment: apartment=%s: %w", apartment, err)
	}
	return deliveries, nil
}
