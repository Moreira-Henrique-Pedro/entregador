package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	interfaces "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/repositories"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories"
	mongodb "github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	residentsCollectionName  = "residents"
	deliveriesCollectionName = "deliveries"
)

type repositoryProviders struct {
	mongoClient        *mongo.Client
	residentRepository interfaces.ResidentRepositoryPort
	deliveryRepository interfaces.DeliveryRepositoryPort
}

func newRepositoryProviders(env *config.Environment) (*repositoryProviders, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(env.MongoDB.URI))
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}

	database := client.Database(env.MongoDB.Database)

	residentRepository, err := repositories.NewMongoDBResidentRepository(
		ctx,
		mongodb.NewMongoCollectionClient(database.Collection(residentsCollectionName)),
	)
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	deliveryRepository, err := repositories.NewMongoDBDeliveryRepository(
		ctx,
		mongodb.NewMongoCollectionClient(database.Collection(deliveriesCollectionName)),
	)
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return &repositoryProviders{
		mongoClient:        client,
		residentRepository: residentRepository,
		deliveryRepository: deliveryRepository,
	}, nil
}
