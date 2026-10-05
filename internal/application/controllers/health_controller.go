package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", c.health)
}

func (c *HealthController) health(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}
