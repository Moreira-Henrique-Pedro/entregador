package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationMessages(t *testing.T) {
	residents := NewResidentsController(ResidentsControllerDependencies{Authorizer: testAuthorizer{}})
	deliveries := NewDeliveriesController(DeliveriesControllerDependencies{Authorizer: testAuthorizer{}})

	tests := []struct {
		name       string
		controller interface{ RegisterRoutes(r gin.IRouter) }
		method     string
		path       string
		body       string
		want       string
	}{
		{name: "create resident without fields", controller: residents, method: http.MethodPost, path: "/v1/residents", body: `{}`, want: "name is required; apartment is required; phone is required"},
		{name: "update resident without fields", controller: residents, method: http.MethodPatch, path: "/v1/residents/r1", body: `{}`, want: "name, apartment or phone is required"},
		{name: "list residents without filter", controller: residents, method: http.MethodGet, path: "/v1/residents", want: "apartment or phone is required"},
		{name: "list residents with both filters", controller: residents, method: http.MethodGet, path: "/v1/residents?apartment=101&phone=1", want: "use either apartment or phone, not both"},
		{name: "register delivery without apartment", controller: deliveries, method: http.MethodPost, path: "/v1/deliveries", body: `{}`, want: "apartment is required"},
		{name: "list deliveries with invalid status", controller: deliveries, method: http.MethodGet, path: "/v1/deliveries?apartment=101&status=lost", want: "status must be one of: pending, deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			serveTo(rec, tt.controller, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))

			require.Equal(t, http.StatusBadRequest, rec.Code)
			var body response.Error
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.want, body.Error)
		})
	}
}

func TestNotificationsController_InvalidCommandIsRejected(t *testing.T) {
	rec := httptest.NewRecorder()
	body := pushBody(t, "NotifyDelivery", map[string]string{"notification_type": "delivery_arrived"})
	serveTo(rec, NewNotificationsController(mocks.NewNotifyDelivery(t), nil), httptest.NewRequest(http.MethodPost, PubSubPushPath, strings.NewReader(body)))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
