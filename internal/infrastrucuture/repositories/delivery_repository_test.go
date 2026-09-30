package repositories

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newDeliveryRepo(t *testing.T, fake *fakeCollection) *MongoDBDeliveryRepository {
	t.Helper()
	repo, err := NewMongoDBDeliveryRepository(context.Background(), fake)
	if err != nil {
		t.Fatalf("NewMongoDBDeliveryRepository() error = %v", err)
	}
	return repo.(*MongoDBDeliveryRepository)
}

func statusPtr(s entities.DeliveryStatus) *entities.DeliveryStatus { return &s }

func TestNewMongoDBDeliveryRepository(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{name: "ensures indexes"},
		{name: "propagates ensure indexes error", err: boom, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{ensureIndexesErr: tt.err}
			repo, err := NewMongoDBDeliveryRepository(context.Background(), fake)

			if tt.wantErr {
				if !errors.Is(err, tt.err) || repo != nil {
					t.Fatalf("got (%v, %v), want (nil, wrapping %v)", repo, err, tt.err)
				}
				if !strings.Contains(err.Error(), "deliveries indexes") {
					t.Errorf("error = %q, want it to mention deliveries indexes", err)
				}
				return
			}
			if err != nil || repo == nil {
				t.Fatalf("got (%v, %v), want repository and nil error", repo, err)
			}
			assertIndexes(t, fake.gotIndexes,
				[]bson.D{
					{{Key: "delivery_id", Value: 1}},
					{{Key: "apartment", Value: 1}, {Key: "createdat", Value: -1}},
					{{Key: "apartment", Value: 1}, {Key: "status", Value: 1}, {Key: "createdat", Value: -1}},
				},
				[]bool{true, false, false},
			)
		})
	}
}

func TestDeliveryRepositoryInsert(t *testing.T) {
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	boom := errors.New("boom")

	tests := []struct {
		name          string
		delivery      *entities.Delivery
		insertErr     error
		wantErr       error
		wantAnyErr    bool
		wantCreatedAt *time.Time
	}{
		{name: "nil delivery", delivery: nil, wantAnyErr: true},
		{name: "sets timestamps", delivery: &entities.Delivery{DeliveryID: "d1", Apartment: "101", Status: entities.DeliveryStatusPending}},
		{name: "keeps existing created at", delivery: &entities.Delivery{DeliveryID: "d1", CreatedAt: created}, wantCreatedAt: &created},
		{name: "duplicate key is swallowed", delivery: &entities.Delivery{DeliveryID: "d1"}, insertErr: duplicateKeyErr()},
		{name: "other errors are returned", delivery: &entities.Delivery{DeliveryID: "d1"}, insertErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{insertErr: tt.insertErr}
			repo := newDeliveryRepo(t, fake)
			before := time.Now().UTC()

			err := repo.Insert(context.Background(), tt.delivery)

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if len(fake.gotInserted) != 0 {
					t.Errorf("InsertOne called %d times, want 0", len(fake.gotInserted))
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if len(fake.gotInserted) != 1 {
				t.Fatalf("InsertOne called %d times, want 1", len(fake.gotInserted))
			}
			model, ok := fake.gotInserted[0].(*models.Delivery)
			if !ok {
				t.Fatalf("inserted type = %T, want *models.Delivery", fake.gotInserted[0])
			}
			if model.DeliveryID != tt.delivery.DeliveryID || model.Status != string(tt.delivery.Status) {
				t.Errorf("inserted = %+v, want converted %+v", model, tt.delivery)
			}
			if tt.wantCreatedAt != nil {
				if !model.CreatedAt.Equal(*tt.wantCreatedAt) {
					t.Errorf("CreatedAt = %v, want %v", model.CreatedAt, *tt.wantCreatedAt)
				}
			} else {
				assertRecentTime(t, "CreatedAt", model.CreatedAt, before)
			}
			assertRecentTime(t, "UpdatedAt", model.UpdatedAt, before)
		})
	}
}

