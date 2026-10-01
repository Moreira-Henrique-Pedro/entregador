package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ListResidentsByPhone struct {
	residentRepository out.ResidentRepository
}

func NewListResidentsByPhone(residentRepository out.ResidentRepository) *ListResidentsByPhone {
	return &ListResidentsByPhone{
		residentRepository: residentRepository,
	}
}

func (uc *ListResidentsByPhone) Execute(ctx context.Context, phone string) ([]*domain.Resident, error) {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Getting residents by phone", "phone", phone)

	if phone == "" {
		return nil, errors.New("phone is required")
	}

	residents, err := uc.residentRepository.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find residents by phone: phone=%s: %w", phone, err)
	}

	return residents, nil
}
