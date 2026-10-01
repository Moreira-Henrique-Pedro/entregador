package transporters

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testInternalTopic = "delivery.internal"
	testSourceTopic   = "delivery.source"
)

type handleFunc func(ctx context.Context, pub pubsub.MessagePublisher[any]) error

func expectPublish(t *testing.T, err error) (*mocks.MessagePublisher[any], **pubsub.Message[any]) {
	t.Helper()
	pub := mocks.NewMessagePublisher[any](t)
	var published *pubsub.Message[any]
	pub.EXPECT().Publish(mock.Anything, testInternalTopic, mock.Anything).
		Run(func(_ context.Context, _ string, messages ...*pubsub.Message[any]) {
			published = messages[0]
		}).
		Return(err).
		Once()
	return pub, &published
}

func publishMessage(t *testing.T, ctx context.Context, handle handleFunc) *pubsub.Message[any] {
	t.Helper()
	pub, published := expectPublish(t, nil)
	require.NoError(t, handle(ctx, pub))
	require.NotNil(t, *published)
	return *published
}

func commandOf[C any](t *testing.T, msg *pubsub.Message[any]) *C {
	t.Helper()
	command, ok := msg.Payload.Data.(*C)
	require.Truef(t, ok, "payload type = %T, want *%T", msg.Payload.Data, *new(C))
	return command
}

func assertHeaders(t *testing.T, msg *pubsub.Message[any], wantType, wantKey string) {
	t.Helper()
	assert.Equal(t, wantType, msg.Headers.EventType)
	assert.Equal(t, wantKey, msg.Headers.Key)
	assert.Equal(t, testSourceTopic, msg.Headers.Source)
	assert.Nil(t, msg.Headers.OriginalTopic)
}

func assertCommandID[C any](t *testing.T, handle handleFunc, idOf func(*C) string) {
	t.Helper()
	commandID := func(ctx context.Context) string {
		return idOf(commandOf[C](t, publishMessage(t, ctx, handle)))
	}

	t.Run("deterministic with source message id", func(t *testing.T) {
		ctx := pubsub.ContextWithSourceMessageID(context.Background(), "topic-0-42")
		first := commandID(ctx)
		assert.Equal(t, first, commandID(ctx), "redelivery produced a different CommandID")
		assert.Equal(t, newCommandID(ctx), first)
		other := commandID(pubsub.ContextWithSourceMessageID(context.Background(), "topic-0-43"))
		assert.NotEqual(t, first, other, "different source messages produced the same CommandID")
	})

	t.Run("random without source message id", func(t *testing.T) {
		first, second := commandID(context.Background()), commandID(context.Background())
		require.NotEmpty(t, first)
		assert.NotEqual(t, first, second, "CommandID without source message id must not repeat")
	})
}

func assertPublishError(t *testing.T, handle handleFunc) {
	t.Helper()
	errBroker := errors.New("broker down")
	pub, _ := expectPublish(t, errBroker)
	require.ErrorIs(t, handle(context.Background(), pub), errBroker)
}
