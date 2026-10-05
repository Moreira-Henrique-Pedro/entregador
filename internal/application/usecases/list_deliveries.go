package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListDeliveries struct {
	deliveryRepository repositories.DeliveryRepository
	residentRepository repositories.ResidentRepository
}

func NewListDeliveries(deliveryRepository repositories.DeliveryRepository, residentRepository repositories.ResidentRepository) *ListDeliveries {
	return &ListDeliveries{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
	}
}

func (uc *ListDeliveries) Execute(ctx context.Context, filter entities.DeliveryFilter) ([]*entities.Delivery, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	logger.GetLoggerFromContext(ctx).Info("Listing deliveries", "apartment", filter.Apartment)

	deliveries, err := uc.deliveryRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find deliveries: apartment=%s: %w", filter.Apartment, err)
	}

	if err := uc.attachResidentNames(ctx, deliveries); err != nil {
		return nil, err
	}
	return deliveries, nil
}

func (uc *ListDeliveries) attachResidentNames(ctx context.Context, deliveries []*entities.Delivery) error {
	residents, err := uc.residentRepository.FindByResidentIDs(ctx, entities.ResidentIDsOf(deliveries))
	if err != nil {
		return fmt.Errorf("failed to find delivery residents: %w", err)
	}
	entities.AttachResidentNames(deliveries, residents)
	return nil
}
