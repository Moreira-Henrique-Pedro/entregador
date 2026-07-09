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
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeleteAt   time.Time
}

func ResidentFromEntity(resident *entities.Resident) *Resident {
	return &Resident{
		ID:         resident.ID,
		ResidentID: resident.ResidentID,
		Apartment:  resident.Apartment,
		Name:       resident.Name,
		Phone:      resident.Phone,
		CreatedAt:  resident.CreatedAt,
		UpdatedAt:  resident.UpdatedAt,
		DeleteAt:   resident.DeleteAt,
	}
}
