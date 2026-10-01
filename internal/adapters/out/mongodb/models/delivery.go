package models

import (
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

type Delivery struct {
	ID                string `bson:"_id"`
	DeliveryID        string `bson:"delivery_id"`
	Apartment         string `bson:"apartment"`
	ResidentID        string `bson:"resident_id"`
	PackageType       string `bson:"package_type"`
	Urgency           string `bson:"urgency"`
	Status            string `bson:"status"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeleteAt          time.Time
	ArrivalNotifiedAt time.Time
	PickupNotifiedAt  time.Time
}

func DeliveryFromEntity(delivery *domain.Delivery) *Delivery {
	return &Delivery{
		ID:                delivery.ID,
		DeliveryID:        delivery.DeliveryID,
		Apartment:         delivery.Apartment,
		ResidentID:        delivery.ResidentID,
		PackageType:       delivery.PackageType,
		Urgency:           delivery.Urgency,
		Status:            string(delivery.Status),
		CreatedAt:         delivery.CreatedAt,
		UpdatedAt:         delivery.UpdatedAt,
		DeleteAt:          delivery.DeleteAt,
		ArrivalNotifiedAt: delivery.ArrivalNotifiedAt,
		PickupNotifiedAt:  delivery.PickupNotifiedAt,
	}
}

func (d *Delivery) ToEntity() *domain.Delivery {
	return &domain.Delivery{
		ID:                d.ID,
		DeliveryID:        d.DeliveryID,
		Apartment:         d.Apartment,
		ResidentID:        d.ResidentID,
		PackageType:       d.PackageType,
		Urgency:           d.Urgency,
		Status:            domain.DeliveryStatus(d.Status),
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
		DeleteAt:          d.DeleteAt,
		ArrivalNotifiedAt: d.ArrivalNotifiedAt,
		PickupNotifiedAt:  d.PickupNotifiedAt,
	}
}
