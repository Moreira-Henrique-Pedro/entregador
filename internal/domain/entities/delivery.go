package entities

import (
	"fmt"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/validation"
)

type Delivery struct {
	ID                string
	DeliveryID        string
	Apartment         string
	ResidentID        string
	PackageType       string
	Urgency           string
	Status            DeliveryStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeleteAt          time.Time
	ArrivalNotifiedAt time.Time
	PickupNotifiedAt  time.Time
}

func NewPendingDelivery(id, residentID string, input *Delivery) *Delivery {
	return &Delivery{
		ID:          id,
		DeliveryID:  id,
		Apartment:   input.Apartment,
		ResidentID:  residentID,
		PackageType: input.PackageType,
		Urgency:     input.Urgency,
		Status:      DeliveryStatusPending,
	}
}

func (d *Delivery) IsPending() bool {
	return d.Status == DeliveryStatusPending
}

func (d *Delivery) IsPickupNotified() bool {
	return !d.PickupNotifiedAt.IsZero()
}

func (d *Delivery) IsArrivalNotified() bool {
	return !d.ArrivalNotifiedAt.IsZero()
}

func (d *Delivery) ValidateForRegister() error {
	return validation.Required(ErrInvalidDelivery, "apartment", d.Apartment)
}

func (d *Delivery) NotificationSkipReason(notificationType NotificationType) string {
	switch {
	case notificationType == NotificationTypeDeliveryArrived && d.IsArrivalNotified():
		return "arrival already notified"
	case notificationType == NotificationTypeDeliveryArrived && !d.IsPending():
		return "delivery already picked up"
	case notificationType == NotificationTypeDeliveryPickedUp && d.IsPickupNotified():
		return "pickup already notified"
	case notificationType == NotificationTypeDeliveryPickedUp && d.IsPending():
		return "delivery not picked up yet"
	}
	return ""
}

func ValidateDeliveryID(deliveryID string) error {
	return validation.Required(ErrInvalidDelivery, "delivery_id", deliveryID)
}

func ValidateDeliveryFilter(apartment string, status *DeliveryStatus) error {
	return validation.First(
		validation.Required(ErrInvalidDelivery, "apartment", apartment),
		validateStatusFilter(status),
	)
}

func validateStatusFilter(status *DeliveryStatus) error {
	if status != nil && !status.IsValid() {
		return fmt.Errorf("%w: invalid delivery status %q", ErrInvalidDelivery, *status)
	}
	return nil
}
