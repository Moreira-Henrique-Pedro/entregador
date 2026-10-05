package request

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

type RegisterDelivery struct {
	Apartment   string `json:"apartment" binding:"required"`
	ResidentID  string `json:"resident_id"`
	PackageType string `json:"package_type"`
	Urgency     string `json:"urgency"`
}

func (d *RegisterDelivery) FromDTO() *entities.Delivery {
	return &entities.Delivery{
		Apartment:   d.Apartment,
		ResidentID:  d.ResidentID,
		PackageType: d.PackageType,
		Urgency:     d.Urgency,
	}
}

type ListDeliveries struct {
	Apartment string `form:"apartment" binding:"required"`
	Status    string `form:"status" binding:"omitempty,oneof=pending deleted"`
}

func (q *ListDeliveries) FromDTO() (string, *entities.DeliveryStatus) {
	if q.Status == "" {
		return q.Apartment, nil
	}
	status := entities.DeliveryStatus(q.Status)
	return q.Apartment, &status
}
