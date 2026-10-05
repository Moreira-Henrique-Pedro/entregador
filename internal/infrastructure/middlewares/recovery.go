package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(recovered)
			}

			logger.GetLoggerFromContext(ctx.Request.Context()).Error("Recovered from panic",
				"panic", fmt.Sprint(recovered),
				"stack", string(debug.Stack()),
			)
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}()
		ctx.Next()
	}
}
