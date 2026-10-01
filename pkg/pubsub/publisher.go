package pubsub

import (
	"context"
)

type MessagePublisher[T any] interface {
	Publish(ctx context.Context, topic string, messages ...*Message[T]) error
	Close(ctx context.Context) error
}
