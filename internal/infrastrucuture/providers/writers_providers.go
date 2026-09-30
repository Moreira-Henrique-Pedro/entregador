package providers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/writers"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
	"go.mongodb.org/mongo-driver/mongo"
)

type WriterProviders struct {
	Registry    *pkgEvents.EventHandlerRegistry
	mongoClient *mongo.Client
}

func NewWriterProviders(env *config.Environment, serviceProviders *ServiceProviders) (*WriterProviders, error) {
	deliveryNotifier, err := newNotifier(env)
	if err != nil {
		return nil, fmt.Errorf("create notifier: %w", err)
	}

	repos, err := newRepositoryProviders(env)
	if err != nil {
		return nil, err
	}

	processCreateResidentWriter := writers.NewProcessCreateResident(repos.residentRepository)
	processUpdateResidentWriter := writers.NewProcessUpdateResident(repos.residentRepository)
	processDeleteResidentWriter := writers.NewProcessDeleteResident(repos.residentRepository)
	processCreateDeliveryWriter := writers.NewProcessCreateDelivery(
		repos.deliveryRepository,
		repos.residentRepository,
		serviceProviders.MessagePublisher,
		deliveryInternalCommands,
	)
	processNotifyDeliveryWriter := writers.NewProcessNotifyDelivery(
		repos.deliveryRepository,
		repos.residentRepository,
		deliveryNotifier,
	)
	processDeleteDeliveryWriter := writers.NewProcessDeleteDelivery(
		repos.deliveryRepository,
		serviceProviders.MessagePublisher,
		deliveryInternalCommands,
	)

	registry := pkgEvents.NewEventHandlerRegistry()
	registerWriter(registry, commands.ProcessCreateResidentCommandType, processCreateResidentWriter.Handle)
	registerWriter(registry, commands.ProcessUpdateResidentCommandType, processUpdateResidentWriter.Handle)
	registerWriter(registry, commands.ProcessDeleteResidentCommandType, processDeleteResidentWriter.Handle)
	registerWriter(registry, commands.ProcessCreateDeliveryCommandType, processCreateDeliveryWriter.Handle)
	registerWriter(registry, commands.ProcessDeleteDeliveryCommandType, processDeleteDeliveryWriter.Handle)
	registerWriter(registry, commands.ProcessNotifyDeliveryCommandType, processNotifyDeliveryWriter.Handle)

	return &WriterProviders{
		Registry:    registry,
		mongoClient: repos.mongoClient,
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
