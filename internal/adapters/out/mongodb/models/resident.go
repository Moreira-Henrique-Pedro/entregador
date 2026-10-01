package models

import (
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
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

// Stored before residents were split into primary and secondary.
const legacyResidentType = "resident"

func ResidentFromEntity(resident *domain.Resident) *Resident {
	residentType := resident.Type
	if residentType == "" {
		residentType = domain.ResidentTypeSecondary
	}

	status := resident.Status
	if status == "" {
		status = domain.ResidentStatusCreated
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

func (r *Resident) ToEntity() *domain.Resident {
	// Empty or legacy types are secondary until EnsurePrimaryResident promotes one.
	residentType := domain.ResidentType(r.Type)
	if residentType == "" || r.Type == legacyResidentType {
		residentType = domain.ResidentTypeSecondary
	}

	// Residents stored before the status field existed get it from deleteat.
	status := domain.ResidentStatus(r.Status)
	if status == "" {
		status = domain.ResidentStatusCreated
		if !r.DeleteAt.IsZero() {
			status = domain.ResidentStatusDeleted
		}
	}

	return &domain.Resident{
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
