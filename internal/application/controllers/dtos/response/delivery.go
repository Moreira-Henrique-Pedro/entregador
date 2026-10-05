package response

import (
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type Delivery struct {
	DeliveryID   string     `json:"delivery_id"`
	Apartment    string     `json:"apartment"`
	ResidentID   string     `json:"resident_id"`
	ResidentName string     `json:"resident_name,omitempty"`
	PackageType  string     `json:"package_type"`
	Urgency      string     `json:"urgency"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func NewDelivery(delivery *entities.Delivery) Delivery {
	response := Delivery{
		DeliveryID:   delivery.DeliveryID,
		Apartment:    delivery.Apartment,
		ResidentID:   delivery.ResidentID,
		ResidentName: delivery.ResidentName,
		PackageType:  delivery.PackageType,
		Urgency:      delivery.Urgency,
		Status:       string(delivery.Status),
		CreatedAt:    delivery.CreatedAt,
		UpdatedAt:    delivery.UpdatedAt,
	}
	if !delivery.DeleteAt.IsZero() {
		deletedAt := delivery.DeleteAt
		response.DeletedAt = &deletedAt
	}
	return response
}

func NewDeliveries(deliveries []*entities.Delivery) []Delivery {
	response := make([]Delivery, 0, len(deliveries))
	for _, delivery := range deliveries {
		response = append(response, NewDelivery(delivery))
	}
	return response
}
