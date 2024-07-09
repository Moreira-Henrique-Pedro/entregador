package controller

import (
	"net/http"
	"strconv"

	"github.com/Moreira-Henrique-Pedro/entregador/src/model"
	"github.com/Moreira-Henrique-Pedro/entregador/src/service"
	"github.com/gin-gonic/gin"
)

type BoxController struct {
	service *service.BoxService
}

func NewBoxController(service *service.BoxService) *BoxController {
	return &BoxController{
		service: service,
	}
}

func (c *BoxController) InitRoutes() {
	app := gin.Default()
	api := app.Group("/api/entregador")

	api.GET("/:id", c.findBoxByID)
	api.POST("/", c.createBox)
	api.PUT("/", c.updateBox)
	api.DELETE("/", c.deleteBoxByID)

	app.Run(":3000")
}

func (c *BoxController) createBox(ctx *gin.Context) {
	box := new(model.Box)
	if err := ctx.ShouldBindJSON(&box); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err},
		)
		return
	}

	id, err := c.service.CreateBox(*box)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err},
		)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func (c *BoxController) findBoxByID(ctx *gin.Context) {
	idString := ctx.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err},
		)
		return
	}

	box, err := c.service.FindBoxByID(id)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err},
		)
		return
	}

	ctx.JSON(http.StatusOK, box)
}

func (c *BoxController) updateBox(ctx *gin.Context) {
	box := new(model.Box)
	if err := ctx.ShouldBindJSON(&box); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err},
		)
		return
	}

	idString := ctx.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err},
		)
		return
	}

	newOne, err := c.service.UpdateBox(*box, id)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err},
		)
		return
	}

	ctx.JSON(http.StatusNoContent, newOne)
}

func (c *BoxController) deleteBoxByID(ctx *gin.Context) {
	idString := ctx.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": err},
		)
		return
	}

	err = c.service.DeleteBoxByID(id)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err},
		)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
