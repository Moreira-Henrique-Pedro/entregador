package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListResidentsByApartment struct {
	residentRepository out.ResidentRepository
}

func NewListResidentsByApartment(residentRepository out.ResidentRepository) *ListResidentsByApartment {
	return &ListResidentsByApartment{
		residentRepository: residentRepository,
	}
}

func (uc *ListResidentsByApartment) Execute(ctx context.Context, apartment string) ([]*domain.Resident, error) {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Getting residents by apartment", "apartment", apartment)

	if apartment == "" {
		return nil, errors.New("apartment is required")
	}

	residents, err := uc.residentRepository.FindByApartment(ctx, apartment)
	if err != nil {
		return nil, fmt.Errorf("failed to find residents by apartment: apartment=%s: %w", apartment, err)
	}

	return residents, nil
}
