package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

func RequestLogger(log logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestLogger := log.With("method", ctx.Request.Method, "path", ctx.Request.URL.Path)
		ctx.Request = ctx.Request.WithContext(requestLogger.AddToContext(ctx.Request.Context(), requestLogger))
		ctx.Next()
	}
}
