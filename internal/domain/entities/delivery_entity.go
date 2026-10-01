package entities

import "time"

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
