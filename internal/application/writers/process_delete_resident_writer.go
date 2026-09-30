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
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Processing ProcessDeleteResident command: commandID=%s", command.CommandID)

	current, err := w.residentRepository.FindByResidentID(ctx, command.ResidentID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for delete: ResidentID=%s", command.ResidentID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find resident: residentID=%s: %w", command.ResidentID, err)
	}
	if current.IsOther() {
		logger.Warn("Discarding delete of other resident: ResidentID=%s", command.ResidentID)
		return nil
	}

	err = w.residentRepository.DeleteByResidentID(ctx, command.ResidentID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for delete: ResidentID=%s", command.ResidentID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete resident: residentID=%s: %w", command.ResidentID, err)
	}

	logger.Info("Resident deleted: ResidentID=%s", command.ResidentID)

	return nil
}
