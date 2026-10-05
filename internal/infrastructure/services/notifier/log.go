package notifier

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type Log struct{}

func NewLog() services.Notifier {
	return &Log{}
}

func (n *Log) Send(ctx context.Context, notification entities.Notification) error {
	logger.GetLoggerFromContext(ctx).Info("Notification sent",
		"type", string(notification.Type),
		"phone", notification.Phone,
		"message", notification.Body,
	)
	return nil
}
