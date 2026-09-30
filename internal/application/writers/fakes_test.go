package writers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
)

type fakeResidentRepository struct {
	residents map[string]*entities.Resident

	insertErr        error
	ensureOtherErr   error
	updateErr        error
	deleteErr        error
	findErr          error
	findApartmentErr error

	inserted      []*entities.Resident
	ensuredOthers []string
	updated       []*entities.Resident
	deletedIDs    []string
}

func newFakeResidentRepository(residents ...*entities.Resident) *fakeResidentRepository {
	repo := &fakeResidentRepository{residents: map[string]*entities.Resident{}}
	for _, resident := range residents {
		repo.residents[resident.ResidentID] = resident
	}
	return repo
}

func (f *fakeResidentRepository) Insert(_ context.Context, resident *entities.Resident) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserted = append(f.inserted, resident)
	return nil
}

func (f *fakeResidentRepository) EnsureOtherResident(_ context.Context, apartment string) error {
	if f.ensureOtherErr != nil {
		return f.ensureOtherErr
	}
	f.ensuredOthers = append(f.ensuredOthers, apartment)
	return nil
}

func (f *fakeResidentRepository) Update(_ context.Context, resident *entities.Resident) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, resident)
	return nil
}

func (f *fakeResidentRepository) DeleteByResidentID(_ context.Context, residentID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
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

func (f *fakeResidentRepository) FindByApartment(_ context.Context, apartment string) ([]*entities.Resident, error) {
	if f.findApartmentErr != nil {
		return nil, f.findApartmentErr
	}
	var residents []*entities.Resident
	for _, resident := range f.residents {
		if resident.Apartment == apartment {
			residents = append(residents, resident)
		}
	}
	return residents, nil
}

func (f *fakeResidentRepository) FindByPhone(context.Context, string) ([]*entities.Resident, error) {
	return nil, nil
}

type fakeDeliveryRepository struct {
	deliveries map[string]*entities.Delivery

	insertErr       error
	findErr         error
	markDeletedErr  error
	markNotifiedErr error

	inserted           []*entities.Delivery
	deletedIDs         []string
	arrivalNotifiedIDs []string
	pickupNotifiedIDs  []string
}

func newFakeDeliveryRepository(deliveries ...*entities.Delivery) *fakeDeliveryRepository {
	repo := &fakeDeliveryRepository{deliveries: map[string]*entities.Delivery{}}
	for _, delivery := range deliveries {
		repo.deliveries[delivery.DeliveryID] = delivery
	}
	return repo
}

func (f *fakeDeliveryRepository) Insert(_ context.Context, delivery *entities.Delivery) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserted = append(f.inserted, delivery)
	return nil
}

func (f *fakeDeliveryRepository) FindByDeliveryID(_ context.Context, deliveryID string) (*entities.Delivery, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	delivery, ok := f.deliveries[deliveryID]
	if !ok {
		return nil, entities.ErrEntityNotFound
	}
	return delivery, nil
}

func (f *fakeDeliveryRepository) FindByApartment(context.Context, string, *entities.DeliveryStatus) ([]*entities.Delivery, error) {
	return nil, nil
}

func (f *fakeDeliveryRepository) MarkAsDeleted(_ context.Context, deliveryID string) error {
	if f.markDeletedErr != nil {
		return f.markDeletedErr
	}
	f.deletedIDs = append(f.deletedIDs, deliveryID)
	return nil
}

func (f *fakeDeliveryRepository) MarkArrivalAsNotified(_ context.Context, deliveryID string) error {
	if f.markNotifiedErr != nil {
		return f.markNotifiedErr
	}
	f.arrivalNotifiedIDs = append(f.arrivalNotifiedIDs, deliveryID)
	return nil
}

func (f *fakeDeliveryRepository) MarkPickupAsNotified(_ context.Context, deliveryID string) error {
	if f.markNotifiedErr != nil {
		return f.markNotifiedErr
	}
	f.pickupNotifiedIDs = append(f.pickupNotifiedIDs, deliveryID)
	return nil
}

type fakePublisher struct {
	err      error
	topics   []string
	messages []*pubsub.Message[any]
}

func (f *fakePublisher) Publish(_ context.Context, topic string, messages ...*pubsub.Message[any]) error {
	if f.err != nil {
		return f.err
	}
	for _, message := range messages {
		f.topics = append(f.topics, topic)
		f.messages = append(f.messages, message)
	}
	return nil
}

func (f *fakePublisher) Close(context.Context) error { return nil }

// fakeNotifier fails for the phones in errByPhone and records the rest.
type fakeNotifier struct {
	errByPhone map[string]error
	sent       []notifier.Notification
}

func (f *fakeNotifier) Send(_ context.Context, notification notifier.Notification) error {
	if err := f.errByPhone[notification.Phone]; err != nil {
		return err
	}
	f.sent = append(f.sent, notification)
	return nil
}

func (f *fakeNotifier) phones() []string {
	phones := make([]string, 0, len(f.sent))
	for _, notification := range f.sent {
		phones = append(phones, notification.Phone)
	}
	return phones
}
