package repositories

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// fakeCollection is a hand-written fake of client.MongoClientCollectionPort that
// records every call and returns configurable results.
type fakeCollection struct {
	ensureIndexesErr error
	gotIndexes       []mongo.IndexModel

	insertErr   error
	gotInserted []interface{}

	findOneDoc    interface{}
	findOneErr    error
	gotFindOne    []interface{}
	findDocs      []interface{}
	findErr       error
	gotFind       []interface{}
	gotFindOpts   []*options.FindOptions
	updateResult  *mongo.UpdateResult
	updateErr     error
	gotUpdFilter  []interface{}
	gotUpdate     []interface{}
	gotUpdateOpts []*options.UpdateOptions
}

func (f *fakeCollection) InsertOne(_ context.Context, document interface{}, _ ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	f.gotInserted = append(f.gotInserted, document)
	if f.insertErr != nil {
		return nil, f.insertErr
	}
	return &mongo.InsertOneResult{}, nil
}

func (f *fakeCollection) FindOne(_ context.Context, filter interface{}, _ ...*options.FindOneOptions) *mongo.SingleResult {
	f.gotFindOne = append(f.gotFindOne, filter)
	doc := f.findOneDoc
	if doc == nil {
		doc = bson.D{}
	}
	return mongo.NewSingleResultFromDocument(doc, f.findOneErr, nil)
}

func (f *fakeCollection) Find(_ context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	f.gotFind = append(f.gotFind, filter)
	f.gotFindOpts = append(f.gotFindOpts, opts...)
	if f.findErr != nil {
		return nil, f.findErr
	}
	return mongo.NewCursorFromDocuments(f.findDocs, nil, nil)
}

func (f *fakeCollection) UpdateOne(_ context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	f.gotUpdFilter = append(f.gotUpdFilter, filter)
	f.gotUpdate = append(f.gotUpdate, update)
	f.gotUpdateOpts = append(f.gotUpdateOpts, opts...)
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updateResult != nil {
		return f.updateResult, nil
	}
	return &mongo.UpdateResult{MatchedCount: 1}, nil
}

func (f *fakeCollection) DeleteOne(_ context.Context, _ interface{}, _ ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{}, nil
}

func (f *fakeCollection) EnsureIndexes(_ context.Context, indexes []mongo.IndexModel) error {
	f.gotIndexes = indexes
	return f.ensureIndexesErr
}

func duplicateKeyErr() error {
	return mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000, Message: "E11000 duplicate key"}}}
}

// lastFilter returns the filter of the last UpdateOne call as bson.M.
func (f *fakeCollection) lastUpdateFilter(t *testing.T) bson.M {
	t.Helper()
	if len(f.gotUpdFilter) == 0 {
		t.Fatal("UpdateOne was not called")
	}
	m, ok := f.gotUpdFilter[len(f.gotUpdFilter)-1].(bson.M)
	if !ok {
		t.Fatalf("filter type = %T, want bson.M", f.gotUpdFilter[len(f.gotUpdFilter)-1])
	}
	return m
}

// lastSet returns the $set document of the last UpdateOne call.
func (f *fakeCollection) lastSet(t *testing.T) bson.M {
	t.Helper()
	if len(f.gotUpdate) == 0 {
		t.Fatal("UpdateOne was not called")
	}
	update, ok := f.gotUpdate[len(f.gotUpdate)-1].(bson.M)
	if !ok {
		t.Fatalf("update type = %T, want bson.M", f.gotUpdate[len(f.gotUpdate)-1])
	}
	set, ok := update["$set"].(bson.M)
	if !ok {
		t.Fatalf("$set type = %T, want bson.M (update: %v)", update["$set"], update)
	}
	return set
}

// assertFilter checks that got has exactly the keys/values in want.
func assertFilter(t *testing.T, got, want bson.M) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("filter = %v, want %v", got, want)
		return
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("filter[%q] = %v, want %v", k, got[k], v)
		}
	}
}

// assertRecentTime checks that v is a UTC time.Time set between before and now.
func assertRecentTime(t *testing.T, name string, v interface{}, before time.Time) {
	t.Helper()
	ts, ok := v.(time.Time)
	if !ok {
		t.Errorf("%s type = %T, want time.Time", name, v)
		return
	}
	if ts.Before(before) || ts.After(time.Now().UTC()) {
		t.Errorf("%s = %v, want between %v and now", name, ts, before)
	}
	if ts.Location() != time.UTC {
		t.Errorf("%s location = %v, want UTC", name, ts.Location())
	}
}

// assertIndexes compares index keys and uniqueness.
func assertIndexes(t *testing.T, got []mongo.IndexModel, wantKeys []bson.D, wantUnique []bool) {
	t.Helper()
	if len(got) != len(wantKeys) {
		t.Fatalf("got %d indexes, want %d", len(got), len(wantKeys))
	}
	for i := range got {
		keys, ok := got[i].Keys.(bson.D)
		if !ok {
			t.Fatalf("index %d keys type = %T, want bson.D", i, got[i].Keys)
		}
		if len(keys) != len(wantKeys[i]) {
			t.Fatalf("index %d keys = %v, want %v", i, keys, wantKeys[i])
		}
		for j := range keys {
			if keys[j] != wantKeys[i][j] {
				t.Errorf("index %d key %d = %v, want %v", i, j, keys[j], wantKeys[i][j])
			}
		}
		unique := got[i].Options != nil && got[i].Options.Unique != nil && *got[i].Options.Unique
		if unique != wantUnique[i] {
			t.Errorf("index %d unique = %v, want %v", i, unique, wantUnique[i])
		}
	}
}
