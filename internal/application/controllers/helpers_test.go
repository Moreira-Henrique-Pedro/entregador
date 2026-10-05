package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func serveTo(rec *httptest.ResponseRecorder, controller interface{ RegisterRoutes(gin.IRouter) }, req *http.Request) {
	engine := gin.New()
	controller.RegisterRoutes(engine)
	engine.ServeHTTP(rec, req)
}

func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	var body response.Error
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body.Error)
}

func ptr[T any](v T) *T { return &v }

const testRoleHeader = "X-Test-Role"

type testAuthorizer struct{}

func (testAuthorizer) Require(roles ...entities.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetHeader(testRoleHeader)
		switch {
		case role == "":
			role = string(entities.RoleAdmin)
		case role == "anonymous":
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Error{Error: "unauthenticated"})
			return
		}
		user := &entities.User{ID: "test-user", Role: entities.Role(role)}
		if !user.HasAnyRole(roles...) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.Error{Error: "forbidden"})
			return
		}
		ctx.Next()
	}
}
