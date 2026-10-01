package pubsub

import "context"

type contextKey string

const sourceMessageIDKey contextKey = "source_message_id"

func ContextWithSourceMessageID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sourceMessageIDKey, id)
}

func SourceMessageIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(sourceMessageIDKey).(string)
	return id, ok && id != ""
}
