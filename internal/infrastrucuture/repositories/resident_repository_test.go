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

func newResidentRepo(t *testing.T, fake *fakeCollection) *MongoDBResidentRepository {
	t.Helper()
	repo, err := NewMongoDBResidentRepository(context.Background(), fake)
	if err != nil {
		t.Fatalf("NewMongoDBResidentRepository() error = %v", err)
	}
	return repo.(*MongoDBResidentRepository)
}

func TestNewMongoDBResidentRepository(t *testing.T) {
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
			fake := &fakeCollection{ensureIndexesErr: tt.err}
			repo, err := NewMongoDBResidentRepository(context.Background(), fake)

			if tt.wantErr {
				if !errors.Is(err, tt.err) || repo != nil {
					t.Fatalf("got (%v, %v), want (nil, wrapping %v)", repo, err, tt.err)
				}
				if !strings.Contains(err.Error(), "residents indexes") {
					t.Errorf("error = %q, want it to mention residents indexes", err)
				}
				return
			}
			if err != nil || repo == nil {
				t.Fatalf("got (%v, %v), want repository and nil error", repo, err)
			}
			assertIndexes(t, fake.gotIndexes,
				[]bson.D{
					{{Key: "resident_id", Value: 1}},
					{{Key: "apartment", Value: 1}, {Key: "deleteat", Value: 1}},
					{{Key: "phone", Value: 1}, {Key: "deleteat", Value: 1}},
				},
				[]bool{true, false, false},
			)
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
		{name: "sets timestamps and defaults type", resident: &entities.Resident{ResidentID: "r1", Name: "Ana"}, wantType: "resident"},
		{name: "keeps existing created at", resident: &entities.Resident{ResidentID: "r1", CreatedAt: created}, wantCreatedAt: &created, wantType: "resident"},
		{name: "keeps other type", resident: &entities.Resident{ResidentID: "other-101", Type: entities.ResidentTypeOther}, wantType: "other"},
		{name: "duplicate key is swallowed", resident: &entities.Resident{ResidentID: "r1"}, insertErr: duplicateKeyErr(), wantType: "resident"},
		{name: "other errors are returned", resident: &entities.Resident{ResidentID: "r1"}, insertErr: boom, wantErr: boom, wantType: "resident"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{insertErr: tt.insertErr}
			repo := newResidentRepo(t, fake)
			before := time.Now().UTC()

			err := repo.Insert(context.Background(), tt.resident)

			switch {
			case tt.wantAnyErr:
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			case !errors.Is(err, tt.wantErr):
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantNoInsert {
				if len(fake.gotInserted) != 0 {
					t.Errorf("InsertOne called %d times, want 0", len(fake.gotInserted))
				}
				return
			}
			if len(fake.gotInserted) != 1 {
				t.Fatalf("InsertOne called %d times, want 1", len(fake.gotInserted))
			}
			model, ok := fake.gotInserted[0].(*models.Resident)
			if !ok {
				t.Fatalf("inserted type = %T, want *models.Resident", fake.gotInserted[0])
			}
			if model.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", model.Type, tt.wantType)
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
			fake := &fakeCollection{updateErr: tt.updateErr}
			repo := newResidentRepo(t, fake)
			before := time.Now().UTC()

			err := repo.EnsureOtherResident(context.Background(), "101")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}

			assertFilter(t, fake.lastUpdateFilter(t), bson.M{"resident_id": "other-101"})

			update, ok := fake.gotUpdate[0].(bson.M)
			if !ok {
				t.Fatalf("update type = %T, want bson.M", fake.gotUpdate[0])
			}
			if len(update) != 1 {
				t.Errorf("update = %v, want only $setOnInsert", update)
			}
			model, ok := update["$setOnInsert"].(*models.Resident)
			if !ok {
				t.Fatalf("$setOnInsert type = %T, want *models.Resident", update["$setOnInsert"])
			}
			if model.ID != "other-101" || model.ResidentID != "other-101" || model.Apartment != "101" || model.Type != "other" {
				t.Errorf("$setOnInsert = %+v, want other resident of apartment 101", model)
			}
			assertRecentTime(t, "CreatedAt", model.CreatedAt, before)
			assertRecentTime(t, "UpdatedAt", model.UpdatedAt, before)

			if len(fake.gotUpdateOpts) != 1 || fake.gotUpdateOpts[0].Upsert == nil || !*fake.gotUpdateOpts[0].Upsert {
				t.Errorf("update options = %v, want upsert true", fake.gotUpdateOpts)
			}
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
			fake := &fakeCollection{updateResult: tt.result, updateErr: tt.updateErr}
			repo := newResidentRepo(t, fake)
			before := time.Now().UTC()

			err := repo.Update(context.Background(), tt.resident)

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if len(fake.gotUpdate) != 0 {
					t.Errorf("UpdateOne called %d times, want 0", len(fake.gotUpdate))
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if errors.Is(tt.wantErr, entities.ErrEntityNotFound) && !strings.Contains(err.Error(), "r1") {
				t.Errorf("error = %q, want it to contain the resident id", err)
			}

			assertFilter(t, fake.lastUpdateFilter(t), bson.M{"resident_id": "r1", "deleteat": time.Time{}})

			set := fake.lastSet(t)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
			delete(set, "updatedat")
			assertFilter(t, set, tt.wantSet)
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
			fake := &fakeCollection{updateResult: tt.result, updateErr: tt.updateErr}
			repo := newResidentRepo(t, fake)
			before := time.Now().UTC()

			err := repo.DeleteByResidentID(context.Background(), "r1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && errors.Is(tt.wantErr, entities.ErrEntityNotFound) && !strings.Contains(err.Error(), "r1") {
				t.Errorf("error = %q, want it to contain the resident id", err)
			}

			assertFilter(t, fake.lastUpdateFilter(t), bson.M{"resident_id": "r1", "deleteat": time.Time{}})
			set := fake.lastSet(t)
			if len(set) != 2 {
				t.Errorf("$set = %v, want deleteat and updatedat", set)
			}
			assertRecentTime(t, "deleteat", set["deleteat"], before)
			assertRecentTime(t, "updatedat", set["updatedat"], before)
		})
	}
}

