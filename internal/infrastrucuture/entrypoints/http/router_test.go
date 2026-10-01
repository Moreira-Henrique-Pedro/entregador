package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/readers/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		setup      func(byApartment *mocks.ResidentsReaderPort, deliveries *mocks.DeliveriesReaderPort)
		wantStatus int
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{
			name: "residents", method: http.MethodGet, path: "/v1/residents?apartment=101",
			setup: func(byApartment *mocks.ResidentsReaderPort, _ *mocks.DeliveriesReaderPort) {
				byApartment.EXPECT().Handle(mock.Anything, "101").Return([]*entities.Resident{}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "deliveries", method: http.MethodGet, path: "/v1/deliveries?apartment=101",
			setup: func(_ *mocks.ResidentsReaderPort, deliveries *mocks.DeliveriesReaderPort) {
				deliveries.EXPECT().Handle(mock.Anything, "101", (*entities.DeliveryStatus)(nil)).Return(nil, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{name: "method not allowed", method: http.MethodPost, path: "/v1/residents", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown path", method: http.MethodGet, path: "/v1/unknown", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byApartment := mocks.NewResidentsReaderPort(t)
			byPhone := mocks.NewResidentsReaderPort(t)
			deliveries := mocks.NewDeliveriesReaderPort(t)
			if tt.setup != nil {
				tt.setup(byApartment, deliveries)
			}
			router := NewRouter(NewResidentHandler(byApartment, byPhone), NewDeliveryHandler(deliveries), logger.NewNoopLogger())

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
