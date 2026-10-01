package events

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventHandlerRegistry(t *testing.T) {
	registry := NewEventHandlerRegistry()
	assert.Empty(t, registry.GetAllEventTypes())

	sentinel := errors.New("called")
	registry.RegisterHandler("a", func(ctx context.Context, payload any) error { return sentinel }, reflect.TypeFor[testPayload]())
	registry.RegisterHandler("b", func(ctx context.Context, payload any) error { return nil }, reflect.TypeFor[string]())

	handler, err := registry.GetEventHandlerByEventType("a")
	require.NoError(t, err)
	require.NotNil(t, handler)
	assert.Equal(t, reflect.TypeFor[testPayload](), handler.PayloadType)
	assert.ErrorIs(t, handler.Handler(context.Background(), nil), sentinel)

	handler, err = registry.GetEventHandlerByEventType("missing")
	assert.Nil(t, handler)
	assert.EqualError(t, err, "event handler not registered for event type: missing")

	assert.ElementsMatch(t, []string{"a", "b"}, registry.GetAllEventTypes())

	registry.RegisterHandler("a", func(ctx context.Context, payload any) error { return nil }, reflect.TypeFor[int]())
	handler, err = registry.GetEventHandlerByEventType("a")
	require.NoError(t, err)
	assert.Equal(t, reflect.TypeFor[int](), handler.PayloadType)
	assert.Len(t, registry.GetAllEventTypes(), 2)
}
