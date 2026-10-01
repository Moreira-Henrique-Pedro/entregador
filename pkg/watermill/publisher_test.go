package watermill

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill/mocks"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewWatermillPublisherRequiresBrokers(t *testing.T) {
	publisher, err := NewWatermillPublisher[samplePayload](nil, nil)

	assert.Nil(t, publisher)
	assert.Error(t, err)
}

func TestWatermillPublisherPublish(t *testing.T) {
	broker := mocks.NewPublisher(t)
	var published []*message.Message
	broker.EXPECT().Publish("topic", mock.Anything, mock.Anything).
		Run(func(_ string, messages ...*message.Message) { published = messages }).
		Return(nil).Once()
	broker.EXPECT().Close().Return(nil).Once()
	publisher := &WatermillPublisher[samplePayload]{publisher: broker}

	err := publisher.Publish(context.Background(), "topic",
		pubsub.NewMessage(context.Background(), pubsub.Headers{EventType: "a", Key: "1"}, samplePayload{ID: "1"}),
		pubsub.NewMessage(context.Background(), pubsub.Headers{EventType: "b"}, samplePayload{ID: "2"}),
	)

	require.NoError(t, err)
	require.Len(t, published, 2)
	assert.JSONEq(t, `{"data":{"id":"1"}}`, string(published[0].Payload))
	assert.Equal(t, []map[string]string{
		{"eventType": "a", "key": "1"},
		{"eventType": "b", "key": ""},
	}, getEventTypeKeyArray(published))

	require.NoError(t, publisher.Close(context.Background()))
}

func TestWatermillPublisherPublishErrors(t *testing.T) {
	t.Run("broker error", func(t *testing.T) {
		brokerErr := errors.New("broker down")
		broker := mocks.NewPublisher(t)
		broker.EXPECT().Publish("topic", mock.Anything).Return(brokerErr).Once()
		publisher := &WatermillPublisher[samplePayload]{publisher: broker}

		err := publisher.Publish(context.Background(), "topic",
			pubsub.NewMessage(context.Background(), pubsub.Headers{EventType: "a"}, samplePayload{ID: "1"}))

		assert.ErrorIs(t, err, brokerErr)
	})

	t.Run("conversion error does not reach the broker", func(t *testing.T) {

		publisher := &WatermillPublisher[any]{publisher: mocks.NewPublisher(t)}

		err := publisher.Publish(context.Background(), "topic",
			pubsub.NewMessage[any](context.Background(), pubsub.Headers{EventType: "a"}, make(chan int)))

		assert.Error(t, err)
	})
}
