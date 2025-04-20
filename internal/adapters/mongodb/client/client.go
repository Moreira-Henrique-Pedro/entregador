package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface para abstrair a collection
type MongoCollectionPort interface {
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) SingleResultPort
	UpdateOne(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

// Interface para o resultado de FindOne
type SingleResultPort interface {
	Decode(val interface{}) error
}

// Implementação do wrapper para Collection
type MongoCollectionAdapter struct {
	collection *mongo.Collection
}

func (mca *MongoCollectionAdapter) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return mca.collection.InsertOne(ctx, document, opts...)
}

func (mca *MongoCollectionAdapter) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) SingleResultPort {
	return mca.collection.FindOne(ctx, filter, opts...)
}

func (mca *MongoCollectionAdapter) UpdateOne(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mca.collection.UpdateOne(ctx, filter, update, opts...)
}

func (mca *MongoCollectionAdapter) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return mca.collection.DeleteOne(ctx, filter, opts...)
}

// Função que cria e retorna o cliente do Mongo
func NewMongoClient(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetMongoCollection(client *mongo.Client, dbName, collectionName string) MongoCollectionPort {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoCollectionAdapter{collection: collection}
}
