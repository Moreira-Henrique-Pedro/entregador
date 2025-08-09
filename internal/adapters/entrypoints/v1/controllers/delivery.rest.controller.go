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
	deleteDeliveryUseCase usecases.DeleteDeliveryUseCasePort
}

// NewDeliveryController cria uma nova instância de DeliveryController
func NewDeliveryController(
	createDeliveryUseCase usecases.CreateDeliveryUseCasePort,
	deleteDeliveryUseCase usecases.DeleteDeliveryUseCasePort,
) *DeliveryController {
	return &DeliveryController{
		createDeliveryUseCase: createDeliveryUseCase,
		deleteDeliveryUseCase: deleteDeliveryUseCase,
	}
}

// RegisterRoutes registra as rotas do controlador de entregas
func (controller *DeliveryController) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/v1/entregador")
	v1.POST("", controller.CreateDelivery)
	v1.DELETE("/:id", controller.DeleteDelivery)
}

// CreateDelivery é o manipulador para criar uma nova entrega
func (controller *DeliveryController) CreateDelivery(c *gin.Context) {
	var deliveryDTO presenters.DeliveryDTO

	if err := c.ShouldBindJSON(&deliveryDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delivery := deliveryDTO.ToEntity()
	createdDelivery, err := controller.createDeliveryUseCase.Execute(c.Request.Context(), delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Entrega registrada com sucesso",
		"id":      createdDelivery.ID,
	})
}

// DeleteDelivery é o manipulador para excluir uma entrega existente
func (controller *DeliveryController) DeleteDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da entrega não pode ser vazio"})
		return
	}

	err := controller.deleteDeliveryUseCase.Execute(c.Request.Context(), deliveryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Entrega excluída com sucesso",
	})
}
