// package entities contém as definições de entidades do domínio, representando os dados e comportamentos principais do sistema.
package entities

import "time"

type DeliveryResponse struct {
	ID          string
	ApNum       string
	PackageType string
	Urgency     string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeleteAt    time.Time
	Erro        DeliveryResponseError
}

// DeliveryResponseError é a estrutura de dados para erros de resposta de entrega
type DeliveryResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
