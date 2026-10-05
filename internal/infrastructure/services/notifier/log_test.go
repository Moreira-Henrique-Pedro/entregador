package notifier

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

func TestLog_NeverFails(t *testing.T) {
	if err := NewLog().Send(context.Background(), entities.Notification{Phone: "111", Body: "oi"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