func TestResidentRepositoryFindByResidentID(t *testing.T) {
	boom := errors.New("boom")
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	stored := models.Resident{ID: "r1", ResidentID: "r1", Apartment: "101", Name: "Ana", Phone: "5511", Type: "resident", CreatedAt: created}
	legacy := models.Resident{ID: "r2", ResidentID: "r2", Apartment: "101", Name: "Bia"}

	tests := []struct {
		name     string
		id       string
		doc      interface{}
		findErr  error
		wantErr  error
		wantName string
		wantType entities.ResidentType
	}{
		{name: "decodes resident", id: "r1", doc: stored, wantName: "Ana", wantType: entities.ResidentTypeResident},
		{name: "legacy resident without type", id: "r2", doc: legacy, wantName: "Bia", wantType: entities.ResidentTypeResident},
		{name: "not found", id: "r1", findErr: mongo.ErrNoDocuments, wantErr: entities.ErrEntityNotFound},
		{name: "find error", id: "r1", findErr: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{findOneDoc: tt.doc, findOneErr: tt.findErr}
			repo := newResidentRepo(t, fake)

			got, err := repo.FindByResidentID(context.Background(), tt.id)

			if len(fake.gotFindOne) != 1 {
				t.Fatalf("FindOne called %d times, want 1", len(fake.gotFindOne))
			}
			filter, ok := fake.gotFindOne[0].(bson.M)
			if !ok {
				t.Fatalf("filter type = %T, want bson.M", fake.gotFindOne[0])
			}
			assertFilter(t, filter, bson.M{"resident_id": tt.id, "deleteat": time.Time{}})

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) || got != nil {
					t.Fatalf("got (%v, %v), want (nil, %v)", got, err, tt.wantErr)
				}
				if errors.Is(tt.wantErr, entities.ErrEntityNotFound) && !strings.Contains(err.Error(), tt.id) {
					t.Errorf("error = %q, want it to contain the resident id", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ResidentID != tt.id || got.Name != tt.wantName || got.Type != tt.wantType {
				t.Errorf("got %+v, want id %q name %q type %q", got, tt.id, tt.wantName, tt.wantType)
			}
		})
	}
}

func TestResidentRepositoryFindMany(t *testing.T) {
	boom := errors.New("boom")
	docs := []interface{}{
		models.Resident{ID: "r1", ResidentID: "r1", Apartment: "101", Phone: "5511", Name: "Ana", Type: "resident"},
		models.Resident{ID: "other-101", ResidentID: "other-101", Apartment: "101", Name: "Outro", Type: "other"},
		models.Resident{ID: "r3", ResidentID: "r3", Apartment: "101", Phone: "5511", Name: "Legacy"},
	}

	tests := []struct {
		name       string
		call       func(*MongoDBResidentRepository) ([]*entities.Resident, error)
		docs       []interface{}
		findErr    error
		wantFilter bson.M
		wantErr    error
		wantTypes  []entities.ResidentType
	}{
		{
			name: "find by apartment",
			call: func(r *MongoDBResidentRepository) ([]*entities.Resident, error) {
				return r.FindByApartment(context.Background(), "101")
			},
			docs:       docs,
			wantFilter: bson.M{"apartment": "101", "deleteat": time.Time{}},
			wantTypes:  []entities.ResidentType{entities.ResidentTypeResident, entities.ResidentTypeOther, entities.ResidentTypeResident},
		},
		{
			name: "find by phone",
			call: func(r *MongoDBResidentRepository) ([]*entities.Resident, error) {
				return r.FindByPhone(context.Background(), "5511")
			},
			docs:       []interface{}{docs[0]},
			wantFilter: bson.M{"phone": "5511", "deleteat": time.Time{}},
			wantTypes:  []entities.ResidentType{entities.ResidentTypeResident},
		},
		{
			name: "no results returns empty slice",
			call: func(r *MongoDBResidentRepository) ([]*entities.Resident, error) {
				return r.FindByPhone(context.Background(), "000")
			},
			docs:       []interface{}{},
			wantFilter: bson.M{"phone": "000", "deleteat": time.Time{}},
			wantTypes:  []entities.ResidentType{},
		},
		{
			name: "find error",
			call: func(r *MongoDBResidentRepository) ([]*entities.Resident, error) {
				return r.FindByApartment(context.Background(), "101")
			},
			findErr:    boom,
			wantFilter: bson.M{"apartment": "101", "deleteat": time.Time{}},
			wantErr:    boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCollection{findDocs: tt.docs, findErr: tt.findErr}
			repo := newResidentRepo(t, fake)

			got, err := tt.call(repo)

			if len(fake.gotFind) != 1 {
				t.Fatalf("Find called %d times, want 1", len(fake.gotFind))
			}
			filter, ok := fake.gotFind[0].(bson.M)
			if !ok {
				t.Fatalf("filter type = %T, want bson.M", fake.gotFind[0])
			}
			assertFilter(t, filter, tt.wantFilter)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got == nil || len(got) != len(tt.wantTypes) {
				t.Fatalf("got %d residents (%v), want %d", len(got), got, len(tt.wantTypes))
			}
			for i, r := range got {
				if r.Type != tt.wantTypes[i] {
					t.Errorf("resident %d type = %q, want %q", i, r.Type, tt.wantTypes[i])
				}
			}
		})
	}
}
