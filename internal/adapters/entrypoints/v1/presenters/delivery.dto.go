package presenters

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

type DeliveryDTO struct {
	ID          string
	ApNum       string
	PackageType string
	Urgency     string
	Status      string
}

func (dto *DeliveryDTO) ToEntity() entities.Delivery {
	return entities.Delivery{
		ID:          dto.ID,
		ApNum:       dto.ApNum,
		PackageType: dto.PackageType,
		Urgency:     dto.Urgency,
		Status:      dto.Status,
	}
}
