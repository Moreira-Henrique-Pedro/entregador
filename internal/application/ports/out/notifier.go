package out

import (
	"context"
	"errors"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

var ErrInvalidRecipient = errors.New("invalid notification recipient")

type Notification struct {
	Type      domain.NotificationType
	Phone     string
	Body      string
	Variables []string
}

// Notifier sends a notification to a resident right away (WhatsApp, log…).
type Notifier interface {
	Send(ctx context.Context, notification Notification) error
}

// NotificationScheduler queues a delivery notification to be sent asynchronously.
type NotificationScheduler interface {
	Schedule(ctx context.Context, deliveryID string, notificationType domain.NotificationType) error
}
