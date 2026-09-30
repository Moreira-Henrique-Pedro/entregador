package models

import (
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type Resident struct {
	ID         string `bson:"_id"`
	ResidentID string `bson:"resident_id"`
	Apartment  string `bson:"apartment"`
	Name       string `bson:"name"`
	Phone      string `bson:"phone"`
	Type       string `bson:"type"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeleteAt   time.Time
}

func ResidentFromEntity(resident *entities.Resident) *Resident {
	residentType := resident.Type
	if residentType == "" {
		residentType = entities.ResidentTypeResident
	}

	return &Resident{
		ID:         resident.ID,
		ResidentID: resident.ResidentID,
		Apartment:  resident.Apartment,
		Name:       resident.Name,
		Phone:      resident.Phone,
		Type:       string(residentType),
		CreatedAt:  resident.CreatedAt,
		UpdatedAt:  resident.UpdatedAt,
		DeleteAt:   resident.DeleteAt,
	}
}

func (r *Resident) ToEntity() *entities.Resident {
	// Residents stored before the type field existed are regular residents.
	residentType := entities.ResidentType(r.Type)
	if residentType == "" {
		residentType = entities.ResidentTypeResident
	}

	return &entities.Resident{
		ID:         r.ID,
		ResidentID: r.ResidentID,
		Apartment:  r.Apartment,
		Name:       r.Name,
		Phone:      r.Phone,
		Type:       residentType,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
		DeleteAt:   r.DeleteAt,
	}
}
