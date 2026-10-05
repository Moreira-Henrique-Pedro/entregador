package controllers

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/request"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases"
	"github.com/gin-gonic/gin"
)

const deliveryResource = "delivery"

type DeliveriesController struct {
	authorizer       Authorizer
	listDeliveries   usecases.ListDeliveries
	registerDelivery usecases.RegisterDelivery
	deleteDelivery   usecases.DeleteDelivery
}

type DeliveriesControllerDependencies struct {
	Authorizer       Authorizer
	ListDeliveries   usecases.ListDeliveries
	RegisterDelivery usecases.RegisterDelivery
	DeleteDelivery   usecases.DeleteDelivery
}

func NewDeliveriesController(deps DeliveriesControllerDependencies) *DeliveriesController {
	return &DeliveriesController{
		authorizer:       deps.Authorizer,
		listDeliveries:   deps.ListDeliveries,
		registerDelivery: deps.RegisterDelivery,
		deleteDelivery:   deps.DeleteDelivery,
	}
}

func (c *DeliveriesController) RegisterRoutes(router gin.IRouter) {
	router.GET("/v1/deliveries", c.authorizer.Require(staff...), c.list)
	router.POST("/v1/deliveries", c.authorizer.Require(staff...), c.register)
	router.DELETE("/v1/deliveries/:delivery_id", c.authorizer.Require(staff...), c.delete)
}

func (c *DeliveriesController) list(ctx *gin.Context) {
	query, ok := bindQuery[request.ListDeliveries](ctx)
	if !ok {
		return
	}

	apartment, status := query.FromDTO()
	deliveries, err := c.listDeliveries.Execute(ctx.Request.Context(), apartment, status)
	if err != nil {
		writeError(ctx, deliveryResource, err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewDeliveries(deliveries))
}

func (c *DeliveriesController) register(ctx *gin.Context) {
	body, ok := bindJSON[request.RegisterDelivery](ctx)
	if !ok {
		return
	}

	delivery, err := c.registerDelivery.Execute(ctx.Request.Context(), body.FromDTO())
	if err != nil {
		writeError(ctx, deliveryResource, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.NewDelivery(delivery))
}

func (c *DeliveriesController) delete(ctx *gin.Context) {
	if err := c.deleteDelivery.Execute(ctx.Request.Context(), ctx.Param("delivery_id")); err != nil {
		writeError(ctx, deliveryResource, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
