package entities

import "time"

type Delivery struct {
	ID          string
	DeliveryID  string
	ApNum       string
	PackageType string
	Urgency     string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeleteAt    time.Time
}
