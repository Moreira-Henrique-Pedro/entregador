package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type UpdateResident struct {
	residentRepository repositories.ResidentRepository
}

func NewUpdateResident(residentRepository repositories.ResidentRepository) *UpdateResident {
	return &UpdateResident{
		residentRepository: residentRepository,
	}
}

func (uc *UpdateResident) Execute(ctx context.Context, input *entities.Resident) (*entities.Resident, error) {
	if err := input.ValidateForUpdate(); err != nil {
		return nil, err
	}

	current, err := findEditableResident(ctx, uc.residentRepository, input.ResidentID)
	if err != nil {
		return nil, err
	}

	if err := uc.update(ctx, current, input); err != nil {
		return nil, err
	}

	if err := uc.rebalanceApartments(ctx, current, input.Apartment); err != nil {
		return nil, err
	}

	logger.GetLoggerFromContext(ctx).Info("Resident updated", "resident_id", input.ResidentID)

	return findResident(ctx, uc.residentRepository, input.ResidentID)
}

func (uc *UpdateResident) update(ctx context.Context, current, input *entities.Resident) error {
	if err := uc.residentRepository.Update(ctx, changes(current, input)); err != nil {
		return fmt.Errorf("failed to update resident: residentID=%s: %w", input.ResidentID, err)
	}
	return nil
}

func (uc *UpdateResident) rebalanceApartments(ctx context.Context, current *entities.Resident, newApartment string) error {
	if !current.MovesTo(newApartment) {
		return nil
	}
	if err := ensureOtherResident(ctx, uc.residentRepository, newApartment); err != nil {
		return err
	}
	for _, apartment := range []string{current.Apartment, newApartment} {
		if err := ensurePrimaryResident(ctx, uc.residentRepository, apartment); err != nil {
			return err
		}
	}
	return nil
}

func changes(current, input *entities.Resident) *entities.Resident {
	resident := &entities.Resident{
		ResidentID: input.ResidentID,
		Apartment:  input.Apartment,
		Name:       input.Name,
		Phone:      input.Phone,
	}
	if current.IsPrimary() && current.MovesTo(input.Apartment) {
		resident.Type = entities.ResidentTypeSecondary
	}
	return resident
}
