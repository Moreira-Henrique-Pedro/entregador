// package presenters contém as definições de estruturas de dados que são usadas para transferir dados entre diferentes partes do sistema, como entre controladores e serviços. Essas estruturas são frequentemente usadas para serializar e desserializar dados em formatos como JSON ou XML, facilitando a comunicação entre componentes do sistema.
package presenters

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

// DeliveryDTO é a estrutura de dados de transferência de entrega
type DeliveryDTO struct {
	ID          string
	ApNum       string
	PackageType string
	Urgency     string
	Status      string
}

// ToEntity converte o DeliveryDTO em uma entities.Delivery
func (dto *DeliveryDTO) ToEntity() entities.Delivery {
	return entities.Delivery{
		ID:          dto.ID,
		ApNum:       dto.ApNum,
		PackageType: dto.PackageType,
		Urgency:     dto.Urgency,
		Status:      dto.Status,
	}
}
