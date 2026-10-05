package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
)

type ListApartments struct {
	residentRepository repositories.ResidentRepository
}

func NewListApartments(residentRepository repositories.ResidentRepository) *ListApartments {
	return &ListApartments{
		residentRepository: residentRepository,
	}
}

func (uc *ListApartments) Execute(ctx context.Context) ([]string, error) {
	apartments, err := uc.residentRepository.ListApartments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list apartments: %w", err)
	}
	return apartments, nil
}
