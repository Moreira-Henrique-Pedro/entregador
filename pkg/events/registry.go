package events

import (
	"context"
	"fmt"
	"reflect"
)

type EventHandlerRegistry struct {
	EventHandlers map[string]EventHandler[any]
}

type EventHandler[T any] struct {
	Handler     func(ctx context.Context, payload T) error
	PayloadType reflect.Type
}

func NewEventHandlerRegistry() *EventHandlerRegistry {
	return &EventHandlerRegistry{
		EventHandlers: make(map[string]EventHandler[any]),
	}
}

func (e *EventHandlerRegistry) RegisterHandler(eventType string, handler func(ctx context.Context, payload any) error, payloadType reflect.Type) {
	e.EventHandlers[eventType] = EventHandler[any]{
		Handler:     handler,
		PayloadType: payloadType,
	}
}

func (e *EventHandlerRegistry) GetEventHandlerByEventType(eventType string) (*EventHandler[any], error) {
	eventHandler, ok := e.EventHandlers[eventType]
	if !ok {
		return nil, fmt.Errorf("event handler not registered for event type: %s", eventType)
	}
	return &eventHandler, nil
}

func (e *EventHandlerRegistry) GetAllEventTypes() []string {
	eventTypes := make([]string, 0, len(e.EventHandlers))
	for eventType := range e.EventHandlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}
