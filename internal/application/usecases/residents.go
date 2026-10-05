package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
)

func findResident(ctx context.Context, residentRepository repositories.ResidentRepository, residentID string) (*entities.Resident, error) {
	resident, err := residentRepository.FindByResidentID(ctx, residentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find resident: residentID=%s: %w", residentID, err)
	}
	return resident, nil
}

func findEditableResident(ctx context.Context, residentRepository repositories.ResidentRepository, residentID string) (*entities.Resident, error) {
	resident, err := findResident(ctx, residentRepository, residentID)
	if err != nil {
		return nil, err
	}
	return resident, resident.EnsureEditable()
}

func ensurePrimaryResident(ctx context.Context, residentRepository repositories.ResidentRepository, apartment string) error {
	if err := residentRepository.EnsurePrimaryResident(ctx, apartment); err != nil {
		return fmt.Errorf("failed to ensure primary resident: apartment=%s: %w", apartment, err)
	}
	return nil
}

func ensureOtherResident(ctx context.Context, residentRepository repositories.ResidentRepository, apartment string) error {
	if err := residentRepository.EnsureOtherResident(ctx, apartment); err != nil {
		return fmt.Errorf("failed to ensure other resident: apartment=%s: %w", apartment, err)
	}
	return nil
}

func ensureApartmentResidents(ctx context.Context, residentRepository repositories.ResidentRepository, apartment string) error {
	if err := ensureOtherResident(ctx, residentRepository, apartment); err != nil {
		return err
	}
	return ensurePrimaryResident(ctx, residentRepository, apartment)
}

func findApartmentResidents(ctx context.Context, residentRepository repositories.ResidentRepository, apartment string) ([]*entities.Resident, error) {
	residents, err := residentRepository.FindByApartment(ctx, apartment)
	if err != nil {
		return nil, fmt.Errorf("failed to find residents by apartment: apartment=%s: %w", apartment, err)
	}
	return residents, nil
}
