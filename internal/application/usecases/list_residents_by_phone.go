package usecases

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListResidentsByPhone struct {
	residentRepository repositories.ResidentRepository
}

func NewListResidentsByPhone(residentRepository repositories.ResidentRepository) *ListResidentsByPhone {
	return &ListResidentsByPhone{
		residentRepository: residentRepository,
	}
}

func (uc *ListResidentsByPhone) Execute(ctx context.Context, phone string) ([]*entities.Resident, error) {
	if err := entities.ValidatePhone(phone); err != nil {
		return nil, err
	}

	logger.GetLoggerFromContext(ctx).Info("Getting residents by phone", "phone", phone)

	residents, err := uc.residentRepository.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find residents by phone: phone=%s: %w", phone, err)
	}
	return residents, nil
}
