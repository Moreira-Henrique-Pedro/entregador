package mongodb

import (
	"context"

	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCollectionClient struct {
	collection *mongo.Collection
}

func NewMongoCollectionClient(collection *mongo.Collection) interfaces.MongoClientCollectionPort {
	return &MongoCollectionClient{collection: collection}
}

func (m *MongoCollectionClient) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.collection.UpdateOne(ctx, filter, update, opts...)
}

func (m *MongoCollectionClient) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.collection.InsertOne(ctx, document, opts...)
}

func (m *MongoCollectionClient) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return m.collection.FindOne(ctx, filter, opts...)
}

func (m *MongoCollectionClient) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return m.collection.Find(ctx, filter, opts...)
}

func (m *MongoCollectionClient) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.collection.DeleteOne(ctx, filter, opts...)
}

func (m *MongoCollectionClient) EnsureUniqueIndex(keys interface{}) error {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true),
	}
	_, err := m.collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}
