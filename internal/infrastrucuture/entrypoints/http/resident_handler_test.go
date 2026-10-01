package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/readers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListResidents(t *testing.T) {
	resident := &entities.Resident{
		ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
		Type: entities.ResidentTypePrimary, Status: entities.ResidentStatusCreated,
	}

	tests := []struct {
		name          string
		query         string
		wantApartment string
		wantPhone     string
		readerOut     []*entities.Resident
		readerErr     error
		wantStatus    int
		wantCount     int
	}{
		{name: "by apartment", query: "?apartment=101", wantApartment: "101", readerOut: []*entities.Resident{resident}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "by phone", query: "?phone=11999999999", wantPhone: "11999999999", readerOut: []*entities.Resident{resident}, wantStatus: http.StatusOK, wantCount: 1},
		{name: "empty result returns empty list", query: "?apartment=999", wantApartment: "999", wantStatus: http.StatusOK, wantCount: 0},
		{name: "missing params", query: "", wantStatus: http.StatusBadRequest},
		{name: "both params", query: "?apartment=101&phone=1", wantStatus: http.StatusBadRequest},
		{name: "reader error", query: "?phone=1", wantPhone: "1", readerErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byApartment := mocks.NewResidentsReaderPort(t)
			byPhone := mocks.NewResidentsReaderPort(t)
			if tt.wantApartment != "" {
				byApartment.EXPECT().Handle(mock.Anything, tt.wantApartment).Return(tt.readerOut, tt.readerErr).Once()
			}
			if tt.wantPhone != "" {
				byPhone.EXPECT().Handle(mock.Anything, tt.wantPhone).Return(tt.readerOut, tt.readerErr).Once()
			}

			rec := httptest.NewRecorder()
			NewResidentHandler(byApartment, byPhone).ListResidents(rec, httptest.NewRequest(http.MethodGet, "/v1/residents"+tt.query, nil))

			require.Equal(t, tt.wantStatus, rec.Code, "body: %s", rec.Body.String())
			if tt.wantStatus != http.StatusOK {
				var body errorResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.NotEmpty(t, body.Error)
				return
			}

			var body []residentResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Len(t, body, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, residentResponse{
					ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "11999999999",
					Type: "resident-primary", Status: "created",
				}, body[0])
			}
		})
	}
}
