package entities

import "time"

type Resident struct {
	ID         string `bson:"_id"`
	ResidentID string `bson:"resident_id"`
	Apartment  string `bson:"apartment"`
	Name       string `bson:"name"`
	Phone      string `bson:"phone"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeleteAt   time.Time
}
