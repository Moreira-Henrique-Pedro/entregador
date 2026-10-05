package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases"
)

const apartmentResource = "apartment"

type ApartmentsController struct {
	authorizer     Authorizer
	listApartments usecases.ListApartments
}

type ApartmentsControllerDependencies struct {
	Authorizer     Authorizer
	ListApartments usecases.ListApartments
}

func NewApartmentsController(deps ApartmentsControllerDependencies) *ApartmentsController {
	return &ApartmentsController{
		authorizer:     deps.Authorizer,
		listApartments: deps.ListApartments,
	}
}

func (c *ApartmentsController) RegisterRoutes(router gin.IRouter) {
	router.GET("/v1/apartments", c.authorizer.Require(staff...), c.list)
}

func (c *ApartmentsController) list(ctx *gin.Context) {
	apartments, err := c.listApartments.Execute(ctx.Request.Context())
	if err != nil {
		writeError(ctx, apartmentResource, err)
		return
	}

	ctx.JSON(http.StatusOK, apartments)
}
