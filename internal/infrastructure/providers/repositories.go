package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/repositories/mongodb"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/repositories/mongodb/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	residentsCollectionName  = "residents"
	deliveriesCollectionName = "deliveries"
)

type mongoRepositories struct {
	mongoClient        *mongo.Client
	residentRepository repositories.ResidentRepository
	deliveryRepository repositories.DeliveryRepository
}

func newMongoRepositories(env *config.Environment) (*mongoRepositories, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(env.MongoDB.URI))
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}

	database := mongoClient.Database(env.MongoDB.Database)

	residentRepository, err := mongodb.NewResidentRepository(
		ctx,
		client.NewMongoCollectionClient(database.Collection(residentsCollectionName)),
	)
	if err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}

	deliveryRepository, err := mongodb.NewDeliveryRepository(
		ctx,
		client.NewMongoCollectionClient(database.Collection(deliveriesCollectionName)),
	)
	if err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}

	return &mongoRepositories{
		mongoClient:        mongoClient,
		residentRepository: residentRepository,
		deliveryRepository: deliveryRepository,
	}, nil
}

func (r *mongoRepositories) close(ctx context.Context) error {
	return r.mongoClient.Disconnect(ctx)
}
