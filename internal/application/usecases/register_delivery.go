package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/google/uuid"
)

type RegisterDelivery struct {
	deliveryRepository repositories.DeliveryRepository
	residentRepository repositories.ResidentRepository
	scheduler          services.NotificationScheduler
	newID              func() string
}

func NewRegisterDelivery(
	deliveryRepository repositories.DeliveryRepository,
	residentRepository repositories.ResidentRepository,
	scheduler services.NotificationScheduler,
) *RegisterDelivery {
	return &RegisterDelivery{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
		scheduler:          scheduler,
		newID:              uuid.NewString,
	}
}

func (uc *RegisterDelivery) Execute(ctx context.Context, input *entities.Delivery) (*entities.Delivery, error) {
	if err := input.ValidateForRegister(); err != nil {
		return nil, err
	}

	if err := uc.ensureApartmentHasResident(ctx, input.Apartment); err != nil {
		return nil, err
	}

	residentID, err := uc.resolveRecipient(ctx, input)
	if err != nil {
		return nil, err
	}

	delivery := entities.NewPendingDelivery(uc.newID(), residentID, input)
	if err := uc.insert(ctx, delivery); err != nil {
		return nil, err
	}

	uc.scheduleArrivalNotification(ctx, delivery)

	return delivery, nil
}

func (uc *RegisterDelivery) ensureApartmentHasResident(ctx context.Context, apartment string) error {
	residents, err := findApartmentResidents(ctx, uc.residentRepository, apartment)
	if err != nil {
		return err
	}
	if !entities.HasResidentToReceive(residents) {
		return fmt.Errorf("apartment %s: %w", apartment, entities.ErrNoResidentInApartment)
	}
	return nil
}

func (uc *RegisterDelivery) resolveRecipient(ctx context.Context, input *entities.Delivery) (string, error) {
	resident, err := uc.findInformedResident(ctx, input)
	if err != nil || resident != nil {
		return residentIDOf(resident), err
	}

	if err := ensureOtherResident(ctx, uc.residentRepository, input.Apartment); err != nil {
		return "", err
	}
	return entities.OtherResidentID(input.Apartment), nil
}

func (uc *RegisterDelivery) findInformedResident(ctx context.Context, input *entities.Delivery) (*entities.Resident, error) {
	if input.ResidentID == "" {
		return nil, nil
	}

	log := logger.GetLoggerFromContext(ctx).With("resident_id", input.ResidentID, "apartment", input.Apartment)

	resident, err := uc.residentRepository.FindByResidentID(ctx, input.ResidentID)
	switch {
	case errors.Is(err, entities.ErrEntityNotFound):
		log.Warn("Resident not found, using other")
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed to find delivery recipient: residentID=%s: %w", input.ResidentID, err)
	case !resident.LivesIn(input.Apartment):
		log.Warn("Resident does not live in the delivery apartment, using other")
		return nil, nil
	}
	return resident, nil
}

func (uc *RegisterDelivery) insert(ctx context.Context, delivery *entities.Delivery) error {
	if err := uc.deliveryRepository.Insert(ctx, delivery); err != nil {
		return fmt.Errorf("failed to insert delivery: deliveryID=%s: %w", delivery.DeliveryID, err)
	}
	logger.GetLoggerFromContext(ctx).Info("Delivery registered", "delivery_id", delivery.DeliveryID, "resident_id", delivery.ResidentID)
	return nil
}

func (uc *RegisterDelivery) scheduleArrivalNotification(ctx context.Context, delivery *entities.Delivery) {
	if err := uc.scheduler.Schedule(ctx, delivery.DeliveryID, entities.NotificationTypeDeliveryArrived); err != nil {
		logger.GetLoggerFromContext(ctx).Error("Failed to schedule arrival notification", "delivery_id", delivery.DeliveryID, "error", err.Error())
	}
}

func residentIDOf(resident *entities.Resident) string {
	if resident == nil {
		return ""
	}
	return resident.ResidentID
}
