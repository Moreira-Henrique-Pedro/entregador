package readers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type GetResidentsByPhone struct {
	residentRepository interfaces.ResidentRepositoryPort
}

func NewGetResidentsByPhone(residentRepository interfaces.ResidentRepositoryPort) *GetResidentsByPhone {
	return &GetResidentsByPhone{
		residentRepository: residentRepository,
	}
}

func (r *GetResidentsByPhone) Handle(ctx context.Context, phone string) ([]*entities.Resident, error) {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Getting residents by phone", "phone", phone)

	if phone == "" {
		return nil, errors.New("phone is required")
	}

	residents, err := r.residentRepository.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find residents by phone: phone=%s: %w", phone, err)
	}

	return residents, nil
}
