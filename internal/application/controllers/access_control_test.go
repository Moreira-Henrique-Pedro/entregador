package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases/mocks"
)

func TestRouteAccess(t *testing.T) {
	deleteResident := mocks.NewDeleteResident(t)
	deleteResident.EXPECT().Execute(mock.Anything, mock.Anything).Return(nil).Maybe()
	residents := NewResidentsController(ResidentsControllerDependencies{Authorizer: testAuthorizer{}, DeleteResident: deleteResident})
	deliveries := NewDeliveriesController(DeliveriesControllerDependencies{Authorizer: testAuthorizer{}})
	users := NewUsersController(UsersControllerDependencies{Authorizer: testAuthorizer{}})

	routes := []struct {
		name       string
		controller interface{ RegisterRoutes(r gin.IRouter) }
		method     string
		path       string
		adminOnly  bool
	}{
		{name: "list residents", controller: residents, method: http.MethodGet, path: "/v1/residents"},
		{name: "create resident", controller: residents, method: http.MethodPost, path: "/v1/residents", adminOnly: true},
		{name: "update resident", controller: residents, method: http.MethodPatch, path: "/v1/residents/r1", adminOnly: true},
		{name: "delete resident", controller: residents, method: http.MethodDelete, path: "/v1/residents/r1", adminOnly: true},
		{name: "list deliveries", controller: deliveries, method: http.MethodGet, path: "/v1/deliveries"},
		{name: "register delivery", controller: deliveries, method: http.MethodPost, path: "/v1/deliveries"},
		{name: "create user", controller: users, method: http.MethodPost, path: "/v1/users", adminOnly: true},
	}

	for _, route := range routes {
		for _, role := range []string{"anonymous", "admin", "doorman", "unknown"} {
			t.Run(route.name+" as "+role, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, strings.NewReader(`{}`))
				req.Header.Set(testRoleHeader, role)
				rec := httptest.NewRecorder()
				serveTo(rec, route.controller, req)

				switch {
				case role == "anonymous":
					assert.Equal(t, http.StatusUnauthorized, rec.Code)
				case role == "unknown", role == "doorman" && route.adminOnly:
					assert.Equal(t, http.StatusForbidden, rec.Code)
				default:
					assert.NotContains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, rec.Code)
				}
			})
		}
	}
}
