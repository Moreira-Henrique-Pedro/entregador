package usecases

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListResidentsByApartment struct {
	residentRepository repositories.ResidentRepository
}

func NewListResidentsByApartment(residentRepository repositories.ResidentRepository) *ListResidentsByApartment {
	return &ListResidentsByApartment{
		residentRepository: residentRepository,
	}
}

func (uc *ListResidentsByApartment) Execute(ctx context.Context, apartment string) ([]*entities.Resident, error) {
	if err := entities.ValidateApartment(apartment); err != nil {
		return nil, err
	}

	logger.GetLoggerFromContext(ctx).Info("Getting residents by apartment", "apartment", apartment)

	return findApartmentResidents(ctx, uc.residentRepository, apartment)
}
