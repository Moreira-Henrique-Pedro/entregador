package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/middlewares"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type Controller interface {
	RegisterRoutes(router gin.IRouter)
}

func SetupGin(debug bool, log logger.Logger) {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
		return
	}
	gin.SetMode(gin.DebugMode)
	gin.DebugPrintFunc = func(format string, values ...any) {
		log.Debug(strings.TrimSpace(fmt.Sprintf(format, values...)), "component", "gin")
	}
	gin.DebugPrintRouteFunc = func(method, path, handler string, _ int) {
		log.Debug("Route registered", "component", "gin", "method", method, "path", path, "handler", handler)
	}
}

func New(log logger.Logger, controllers ...Controller) http.Handler {
	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	engine.Use(middlewares.RequestLogger(log), middlewares.Recovery())

	for _, controller := range controllers {
		controller.RegisterRoutes(engine)
	}
	return engine
}
