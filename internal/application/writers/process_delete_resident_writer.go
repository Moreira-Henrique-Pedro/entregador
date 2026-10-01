package writers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ProcessDeleteResident struct {
	residentRepository interfaces.ResidentRepositoryPort
}

func NewProcessDeleteResident(residentRepository interfaces.ResidentRepositoryPort) *ProcessDeleteResident {
	return &ProcessDeleteResident{
		residentRepository: residentRepository,
	}
}

func (w *ProcessDeleteResident) Handle(ctx context.Context, command *commands.ProcessDeleteResidentCommand) error {
	logger := logger.GetLoggerFromContext(ctx).With("resident_id", command.ResidentID)
	logger.Info("Processing ProcessDeleteResident command", "command_id", command.CommandID)

	current, err := w.residentRepository.FindByResidentID(ctx, command.ResidentID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for delete")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find resident: residentID=%s: %w", command.ResidentID, err)
	}
	if current.IsOther() {
		logger.Warn("Discarding delete of other resident")
		return nil
	}

	err = w.residentRepository.DeleteByResidentID(ctx, command.ResidentID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for delete")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete resident: residentID=%s: %w", command.ResidentID, err)
	}

	if current.IsPrimary() {
		if err := w.residentRepository.EnsurePrimaryResident(ctx, current.Apartment); err != nil {
			return fmt.Errorf("failed to promote new primary resident: apartment=%s: %w", current.Apartment, err)
		}
	}

	logger.Info("Resident deleted", "apartment", current.Apartment)

	return nil
}
