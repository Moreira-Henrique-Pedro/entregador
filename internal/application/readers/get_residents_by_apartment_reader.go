package readers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type GetResidentsByApartment struct {
	residentRepository interfaces.ResidentRepositoryPort
}

func NewGetResidentsByApartment(residentRepository interfaces.ResidentRepositoryPort) *GetResidentsByApartment {
	return &GetResidentsByApartment{
		residentRepository: residentRepository,
	}
}

func (r *GetResidentsByApartment) Handle(ctx context.Context, apartment string) ([]*entities.Resident, error) {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Getting residents by apartment: Apartment=%s", apartment)

	if apartment == "" {
		return nil, errors.New("apartment is required")
	}

	residents, err := r.residentRepository.FindByApartment(ctx, apartment)
	if err != nil {
		return nil, fmt.Errorf("failed to find residents by apartment: apartment=%s: %w", apartment, err)
	}

	return residents, nil
}
