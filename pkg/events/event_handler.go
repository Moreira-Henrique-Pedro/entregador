package events

import "context"

type EventHandlerInterface[Event any] interface {
	Handle(ctx context.Context, events *Event) error
}
