package controller

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/entrypoints/v1/presenters"
	"github.com/gin-gonic/gin"
)

type DeliveryController struct {
	//
}

func NewDeliveryController() *DeliveryController {
	return &DeliveryController{
		//
	}
}

func (d *DeliveryController) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/v1/entregador")
	v1.POST("/box", d.CreateDelivery)
}

func (d *DeliveryController) CreateDelivery(c *gin.Context) {
	var box presenters.DeliveryDTO

	if err := c.ShouldBindJSON(&box); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": ""})
}
