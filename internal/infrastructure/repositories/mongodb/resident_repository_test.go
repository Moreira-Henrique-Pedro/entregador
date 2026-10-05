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

func newResidentRepo(t *testing.T) (*ResidentRepository, *mocks.MongoClientCollectionPort) {
	t.Helper()
	coll := mocks.NewMongoClientCollectionPort(t)
	expectIndexes(coll)
	repo, err := NewResidentRepository(context.Background(), coll)
	require.NoError(t, err)
	return repo.(*ResidentRepository), coll
}

func TestNewResidentRepository(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{name: "ensures indexes", err: nil},
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

			repo, err := NewResidentRepository(context.Background(), coll)

			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				assert.Nil(t, repo)
				assert.Contains(t, err.Error(), "residents indexes")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)
			assertIndexes(t, gotIndexes,
				[]bson.D{
					{{Key: "resident_id", Value: 1}},
					{{Key: "apartment", Value: 1}, {Key: "deleteat", Value: 1}},
					{{Key: "phone", Value: 1}, {Key: "deleteat", Value: 1}},
					{{Key: "apartment", Value: 1}},
				},
				[]bool{true, false, false, true},
			)
			primaryIndex := gotIndexes[3].Options
			require.NotNil(t, primaryIndex.Name)
			assert.Equal(t, "apartment_primary_unique", *primaryIndex.Name)
			assert.Equal(t, bson.M{"type": "resident-primary", "deleteat": time.Time{}}, primaryIndex.PartialFilterExpression)
		})
	}
}

func TestResidentRepositoryInsert(t *testing.T) {
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	boom := errors.New("boom")

	tests := []struct {
		name          string
		resident      *entities.Resident
		insertErr     error
		wantErr       error
		wantAnyErr    bool
		wantCreatedAt *time.Time
		wantType      string
		wantNoInsert  bool
	}{
		{name: "nil resident", resident: nil, wantAnyErr: true, wantNoInsert: true},
		{name: "sets timestamps and defaults type", resident: &entities.Resident{ResidentID: "r1", Name: "Ana"}, wantType: "resident-secondary"},
		{name: "keeps existing created at", resident: &entities.Resident{ResidentID: "r1", CreatedAt: created}, wantCreatedAt: &created, wantType: "resident-secondary"},
		{name: "keeps other type", resident: &entities.Resident{ResidentID: "other-101", Type: entities.ResidentTypeOther}, wantType: "other"},
		{name: "duplicate key is swallowed", resident: &entities.Resident{ResidentID: "r1"}, insertErr: duplicateKeyErr(), wantType: "resident-secondary"},
		{name: "other errors are returned", resident: &entities.Resident{ResidentID: "r1"}, insertErr: boom, wantErr: boom, wantType: "resident-secondary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			var model *models.Resident
			if !tt.wantNoInsert {
				coll.EXPECT().InsertOne(mock.Anything, mock.AnythingOfType("*models.Resident")).
					Run(func(_ context.Context, document any, _ ...*options.InsertOneOptions) {
						model = document.(*models.Resident)
					}).
					Return(&mongo.InsertOneResult{}, tt.insertErr).
					Once()
			}
			before := time.Now().UTC()

			err := repo.Insert(context.Background(), tt.resident)

			if tt.wantAnyErr {
				require.Error(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			if tt.wantNoInsert {
				return
			}
			require.NotNil(t, model)
			assert.Equal(t, tt.wantType, model.Type)
			if tt.wantCreatedAt != nil {
				assert.True(t, model.CreatedAt.Equal(*tt.wantCreatedAt), "CreatedAt = %v, want %v", model.CreatedAt, *tt.wantCreatedAt)
			} else {
				assertRecentTime(t, "CreatedAt", model.CreatedAt, before)
			}
			assertRecentTime(t, "UpdatedAt", model.UpdatedAt, before)
		})
	}
}

func TestResidentRepositoryEnsureOtherResident(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name      string
		updateErr error
		wantErr   error
	}{
		{name: "upserts other resident"},
		{name: "duplicate key is swallowed", updateErr: duplicateKeyErr()},
		{name: "other errors are returned", updateErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			var gotUpdate any
			var gotOpts []*options.UpdateOptions
			coll.EXPECT().UpdateOne(mock.Anything, bson.M{"resident_id": "other-101"}, mock.Anything, mock.Anything).
				Run(func(_ context.Context, _ any, update any, opts ...*options.UpdateOptions) {
					gotUpdate = update
					gotOpts = opts
				}).
				Return(updateResult(nil, tt.updateErr)).
				Once()
			before := time.Now().UTC()

			err := repo.EnsureOtherResident(context.Background(), "101")
			require.ErrorIs(t, err, tt.wantErr)

			update, ok := gotUpdate.(bson.M)
			require.Truef(t, ok, "update type = %T, want bson.M", gotUpdate)
			assert.Len(t, update, 1, "update = %v, want only $setOnInsert", update)
			model, ok := update["$setOnInsert"].(*models.Resident)
			require.Truef(t, ok, "$setOnInsert type = %T, want *models.Resident", update["$setOnInsert"])
			assert.Equal(t, "other-101", model.ID)
			assert.Equal(t, "other-101", model.ResidentID)
			assert.Equal(t, "101", model.Apartment)
			assert.Equal(t, "other", model.Type)
			assertRecentTime(t, "CreatedAt", model.CreatedAt, before)
			assertRecentTime(t, "UpdatedAt", model.UpdatedAt, before)

			require.Len(t, gotOpts, 1)
			require.NotNil(t, gotOpts[0].Upsert)
			assert.True(t, *gotOpts[0].Upsert)
		})
	}
}

