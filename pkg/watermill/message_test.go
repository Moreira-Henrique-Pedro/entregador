package watermill

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type samplePayload struct {
	ID string `json:"id"`
}

func TestConvertPubsubToWatermill(t *testing.T) {
	topic := "orig.topic"

	tests := []struct {
		name        string
		headers     pubsub.Headers
		wantKey     bool
		wantOrigTop bool
	}{
		{
			name:        "all headers",
			headers:     pubsub.Headers{EventType: "evt", Key: "k-1", Source: "src", OriginalTopic: &topic},
			wantKey:     true,
			wantOrigTop: true,
		},
		{
			name:    "without key and original topic",
			headers: pubsub.Headers{EventType: "evt", Source: "src"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := pubsub.NewMessage(context.Background(), tt.headers, samplePayload{ID: "42"})

			msg, err := ConvertPubsubToWatermill(in, logger.NewNoopLogger())

			require.NoError(t, err)
			assert.NotEmpty(t, msg.UUID)
			assert.JSONEq(t, `{"data":{"id":"42"}}`, string(msg.Payload))
			assert.Equal(t, "evt", msg.Metadata.Get(pubsub.EventTypeHeader))
			assert.Equal(t, "src", msg.Metadata.Get(pubsub.SourceHeader))

			_, hasKey := msg.Metadata[pubsub.KeyHeader]
			assert.Equal(t, tt.wantKey, hasKey)
			_, hasTopic := msg.Metadata[pubsub.OriginalTopicHeader]
			assert.Equal(t, tt.wantOrigTop, hasTopic)
			if tt.wantKey {
				assert.Equal(t, "k-1", msg.Metadata.Get(pubsub.KeyHeader))
			}
			if tt.wantOrigTop {
				assert.Equal(t, topic, msg.Metadata.Get(pubsub.OriginalTopicHeader))
			}
		})
	}
}

func TestConvertPubsubToWatermillMarshalError(t *testing.T) {
	in := pubsub.NewMessage[any](context.Background(), pubsub.Headers{EventType: "evt"}, make(chan int))

	msg, err := ConvertPubsubToWatermill(in, logger.NewNoopLogger())

	assert.Nil(t, msg)
	assert.Error(t, err)
}

func TestConvertWatermillToPubsub(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		metadata message.Metadata
		wantData any
		wantTop  *string
	}{
		{
			name:     "data envelope",
			payload:  `{"data":{"id":"42"}}`,
			metadata: message.Metadata{pubsub.EventTypeHeader: "evt", pubsub.KeyHeader: "k", pubsub.SourceHeader: "src"},
			wantData: map[string]any{"id": "42"},
		},
		{
			name:     "object without envelope",
			payload:  `{"id":"42"}`,
			metadata: message.Metadata{pubsub.EventTypeHeader: "evt", pubsub.OriginalTopicHeader: "orig.topic"},
			wantData: map[string]any{"id": "42"},
			wantTop:  ptr("orig.topic"),
		},
		{
			name:     "non object payload",
			payload:  `[1,2]`,
			metadata: message.Metadata{},
			wantData: []any{float64(1), float64(2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := message.NewMessage("uuid-1", []byte(tt.payload))
			msg.Metadata = tt.metadata

			out, err := ConvertWatermillToPubsub(msg, nil)

			require.NoError(t, err)
			assert.Equal(t, tt.wantData, out.Payload.Data)
			assert.Equal(t, tt.metadata.Get(pubsub.EventTypeHeader), out.Headers.EventType)
			assert.Equal(t, tt.metadata.Get(pubsub.KeyHeader), out.Headers.Key)
			assert.Equal(t, tt.metadata.Get(pubsub.SourceHeader), out.Headers.Source)
			assert.Equal(t, tt.wantTop, out.Headers.OriginalTopic)
		})
	}
}

func TestConvertWatermillToPubsubInvalidJSON(t *testing.T) {
	msg := message.NewMessage("uuid-1", []byte(`{not json`))

	out, err := ConvertWatermillToPubsub(msg, nil)

	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestBuildRawDLQMessage(t *testing.T) {
	raw := []byte(`{broken`)
	msg := message.NewMessage("uuid-1", raw)
	msg.Metadata = message.Metadata{
		pubsub.EventTypeHeader:     "evt",
		pubsub.KeyHeader:           "k",
		pubsub.SourceHeader:        "src",
		pubsub.OriginalTopicHeader: "orig.topic",
	}

	out := BuildRawDLQMessage(msg, errors.New("cannot parse"))

	require.NotNil(t, out)
	assert.Equal(t, "evt", out.Headers.EventType)
	assert.Equal(t, "k", out.Headers.Key)
	assert.Equal(t, "src", out.Headers.Source)
	require.NotNil(t, out.Headers.OriginalTopic)
	assert.Equal(t, "orig.topic", *out.Headers.OriginalTopic)

	data, ok := out.Payload.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, base64.StdEncoding.EncodeToString(raw), data["raw_payload_base64"])
	assert.Equal(t, string(raw), data["raw_payload_string"])
	assert.Equal(t, "uuid-1", data["message_uuid"])
	assert.Equal(t, "cannot parse", data["error"])
}

func TestBuildRawDLQMessageWithoutOriginalTopic(t *testing.T) {
	msg := message.NewMessage("uuid-2", nil)

	out := BuildRawDLQMessage(msg, nil)

	assert.Nil(t, out.Headers.OriginalTopic)
	data, ok := out.Payload.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "", data["raw_payload_string"])
	assert.Equal(t, "<nil>", data["error"])
}

func TestIsValidJSONPayload(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    bool
	}{
		{name: "nil", payload: nil, want: false},
		{name: "empty", payload: []byte{}, want: false},
		{name: "object", payload: []byte(`{"a":1}`), want: true},
		{name: "array", payload: []byte(`[1]`), want: true},
		{name: "string", payload: []byte(`"x"`), want: true},
		{name: "broken", payload: []byte(`{"a":`), want: false},
		{name: "plain text", payload: []byte(`hello`), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidJSONPayload(tt.payload))
		})
	}
}

func ptr(s string) *string { return &s }
