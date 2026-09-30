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

type ProcessUpdateResident struct {
	residentRepository interfaces.ResidentRepositoryPort
}

func NewProcessUpdateResident(residentRepository interfaces.ResidentRepositoryPort) *ProcessUpdateResident {
	return &ProcessUpdateResident{
		residentRepository: residentRepository,
	}
}

func (w *ProcessUpdateResident) Handle(ctx context.Context, command *commands.ProcessUpdateResidentCommand) error {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Processing ProcessUpdateResident command: commandID=%s", command.CommandID)

	if command.ResidentID == "" {
		logger.Warn("Discarding ProcessUpdateResident command without resident_id: commandID=%s", command.CommandID)
		return nil
	}

	current, err := w.residentRepository.FindByResidentID(ctx, command.ResidentID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for update: ResidentID=%s", command.ResidentID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find resident: residentID=%s: %w", command.ResidentID, err)
	}
	if current.IsOther() {
		logger.Warn("Discarding update of other resident: ResidentID=%s", command.ResidentID)
		return nil
	}

	resident := w.buildResidentEntity(command)
	err = w.residentRepository.Update(ctx, resident)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for update: ResidentID=%s", command.ResidentID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to update resident: residentID=%s: %w", command.ResidentID, err)
	}

	if command.Apartment != "" && command.Apartment != current.Apartment {
		if err := w.residentRepository.EnsureOtherResident(ctx, command.Apartment); err != nil {
			return fmt.Errorf("failed to ensure other resident: apartment=%s: %w", command.Apartment, err)
		}
	}

	logger.Info("Resident updated: ResidentID=%s", command.ResidentID)

	return nil
}

func (w *ProcessUpdateResident) buildResidentEntity(command *commands.ProcessUpdateResidentCommand) *entities.Resident {
	return &entities.Resident{
		ResidentID: command.ResidentID,
		Apartment:  command.Apartment,
		Name:       command.Name,
		Phone:      command.Phone,
	}
}
