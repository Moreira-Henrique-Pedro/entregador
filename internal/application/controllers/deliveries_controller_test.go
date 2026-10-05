package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers/dtos/response"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListDeliveries(t *testing.T) {
	pending := &entities.Delivery{DeliveryID: "d1", Apartment: "101", ResidentName: "Ana", Status: entities.DeliveryStatusPending}
	deleted := &entities.Delivery{DeliveryID: "d2", Apartment: "101", Status: entities.DeliveryStatusDeleted, DeleteAt: time.Now()}

	tests := []struct {
		name       string
		query      string
		expect     bool
		wantFilter entities.DeliveryFilter
		readerOut  []*entities.Delivery
		readerErr  error
		wantStatus int
		wantCount  int
	}{
		{name: "all deliveries", query: "", expect: true, readerOut: []*entities.Delivery{pending, deleted}, wantStatus: http.StatusOK, wantCount: 2},
		{name: "by apartment", query: "?apartment=101", expect: true, wantFilter: entities.DeliveryFilter{Apartment: "101"}, readerOut: []*entities.Delivery{pending, deleted}, wantStatus: http.StatusOK, wantCount: 2},
		{name: "by apartment and status", query: "?apartment=101&status=pending", expect: true, wantFilter: entities.DeliveryFilter{Apartment: "101", Status: ptr(entities.DeliveryStatusPending)}, readerOut: []*entities.Delivery{pending}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "invalid status", query: "?status=lost", wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "", expect: true, readerErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := mocks.NewListDeliveries(t)
			if tt.expect {
				reader.EXPECT().Execute(mock.Anything, tt.wantFilter).Return(tt.readerOut, tt.readerErr).Once()
			}

			rec := httptest.NewRecorder()
			serveTo(rec, NewDeliveriesController(DeliveriesControllerDependencies{Authorizer: testAuthorizer{}, ListDeliveries: reader}), httptest.NewRequest(http.MethodGet, "/v1/deliveries"+tt.query, nil))

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				var body response.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.NotEmpty(t, body.Error)
				return
			}

			var body []map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Len(t, body, tt.wantCount)
			for i, item := range body {
				assert.Equal(t, tt.readerOut[i].DeliveryID, item["delivery_id"])
				assert.Equal(t, string(tt.readerOut[i].Status), item["status"])
				if tt.readerOut[i].ResidentName != "" {
					assert.Equal(t, tt.readerOut[i].ResidentName, item["resident_name"])
				}
				_, hasDeletedAt := item["deleted_at"]
				assert.Equal(t, item["status"] == "deleted", hasDeletedAt, "deleted_at presence mismatch for %v", item)
			}
		})
	}
}

func TestRegisterDelivery(t *testing.T) {
	created := &entities.Delivery{DeliveryID: "d1", Apartment: "101", ResidentID: "ana", PackageType: "caixa", Status: entities.DeliveryStatusPending}

	tests := []struct {
		name       string
		body       string
		wantInput  *entities.Delivery
		useCaseErr error
		wantStatus int
	}{
		{
			name:       "registered",
			body:       `{"apartment":"101","resident_id":"ana","package_type":"caixa","urgency":"alta"}`,
			wantInput:  &entities.Delivery{Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"},
			wantStatus: http.StatusCreated,
		},
		{name: "invalid JSON", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"apartment":"101","color":"blue"}`, wantStatus: http.StatusBadRequest},
		{
			name:       "missing apartment is rejected by the DTO",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid delivery",
			body:       `{"apartment":"101"}`,
			wantInput:  &entities.Delivery{Apartment: "101"},
			useCaseErr: fmt.Errorf("%w: apartment is required", entities.ErrInvalidDelivery),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "apartment without residents",
			body:       `{"apartment":"303"}`,
			wantInput:  &entities.Delivery{Apartment: "303"},
			useCaseErr: fmt.Errorf("apartment 303: %w", entities.ErrNoResidentInApartment),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unexpected error",
			body:       `{"apartment":"101"}`,
			wantInput:  &entities.Delivery{Apartment: "101"},
			useCaseErr: errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			register := mocks.NewRegisterDelivery(t)
			if tt.wantInput != nil {
				var out *entities.Delivery
				if tt.useCaseErr == nil {
					out = created
				}
				register.EXPECT().Execute(mock.Anything, tt.wantInput).Return(out, tt.useCaseErr).Once()
			}

			rec := httptest.NewRecorder()
			serveTo(rec, NewDeliveriesController(DeliveriesControllerDependencies{Authorizer: testAuthorizer{}, RegisterDelivery: register}), httptest.NewRequest(http.MethodPost, "/v1/deliveries", strings.NewReader(tt.body)))

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusCreated {
				var body response.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.NotEmpty(t, body.Error)
				return
			}

			var body response.Delivery
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "d1", body.DeliveryID)
			assert.Equal(t, "ana", body.ResidentID)
			assert.Equal(t, "pending", body.Status)
			assert.Nil(t, body.DeletedAt)
		})
	}
}

func TestDeleteDelivery(t *testing.T) {
	tests := []struct {
		name       string
		useCaseErr error
		wantStatus int
	}{
		{name: "deleted", wantStatus: http.StatusNoContent},
		{name: "not found", useCaseErr: fmt.Errorf("find: %w", entities.ErrEntityNotFound), wantStatus: http.StatusNotFound},
		{name: "unexpected error", useCaseErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			del := mocks.NewDeleteDelivery(t)
			del.EXPECT().Execute(mock.Anything, "d1").Return(tt.useCaseErr).Once()

			req := httptest.NewRequest(http.MethodDelete, "/v1/deliveries/d1", nil)
			rec := httptest.NewRecorder()
			serveTo(rec, NewDeliveriesController(DeliveriesControllerDependencies{Authorizer: testAuthorizer{}, DeleteDelivery: del}), req)

			assert.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
		})
	}
}
