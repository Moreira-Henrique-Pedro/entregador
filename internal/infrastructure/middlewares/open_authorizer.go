package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type OpenAuthorizer struct{}

func NewOpenAuthorizer() *OpenAuthorizer {
	return &OpenAuthorizer{}
}

func (a *OpenAuthorizer) Require(...entities.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}
