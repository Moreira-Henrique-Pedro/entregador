package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeleteResident struct {
	residentRepository out.ResidentRepository
}

func NewDeleteResident(residentRepository out.ResidentRepository) *DeleteResident {
	return &DeleteResident{
		residentRepository: residentRepository,
	}
}

func (uc *DeleteResident) Execute(ctx context.Context, residentID string) error {
	if residentID == "" {
		return fmt.Errorf("%w: resident_id is required", domain.ErrInvalidResident)
	}

	logger := logger.GetLoggerFromContext(ctx).With("resident_id", residentID)

	current, err := uc.residentRepository.FindByResidentID(ctx, residentID)
	if err != nil {
		return fmt.Errorf("failed to find resident: residentID=%s: %w", residentID, err)
	}
	if current.IsOther() {
		return fmt.Errorf("resident %s: %w", residentID, domain.ErrOtherResidentReadOnly)
	}

	if err := uc.residentRepository.DeleteByResidentID(ctx, residentID); err != nil {
		return fmt.Errorf("failed to delete resident: residentID=%s: %w", residentID, err)
	}

	if current.IsPrimary() {
		if err := uc.residentRepository.EnsurePrimaryResident(ctx, current.Apartment); err != nil {
			return fmt.Errorf("failed to promote new primary resident: apartment=%s: %w", current.Apartment, err)
		}
	}

	logger.Info("Resident deleted", "apartment", current.Apartment)

	return nil
}
