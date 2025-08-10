// package controller contém a lógica de controle do sistema, manipulando as requisições e respostas HTTP, e interagindo com os casos de uso e serviços.
package controller

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/entrypoints/v1/presenters"
	deliveryUsecases "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/usecases/delivery"
	"github.com/gin-gonic/gin"
)

// DeliveryController é a estrutura que representa a controller de entregas
type DeliveryController struct {
	createDeliveryUseCase deliveryUsecases.CreateDeliveryUseCasePort
	deleteDeliveryUseCase deliveryUsecases.DeleteDeliveryUseCasePort
}

// NewDeliveryController cria uma nova instância de DeliveryController
func NewDeliveryController(
	createDeliveryUseCase deliveryUsecases.CreateDeliveryUseCasePort,
	deleteDeliveryUseCase deliveryUsecases.DeleteDeliveryUseCasePort,
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
		responseError(c, http.StatusBadRequest, err.Error())
		return
	}

	delivery := deliveryDTO.ToEntity()
	createdDelivery, err := controller.createDeliveryUseCase.Execute(c.Request.Context(), delivery)
	if err != nil {
		responseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Converte a entrega criada para o formato de resposta
	responseDTO := &presenters.DeliveryResponseDTO{}
	responseDTO.ToResponseDTO(*createdDelivery)
	c.JSON(http.StatusCreated, responseDTO)
}

// DeleteDelivery é o manipulador para excluir uma entrega existente
func (controller *DeliveryController) DeleteDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	if deliveryID == "" {
		responseError(c, http.StatusBadRequest, "ID da entrega não pode ser vazio")
		return
	}

	err := controller.deleteDeliveryUseCase.Execute(c.Request.Context(), deliveryID)
	if err != nil {
		responseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Entrega excluída com sucesso",
	})
}

func responseError(c *gin.Context, statusCode int, message string) {
	errorDTO := presenters.ErrorDTO{
		Code:    statusCode,
		Message: message,
	}
	c.JSON(statusCode, errorDTO)
}
