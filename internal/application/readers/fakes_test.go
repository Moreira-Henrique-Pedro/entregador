package readers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type fakeResidentRepository struct {
	residents    []*entities.Resident
	err          error
	calls        int
	gotApartment string
	gotPhone     string
}

func (f *fakeResidentRepository) Insert(context.Context, *entities.Resident) error { return nil }

func (f *fakeResidentRepository) EnsureOtherResident(context.Context, string) error { return nil }

func (f *fakeResidentRepository) EnsurePrimaryResident(context.Context, string) error { return nil }

func (f *fakeResidentRepository) Update(context.Context, *entities.Resident) error { return nil }

func (f *fakeResidentRepository) DeleteByResidentID(context.Context, string) error { return nil }

func (f *fakeResidentRepository) FindByResidentID(context.Context, string) (*entities.Resident, error) {
	return nil, nil
}

func (f *fakeResidentRepository) FindByApartment(_ context.Context, apartment string) ([]*entities.Resident, error) {
	f.calls++
	f.gotApartment = apartment
	return f.residents, f.err
}

func (f *fakeResidentRepository) FindByPhone(_ context.Context, phone string) ([]*entities.Resident, error) {
	f.calls++
	f.gotPhone = phone
	return f.residents, f.err
}

type fakeDeliveryRepository struct {
	deliveries   []*entities.Delivery
	err          error
	calls        int
	gotApartment string
	gotStatus    *entities.DeliveryStatus
}

func (f *fakeDeliveryRepository) Insert(context.Context, *entities.Delivery) error { return nil }

func (f *fakeDeliveryRepository) FindByDeliveryID(context.Context, string) (*entities.Delivery, error) {
	return nil, nil
}

func (f *fakeDeliveryRepository) FindByApartment(_ context.Context, apartment string, status *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	f.calls++
	f.gotApartment = apartment
	f.gotStatus = status
	return f.deliveries, f.err
}

func (f *fakeDeliveryRepository) MarkAsDeleted(context.Context, string) error { return nil }

func (f *fakeDeliveryRepository) MarkArrivalAsNotified(context.Context, string) error { return nil }

func (f *fakeDeliveryRepository) MarkPickupAsNotified(context.Context, string) error { return nil }

func ptr[T any](v T) *T { return &v }
