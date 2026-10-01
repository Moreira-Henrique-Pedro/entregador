package domain

import "time"

type Resident struct {
	ID         string       `bson:"_id"`
	ResidentID string       `bson:"resident_id"`
	Apartment  string       `bson:"apartment"`
	Name       string       `bson:"name"`
	Phone      string       `bson:"phone"`
	Type       ResidentType `bson:"type"`
	Status     ResidentStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeleteAt   time.Time
}

func (r *Resident) IsOther() bool {
	return r.Type == ResidentTypeOther
}

func (r *Resident) IsPrimary() bool {
	return r.Type == ResidentTypePrimary
}
