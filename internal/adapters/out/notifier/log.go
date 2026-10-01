package notifier

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type Log struct{}

func NewLog() out.Notifier {
	return &Log{}
}

func (n *Log) Send(ctx context.Context, notification out.Notification) error {
	logger.GetLoggerFromContext(ctx).Info("Notification sent",
		"type", string(notification.Type),
		"phone", notification.Phone,
		"message", notification.Body,
	)
	return nil
}
