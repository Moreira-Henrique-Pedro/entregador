package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/google/uuid"
)

type CreateResident struct {
	residentRepository out.ResidentRepository
	newID              func() string
}

func NewCreateResident(residentRepository out.ResidentRepository) *CreateResident {
	return &CreateResident{
		residentRepository: residentRepository,
		newID:              uuid.NewString,
	}
}

func (uc *CreateResident) Execute(ctx context.Context, input *domain.Resident) (*domain.Resident, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidResident)
	}
	if input.Apartment == "" {
		return nil, fmt.Errorf("%w: apartment is required", domain.ErrInvalidResident)
	}

	resident := uc.buildResidentEntity(input)
	logger := logger.GetLoggerFromContext(ctx).With("resident_id", resident.ResidentID, "apartment", resident.Apartment)

	if err := uc.residentRepository.Insert(ctx, resident); err != nil {
		return nil, fmt.Errorf("failed to insert resident: %w", err)
	}

	if err := uc.residentRepository.EnsureOtherResident(ctx, resident.Apartment); err != nil {
		return nil, fmt.Errorf("failed to ensure other resident: apartment=%s: %w", resident.Apartment, err)
	}

	if err := uc.residentRepository.EnsurePrimaryResident(ctx, resident.Apartment); err != nil {
		return nil, fmt.Errorf("failed to ensure primary resident: apartment=%s: %w", resident.Apartment, err)
	}

	// Re-read: the resident may have been promoted to primary.
	created, err := uc.residentRepository.FindByResidentID(ctx, resident.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find created resident: residentID=%s: %w", resident.ResidentID, err)
	}

	logger.Info("Resident created")

	return created, nil
}

func (uc *CreateResident) buildResidentEntity(input *domain.Resident) *domain.Resident {
	id := uc.newID()
	return &domain.Resident{
		ID:         id,
		ResidentID: id,
		Apartment:  input.Apartment,
		Name:       input.Name,
		Phone:      input.Phone,
		Type:       domain.ResidentTypeSecondary,
		Status:     domain.ResidentStatusCreated,
	}
}
