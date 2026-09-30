package notifiers

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

// LogNotifier only logs the notification; used in development when no provider is configured.
type LogNotifier struct{}

func NewLogNotifier() notifier.NotifierPort {
	return &LogNotifier{}
}

func (n *LogNotifier) Send(ctx context.Context, notification notifier.Notification) error {
	logger.GetLoggerFromContext(ctx).Info("Notification sent",
		"type", string(notification.Type),
		"phone", notification.Phone,
		"message", notification.Body,
	)
	return nil
}
