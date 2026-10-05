package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	repomocks "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errRepo = errors.New("mongo down")

func TestListResidentsByApartment(t *testing.T) {
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
			repo := repomocks.NewResidentRepository(t)
			if tt.expect {
				repo.EXPECT().FindByApartment(mock.Anything, tt.apartment).Return(tt.repoOut, tt.repoErr).Once()
			}

			got, err := NewListResidentsByApartment(repo).Execute(context.Background(), tt.apartment)

			checkErr(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListResidentsByPhone(t *testing.T) {
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
			repo := repomocks.NewResidentRepository(t)
			if tt.expect {
				repo.EXPECT().FindByPhone(mock.Anything, tt.phone).Return(tt.repoOut, tt.repoErr).Once()
			}

			got, err := NewListResidentsByPhone(repo).Execute(context.Background(), tt.phone)

			checkErr(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListDeliveries(t *testing.T) {
	pending := entities.DeliveryStatusPending

	tests := []struct {
		name         string
		filter       entities.DeliveryFilter
		expectFind   bool
		repoErr      error
		expectNames  bool
		residentsErr error
		wantNames    []string
		wantErr      error
	}{
		{name: "invalid status", filter: entities.DeliveryFilter{Status: ptr(entities.DeliveryStatus("lost"))}, wantErr: entities.ErrInvalidDelivery},
		{name: "all deliveries with resident names", expectFind: true, expectNames: true, wantNames: []string{"Ana", "Outro"}},
		{name: "filtered by apartment and status", filter: entities.DeliveryFilter{Apartment: "101", Status: &pending}, expectFind: true, expectNames: true, wantNames: []string{"Ana", "Outro"}},
		{name: "delivery repository error", expectFind: true, repoErr: errRepo, wantErr: errRepo},
		{name: "resident repository error", expectFind: true, expectNames: true, residentsErr: errRepo, wantErr: errRepo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveries := []*entities.Delivery{
				{DeliveryID: "d1", Apartment: "101", ResidentID: "ana"},
				{DeliveryID: "d2", Apartment: "101", ResidentID: "other-101"},
			}
			deliveryRepo := repomocks.NewDeliveryRepository(t)
			residentRepo := repomocks.NewResidentRepository(t)
			if tt.expectFind {
				out := deliveries
				if tt.repoErr != nil {
					out = nil
				}
				deliveryRepo.EXPECT().Find(mock.Anything, tt.filter).Return(out, tt.repoErr).Once()
			}
			if tt.expectNames {
				residentRepo.EXPECT().FindByResidentIDs(mock.Anything, []string{"ana", "other-101"}).
					Return([]*entities.Resident{{ResidentID: "ana", Name: "Ana"}, entities.NewOtherResident("101")}, tt.residentsErr).Once()
			}

			got, err := NewListDeliveries(deliveryRepo, residentRepo).Execute(context.Background(), tt.filter)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, len(tt.wantNames))
			for i, name := range tt.wantNames {
				assert.Equal(t, name, got[i].ResidentName)
			}
		})
	}
}

func TestListApartments(t *testing.T) {
	repo := repomocks.NewResidentRepository(t)
	repo.EXPECT().ListApartments(mock.Anything).Return([]string{"63", "101"}, nil).Once()

	got, err := NewListApartments(repo).Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"63", "101"}, got)

	failing := repomocks.NewResidentRepository(t)
	failing.EXPECT().ListApartments(mock.Anything).Return(nil, errRepo).Once()

	_, err = NewListApartments(failing).Execute(context.Background())

	assert.ErrorIs(t, err, errRepo)
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
