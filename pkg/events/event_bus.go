package events

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type ContextKey string

const CorrelationIDKey ContextKey = "correlation_id"

type EventBus struct {
	eventHandlerRegistry *EventHandlerRegistry
}

type EventBusDependencies struct {
	EventHandlerRegistry *EventHandlerRegistry
}

func NewEventBus(props EventBusDependencies) *EventBus {
	return &EventBus{
		eventHandlerRegistry: props.EventHandlerRegistry,
	}
}

func (e *EventBus) Handle(ctx context.Context, msg *pubsub.Message[any]) error {
	logger := logger.GetLoggerFromContext(ctx)

	if msg == nil || msg.Headers.EventType == "" {
		logger.Debug("Unrecognized event")
		return nil
	}

	handler, err := e.eventHandlerRegistry.GetEventHandlerByEventType(msg.Headers.EventType)
	if handler == nil || err != nil {
		logger.Debug("No Handler registered for this event type")
		return nil
	}

	payload, err := e.processPayload(msg, handler, logger)
	if err != nil {
		logger.Error("Failed to process payload", map[string]any{
			"error":        err.Error(),
			"payload_type": handler.PayloadType.String(),
		})
		return err
	}

	return e.executeHandler(ctx, handler, payload, logger)
}

func (e *EventBus) processPayload(msg *pubsub.Message[any], handler *EventHandler[any], logger logger.Logger) (any, error) {
	payload := reflect.New(handler.PayloadType).Interface()

	dataBytes, err := e.convertToBytes(msg.Payload.Data, logger)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dataBytes, payload); err != nil {
		logger.Error("Error unmarshalling event payload", map[string]any{
			"error":        err.Error(),
			"payload_type": handler.PayloadType.String(),
		})
		return nil, fmt.Errorf("error unmarshalling event payload for type %s: %w", msg.Headers.EventType, err)
	}

	if err := e.validatePayloadType(payload, handler, logger); err != nil {
		return nil, err
	}

	return payload, nil
}

func (e *EventBus) convertToBytes(data any, logger logger.Logger) ([]byte, error) {
	switch d := data.(type) {
	case []byte:
		return d, nil
	case string:
		return []byte(d), nil
	case nil:
		return []byte("{}"), nil
	default:
		dataBytes, err := json.Marshal(d)
		if err != nil {
			logger.Error("Error marshalling payload data to JSON", map[string]any{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("error marshalling payload data to JSON: %w", err)
		}
		return dataBytes, nil
	}
}

func (e *EventBus) validatePayloadType(payload any, handler *EventHandler[any], logger logger.Logger) error {
	payloadValue := reflect.ValueOf(payload)
	if payloadValue.Kind() == reflect.Ptr {
		payloadValue = payloadValue.Elem()
	}

	if !payloadValue.Type().AssignableTo(handler.PayloadType) && payloadValue.Type() != handler.PayloadType {
		logger.Error("Payload type mismatch", map[string]interface{}{
			"expected_type": handler.PayloadType.String(),
			"actual_type":   payloadValue.Type().String(),
		})
		return fmt.Errorf("payload type mismatch: expected %s, got %s", handler.PayloadType, payloadValue.Type())
	}
	return nil
}

func (e *EventBus) executeHandler(ctx context.Context, handler *EventHandler[any], payload interface{}, logger logger.Logger) error {
	logger.Debug("Calling event handler")

	err := handler.Handler(ctx, payload)
	if err != nil {
		logger.Error("Event handler execution failed", map[string]any{
			"error": err.Error(),
		})
		return err
	}

	logger.Debug("Event handler executed successfully")
	return nil
}