func TestDeliveryRepositoryFindByDeliveryID(t *testing.T) {
	boom := errors.New("boom")
	stored := models.Delivery{ID: "d1", DeliveryID: "d1", Apartment: "101", ResidentID: "r1", PackageType: "box", Urgency: "high", Status: "pending"}

	tests := []struct {
		name    string
		doc     interface{}
		findErr error
		wantErr error
	}{
		{name: "decodes delivery", doc: stored},
		{name: "not found", findErr: mongo.ErrNoDocuments, wantErr: entities.ErrEntityNotFound},
		{name: "find error", findErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{findOneDoc: tt.doc, findOneErr: tt.findErr}
			repo := newDeliveryRepo(t, fake)

			got, err := repo.FindByDeliveryID(context.Background(), "d1")

			if len(fake.gotFindOne) != 1 {
				t.Fatalf("FindOne called %d times, want 1", len(fake.gotFindOne))
			}
			filter, ok := fake.gotFindOne[0].(bson.M)
			if !ok {
				t.Fatalf("filter type = %T, want bson.M", fake.gotFindOne[0])
			}
			assertFilter(t, filter, bson.M{"delivery_id": "d1"})

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) || got != nil {
					t.Fatalf("got (%v, %v), want (nil, %v)", got, err, tt.wantErr)
				}
				if errors.Is(tt.wantErr, entities.ErrEntityNotFound) && !strings.Contains(err.Error(), "d1") {
					t.Errorf("error = %q, want it to contain the delivery id", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := stored.ToEntity()
			if *got != *want {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func TestDeliveryRepositoryFindByApartment(t *testing.T) {
	boom := errors.New("boom")
	docs := []interface{}{
		models.Delivery{ID: "d2", DeliveryID: "d2", Apartment: "101", Status: "deleted"},
		models.Delivery{ID: "d1", DeliveryID: "d1", Apartment: "101", Status: "pending"},
	}

	tests := []struct {
		name       string
		status     *entities.DeliveryStatus
		docs       []interface{}
		findErr    error
		wantFilter bson.M
		wantErr    error
		wantIDs    []string
	}{
		{name: "without status", docs: docs, wantFilter: bson.M{"apartment": "101"}, wantIDs: []string{"d2", "d1"}},
		{name: "with status", status: statusPtr(entities.DeliveryStatusPending), docs: docs[1:], wantFilter: bson.M{"apartment": "101", "status": "pending"}, wantIDs: []string{"d1"}},
		{name: "no results", docs: []interface{}{}, wantFilter: bson.M{"apartment": "101"}, wantIDs: []string{}},
		{name: "find error", findErr: boom, wantFilter: bson.M{"apartment": "101"}, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{findDocs: tt.docs, findErr: tt.findErr}
			repo := newDeliveryRepo(t, fake)

			got, err := repo.FindByApartment(context.Background(), "101", tt.status)

			if len(fake.gotFind) != 1 {
				t.Fatalf("Find called %d times, want 1", len(fake.gotFind))
			}
			filter, ok := fake.gotFind[0].(bson.M)
			if !ok {
				t.Fatalf("filter type = %T, want bson.M", fake.gotFind[0])
			}
			assertFilter(t, filter, tt.wantFilter)

			if len(fake.gotFindOpts) != 1 {
				t.Fatalf("got %d find options, want 1", len(fake.gotFindOpts))
			}
			sort, ok := fake.gotFindOpts[0].Sort.(bson.D)
			if !ok || len(sort) != 1 || sort[0] != (bson.E{Key: "createdat", Value: -1}) {
				t.Errorf("sort = %v, want createdat desc", fake.gotFindOpts[0].Sort)
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got == nil || len(got) != len(tt.wantIDs) {
				t.Fatalf("got %d deliveries (%v), want %d", len(got), got, len(tt.wantIDs))
			}
			for i, d := range got {
				if d.DeliveryID != tt.wantIDs[i] {
					t.Errorf("delivery %d id = %q, want %q", i, d.DeliveryID, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestDeliveryRepositoryMarkAsDeleted(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name      string
		result    *mongo.UpdateResult
		updateErr error
		wantErr   error
	}{
		{name: "marks pending delivery as deleted"},
		{name: "not found or not pending", result: &mongo.UpdateResult{MatchedCount: 0}, wantErr: entities.ErrEntityNotFound},
		{name: "update error", updateErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{updateResult: tt.result, updateErr: tt.updateErr}
			repo := newDeliveryRepo(t, fake)
			before := time.Now().UTC()

			err := repo.MarkAsDeleted(context.Background(), "d1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) && !strings.Contains(err.Error(), "d1") {
				t.Errorf("error = %q, want it to contain the delivery id", err)
			}

			assertFilter(t, fake.lastUpdateFilter(t), bson.M{"delivery_id": "d1", "status": "pending", "deleteat": time.Time{}})
			set := fake.lastSet(t)
			if len(set) != 3 || set["status"] != "deleted" {
				t.Errorf("$set = %v, want status deleted, deleteat and updatedat", set)
			}
			assertRecentTime(t, "deleteat", set["deleteat"], before)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
		})
	}
}

func TestDeliveryRepositoryMarkAsNotified(t *testing.T) {
	boom := errors.New("boom")
	arrival := func(r *MongoDBDeliveryRepository) error { return r.MarkArrivalAsNotified(context.Background(), "d1") }
	pickup := func(r *MongoDBDeliveryRepository) error { return r.MarkPickupAsNotified(context.Background(), "d1") }

	tests := []struct {
		name      string
		call      func(*MongoDBDeliveryRepository) error
		field     string
		result    *mongo.UpdateResult
		updateErr error
		wantErr   error
	}{
		{name: "arrival", call: arrival, field: "arrivalnotifiedat"},
		{name: "arrival not found", call: arrival, field: "arrivalnotifiedat", result: &mongo.UpdateResult{MatchedCount: 0}, wantErr: entities.ErrEntityNotFound},
		{name: "arrival update error", call: arrival, field: "arrivalnotifiedat", updateErr: boom, wantErr: boom},
		{name: "pickup", call: pickup, field: "pickupnotifiedat"},
		{name: "pickup not found", call: pickup, field: "pickupnotifiedat", result: &mongo.UpdateResult{MatchedCount: 0}, wantErr: entities.ErrEntityNotFound},
		{name: "pickup update error", call: pickup, field: "pickupnotifiedat", updateErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{updateResult: tt.result, updateErr: tt.updateErr}
			repo := newDeliveryRepo(t, fake)
			before := time.Now().UTC()

			err := tt.call(repo)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) && !strings.Contains(err.Error(), "d1") {
				t.Errorf("error = %q, want it to contain the delivery id", err)
			}

			assertFilter(t, fake.lastUpdateFilter(t), bson.M{"delivery_id": "d1"})
			set := fake.lastSet(t)
			if len(set) != 2 {
				t.Errorf("$set = %v, want %s and updatedat", set, tt.field)
			}
			assertRecentTime(t, tt.field, set[tt.field], before)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
		})
	}
}