type findOneResult struct {
	doc any
	err error
}

func TestResidentRepositoryEnsurePrimaryResident(t *testing.T) {
	boom := errors.New("boom")
	noDocs := findOneResult{err: mongo.ErrNoDocuments}
	primary := findOneResult{doc: models.Resident{ResidentID: "ana", Apartment: "101", Type: "resident-primary"}}
	candidate := findOneResult{doc: models.Resident{ResidentID: "bia", Apartment: "101", Type: "resident-secondary"}}

	primaryFilter := bson.M{"apartment": "101", "type": "resident-primary", "deleteat": time.Time{}}
	candidateFilter := bson.M{"apartment": "101", "type": bson.M{"$ne": "other"}, "deleteat": time.Time{}}
	wantSort := bson.D{{Key: "createdat", Value: 1}, {Key: "resident_id", Value: 1}}
	sortedByCreation := mock.MatchedBy(func(o *options.FindOneOptions) bool {
		return o != nil && reflect.DeepEqual(o.Sort, wantSort)
	})

	tests := []struct {
		name        string
		finds       []findOneResult
		updateErr   error
		wantErr     error
		wantPromote bool
	}{
		{name: "apartment with primary is left untouched", finds: []findOneResult{primary}},
		{name: "oldest resident is promoted when there is no primary", finds: []findOneResult{noDocs, candidate}, wantPromote: true},
		{name: "apartment without residents has nothing to promote", finds: []findOneResult{noDocs, noDocs}},
		{name: "concurrent promotion is swallowed", finds: []findOneResult{noDocs, candidate}, updateErr: duplicateKeyErr(), wantPromote: true},
		{name: "primary lookup error is returned", finds: []findOneResult{{err: boom}}, wantErr: boom},
		{name: "candidate lookup error is returned", finds: []findOneResult{noDocs, {err: boom}}, wantErr: boom},
		{name: "promotion error is returned", finds: []findOneResult{noDocs, candidate}, updateErr: boom, wantErr: boom, wantPromote: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			coll.EXPECT().FindOne(mock.Anything, primaryFilter).
				Return(singleResult(tt.finds[0].doc, tt.finds[0].err)).
				Once()
			if len(tt.finds) == 2 {
				coll.EXPECT().FindOne(mock.Anything, candidateFilter, sortedByCreation).
					Return(singleResult(tt.finds[1].doc, tt.finds[1].err)).
					Once()
			}
			var gotUpdate any
			if tt.wantPromote {
				expectUpdate(coll, bson.M{"resident_id": "bia", "deleteat": time.Time{}}, nil, tt.updateErr, &gotUpdate)
			}
			before := time.Now().UTC()

			err := repo.EnsurePrimaryResident(context.Background(), "101")

			require.ErrorIs(t, err, tt.wantErr)
			if !tt.wantPromote {
				return
			}
			set := setOf(t, gotUpdate)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
			delete(set, "updatedat")
			assert.Equal(t, bson.M{"type": "resident-primary"}, set)
		})
	}
}

