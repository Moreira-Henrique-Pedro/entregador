package kafka

import (
	"context"
	"reflect"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
)

type UseCases struct {
	NotifyDelivery in.NotifyDelivery
}

// NewHandlerRegistry maps each message type of the topic to the use case that handles it.
func NewHandlerRegistry(useCases UseCases) *pkgEvents.EventHandlerRegistry {
	registry := pkgEvents.NewEventHandlerRegistry()

	register(registry, messages.NotifyDeliveryType, func(ctx context.Context, message *messages.NotifyDelivery) error {
		return useCases.NotifyDelivery.Execute(ctx, message.DeliveryID, message.NotificationType)
	})

	return registry
}

func register[T any](
	registry *pkgEvents.EventHandlerRegistry,
	messageType string,
	handlerFunc func(context.Context, *T) error,
) {
	var zero T

	registry.RegisterHandler(
		messageType,
		func(ctx context.Context, payload any) error {
			return handlerFunc(ctx, payload.(*T))
		},
		reflect.TypeOf(zero),
	)
}
