package writers

import (
	"context"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ProcessCreateResident struct {
	residentRepository interfaces.ResidentRepositoryPort
}

func NewProcessCreateResident(residentRepository interfaces.ResidentRepositoryPort) *ProcessCreateResident {
	return &ProcessCreateResident{
		residentRepository: residentRepository,
	}
}

func (w *ProcessCreateResident) Handle(ctx context.Context, command *commands.ProcessCreateResidentCommand) error {
	logger := logger.GetLoggerFromContext(ctx)
	logger.Info("Processing ProcessCreateResident command: commandID=%s", command.CommandID)

	resident := w.buildResidentEntity(command)
	if err := w.residentRepository.Insert(ctx, resident); err != nil {
		return fmt.Errorf("failed to insert resident: %w", err)
	}

	if err := w.residentRepository.EnsureOtherResident(ctx, command.Apartment); err != nil {
		return fmt.Errorf("failed to ensure other resident: apartment=%s: %w", command.Apartment, err)
	}

	logger.Info("Resident created: Name=%s, Apartment=%s, Phone=%s", command.Name, command.Apartment, command.Phone)

	return nil
}

func (w *ProcessCreateResident) buildResidentEntity(command *commands.ProcessCreateResidentCommand) *entities.Resident {
	return &entities.Resident{
		ID:         command.CommandID,
		ResidentID: command.CommandID,
		Apartment:  command.Apartment,
		Name:       command.Name,
		Phone:      command.Phone,
		Type:       entities.ResidentTypeResident,
	}
}
