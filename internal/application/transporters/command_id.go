package transporters

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/google/uuid"
)

var commandIDNamespace = uuid.MustParse("6f1c2a4e-3b8d-4c5e-9a7f-2d1e0b9c8a76")

// Deterministic per source message: a redelivered event yields the same command, deduplicated by the writers.
func newCommandID(ctx context.Context) string {
	if sourceID, ok := pubsub.SourceMessageIDFromContext(ctx); ok {
		return uuid.NewSHA1(commandIDNamespace, []byte(sourceID)).String()
	}
	return uuid.New().String()
}
