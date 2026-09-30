package notifiers

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
)

func TestLogNotifier_NeverFails(t *testing.T) {
	if err := NewLogNotifier().Send(context.Background(), notifier.Notification{Phone: "111", Body: "oi"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
