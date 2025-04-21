package controller

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/entrypoints/v1/presenters"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/usecases"
	"github.com/gin-gonic/gin"
)

type DeliveryController struct {
	createDeliveryUseCase usecases.CreateDeliveryUseCasePort
}

func NewDeliveryController(createDeliveryUseCase usecases.CreateDeliveryUseCasePort) *DeliveryController {
	return &DeliveryController{
		createDeliveryUseCase: createDeliveryUseCase,
	}
}

func (d *DeliveryController) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/v1/entregador")
	v1.POST("", d.CreateDelivery)
}

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
