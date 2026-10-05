package services

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
)

type Notifier interface {
	Send(ctx context.Context, notification entities.Notification) error
}

type NotificationScheduler interface {
	Schedule(ctx context.Context, deliveryID string, notificationType entities.NotificationType) error
}
