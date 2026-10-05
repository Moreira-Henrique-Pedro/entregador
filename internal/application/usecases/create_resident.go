package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/google/uuid"
)

type CreateResident struct {
	residentRepository repositories.ResidentRepository
	newID              func() string
}

func NewCreateResident(residentRepository repositories.ResidentRepository) *CreateResident {
	return &CreateResident{
		residentRepository: residentRepository,
		newID:              uuid.NewString,
	}
}

func (uc *CreateResident) Execute(ctx context.Context, input *entities.Resident) (*entities.Resident, error) {
	if err := input.ValidateForCreate(); err != nil {
		return nil, err
	}

	resident := entities.NewResident(uc.newID(), input.Name, input.Apartment, input.Phone)
	if err := uc.insert(ctx, resident); err != nil {
		return nil, err
	}

	logger.GetLoggerFromContext(ctx).Info("Resident created", "resident_id", resident.ResidentID, "apartment", resident.Apartment)

	return findResident(ctx, uc.residentRepository, resident.ResidentID)
}

func (uc *CreateResident) insert(ctx context.Context, resident *entities.Resident) error {
	if err := uc.residentRepository.Insert(ctx, resident); err != nil {
		return fmt.Errorf("failed to insert resident: %w", err)
	}
	return ensureApartmentResidents(ctx, uc.residentRepository, resident.Apartment)
}
