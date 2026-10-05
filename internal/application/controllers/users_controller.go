package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/request"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases"
)

const userResource = "user"

type UsersController struct {
	authorizer Authorizer
	createUser usecases.CreateUser
}

type UsersControllerDependencies struct {
	Authorizer Authorizer
	CreateUser usecases.CreateUser
}

func NewUsersController(deps UsersControllerDependencies) *UsersController {
	return &UsersController{
		authorizer: deps.Authorizer,
		createUser: deps.CreateUser,
	}
}

func (c *UsersController) RegisterRoutes(router gin.IRouter) {
	router.POST("/v1/users", c.authorizer.Require(adminOnly...), c.create)
}

func (c *UsersController) create(ctx *gin.Context) {
	body, ok := bindJSON[request.CreateUser](ctx)
	if !ok {
		return
	}

	user, password := body.FromDTO()
	created, err := c.createUser.Execute(ctx.Request.Context(), user, password)
	if err != nil {
		writeError(ctx, userResource, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.NewUser(created))
}
