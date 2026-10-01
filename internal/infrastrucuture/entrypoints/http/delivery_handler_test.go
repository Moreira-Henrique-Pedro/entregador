package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/readers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListDeliveries(t *testing.T) {
	pending := &entities.Delivery{DeliveryID: "d1", Apartment: "101", Status: entities.DeliveryStatusPending}
	deleted := &entities.Delivery{DeliveryID: "d2", Apartment: "101", Status: entities.DeliveryStatusDeleted, DeleteAt: time.Now()}

	tests := []struct {
		name       string
		query      string
		expect     bool
		wantFilter *entities.DeliveryStatus
		readerOut  []*entities.Delivery
		readerErr  error
		wantStatus int
		wantCount  int
	}{
		{name: "all statuses", query: "?apartment=101", expect: true, readerOut: []*entities.Delivery{pending, deleted}, wantStatus: http.StatusOK, wantCount: 2},
		{name: "filter by status", query: "?apartment=101&status=pending", expect: true, wantFilter: ptr(entities.DeliveryStatusPending), readerOut: []*entities.Delivery{pending}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "invalid status", query: "?apartment=101&status=lost", wantStatus: http.StatusBadRequest},
		{name: "missing apartment", query: "", wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "?apartment=101", expect: true, readerErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := mocks.NewDeliveriesReaderPort(t)
			if tt.expect {

				reader.EXPECT().Handle(mock.Anything, "101", tt.wantFilter).Return(tt.readerOut, tt.readerErr).Once()
			}

			rec := httptest.NewRecorder()
			NewDeliveryHandler(reader).ListDeliveries(rec, httptest.NewRequest(http.MethodGet, "/v1/deliveries"+tt.query, nil))

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

func ptr[T any](v T) *T { return &v }
