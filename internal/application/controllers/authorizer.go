package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type Authorizer interface {
	Require(roles ...entities.Role) gin.HandlerFunc
}

var (
	adminOnly = []entities.Role{entities.RoleAdmin}
	staff     = []entities.Role{entities.RoleAdmin, entities.RoleDoorman}
)
