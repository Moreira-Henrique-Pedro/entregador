package watermill

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWatermillMarshalerRoundTrip(t *testing.T) {
	m := NewWatermillMarshaler()
	assert.IsType(t, kafka.DefaultMarshaler{}, m)

	msg := message.NewMessage("uuid-1", []byte(`{"data":{}}`))
	msg.Metadata.Set("EventType", "evt")

	producerMsg, err := m.Marshal("topic", msg)
	require.NoError(t, err)
	assert.Equal(t, "topic", producerMsg.Topic)

	payload, err := producerMsg.Value.Encode()
	require.NoError(t, err)
	assert.Equal(t, msg.Payload, message.Payload(payload))
}
