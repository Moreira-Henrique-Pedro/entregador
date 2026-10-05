package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type spyLogger struct {
	logger.Logger
	fields []any
}

func (s *spyLogger) With(fields ...any) logger.Logger {
	s.fields = fields
	return s
}

func (s *spyLogger) AddToContext(ctx context.Context, l logger.Logger) context.Context {
	return context.WithValue(ctx, logger.LoggerContextKey, l)
}

func TestRequestLogger(t *testing.T) {
	log := &spyLogger{Logger: logger.NewNoopLogger()}

	var got logger.Logger
	engine := gin.New()
	engine.Use(RequestLogger(log))
	engine.GET("/health", func(ctx *gin.Context) {
		got = logger.GetLoggerFromContext(ctx.Request.Context())
	})

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	assert.Same(t, log, got)
	assert.Equal(t, []any{"method", "GET", "path", "/health"}, log.fields)
}
