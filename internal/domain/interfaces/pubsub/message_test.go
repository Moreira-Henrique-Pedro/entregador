package pubsub

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
)

func TestNewHeaders(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		key       string
	}{
		{name: "event with key", eventType: "ProcessCreateDelivery", key: "101"},
		{name: "empty key", eventType: "ProcessDeleteDelivery", key: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHeaders(tt.eventType, tt.key)
			if got.EventType != tt.eventType {
				t.Errorf("EventType = %q, want %q", got.EventType, tt.eventType)
			}
			if got.Key != tt.key {
				t.Errorf("Key = %q, want %q", got.Key, tt.key)
			}
			if got.Source != config.AppName {
				t.Errorf("Source = %q, want %q", got.Source, config.AppName)
			}
			if got.OriginalTopic != nil {
				t.Errorf("OriginalTopic = %q, want nil", *got.OriginalTopic)
			}
		})
	}
}

func TestNewMessage(t *testing.T) {
	type payload struct{ ID string }

	tests := []struct {
		name    string
		headers Headers
		data    payload
	}{
		{name: "with headers", headers: NewHeaders("SomeEvent", "k1"), data: payload{ID: "p1"}},
		{name: "zero headers", headers: Headers{}, data: payload{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(context.Background(), tt.headers, tt.data)
			if msg.Headers != tt.headers {
				t.Errorf("Headers = %+v, want %+v", msg.Headers, tt.headers)
			}
			if msg.Payload.Data != tt.data {
				t.Errorf("Payload.Data = %+v, want %+v", msg.Payload.Data, tt.data)
			}
			if msg.GetEventType() != tt.headers.EventType {
				t.Errorf("GetEventType() = %q, want %q", msg.GetEventType(), tt.headers.EventType)
			}
		})
	}
}
