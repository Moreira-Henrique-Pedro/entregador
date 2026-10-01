package events

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newMessage(eventType string, data any) *pubsub.Message[any] {
	return pubsub.NewMessage[any](context.Background(), pubsub.Headers{EventType: eventType}, data)
}

func TestEventBusHandleIgnoresUnroutableMessages(t *testing.T) {
	called := false
	registry := NewEventHandlerRegistry()
	registry.RegisterHandler("known", func(ctx context.Context, payload any) error {
		called = true
		return nil
	}, reflect.TypeFor[testPayload]())
	bus := NewEventBus(EventBusDependencies{EventHandlerRegistry: registry})

	tests := []struct {
		name string
		msg  *pubsub.Message[any]
	}{
		{name: "nil message", msg: nil},
		{name: "empty event type", msg: newMessage("", map[string]any{"id": "1"})},
		{name: "unregistered event type", msg: newMessage("unknown", map[string]any{"id": "1"})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, bus.Handle(context.Background(), tt.msg))
		})
	}
	assert.False(t, called)
}

func TestEventBusHandleDecodesPayload(t *testing.T) {
	tests := []struct {
		name string
		data any
		want testPayload
	}{
		{name: "map", data: map[string]any{"id": "1", "name": "box"}, want: testPayload{ID: "1", Name: "box"}},
		{name: "bytes", data: []byte(`{"id":"2","name":"letter"}`), want: testPayload{ID: "2", Name: "letter"}},
		{name: "string", data: `{"id":"3"}`, want: testPayload{ID: "3"}},
		{name: "nil", data: nil, want: testPayload{}},
		{name: "struct", data: testPayload{ID: "4", Name: "bag"}, want: testPayload{ID: "4", Name: "bag"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var received any
			registry := NewEventHandlerRegistry()
			registry.RegisterHandler("evt", func(ctx context.Context, payload any) error {
				received = payload
				return nil
			}, reflect.TypeFor[testPayload]())
			bus := NewEventBus(EventBusDependencies{EventHandlerRegistry: registry})

			require.NoError(t, bus.Handle(context.Background(), newMessage("evt", tt.data)))

			got, ok := received.(*testPayload)
			require.True(t, ok, "handler must receive *testPayload, got %T", received)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestEventBusHandleErrors(t *testing.T) {
	handlerErr := errors.New("boom")

	tests := []struct {
		name       string
		data       any
		handlerErr error
		wantErr    error
		wantMsg    string
		wantCalled bool
	}{
		{name: "invalid json", data: "not-json", wantMsg: "error unmarshalling event payload for type evt"},
		{name: "wrong json shape", data: `[1,2]`, wantMsg: "error unmarshalling event payload for type evt"},
		{name: "unmarshalable data", data: map[string]any{"ch": make(chan int)}, wantMsg: "error marshalling payload data to JSON"},
		{name: "handler error", data: `{"id":"1"}`, handlerErr: handlerErr, wantErr: handlerErr, wantCalled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			registry := NewEventHandlerRegistry()
			registry.RegisterHandler("evt", func(ctx context.Context, payload any) error {
				called = true
				return tt.handlerErr
			}, reflect.TypeFor[testPayload]())
			bus := NewEventBus(EventBusDependencies{EventHandlerRegistry: registry})

			err := bus.Handle(context.Background(), newMessage("evt", tt.data))

			require.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			if tt.wantMsg != "" {
				assert.Contains(t, err.Error(), tt.wantMsg)
			}
			assert.Equal(t, tt.wantCalled, called)
		})
	}
}
