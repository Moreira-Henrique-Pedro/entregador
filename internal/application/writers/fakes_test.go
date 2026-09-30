package writers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type fakeResidentRepository struct {
	residents      map[string]*entities.Resident
	findErr        error
	ensuredOthers  []string
	updated        []*entities.Resident
	deletedIDs     []string
	insertedResult []*entities.Resident
}

func newFakeResidentRepository(residents ...*entities.Resident) *fakeResidentRepository {
	repo := &fakeResidentRepository{residents: map[string]*entities.Resident{}}
	for _, resident := range residents {
		repo.residents[resident.ResidentID] = resident
	}
	return repo
}

func (f *fakeResidentRepository) Insert(_ context.Context, resident *entities.Resident) error {
	f.insertedResult = append(f.insertedResult, resident)
	return nil
}

func (f *fakeResidentRepository) EnsureOtherResident(_ context.Context, apartment string) error {
	f.ensuredOthers = append(f.ensuredOthers, apartment)
	return nil
}

func (f *fakeResidentRepository) Update(_ context.Context, resident *entities.Resident) error {
	f.updated = append(f.updated, resident)
	return nil
}

func (f *fakeResidentRepository) DeleteByResidentID(_ context.Context, residentID string) error {
	f.deletedIDs = append(f.deletedIDs, residentID)
	return nil
}

func (f *fakeResidentRepository) FindByResidentID(_ context.Context, residentID string) (*entities.Resident, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	resident, ok := f.residents[residentID]
	if !ok {
		return nil, entities.ErrEntityNotFound
	}
	return resident, nil
}

func (f *fakeResidentRepository) FindByApartment(context.Context, string) ([]*entities.Resident, error) {
	return nil, nil
}

func (f *fakeResidentRepository) FindByPhone(context.Context, string) ([]*entities.Resident, error) {
	return nil, nil
}

type fakeDeliveryRepository struct {
	inserted   []*entities.Delivery
	deleteErr  error
	deletedIDs []string
}

func (f *fakeDeliveryRepository) Insert(_ context.Context, delivery *entities.Delivery) error {
	f.inserted = append(f.inserted, delivery)
	return nil
}

func (f *fakeDeliveryRepository) FindByApartment(context.Context, string, *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	return nil, nil
}

func (f *fakeDeliveryRepository) MarkAsDeleted(_ context.Context, deliveryID string) error {
	f.deletedIDs = append(f.deletedIDs, deliveryID)
	return f.deleteErr
}
