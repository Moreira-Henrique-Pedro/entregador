package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type routerDeps struct {
	byApartment      *mocks.ListResidents
	create           *mocks.CreateResident
	update           *mocks.UpdateResident
	delete           *mocks.DeleteResident
	deliveries       *mocks.ListDeliveries
	registerDelivery *mocks.RegisterDelivery
	deleteDelivery   *mocks.DeleteDelivery
}

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		setup      func(d routerDeps)
		wantStatus int
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{
			name: "residents", method: http.MethodGet, path: "/v1/residents?apartment=101",
			setup: func(d routerDeps) {
				d.byApartment.EXPECT().Execute(mock.Anything, "101").Return([]*domain.Resident{}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "create resident", method: http.MethodPost, path: "/v1/residents", body: `{"name":"Ana","apartment":"101"}`,
			setup: func(d routerDeps) {
				d.create.EXPECT().Execute(mock.Anything, mock.Anything).Return(&domain.Resident{}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "update resident", method: http.MethodPatch, path: "/v1/residents/r1", body: `{"name":"Ana"}`,
			setup: func(d routerDeps) {
				d.update.EXPECT().Execute(mock.Anything, &domain.Resident{ResidentID: "r1", Name: "Ana"}).Return(&domain.Resident{}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "delete resident", method: http.MethodDelete, path: "/v1/residents/r1",
			setup: func(d routerDeps) {
				d.delete.EXPECT().Execute(mock.Anything, "r1").Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "deliveries", method: http.MethodGet, path: "/v1/deliveries?apartment=101",
			setup: func(d routerDeps) {
				d.deliveries.EXPECT().Execute(mock.Anything, "101", (*domain.DeliveryStatus)(nil)).Return(nil, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "register delivery", method: http.MethodPost, path: "/v1/deliveries", body: `{"apartment":"101"}`,
			setup: func(d routerDeps) {
				d.registerDelivery.EXPECT().Execute(mock.Anything, in.RegisterDeliveryInput{Apartment: "101"}).Return(&domain.Delivery{}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "delete delivery", method: http.MethodDelete, path: "/v1/deliveries/d1",
			setup: func(d routerDeps) {
				d.deleteDelivery.EXPECT().Execute(mock.Anything, "d1").Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{name: "method not allowed", method: http.MethodPut, path: "/v1/residents", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown path", method: http.MethodGet, path: "/v1/unknown", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := routerDeps{
				byApartment:      mocks.NewListResidents(t),
				create:           mocks.NewCreateResident(t),
				update:           mocks.NewUpdateResident(t),
				delete:           mocks.NewDeleteResident(t),
				deliveries:       mocks.NewListDeliveries(t),
				registerDelivery: mocks.NewRegisterDelivery(t),
				deleteDelivery:   mocks.NewDeleteDelivery(t),
			}
			if tt.setup != nil {
				tt.setup(d)
			}
			residentHandler := NewResidentHandler(ResidentHandlerDependencies{
				ListResidentsByApartment: d.byApartment,
				ListResidentsByPhone:     mocks.NewListResidents(t),
				CreateResident:           d.create,
				UpdateResident:           d.update,
				DeleteResident:           d.delete,
			})
			deliveryHandler := NewDeliveryHandler(DeliveryHandlerDependencies{
				ListDeliveries:   d.deliveries,
				RegisterDelivery: d.registerDelivery,
				DeleteDelivery:   d.deleteDelivery,
			})
			router := NewRouter(residentHandler, deliveryHandler, logger.NewNoopLogger())

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
