package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListDeliveries(t *testing.T) {
	pending := &domain.Delivery{DeliveryID: "d1", Apartment: "101", Status: domain.DeliveryStatusPending}
	deleted := &domain.Delivery{DeliveryID: "d2", Apartment: "101", Status: domain.DeliveryStatusDeleted, DeleteAt: time.Now()}

	tests := []struct {
		name       string
		query      string
		expect     bool
		wantFilter *domain.DeliveryStatus
		readerOut  []*domain.Delivery
		readerErr  error
		wantStatus int
		wantCount  int
	}{
		{name: "all statuses", query: "?apartment=101", expect: true, readerOut: []*domain.Delivery{pending, deleted}, wantStatus: http.StatusOK, wantCount: 2},
		{name: "filter by status", query: "?apartment=101&status=pending", expect: true, wantFilter: ptr(domain.DeliveryStatusPending), readerOut: []*domain.Delivery{pending}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "invalid status", query: "?apartment=101&status=lost", wantStatus: http.StatusBadRequest},
		{name: "missing apartment", query: "", wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "?apartment=101", expect: true, readerErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := mocks.NewListDeliveries(t)
			if tt.expect {

				reader.EXPECT().Execute(mock.Anything, "101", tt.wantFilter).Return(tt.readerOut, tt.readerErr).Once()
			}

			rec := httptest.NewRecorder()
			NewDeliveryHandler(DeliveryHandlerDependencies{ListDeliveries: reader}).ListDeliveries(rec, httptest.NewRequest(http.MethodGet, "/v1/deliveries"+tt.query, nil))

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				var body errorResponse
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
				_, hasDeletedAt := item["deleted_at"]
				assert.Equal(t, item["status"] == "deleted", hasDeletedAt, "deleted_at presence mismatch for %v", item)
			}
		})
	}
}

func TestRegisterDelivery(t *testing.T) {
	created := &domain.Delivery{DeliveryID: "d1", Apartment: "101", ResidentID: "ana", PackageType: "caixa", Status: domain.DeliveryStatusPending}

	tests := []struct {
		name       string
		body       string
		wantInput  *in.RegisterDeliveryInput
		useCaseErr error
		wantStatus int
	}{
		{
			name:       "registered",
			body:       `{"apartment":"101","resident_id":"ana","package_type":"caixa","urgency":"alta"}`,
			wantInput:  &in.RegisterDeliveryInput{Apartment: "101", ResidentID: "ana", PackageType: "caixa", Urgency: "alta"},
			wantStatus: http.StatusCreated,
		},
		{name: "invalid JSON", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"apartment":"101","color":"blue"}`, wantStatus: http.StatusBadRequest},
		{
			name:       "invalid delivery",
			body:       `{}`,
			wantInput:  &in.RegisterDeliveryInput{},
			useCaseErr: fmt.Errorf("%w: apartment is required", domain.ErrInvalidDelivery),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "apartment without residents",
			body:       `{"apartment":"303"}`,
			wantInput:  &in.RegisterDeliveryInput{Apartment: "303"},
			useCaseErr: fmt.Errorf("apartment 303: %w", domain.ErrNoResidentInApartment),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unexpected error",
			body:       `{"apartment":"101"}`,
			wantInput:  &in.RegisterDeliveryInput{Apartment: "101"},
			useCaseErr: errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			register := mocks.NewRegisterDelivery(t)
			if tt.wantInput != nil {
				var out *domain.Delivery
				if tt.useCaseErr == nil {
					out = created
				}
				register.EXPECT().Execute(mock.Anything, *tt.wantInput).Return(out, tt.useCaseErr).Once()
			}

			rec := httptest.NewRecorder()
			NewDeliveryHandler(DeliveryHandlerDependencies{RegisterDelivery: register}).
				RegisterDelivery(rec, httptest.NewRequest(http.MethodPost, "/v1/deliveries", strings.NewReader(tt.body)))

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusCreated {
				var body errorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.NotEmpty(t, body.Error)
				return
			}

			var body deliveryResponse
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
		{name: "not found", useCaseErr: fmt.Errorf("find: %w", domain.ErrEntityNotFound), wantStatus: http.StatusNotFound},
		{name: "unexpected error", useCaseErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			del := mocks.NewDeleteDelivery(t)
			del.EXPECT().Execute(mock.Anything, "d1").Return(tt.useCaseErr).Once()

			req := httptest.NewRequest(http.MethodDelete, "/v1/deliveries/d1", nil)
			req.SetPathValue("delivery_id", "d1")
			rec := httptest.NewRecorder()
			NewDeliveryHandler(DeliveryHandlerDependencies{DeleteDelivery: del}).DeleteDelivery(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
		})
	}
}

func ptr[T any](v T) *T { return &v }
