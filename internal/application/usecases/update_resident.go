package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type UpdateResident struct {
	residentRepository out.ResidentRepository
}

func NewUpdateResident(residentRepository out.ResidentRepository) *UpdateResident {
	return &UpdateResident{
		residentRepository: residentRepository,
	}
}

// Handle applies a partial update: empty fields are left unchanged.
func (uc *UpdateResident) Execute(ctx context.Context, input *domain.Resident) (*domain.Resident, error) {
	if input.ResidentID == "" {
		return nil, fmt.Errorf("%w: resident_id is required", domain.ErrInvalidResident)
	}
	if input.Name == "" && input.Apartment == "" && input.Phone == "" {
		return nil, fmt.Errorf("%w: at least one of name, apartment or phone is required", domain.ErrInvalidResident)
	}

	logger := logger.GetLoggerFromContext(ctx).With("resident_id", input.ResidentID)

	current, err := uc.residentRepository.FindByResidentID(ctx, input.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find resident: residentID=%s: %w", input.ResidentID, err)
	}
	if current.IsOther() {
		return nil, fmt.Errorf("resident %s: %w", input.ResidentID, domain.ErrOtherResidentReadOnly)
	}

	previousApartment := current.Apartment
	movingApartment := input.Apartment != "" && input.Apartment != previousApartment

	resident := uc.buildResidentEntity(input)
	// A moving primary arrives as secondary: the new apartment may already have one (unique index).
	if movingApartment && current.IsPrimary() {
		resident.Type = domain.ResidentTypeSecondary
	}

	if err := uc.residentRepository.Update(ctx, resident); err != nil {
		return nil, fmt.Errorf("failed to update resident: residentID=%s: %w", input.ResidentID, err)
	}

	if movingApartment {
		if err := uc.residentRepository.EnsureOtherResident(ctx, input.Apartment); err != nil {
			return nil, fmt.Errorf("failed to ensure other resident: apartment=%s: %w", input.Apartment, err)
		}
		for _, apartment := range []string{previousApartment, input.Apartment} {
			if err := uc.residentRepository.EnsurePrimaryResident(ctx, apartment); err != nil {
				return nil, fmt.Errorf("failed to ensure primary resident: apartment=%s: %w", apartment, err)
			}
		}
	}

	updated, err := uc.residentRepository.FindByResidentID(ctx, input.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find updated resident: residentID=%s: %w", input.ResidentID, err)
	}

	logger.Info("Resident updated")

	return updated, nil
}

func (uc *UpdateResident) buildResidentEntity(input *domain.Resident) *domain.Resident {
	return &domain.Resident{
		ResidentID: input.ResidentID,
		Apartment:  input.Apartment,
		Name:       input.Name,
		Phone:      input.Phone,
	}
}
