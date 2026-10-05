package middlewares

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services/mocks"
)

func TestAuthorizerRequire(t *testing.T) {
	admin := &entities.User{ID: "u1", Role: entities.RoleAdmin}
	doorman := &entities.User{ID: "u2", Role: entities.RoleDoorman}

	tests := []struct {
		name       string
		header     string
		roles      []entities.Role
		verifyOut  *entities.User
		verifyErr  error
		wantStatus int
		wantError  string
	}{
		{name: "any authenticated user", header: "Bearer t", verifyOut: doorman, wantStatus: http.StatusNoContent},
		{name: "allowed role", header: "Bearer t", roles: []entities.Role{entities.RoleAdmin}, verifyOut: admin, wantStatus: http.StatusNoContent},
		{name: "one of the allowed roles", header: "Bearer t", roles: []entities.Role{entities.RoleAdmin, entities.RoleDoorman}, verifyOut: doorman, wantStatus: http.StatusNoContent},
		{name: "role not allowed", header: "Bearer t", roles: []entities.Role{entities.RoleAdmin}, verifyOut: doorman, wantStatus: http.StatusForbidden, wantError: "forbidden"},
		{name: "user without role", header: "Bearer t", roles: []entities.Role{entities.RoleDoorman}, verifyOut: &entities.User{ID: "u3"}, wantStatus: http.StatusForbidden, wantError: "forbidden"},
		{name: "missing token", wantStatus: http.StatusUnauthorized, wantError: "unauthenticated"},
		{name: "not a bearer token", header: "Basic abc", wantStatus: http.StatusUnauthorized, wantError: "unauthenticated"},
		{name: "invalid token", header: "Bearer bad", verifyErr: errors.New("expired"), wantStatus: http.StatusUnauthorized, wantError: "unauthenticated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := mocks.NewTokenVerifier(t)
			if tt.verifyOut != nil || tt.verifyErr != nil {
				verifier.EXPECT().VerifyToken(mock.Anything, tokenOf(tt.header)).Return(tt.verifyOut, tt.verifyErr).Once()
			}

			engine := gin.New()
			engine.GET("/protected", NewAuthorizer(verifier).Require(tt.roles...), func(ctx *gin.Context) {
				ctx.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantError != "" {
				var body response.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.wantError, body.Error)
			}
		})
	}
}

func tokenOf(header string) string {
	return header[len("Bearer "):]
}

func TestOpenAuthorizerAcceptsEveryRequest(t *testing.T) {
	engine := gin.New()
	engine.GET("/protected", NewOpenAuthorizer().Require(entities.RoleAdmin), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
