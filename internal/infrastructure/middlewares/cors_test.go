package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	engine := gin.New()
	engine.Use(CORS([]string{"http://localhost:5173"}))
	engine.GET("/v1/deliveries", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	tests := []struct {
		name        string
		method      string
		origin      string
		wantAllowed bool
	}{
		{name: "preflight from allowed origin", method: http.MethodOptions, origin: "http://localhost:5173", wantAllowed: true},
		{name: "request from allowed origin", method: http.MethodGet, origin: "http://localhost:5173", wantAllowed: true},
		{name: "other origin", method: http.MethodGet, origin: "http://evil.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/deliveries", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.method == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", http.MethodGet)
				req.Header.Set("Access-Control-Request-Headers", "Authorization")
			}
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if tt.wantAllowed {
				assert.Equal(t, tt.origin, rec.Header().Get("Access-Control-Allow-Origin"))
				return
			}
			assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
