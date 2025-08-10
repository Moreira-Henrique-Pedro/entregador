// package presenters contém as definições de estruturas de dados que são usadas para transferir dados entre diferentes partes do sistema, como entre controladores e serviços. Essas estruturas são frequentemente usadas para serializar e desserializar dados em formatos como JSON ou XML, facilitando a comunicação entre componentes do sistema.
package presenters

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"

// DeliveryResponseDTO é a estrutura de dados de transferência de entrega
type DeliveryResponseDTO struct {
	ID          string
	ApNum       string
	PackageType string
	Urgency     string
	Status      string
	Erro        ErrorDTO
}

// ErrorDTO é a estrutura de dados para erros
type ErrorDTO struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToResponseDTO converte uma entities.DeliveryResponse em DeliveryResponseDTO
func (dto *DeliveryResponseDTO) ToResponseDTO(delivery entities.DeliveryResponse) {
	dto.ID = delivery.ID
	dto.ApNum = delivery.ApNum
	dto.PackageType = delivery.PackageType
	dto.Urgency = delivery.Urgency
	dto.Status = delivery.Status
	dto.Erro = ErrorDTO(delivery.Erro)
}
