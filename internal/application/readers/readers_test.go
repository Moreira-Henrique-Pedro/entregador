package readers

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

var errRepo = errors.New("mongo down")

func TestGetResidentsByApartment(t *testing.T) {
	residents := []*entities.Resident{{ResidentID: "r1", Apartment: "101"}, entities.NewOtherResident("101")}

	tests := []struct {
		name      string
		apartment string
		repo      *fakeResidentRepository
		want      []*entities.Resident
		wantErr   error
		wantCalls int
	}{
		{name: "empty apartment", apartment: "", repo: &fakeResidentRepository{}, wantErr: errAny, wantCalls: 0},
		{name: "repository error", apartment: "101", repo: &fakeResidentRepository{err: errRepo}, wantErr: errRepo, wantCalls: 1},
		{name: "returns residents", apartment: "101", repo: &fakeResidentRepository{residents: residents}, want: residents, wantCalls: 1},
		{name: "no residents", apartment: "101", repo: &fakeResidentRepository{}, want: nil, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGetResidentsByApartment(tt.repo).Handle(context.Background(), tt.apartment)

			checkErr(t, err, tt.wantErr)
			if tt.repo.calls != tt.wantCalls {
				t.Fatalf("repository calls = %d, want %d", tt.repo.calls, tt.wantCalls)
			}
			if tt.wantCalls > 0 && tt.repo.gotApartment != tt.apartment {
				t.Errorf("repository apartment = %q, want %q", tt.repo.gotApartment, tt.apartment)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("residents = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResidentsByPhone(t *testing.T) {
	residents := []*entities.Resident{{ResidentID: "r1", Phone: "11999999999"}, {ResidentID: "r2", Phone: "11999999999"}}

	tests := []struct {
		name      string
		phone     string
		repo      *fakeResidentRepository
		want      []*entities.Resident
		wantErr   error
		wantCalls int
	}{
		{name: "empty phone", phone: "", repo: &fakeResidentRepository{}, wantErr: errAny, wantCalls: 0},
		{name: "repository error", phone: "11999999999", repo: &fakeResidentRepository{err: errRepo}, wantErr: errRepo, wantCalls: 1},
		{name: "returns residents", phone: "11999999999", repo: &fakeResidentRepository{residents: residents}, want: residents, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGetResidentsByPhone(tt.repo).Handle(context.Background(), tt.phone)

			checkErr(t, err, tt.wantErr)
			if tt.repo.calls != tt.wantCalls {
				t.Fatalf("repository calls = %d, want %d", tt.repo.calls, tt.wantCalls)
			}
			if tt.wantCalls > 0 && tt.repo.gotPhone != tt.phone {
				t.Errorf("repository phone = %q, want %q", tt.repo.gotPhone, tt.phone)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("residents = %v, want %v", got, tt.want)
			}
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
		repo      *fakeDeliveryRepository
		want      []*entities.Delivery
		wantErr   error
		wantCalls int
	}{
		{name: "empty apartment", apartment: "", repo: &fakeDeliveryRepository{}, wantErr: errAny},
		{name: "invalid status", apartment: "101", status: ptr(entities.DeliveryStatus("lost")), repo: &fakeDeliveryRepository{}, wantErr: errAny},
		{name: "empty status is invalid", apartment: "101", status: ptr(entities.DeliveryStatus("")), repo: &fakeDeliveryRepository{}, wantErr: errAny},
		{name: "repository error", apartment: "101", repo: &fakeDeliveryRepository{err: errRepo}, wantErr: errRepo, wantCalls: 1},
		{name: "nil status means any", apartment: "101", repo: &fakeDeliveryRepository{deliveries: deliveries}, want: deliveries, wantCalls: 1},
		{name: "pending status", apartment: "101", status: ptr(entities.DeliveryStatusPending), repo: &fakeDeliveryRepository{deliveries: deliveries[1:]}, want: deliveries[1:], wantCalls: 1},
		{name: "deleted status", apartment: "101", status: ptr(entities.DeliveryStatusDeleted), repo: &fakeDeliveryRepository{deliveries: deliveries[:1]}, want: deliveries[:1], wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGetDeliveriesByApartment(tt.repo).Handle(context.Background(), tt.apartment, tt.status)

			checkErr(t, err, tt.wantErr)
			if tt.repo.calls != tt.wantCalls {
				t.Fatalf("repository calls = %d, want %d", tt.repo.calls, tt.wantCalls)
			}
			if tt.wantCalls == 0 {
				return
			}
			if tt.repo.gotApartment != tt.apartment {
				t.Errorf("repository apartment = %q, want %q", tt.repo.gotApartment, tt.apartment)
			}
			if tt.repo.gotStatus != tt.status {
				t.Errorf("repository status = %v, want %v", tt.repo.gotStatus, tt.status)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deliveries = %v, want %v", got, tt.want)
			}
		})
	}
}

// errAny marks a case where any non-nil error is expected.
var errAny = errors.New("any error")

func checkErr(t *testing.T, err, want error) {
	t.Helper()
	switch {
	case want == nil && err != nil:
		t.Fatalf("unexpected error: %v", err)
	case want == errAny && err == nil:
		t.Fatal("expected an error, got nil")
	case want != nil && want != errAny && !errors.Is(err, want):
		t.Fatalf("error = %v, want wrapping %v", err, want)
	}
}
