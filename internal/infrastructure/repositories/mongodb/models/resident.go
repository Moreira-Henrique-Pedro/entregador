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
	Status     string `bson:"status"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeleteAt   time.Time
}

const legacyResidentType = "resident"

func ResidentFromEntity(resident *entities.Resident) *Resident {
	residentType := resident.Type
	if residentType == "" {
		residentType = entities.ResidentTypeSecondary
	}

	status := resident.Status
	if status == "" {
		status = entities.ResidentStatusCreated
	}

	return &Resident{
		ID:         resident.ID,
		ResidentID: resident.ResidentID,
		Apartment:  resident.Apartment,
		Name:       resident.Name,
		Phone:      resident.Phone,
		Type:       string(residentType),
		Status:     string(status),
		CreatedAt:  resident.CreatedAt,
		UpdatedAt:  resident.UpdatedAt,
		DeleteAt:   resident.DeleteAt,
	}
}

func (r *Resident) ToEntity() *entities.Resident {
	residentType := entities.ResidentType(r.Type)
	if residentType == "" || r.Type == legacyResidentType {
		residentType = entities.ResidentTypeSecondary
	}

	status := entities.ResidentStatus(r.Status)
	if status == "" {
		status = entities.ResidentStatusCreated
		if !r.DeleteAt.IsZero() {
			status = entities.ResidentStatusDeleted
		}
	}

	return &entities.Resident{
		ID:         r.ID,
		ResidentID: r.ResidentID,
		Apartment:  r.Apartment,
		Name:       r.Name,
		Phone:      r.Phone,
		Type:       residentType,
		Status:     status,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
		DeleteAt:   r.DeleteAt,
	}
}
