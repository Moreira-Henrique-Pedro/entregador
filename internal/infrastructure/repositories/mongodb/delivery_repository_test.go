package mongodb

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/repositories/mongodb/client/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/repositories/mongodb/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newDeliveryRepo(t *testing.T) (*DeliveryRepository, *mocks.MongoClientCollectionPort) {
	t.Helper()
	coll := mocks.NewMongoClientCollectionPort(t)
	expectIndexes(coll)
	repo, err := NewDeliveryRepository(context.Background(), coll)
	require.NoError(t, err)
	return repo.(*DeliveryRepository), coll
}

func statusPtr(s entities.DeliveryStatus) *entities.DeliveryStatus { return &s }

func TestNewDeliveryRepository(t *testing.T) {
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
			coll := mocks.NewMongoClientCollectionPort(t)
			var gotIndexes []mongo.IndexModel
			coll.EXPECT().EnsureIndexes(mock.Anything, mock.Anything).
				Run(func(_ context.Context, indexes []mongo.IndexModel) { gotIndexes = indexes }).
				Return(tt.err).
				Once()

			repo, err := NewDeliveryRepository(context.Background(), coll)

			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				assert.Nil(t, repo)
				assert.Contains(t, err.Error(), "deliveries indexes")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)
			assertIndexes(t, gotIndexes,
				[]bson.D{
					{{Key: "delivery_id", Value: 1}},
					{{Key: "apartment", Value: 1}, {Key: "createdat", Value: -1}},
					{{Key: "apartment", Value: 1}, {Key: "status", Value: 1}, {Key: "createdat", Value: -1}},
					{{Key: "createdat", Value: -1}},
					{{Key: "status", Value: 1}, {Key: "createdat", Value: -1}},
				},
				[]bool{true, false, false, false, false},
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
			repo, coll := newDeliveryRepo(t)
			var model *models.Delivery
			if !tt.wantAnyErr {
				coll.EXPECT().InsertOne(mock.Anything, mock.AnythingOfType("*models.Delivery")).
					Run(func(_ context.Context, document any, _ ...*options.InsertOneOptions) {
						model = document.(*models.Delivery)
					}).
					Return(&mongo.InsertOneResult{}, tt.insertErr).
					Once()
			}
			before := time.Now().UTC()

			err := repo.Insert(context.Background(), tt.delivery)

			if tt.wantAnyErr {
				require.Error(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
			require.NotNil(t, model)
			if tt.insertErr == nil {
				assert.True(t, tt.delivery.CreatedAt.Equal(model.CreatedAt), "entity CreatedAt = %v, want %v", tt.delivery.CreatedAt, model.CreatedAt)
				assert.True(t, tt.delivery.UpdatedAt.Equal(model.UpdatedAt), "entity UpdatedAt = %v, want %v", tt.delivery.UpdatedAt, model.UpdatedAt)
			}
			assert.Equal(t, tt.delivery.DeliveryID, model.DeliveryID)
			assert.Equal(t, string(tt.delivery.Status), model.Status)
			if tt.wantCreatedAt != nil {
				assert.True(t, model.CreatedAt.Equal(*tt.wantCreatedAt), "CreatedAt = %v, want %v", model.CreatedAt, *tt.wantCreatedAt)
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
		doc     any
		findErr error
		wantErr error
	}{
		{name: "decodes delivery", doc: stored},
		{name: "not found", findErr: mongo.ErrNoDocuments, wantErr: entities.ErrEntityNotFound},
		{name: "find error", findErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newDeliveryRepo(t)
			coll.EXPECT().FindOne(mock.Anything, bson.M{"delivery_id": "d1"}).
				Return(singleResult(tt.doc, tt.findErr)).
				Once()

			got, err := repo.FindByDeliveryID(context.Background(), "d1")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				if errors.Is(tt.wantErr, entities.ErrEntityNotFound) {
					assert.Contains(t, err.Error(), "d1")
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, stored.ToEntity(), got)
		})
	}
}

func TestDeliveryRepositoryFind(t *testing.T) {
	boom := errors.New("boom")
	docs := []any{
		models.Delivery{ID: "d2", DeliveryID: "d2", Apartment: "101", Status: "deleted"},
		models.Delivery{ID: "d1", DeliveryID: "d1", Apartment: "101", Status: "pending"},
	}
	sortedByCreationDesc := mock.MatchedBy(func(o *options.FindOptions) bool {
		return o != nil && reflect.DeepEqual(o.Sort, bson.D{{Key: "createdat", Value: -1}})
	})

	tests := []struct {
		name       string
		filter     entities.DeliveryFilter
		docs       []any
		findErr    error
		wantFilter bson.M
		wantErr    error
		wantIDs    []string
	}{
		{name: "all deliveries", docs: docs, wantFilter: bson.M{}, wantIDs: []string{"d2", "d1"}},
		{name: "by apartment", filter: entities.DeliveryFilter{Apartment: "101"}, docs: docs, wantFilter: bson.M{"apartment": "101"}, wantIDs: []string{"d2", "d1"}},
		{name: "by status", filter: entities.DeliveryFilter{Status: statusPtr(entities.DeliveryStatusPending)}, docs: docs[1:], wantFilter: bson.M{"status": "pending"}, wantIDs: []string{"d1"}},
		{name: "by apartment and status", filter: entities.DeliveryFilter{Apartment: "101", Status: statusPtr(entities.DeliveryStatusPending)}, docs: docs[1:], wantFilter: bson.M{"apartment": "101", "status": "pending"}, wantIDs: []string{"d1"}},
		{name: "no results", docs: []any{}, wantFilter: bson.M{}, wantIDs: []string{}},
		{name: "find error", findErr: boom, wantFilter: bson.M{}, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newDeliveryRepo(t)
			call := coll.EXPECT().Find(mock.Anything, tt.wantFilter, sortedByCreationDesc).Once()
			if tt.findErr != nil {
				call.Return(nil, tt.findErr)
			} else {
				call.Return(cursor(t, tt.docs), nil)
			}

			got, err := repo.Find(context.Background(), tt.filter)

			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			require.NotNil(t, got)
			require.Len(t, got, len(tt.wantIDs))
			for i, d := range got {
				assert.Equalf(t, tt.wantIDs[i], d.DeliveryID, "delivery %d id", i)
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
			repo, coll := newDeliveryRepo(t)
			var gotUpdate any
			expectUpdate(coll, bson.M{"delivery_id": "d1", "status": "pending", "deleteat": time.Time{}}, tt.result, tt.updateErr, &gotUpdate)
			before := time.Now().UTC()

			err := repo.MarkAsDeleted(context.Background(), "d1")
			require.ErrorIs(t, err, tt.wantErr)
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) {
				assert.Contains(t, err.Error(), "d1")
			}

			set := setOf(t, gotUpdate)
			assert.Len(t, set, 3, "$set = %v, want status, deleteat and updatedat", set)
			assert.Equal(t, "deleted", set["status"])
			assertRecentTime(t, "deleteat", set["deleteat"], before)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
		})
	}
}

func TestDeliveryRepositoryMarkAsNotified(t *testing.T) {
	boom := errors.New("boom")
	arrival := func(r *DeliveryRepository) error { return r.MarkArrivalAsNotified(context.Background(), "d1") }
	pickup := func(r *DeliveryRepository) error { return r.MarkPickupAsNotified(context.Background(), "d1") }

	tests := []struct {
		name      string
		call      func(*DeliveryRepository) error
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
			repo, coll := newDeliveryRepo(t)
			var gotUpdate any
			expectUpdate(coll, bson.M{"delivery_id": "d1"}, tt.result, tt.updateErr, &gotUpdate)
			before := time.Now().UTC()

			err := tt.call(repo)
			require.ErrorIs(t, err, tt.wantErr)
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) {
				assert.Contains(t, err.Error(), "d1")
			}

			set := setOf(t, gotUpdate)
			assert.Len(t, set, 2, "$set = %v, want %s and updatedat", set, tt.field)
			assertRecentTime(t, tt.field, set[tt.field], before)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
		})
	}
}
