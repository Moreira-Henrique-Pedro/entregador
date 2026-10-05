package response

import (
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type Resident struct {
	ResidentID string    `json:"resident_id"`
	Name       string    `json:"name"`
	Apartment  string    `json:"apartment"`
	Phone      string    `json:"phone"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewResident(resident *entities.Resident) Resident {
	return Resident{
		ResidentID: resident.ResidentID,
		Name:       resident.Name,
		Apartment:  resident.Apartment,
		Phone:      resident.Phone,
		Type:       string(resident.Type),
		Status:     string(resident.Status),
		CreatedAt:  resident.CreatedAt,
		UpdatedAt:  resident.UpdatedAt,
	}
}

func NewResidents(residents []*entities.Resident) []Resident {
	response := make([]Resident, 0, len(residents))
	for _, resident := range residents {
		response = append(response, NewResident(resident))
	}
	return response
}
