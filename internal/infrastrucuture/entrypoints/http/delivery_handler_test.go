package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type fakeDeliveryReader struct {
	deliveries   []*entities.Delivery
	err          error
	gotApartment string
	gotStatus    *entities.DeliveryStatus
}

func (f *fakeDeliveryReader) Handle(_ context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	f.gotApartment = apartment
	f.gotStatus = status
	return f.deliveries, f.err
}

func TestListDeliveries(t *testing.T) {
	pending := &entities.Delivery{DeliveryID: "d1", Apartment: "101", Status: entities.DeliveryStatusPending}
	deleted := &entities.Delivery{DeliveryID: "d2", Apartment: "101", Status: entities.DeliveryStatusDeleted, DeleteAt: time.Now()}

	tests := []struct {
		name       string
		query      string
		reader     *fakeDeliveryReader
		wantStatus int
		wantFilter *entities.DeliveryStatus
		wantCount  int
	}{
		{name: "all statuses", query: "?apartment=101", reader: &fakeDeliveryReader{deliveries: []*entities.Delivery{pending, deleted}}, wantStatus: http.StatusOK, wantCount: 2},
		{name: "filter by status", query: "?apartment=101&status=pending", reader: &fakeDeliveryReader{deliveries: []*entities.Delivery{pending}}, wantStatus: http.StatusOK, wantFilter: ptr(entities.DeliveryStatusPending), wantCount: 1},
		{name: "invalid status", query: "?apartment=101&status=lost", reader: &fakeDeliveryReader{}, wantStatus: http.StatusBadRequest},
		{name: "missing apartment", query: "", reader: &fakeDeliveryReader{}, wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "?apartment=101", reader: &fakeDeliveryReader{err: errors.New("boom")}, wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewDeliveryHandler(tt.reader).ListDeliveries(rec, httptest.NewRequest(http.MethodGet, "/v1/deliveries"+tt.query, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			if (tt.reader.gotStatus == nil) != (tt.wantFilter == nil) || (tt.wantFilter != nil && *tt.reader.gotStatus != *tt.wantFilter) {
				t.Errorf("status filter = %v, want %v", tt.reader.gotStatus, tt.wantFilter)
			}

			var body []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if len(body) != tt.wantCount {
				t.Fatalf("got %d deliveries, want %d", len(body), tt.wantCount)
			}
			for _, item := range body {
				_, hasDeletedAt := item["deleted_at"]
				if (item["status"] == "deleted") != hasDeletedAt {
					t.Errorf("deleted_at presence mismatch for %v", item)
				}
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
