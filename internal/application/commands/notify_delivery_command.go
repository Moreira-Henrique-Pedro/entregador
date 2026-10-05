package commands

import (
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/google/uuid"
)

const (
	AttributeEventType = "EventType"
	AttributeKey       = "Key"
	AttributeSource    = "Source"
)

const NotifyDeliveryCommandType = "NotifyDelivery"

type NotifyDeliveryCommand struct {
	CommandID        string                    `json:"command_id"`
	DeliveryID       string                    `json:"delivery_id" binding:"required"`
	NotificationType entities.NotificationType `json:"notification_type"`
}

func NewNotifyDeliveryCommand(deliveryID string, notificationType entities.NotificationType) *NotifyDeliveryCommand {
	return &NotifyDeliveryCommand{
		CommandID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(string(notificationType)+"/"+deliveryID)).String(),
		DeliveryID:       deliveryID,
		NotificationType: notificationType,
	}
}
