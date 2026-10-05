package controllers

import (
	"net/http"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/request"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases"
	"github.com/gin-gonic/gin"
)

const residentResource = "resident"

type ResidentsController struct {
	authorizer               Authorizer
	listResidentsByApartment usecases.ListResidents
	listResidentsByPhone     usecases.ListResidents
	createResident           usecases.CreateResident
	updateResident           usecases.UpdateResident
	deleteResident           usecases.DeleteResident
}

type ResidentsControllerDependencies struct {
	Authorizer               Authorizer
	ListResidentsByApartment usecases.ListResidents
	ListResidentsByPhone     usecases.ListResidents
	CreateResident           usecases.CreateResident
	UpdateResident           usecases.UpdateResident
	DeleteResident           usecases.DeleteResident
}

func NewResidentsController(deps ResidentsControllerDependencies) *ResidentsController {
	return &ResidentsController{
		authorizer:               deps.Authorizer,
		listResidentsByApartment: deps.ListResidentsByApartment,
		listResidentsByPhone:     deps.ListResidentsByPhone,
		createResident:           deps.CreateResident,
		updateResident:           deps.UpdateResident,
		deleteResident:           deps.DeleteResident,
	}
}

func (c *ResidentsController) RegisterRoutes(router gin.IRouter) {
	router.GET("/v1/residents", c.authorizer.Require(staff...), c.list)
	router.POST("/v1/residents", c.authorizer.Require(adminOnly...), c.create)
	router.PATCH("/v1/residents/:resident_id", c.authorizer.Require(adminOnly...), c.update)
	router.DELETE("/v1/residents/:resident_id", c.authorizer.Require(adminOnly...), c.delete)
}

func (c *ResidentsController) list(ctx *gin.Context) {
	query, ok := bindQuery[request.ListResidents](ctx)
	if !ok {
		return
	}

	listResidents, value := c.listResidentsByPhone, query.Phone
	if query.ByApartment() {
		listResidents, value = c.listResidentsByApartment, query.Apartment
	}

	residents, err := listResidents.Execute(ctx.Request.Context(), value)
	if err != nil {
		writeError(ctx, residentResource, err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewResidents(residents))
}

func (c *ResidentsController) create(ctx *gin.Context) {
	body, ok := bindJSON[request.CreateResident](ctx)
	if !ok {
		return
	}

	resident, err := c.createResident.Execute(ctx.Request.Context(), body.FromDTO())
	if err != nil {
		writeError(ctx, residentResource, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.NewResident(resident))
}

func (c *ResidentsController) update(ctx *gin.Context) {
	body, ok := bindJSON[request.UpdateResident](ctx)
	if !ok {
		return
	}

	resident, err := c.updateResident.Execute(ctx.Request.Context(), body.FromDTO(ctx.Param("resident_id")))
	if err != nil {
		writeError(ctx, residentResource, err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewResident(resident))
}

func (c *ResidentsController) delete(ctx *gin.Context) {
	if err := c.deleteResident.Execute(ctx.Request.Context(), ctx.Param("resident_id")); err != nil {
		writeError(ctx, residentResource, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
