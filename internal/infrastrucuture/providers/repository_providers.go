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

	return &repositoryProviders{
		mongoClient: client,
		residentRepository: repositories.NewMongoDBResidentRepository(
			mongodb.NewMongoCollectionClient(database.Collection(residentsCollectionName)),
		),
		deliveryRepository: repositories.NewMongoDBDeliveryRepository(
			mongodb.NewMongoCollectionClient(database.Collection(deliveriesCollectionName)),
		),
	}, nil
}
