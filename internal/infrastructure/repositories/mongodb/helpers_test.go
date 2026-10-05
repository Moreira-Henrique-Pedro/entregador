package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/repositories/mongodb/client/mocks"
)

func duplicateKeyErr() error {
	return mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000, Message: "E11000 duplicate key"}}}
}

func expectIndexes(coll *mocks.MongoClientCollectionPort) {
	coll.EXPECT().EnsureIndexes(mock.Anything, mock.Anything).Return(nil).Once()
}

func singleResult(doc any, err error) *mongo.SingleResult {
	if doc == nil {
		doc = bson.D{}
	}
	return mongo.NewSingleResultFromDocument(doc, err, nil)
}

func cursor(t *testing.T, docs []any) *mongo.Cursor {
	t.Helper()
	c, err := mongo.NewCursorFromDocuments(docs, nil, nil)
	require.NoError(t, err)
	return c
}

func updateResult(result *mongo.UpdateResult, err error) (*mongo.UpdateResult, error) {
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &mongo.UpdateResult{MatchedCount: 1}, nil
	}
	return result, nil
}

func expectUpdate(coll *mocks.MongoClientCollectionPort, filter bson.M, result *mongo.UpdateResult, err error, got *any) {
	coll.EXPECT().UpdateOne(mock.Anything, filter, mock.Anything).
		Run(func(_ context.Context, _ any, update any, _ ...*options.UpdateOptions) {
			*got = update
		}).
		Return(updateResult(result, err)).
		Once()
}

func setOf(t *testing.T, update any) bson.M {
	t.Helper()
	m, ok := update.(bson.M)
	require.Truef(t, ok, "update type = %T, want bson.M", update)
	set, ok := m["$set"].(bson.M)
	require.Truef(t, ok, "$set type = %T, want bson.M (update: %v)", m["$set"], m)
	return set
}

func assertRecentTime(t *testing.T, name string, v any, before time.Time) {
	t.Helper()
	ts, ok := v.(time.Time)
	if !assert.Truef(t, ok, "%s type = %T, want time.Time", name, v) {
		return
	}
	assert.Falsef(t, ts.Before(before) || ts.After(time.Now().UTC()), "%s = %v, want between %v and now", name, ts, before)
	assert.Equalf(t, time.UTC, ts.Location(), "%s location", name)
}

func assertIndexes(t *testing.T, got []mongo.IndexModel, wantKeys []bson.D, wantUnique []bool) {
	t.Helper()
	require.Len(t, got, len(wantKeys))
	for i := range got {
		assert.Equalf(t, wantKeys[i], got[i].Keys, "index %d keys", i)
		unique := got[i].Options != nil && got[i].Options.Unique != nil && *got[i].Options.Unique
		assert.Equalf(t, wantUnique[i], unique, "index %d unique", i)
	}
}
