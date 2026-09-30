package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type fakeReader struct {
	residents []*entities.Resident
	err       error
	gotValue  string
}

func (f *fakeReader) Handle(_ context.Context, value string) ([]*entities.Resident, error) {
	f.gotValue = value
	return f.residents, f.err
}

func TestListResidents(t *testing.T) {
	resident := &entities.Resident{ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999"}

	tests := []struct {
		name          string
		query         string
		byApartment   *fakeReader
		byPhone       *fakeReader
		wantStatus    int
		wantCount     int
		wantApartment string
		wantPhone     string
	}{
		{name: "by apartment", query: "?apartment=101", byApartment: &fakeReader{residents: []*entities.Resident{resident}}, byPhone: &fakeReader{}, wantStatus: http.StatusOK, wantCount: 1, wantApartment: "101"},
		{name: "by phone", query: "?phone=11999999999", byApartment: &fakeReader{}, byPhone: &fakeReader{residents: []*entities.Resident{resident}}, wantStatus: http.StatusOK, wantCount: 1, wantPhone: "11999999999"},
		{name: "empty result returns empty list", query: "?apartment=999", byApartment: &fakeReader{}, byPhone: &fakeReader{}, wantStatus: http.StatusOK, wantCount: 0, wantApartment: "999"},
		{name: "missing params", query: "", byApartment: &fakeReader{}, byPhone: &fakeReader{}, wantStatus: http.StatusBadRequest},
		{name: "both params", query: "?apartment=101&phone=1", byApartment: &fakeReader{}, byPhone: &fakeReader{}, wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "?phone=1", byApartment: &fakeReader{}, byPhone: &fakeReader{err: errors.New("boom")}, wantStatus: http.StatusInternalServerError, wantPhone: "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewResidentHandler(tt.byApartment, tt.byPhone)
			rec := httptest.NewRecorder()

			handler.ListResidents(rec, httptest.NewRequest(http.MethodGet, "/v1/residents"+tt.query, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.byApartment.gotValue != tt.wantApartment {
				t.Errorf("apartment reader got %q, want %q", tt.byApartment.gotValue, tt.wantApartment)
			}
			if tt.byPhone.gotValue != tt.wantPhone {
				t.Errorf("phone reader got %q, want %q", tt.byPhone.gotValue, tt.wantPhone)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var body []residentResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if len(body) != tt.wantCount {
				t.Errorf("got %d residents, want %d", len(body), tt.wantCount)
			}
		})
	}
}
