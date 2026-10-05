package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocsController(t *testing.T) {
	controller := NewDocsController([]byte("openapi: 3.0.3"))

	page := httptest.NewRecorder()
	serveTo(page, controller, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	assert.Equal(t, http.StatusOK, page.Code)
	assert.Contains(t, page.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, page.Body.String(), `url: "/swagger/openapi.yaml"`)

	spec := httptest.NewRecorder()
	serveTo(spec, controller, httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil))
	assert.Equal(t, http.StatusOK, spec.Code)
	assert.Equal(t, "application/yaml", spec.Header().Get("Content-Type"))
	assert.Equal(t, "openapi: 3.0.3", spec.Body.String())
}
