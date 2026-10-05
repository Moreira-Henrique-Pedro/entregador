package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type DeleteResident struct {
	residentRepository repositories.ResidentRepository
}

func NewDeleteResident(residentRepository repositories.ResidentRepository) *DeleteResident {
	return &DeleteResident{
		residentRepository: residentRepository,
	}
}

func (uc *DeleteResident) Execute(ctx context.Context, residentID string) error {
	if err := entities.ValidateResidentID(residentID); err != nil {
		return err
	}

	current, err := findEditableResident(ctx, uc.residentRepository, residentID)
	if err != nil {
		return err
	}

	if err := uc.delete(ctx, current); err != nil {
		return err
	}

	logger.GetLoggerFromContext(ctx).Info("Resident deleted", "resident_id", residentID, "apartment", current.Apartment)

	return nil
}

func (uc *DeleteResident) delete(ctx context.Context, resident *entities.Resident) error {
	if err := uc.residentRepository.DeleteByResidentID(ctx, resident.ResidentID); err != nil {
		return fmt.Errorf("failed to delete resident: residentID=%s: %w", resident.ResidentID, err)
	}
	if !resident.IsPrimary() {
		return nil
	}
	return ensurePrimaryResident(ctx, uc.residentRepository, resident.Apartment)
}
