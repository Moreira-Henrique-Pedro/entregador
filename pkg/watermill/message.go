package watermill

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
)

func ConvertPubsubToWatermill[T any](pubsubMessage *pubsub.Message[T], logger appLogger.Logger) (*message.Message, error) {
	payloadBytes, err := json.Marshal(pubsubMessage.Payload)
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(uuid.New().String(), payloadBytes)
	msg.Metadata.Set(pubsub.EventTypeHeader, pubsubMessage.Headers.EventType)
	if pubsubMessage.Headers.Key != "" {
		msg.Metadata.Set(pubsub.KeyHeader, pubsubMessage.Headers.Key)
	}
	msg.Metadata.Set(pubsub.SourceHeader, pubsubMessage.Headers.Source)
	if pubsubMessage.Headers.OriginalTopic != nil {
		msg.Metadata.Set(pubsub.OriginalTopicHeader, *pubsubMessage.Headers.OriginalTopic)
	}

	return msg, nil
}

func ConvertWatermillToPubsub(msg *message.Message, err *error) (*pubsub.Message[any], error) {
	var rawData any
	if unmarshalErr := json.Unmarshal(msg.Payload, &rawData); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	data := extractDataFromPayload(rawData)
	headers := pubsub.Headers{
		EventType: msg.Metadata.Get(pubsub.EventTypeHeader),
		Key:       msg.Metadata.Get(pubsub.KeyHeader),
		Source:    msg.Metadata.Get(pubsub.SourceHeader),
	}

	if originalTopic := msg.Metadata.Get(pubsub.OriginalTopicHeader); originalTopic != "" {
		headers.OriginalTopic = &originalTopic
	}

	convertedMessage := pubsub.NewMessage(msg.Context(), headers, data)
	return convertedMessage, nil
}

func BuildRawDLQMessage(msg *message.Message, handlerErr, convertErr error) *pubsub.Message[any] {
	headers := pubsub.Headers{
		EventType: msg.Metadata.Get(pubsub.EventTypeHeader),
		Key:       msg.Metadata.Get(pubsub.KeyHeader),
		Source:    msg.Metadata.Get(pubsub.SourceHeader),
	}

	if originalTopic := msg.Metadata.Get(pubsub.OriginalTopicHeader); originalTopic != "" {
		headers.OriginalTopic = &originalTopic
	}

	rawData := map[string]any{
		"raw_payload_base64": base64.StdEncoding.EncodeToString(msg.Payload),
		"raw_payload_string": string(msg.Payload),
		"message_uuid":       msg.UUID,
		"error":              fmt.Sprintf("%v (payload convert error: %v)", handlerErr, convertErr),
	}

	return pubsub.NewMessage[any](msg.Context(), headers, rawData)
}

// IsValidJSONPayload reports whether the message payload can be parsed as JSON.
func IsValidJSONPayload(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	return json.Valid(payload)
}

func extractDataFromPayload(rawData any) any {
	payloadMap, ok := rawData.(map[string]any)
	if !ok {
		return rawData
	}

	data, exists := payloadMap["data"]
	if !exists {
		return rawData
	}

	return data
}
