package readers

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errRepo = errors.New("mongo down")

func TestGetResidentsByApartment(t *testing.T) {
	residents := []*entities.Resident{{ResidentID: "r1", Apartment: "101"}, entities.NewOtherResident("101")}

	tests := []struct {
		name      string
		apartment string
		expect    bool
		repoOut   []*entities.Resident
		repoErr   error
		want      []*entities.Resident
		wantErr   error
	}{
		{name: "empty apartment", apartment: "", wantErr: errAny},
		{name: "repository error", apartment: "101", expect: true, repoErr: errRepo, wantErr: errRepo},
		{name: "returns residents", apartment: "101", expect: true, repoOut: residents, want: residents},
		{name: "no residents", apartment: "101", expect: true, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewResidentRepositoryPort(t)
			if tt.expect {
				repo.EXPECT().FindByApartment(mock.Anything, tt.apartment).Return(tt.repoOut, tt.repoErr).Once()
			}

			got, err := NewGetResidentsByApartment(repo).Handle(context.Background(), tt.apartment)

			checkErr(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetResidentsByPhone(t *testing.T) {
	residents := []*entities.Resident{{ResidentID: "r1", Phone: "11999999999"}, {ResidentID: "r2", Phone: "11999999999"}}

	tests := []struct {
		name    string
		phone   string
		expect  bool
		repoOut []*entities.Resident
		repoErr error
		want    []*entities.Resident
		wantErr error
	}{
		{name: "empty phone", phone: "", wantErr: errAny},
		{name: "repository error", phone: "11999999999", expect: true, repoErr: errRepo, wantErr: errRepo},
		{name: "returns residents", phone: "11999999999", expect: true, repoOut: residents, want: residents},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewResidentRepositoryPort(t)
			if tt.expect {
				repo.EXPECT().FindByPhone(mock.Anything, tt.phone).Return(tt.repoOut, tt.repoErr).Once()
			}

			got, err := NewGetResidentsByPhone(repo).Handle(context.Background(), tt.phone)

			checkErr(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDeliveriesByApartment(t *testing.T) {
	deliveries := []*entities.Delivery{
		{DeliveryID: "d2", Apartment: "101", Status: entities.DeliveryStatusDeleted},
		{DeliveryID: "d1", Apartment: "101", Status: entities.DeliveryStatusPending},
	}

	tests := []struct {
		name      string
		apartment string
		status    *entities.DeliveryStatus
		expect    bool
		repoOut   []*entities.Delivery
		repoErr   error
		want      []*entities.Delivery
		wantErr   error
	}{
		{name: "empty apartment", apartment: "", wantErr: errAny},
		{name: "invalid status", apartment: "101", status: ptr(entities.DeliveryStatus("lost")), wantErr: errAny},
		{name: "empty status is invalid", apartment: "101", status: ptr(entities.DeliveryStatus("")), wantErr: errAny},
		{name: "repository error", apartment: "101", expect: true, repoErr: errRepo, wantErr: errRepo},
		{name: "nil status means any", apartment: "101", expect: true, repoOut: deliveries, want: deliveries},
		{name: "pending status", apartment: "101", status: ptr(entities.DeliveryStatusPending), expect: true, repoOut: deliveries[1:], want: deliveries[1:]},
		{name: "deleted status", apartment: "101", status: ptr(entities.DeliveryStatusDeleted), expect: true, repoOut: deliveries[:1], want: deliveries[:1]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDeliveryRepositoryPort(t)
			if tt.expect {

				sameStatus := mock.MatchedBy(func(s *entities.DeliveryStatus) bool { return s == tt.status })
				repo.EXPECT().FindByApartment(mock.Anything, tt.apartment, sameStatus).Return(tt.repoOut, tt.repoErr).Once()
			}

			got, err := NewGetDeliveriesByApartment(repo).Handle(context.Background(), tt.apartment, tt.status)

			checkErr(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

var errAny = errors.New("any error")

func checkErr(t *testing.T, err, want error) {
	t.Helper()
	switch want {
	case nil:
		require.NoError(t, err)
	case errAny:
		require.Error(t, err)
	default:
		require.ErrorIs(t, err, want)
	}
}

func ptr[T any](v T) *T { return &v }
