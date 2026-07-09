package providers

import (
	"context"
	"reflect"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	subscriberConfig "github.com/Moreira-Henrique-Pedro/entregador/config/subscriber"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/events"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/transporters"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
)

const (
	deliveryInternalCommands = "delivery-internal.commands"
)

type TransporterProviders struct {
	Registry *pkgEvents.EventHandlerRegistry
}

func NewTransporterProviders(
	env *config.Environment,
	subscriberCfg *subscriberConfig.SubscriberConfig,
	serviceProviders *ServiceProviders,
) (*TransporterProviders, error) {
	publisher := serviceProviders.MessagePublisher

	residentTransporter := transporters.NewCreateResidentTransporter(
		publisher,
		deliveryInternalCommands,
		subscriberCfg.Topic,
	)

	registry := pkgEvents.NewEventHandlerRegistry()

	register(registry, events.CreateResidentEventType, residentTransporter.Handle)

	return &TransporterProviders{
		Registry: registry,
	}, nil
}

func register[T any](
	registry *pkgEvents.EventHandlerRegistry,
	eventType string,
	handlerFunc func(context.Context, *T) error,
) {
	var zero T

	registry.RegisterHandler(
		eventType,
		func(ctx context.Context, payload any) error {
			return handlerFunc(ctx, payload.(*T))
		},
		reflect.TypeOf(zero),
	)
}
