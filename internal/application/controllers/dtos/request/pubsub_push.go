package request

import (
	"encoding/json"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
)

type PubSubPush struct {
	Message         PubSubMessage `json:"message"`
	Subscription    string        `json:"subscription"`
	DeliveryAttempt int           `json:"deliveryAttempt"`
}

type PubSubMessage struct {
	Data       []byte            `json:"data"`
	Attributes map[string]string `json:"attributes"`
	MessageID  string            `json:"messageId"`
}

func (p *PubSubPush) EventType() string {
	return p.Message.Attributes[commands.AttributeEventType]
}

func (p *PubSubPush) Key() string {
	return p.Message.Attributes[commands.AttributeKey]
}

func (p *PubSubPush) DecodeData(target any) error {
	return json.Unmarshal(p.Message.Data, target)
}
