// Package messages holds the Kafka message contracts shared by the inbound and outbound Kafka adapters.
package messages

import (
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/google/uuid"
)

const NotifyDeliveryType = "NotifyDelivery"

type NotifyDelivery struct {
	CommandID        string                  `json:"command_id"`
	DeliveryID       string                  `json:"delivery_id"`
	NotificationType domain.NotificationType `json:"notification_type"`
}

func NewNotifyDelivery(deliveryID string, notificationType domain.NotificationType) *NotifyDelivery {
	return &NotifyDelivery{
		// Deterministic: scheduling the same notification twice yields the same command.
		CommandID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(string(notificationType)+"/"+deliveryID)).String(),
		DeliveryID:       deliveryID,
		NotificationType: notificationType,
	}
}
