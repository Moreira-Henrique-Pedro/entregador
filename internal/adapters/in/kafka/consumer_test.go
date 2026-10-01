package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	pubsubmocks "github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub/mocks"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
)

type testPayload struct {
	ID string `json:"id"`
}

func TestConsumer_HandleMessage(t *testing.T) {
	failure := errors.New("boom")

	tests := []struct {
		name       string
		payload    string
		handlerErr error
		wantCalls  int
		wantDLQ    bool
		dlqErr     error
		wantAck    bool
	}{
		{name: "processed message is acked", payload: `{"data":{"id":"1"}}`, wantCalls: 1, wantAck: true},
		{name: "failure is retried then sent to the DLQ", payload: `{"data":{"id":"1"}}`, handlerErr: failure, wantCalls: 3, wantDLQ: true, wantAck: true},
		{name: "permanent failure goes to the DLQ without retry", payload: `{"data":{"id":"1"}}`, handlerErr: pkgEvents.Permanent(failure), wantCalls: 1, wantDLQ: true, wantAck: true},
		{name: "invalid JSON goes to the DLQ without calling the handler", payload: `not json`, wantDLQ: true, wantAck: true},
		{name: "DLQ failure nacks the message", payload: `{"data":{"id":"1"}}`, handlerErr: failure, wantCalls: 3, wantDLQ: true, dlqErr: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			registry := pkgEvents.NewEventHandlerRegistry()
			register(registry, "Test", func(_ context.Context, payload *testPayload) error {
				calls++
				assert.Equal(t, "1", payload.ID)
				return tt.handlerErr
			})

			dlq := pubsubmocks.NewMessagePublisher[any](t)
			if tt.wantDLQ {
				dlq.EXPECT().Publish(mock.Anything, "test.dlq", mock.Anything).
					Run(func(_ context.Context, _ string, messages ...*pubsub.Message[any]) {
						require.Len(t, messages, 1)
						require.NotNil(t, messages[0].Headers.OriginalTopic)
						assert.Equal(t, "test.topic", *messages[0].Headers.OriginalTopic)
					}).
					Return(tt.dlqErr).Once()
			}

			consumer := NewConsumer(nil, dlq, registry, ConsumerConfig{
				Topic:    "test.topic",
				DLQTopic: "test.dlq",
				Timeout:  time.Second,
				Retry:    RetryPolicy{MaxRetries: 2, InitialInterval: time.Millisecond, Multiplier: 1},
			}, logger.NewNoopLogger())

			msg := watermillMessage.NewMessage("uuid-1", []byte(tt.payload))
			msg.Metadata.Set(pubsub.EventTypeHeader, "Test")

			consumer.handleMessage(context.Background(), msg)

			assert.Equal(t, tt.wantCalls, calls)
			if tt.wantAck {
				assertClosed(t, msg.Acked(), "acked")
			} else {
				assertClosed(t, msg.Nacked(), "nacked")
			}
		})
	}
}

func assertClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	default:
		t.Fatalf("message was not %s", what)
	}
}
