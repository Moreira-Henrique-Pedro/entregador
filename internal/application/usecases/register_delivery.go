package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/google/uuid"
)

type RegisterDelivery struct {
	deliveryRepository out.DeliveryRepository
	residentRepository out.ResidentRepository
	scheduler          out.NotificationScheduler
	newID              func() string
}

func NewRegisterDelivery(
	deliveryRepository out.DeliveryRepository,
	residentRepository out.ResidentRepository,
	scheduler out.NotificationScheduler,
) *RegisterDelivery {
	return &RegisterDelivery{
		deliveryRepository: deliveryRepository,
		residentRepository: residentRepository,
		scheduler:          scheduler,
		newID:              uuid.NewString,
	}
}

// Execute saves the delivery synchronously and only schedules the arrival notification,
// which the worker sends asynchronously.
func (uc *RegisterDelivery) Execute(ctx context.Context, input in.RegisterDeliveryInput) (*domain.Delivery, error) {
	if input.Apartment == "" {
		return nil, fmt.Errorf("%w: apartment is required", domain.ErrInvalidDelivery)
	}

	logger := logger.GetLoggerFromContext(ctx).With("apartment", input.Apartment)

	hasResident, err := uc.apartmentHasResident(ctx, input.Apartment)
	if err != nil {
		return nil, fmt.Errorf("failed to find apartment residents: apartment=%s: %w", input.Apartment, err)
	}
	if !hasResident {
		return nil, fmt.Errorf("apartment %s: %w", input.Apartment, domain.ErrNoResidentInApartment)
	}

	residentID, err := uc.resolveRecipient(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve delivery recipient: apartment=%s: %w", input.Apartment, err)
	}

	delivery := uc.buildDeliveryEntity(input, residentID)
	if err := uc.deliveryRepository.Insert(ctx, delivery); err != nil {
		return nil, fmt.Errorf("failed to insert delivery: deliveryID=%s: %w", delivery.DeliveryID, err)
	}

	logger = logger.With("delivery_id", delivery.DeliveryID, "resident_id", delivery.ResidentID)
	logger.Info("Delivery registered")

	// The delivery is already saved: failing here would make the client retry and duplicate it.
	if err := uc.scheduler.Schedule(ctx, delivery.DeliveryID, domain.NotificationTypeDeliveryArrived); err != nil {
		logger.Error("Failed to schedule arrival notification", "error", err.Error())
	}

	return delivery, nil
}

func (uc *RegisterDelivery) apartmentHasResident(ctx context.Context, apartment string) (bool, error) {
	residents, err := uc.residentRepository.FindByApartment(ctx, apartment)
	if err != nil {
		return false, err
	}
	for _, resident := range residents {
		if !resident.IsOther() {
			return true, nil
		}
	}
	return false, nil
}

func (uc *RegisterDelivery) resolveRecipient(ctx context.Context, input in.RegisterDeliveryInput) (string, error) {
	logger := logger.GetLoggerFromContext(ctx)

	if input.ResidentID != "" {
		resident, err := uc.residentRepository.FindByResidentID(ctx, input.ResidentID)
		switch {
		case err == nil && resident.Apartment == input.Apartment:
			return resident.ResidentID, nil
		case err == nil:
			logger.Warn("Resident does not live in the delivery apartment, using other", "resident_id", input.ResidentID, "apartment", input.Apartment)
		case errors.Is(err, domain.ErrEntityNotFound):
			logger.Warn("Resident not found, using other", "resident_id", input.ResidentID, "apartment", input.Apartment)
		default:
			return "", err
		}
	}

	if err := uc.residentRepository.EnsureOtherResident(ctx, input.Apartment); err != nil {
		return "", err
	}
	return domain.OtherResidentID(input.Apartment), nil
}

func (uc *RegisterDelivery) buildDeliveryEntity(input in.RegisterDeliveryInput, residentID string) *domain.Delivery {
	id := uc.newID()
	return &domain.Delivery{
		ID:          id,
		DeliveryID:  id,
		Apartment:   input.Apartment,
		ResidentID:  residentID,
		PackageType: input.PackageType,
		Urgency:     input.Urgency,
		Status:      domain.DeliveryStatusPending,
	}
}
