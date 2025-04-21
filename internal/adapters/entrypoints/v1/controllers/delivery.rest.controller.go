// package controller contém a lógica de controle do sistema, manipulando as requisições e respostas HTTP, e interagindo com os casos de uso e serviços.
package controller

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/entrypoints/v1/presenters"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/usecases"
	"github.com/gin-gonic/gin"
)

// DeliveryController é a estrutura que representa a controller de entregas
type DeliveryController struct {
	createDeliveryUseCase usecases.CreateDeliveryUseCasePort
}

// NewDeliveryController cria uma nova instância de DeliveryController
func NewDeliveryController(createDeliveryUseCase usecases.CreateDeliveryUseCasePort) *DeliveryController {
	return &DeliveryController{
		createDeliveryUseCase: createDeliveryUseCase,
	}
}

// RegisterRoutes registra as rotas do controlador de entregas
func (d *DeliveryController) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/v1/entregador")
	v1.POST("", d.CreateDelivery)
}

// CreateDelivery é o manipulador para criar uma nova entrega
func (d *DeliveryController) CreateDelivery(c *gin.Context) {
	var deliveryDTO presenters.DeliveryDTO

	if err := c.ShouldBindJSON(&deliveryDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delivery := deliveryDTO.ToEntity()
	createdDelivery, err := d.createDeliveryUseCase.Execute(c.Request.Context(), delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Entrega registrada com sucesso",
		"id":      createdDelivery.ID,
	})
}