func TestResidentRepositoryUpdate(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name       string
		resident   *entities.Resident
		result     *mongo.UpdateResult
		updateErr  error
		wantErr    error
		wantAnyErr bool
		wantSet    bson.M
	}{
		{name: "nil resident", resident: nil, wantAnyErr: true},
		{
			name:     "sets all non-empty fields",
			resident: &entities.Resident{ResidentID: "r1", Name: "Ana", Apartment: "101", Phone: "5511"},
			wantSet:  bson.M{"name": "Ana", "apartment": "101", "phone": "5511"},
		},
		{
			name:     "skips empty fields",
			resident: &entities.Resident{ResidentID: "r1", Phone: "5511"},
			wantSet:  bson.M{"phone": "5511"},
		},
		{
			name:     "sets the type when informed",
			resident: &entities.Resident{ResidentID: "r1", Apartment: "202", Type: entities.ResidentTypeSecondary},
			wantSet:  bson.M{"apartment": "202", "type": "resident-secondary"},
		},
		{
			name:     "only updatedat when nothing to change",
			resident: &entities.Resident{ResidentID: "r1"},
			wantSet:  bson.M{},
		},
		{
			name:     "not found",
			resident: &entities.Resident{ResidentID: "r1", Name: "Ana"},
			result:   &mongo.UpdateResult{MatchedCount: 0},
			wantErr:  entities.ErrEntityNotFound,
			wantSet:  bson.M{"name": "Ana"},
		},
		{
			name:      "update error",
			resident:  &entities.Resident{ResidentID: "r1", Name: "Ana"},
			updateErr: boom,
			wantErr:   boom,
			wantSet:   bson.M{"name": "Ana"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			var gotUpdate any
			if !tt.wantAnyErr {
				expectUpdate(coll, bson.M{"resident_id": "r1", "deleteat": time.Time{}}, tt.result, tt.updateErr, &gotUpdate)
			}
			before := time.Now().UTC()

			err := repo.Update(context.Background(), tt.resident)

			if tt.wantAnyErr {
				require.Error(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) {
				assert.Contains(t, err.Error(), "r1")
			}

			set := setOf(t, gotUpdate)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
			delete(set, "updatedat")
			assert.Equal(t, tt.wantSet, set)
		})
	}
}

func TestResidentRepositoryDeleteByResidentID(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name      string
		result    *mongo.UpdateResult
		updateErr error
		wantErr   error
	}{
		{name: "soft deletes"},
		{name: "not found", result: &mongo.UpdateResult{MatchedCount: 0}, wantErr: entities.ErrEntityNotFound},
		{name: "update error", updateErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			var gotUpdate any
			expectUpdate(coll, bson.M{"resident_id": "r1", "deleteat": time.Time{}}, tt.result, tt.updateErr, &gotUpdate)
			before := time.Now().UTC()

			err := repo.DeleteByResidentID(context.Background(), "r1")
			require.ErrorIs(t, err, tt.wantErr)
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) {
				assert.Contains(t, err.Error(), "r1")
			}

			set := setOf(t, gotUpdate)
			assert.Len(t, set, 3, "$set = %v, want status, deleteat and updatedat", set)
			assert.Equal(t, "deleted", set["status"])
			assertRecentTime(t, "deleteat", set["deleteat"], before)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
		})
	}
}

