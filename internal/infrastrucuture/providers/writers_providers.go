package providers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/writers"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories"
	mongodb "github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/repositories/client"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const residentsCollectionName = "residents"

type WriterProviders struct {
	Registry    *pkgEvents.EventHandlerRegistry
	mongoClient *mongo.Client
}

func NewWriterProviders(env *config.Environment, serviceProviders *ServiceProviders) (*WriterProviders, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(env.MongoDB.URI))
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}

	collection := client.Database(env.MongoDB.Database).Collection(residentsCollectionName)
	residentRepository := repositories.NewMongoDBResidentRepository(mongodb.NewMongoCollectionClient(collection))
	processCreateResidentWriter := writers.NewProcessCreateResident(residentRepository)

	registry := pkgEvents.NewEventHandlerRegistry()
	registerWriter(registry, commands.ProcessCreateResidentCommandType, processCreateResidentWriter.Handle)

	return &WriterProviders{
		Registry:    registry,
		mongoClient: client,
	}, nil
}

func (w *WriterProviders) Close(ctx context.Context) error {
	if w == nil || w.mongoClient == nil {
		return nil
	}
	return w.mongoClient.Disconnect(ctx)
}

func registerWriter[T any](
	registry *pkgEvents.EventHandlerRegistry,
	commandType string,
	handlerFunc func(context.Context, *T) error,
) {
	var zero T

	registry.RegisterHandler(
		commandType,
		func(ctx context.Context, payload any) error {
			return handlerFunc(ctx, payload.(*T))
		},
		reflect.TypeOf(zero),
	)
}
