package commands

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"

const (
	ProcessNotifyDeliveryCommandType = "ProcessNotifyDelivery"
)

type ProcessNotifyDeliveryCommand struct {
	CommandID        string                    `json:"command_id"`
	DeliveryID       string                    `json:"delivery_id"`
	NotificationType notifier.NotificationType `json:"notification_type"`
}