func TestResidentRepositoryFindByResidentID(t *testing.T) {
	boom := errors.New("boom")
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	stored := models.Resident{ID: "r1", ResidentID: "r1", Apartment: "101", Name: "Ana", Phone: "5511", Type: "resident-primary", CreatedAt: created}
	legacy := models.Resident{ID: "r2", ResidentID: "r2", Apartment: "101", Name: "Bia"}

	tests := []struct {
		name     string
		id       string
		doc      any
		findErr  error
		wantErr  error
		wantName string
		wantType entities.ResidentType
	}{
		{name: "decodes resident", id: "r1", doc: stored, wantName: "Ana", wantType: entities.ResidentTypePrimary},
		{name: "legacy resident without type is secondary", id: "r2", doc: legacy, wantName: "Bia", wantType: entities.ResidentTypeSecondary},
		{name: "not found", id: "r1", findErr: mongo.ErrNoDocuments, wantErr: entities.ErrEntityNotFound},
		{name: "find error", id: "r1", findErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			coll.EXPECT().FindOne(mock.Anything, bson.M{"resident_id": tt.id, "deleteat": time.Time{}}).
				Return(singleResult(tt.doc, tt.findErr)).
				Once()

			got, err := repo.FindByResidentID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				if errors.Is(tt.wantErr, entities.ErrEntityNotFound) {
					assert.Contains(t, err.Error(), tt.id)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ResidentID)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantType, got.Type)
		})
	}
}

func TestResidentRepositoryFindMany(t *testing.T) {
	boom := errors.New("boom")
	docs := []any{
		models.Resident{ID: "r1", ResidentID: "r1", Apartment: "101", Phone: "5511", Name: "Ana", Type: "resident-primary"},
		models.Resident{ID: "other-101", ResidentID: "other-101", Apartment: "101", Name: "Outro", Type: "other"},
		models.Resident{ID: "r3", ResidentID: "r3", Apartment: "101", Phone: "5511", Name: "Legacy"},
	}

	tests := []struct {
		name       string
		call       func(*ResidentRepository) ([]*entities.Resident, error)
		docs       []any
		findErr    error
		wantFilter bson.M
		wantErr    error
		wantTypes  []entities.ResidentType
	}{
		{
			name: "find by apartment",
			call: func(r *ResidentRepository) ([]*entities.Resident, error) {
				return r.FindByApartment(context.Background(), "101")
			},
			docs:       docs,
			wantFilter: bson.M{"apartment": "101", "deleteat": time.Time{}},
			wantTypes:  []entities.ResidentType{entities.ResidentTypePrimary, entities.ResidentTypeOther, entities.ResidentTypeSecondary},
		},
		{
			name: "find by phone",
			call: func(r *ResidentRepository) ([]*entities.Resident, error) {
				return r.FindByPhone(context.Background(), "5511")
			},
			docs:       []any{docs[0]},
			wantFilter: bson.M{"phone": "5511", "deleteat": time.Time{}},
			wantTypes:  []entities.ResidentType{entities.ResidentTypePrimary},
		},
		{
			name: "no results returns empty slice",
			call: func(r *ResidentRepository) ([]*entities.Resident, error) {
				return r.FindByPhone(context.Background(), "000")
			},
			docs:       []any{},
			wantFilter: bson.M{"phone": "000", "deleteat": time.Time{}},
			wantTypes:  []entities.ResidentType{},
		},
		{
			name: "find error",
			call: func(r *ResidentRepository) ([]*entities.Resident, error) {
				return r.FindByApartment(context.Background(), "101")
			},
			findErr:    boom,
			wantFilter: bson.M{"apartment": "101", "deleteat": time.Time{}},
			wantErr:    boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, coll := newResidentRepo(t)
			call := coll.EXPECT().Find(mock.Anything, tt.wantFilter).Once()
			if tt.findErr != nil {
				call.Return(nil, tt.findErr)
			} else {
				call.Return(cursor(t, tt.docs), nil)
			}

			got, err := tt.call(repo)

			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr != nil {
				return
			}
			require.NotNil(t, got)
			require.Len(t, got, len(tt.wantTypes))
			for i, r := range got {
				assert.Equalf(t, tt.wantTypes[i], r.Type, "resident %d type", i)
			}
		})
	}
}
