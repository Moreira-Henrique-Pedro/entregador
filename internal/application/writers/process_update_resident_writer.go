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
	logger := logger.GetLoggerFromContext(ctx).With("resident_id", command.ResidentID)
	logger.Info("Processing ProcessUpdateResident command", "command_id", command.CommandID)

	if command.ResidentID == "" {
		logger.Warn("Discarding ProcessUpdateResident command without resident_id", "command_id", command.CommandID)
		return nil
	}

	current, err := w.residentRepository.FindByResidentID(ctx, command.ResidentID)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for update")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find resident: residentID=%s: %w", command.ResidentID, err)
	}
	if current.IsOther() {
		logger.Warn("Discarding update of other resident")
		return nil
	}

	previousApartment := current.Apartment
	movingApartment := command.Apartment != "" && command.Apartment != previousApartment

	resident := w.buildResidentEntity(command)
	// A primary that moves out leaves the old apartment's primary and arrives as
	// secondary, so the new apartment keeps its current primary.
	if movingApartment && current.IsPrimary() {
		resident.Type = entities.ResidentTypeSecondary
	}

	err = w.residentRepository.Update(ctx, resident)
	if errors.Is(err, entities.ErrEntityNotFound) {
		logger.Warn("Resident not found for update")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to update resident: residentID=%s: %w", command.ResidentID, err)
	}

	if movingApartment {
		if err := w.residentRepository.EnsureOtherResident(ctx, command.Apartment); err != nil {
			return fmt.Errorf("failed to ensure other resident: apartment=%s: %w", command.Apartment, err)
		}
		for _, apartment := range []string{previousApartment, command.Apartment} {
			if err := w.residentRepository.EnsurePrimaryResident(ctx, apartment); err != nil {
				return fmt.Errorf("failed to ensure primary resident: apartment=%s: %w", apartment, err)
			}
		}
	}

	logger.Info("Resident updated")

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
