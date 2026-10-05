package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type errorSpyLogger struct {
	logger.Logger
	message string
	fields  []any
}

func (s *errorSpyLogger) Error(message string, fields ...any) {
	s.message, s.fields = message, fields
}

func TestRecovery(t *testing.T) {
	log := &errorSpyLogger{Logger: logger.NewNoopLogger()}

	engine := gin.New()
	engine.Use(Recovery())
	engine.GET("/panic", func(*gin.Context) { panic("boom") })
	engine.GET("/ok", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req = req.WithContext(context.WithValue(req.Context(), logger.LoggerContextKey, logger.Logger(log)))
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "Recovered from panic", log.message)
	assert.Equal(t, []any{"panic", "boom"}, log.fields[:2])

	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecovery_RepanicsOnAbortHandler(t *testing.T) {
	engine := gin.New()
	engine.Use(Recovery())
	engine.GET("/abort", func(*gin.Context) { panic(http.ErrAbortHandler) })

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
	})
}
