package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type fakeController struct {
	path string
}

func (c fakeController) RegisterRoutes(router gin.IRouter) {
	router.GET(c.path, func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	router.GET(c.path+"/panic", func(*gin.Context) {
		panic("boom")
	})
}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "first controller", method: http.MethodGet, path: "/a", wantStatus: http.StatusOK},
		{name: "second controller", method: http.MethodGet, path: "/b", wantStatus: http.StatusOK},
		{name: "method not allowed", method: http.MethodPost, path: "/a", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown path", method: http.MethodGet, path: "/c", wantStatus: http.StatusNotFound},
		{name: "panic is recovered", method: http.MethodGet, path: "/a/panic", wantStatus: http.StatusInternalServerError},
	}

	handler := New(logger.NewNoopLogger(), nil, fakeController{path: "/a"}, fakeController{path: "/b"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
