package transporters

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
)

func TestNewCommandID(t *testing.T) {
	ctxA := pubsub.ContextWithSourceMessageID(context.Background(), "msg-a")
	ctxB := pubsub.ContextWithSourceMessageID(context.Background(), "msg-b")

	first, redelivered := newCommandID(ctxA), newCommandID(ctxA)
	if first != redelivered {
		t.Error("same source message must produce the same command id")
	}
	if first == newCommandID(ctxB) {
		t.Error("different source messages must produce different command ids")
	}
	randomA, randomB := newCommandID(context.Background()), newCommandID(context.Background())
	if randomA == randomB {
		t.Error("without a source message the command id must be random")
	}
}
