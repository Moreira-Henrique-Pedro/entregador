package pubsub

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
)

const (
	EventTypeHeader     = "EventType"
	KeyHeader           = "Key"
	SourceHeader        = "Source"
	OriginalTopicHeader = "OriginalTopic"
)

type Headers struct {
	EventType     string
	Key           string
	Source        string
	OriginalTopic *string
}

type Payload[T any] struct {
	Data T `json:"data"`
}

type Message[T any] struct {
	Headers Headers
	Payload Payload[T]
}

func NewHeaders(eventType, key string) Headers {
	return Headers{
		EventType: eventType,
		Key:       key,
		Source:    config.AppName,
	}
}

func NewMessage[T any](ctx context.Context, headers Headers, payload T) *Message[T] {
	return &Message[T]{
		Headers: headers,
		Payload: Payload[T]{Data: payload},
	}
}

func (m *Message[T]) GetEventType() string {
	return m.Headers.EventType
}
