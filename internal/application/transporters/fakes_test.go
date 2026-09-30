package transporters

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
)

type publishCall struct {
	topic    string
	messages []*pubsub.Message[any]
}

type fakePublisher struct {
	calls []publishCall
	err   error
}

func (f *fakePublisher) Publish(_ context.Context, topic string, messages ...*pubsub.Message[any]) error {
	f.calls = append(f.calls, publishCall{topic: topic, messages: messages})
	return f.err
}

func (f *fakePublisher) Close(context.Context) error { return nil }
