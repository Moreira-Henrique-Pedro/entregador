package request

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

type CreateResident struct {
	Name      string `json:"name" binding:"required"`
	Apartment string `json:"apartment" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
}

func (r *CreateResident) FromDTO() *entities.Resident {
	return &entities.Resident{
		Name:      r.Name,
		Apartment: r.Apartment,
		Phone:     r.Phone,
	}
}

type UpdateResident struct {
	Name      string `json:"name" binding:"required_without_all=Apartment Phone"`
	Apartment string `json:"apartment"`
	Phone     string `json:"phone"`
}

func (r *UpdateResident) FromDTO(residentID string) *entities.Resident {
	return &entities.Resident{
		ResidentID: residentID,
		Name:       r.Name,
		Apartment:  r.Apartment,
		Phone:      r.Phone,
	}
}

type ListResidents struct {
	Apartment string `form:"apartment" binding:"required_without=Phone,excluded_with=Phone"`
	Phone     string `form:"phone"`
}

func (q *ListResidents) ByApartment() bool {
	return q.Apartment != ""
}
