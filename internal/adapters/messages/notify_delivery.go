// Package messages holds the Kafka message contracts shared by the inbound and outbound Kafka adapters.
package messages

import "github.com/Moreira-Henrique-Pedro/entregador/internal/domain"

const NotifyDeliveryType = "NotifyDelivery"

type NotifyDelivery struct {
	CommandID        string                  `json:"command_id"`
	DeliveryID       string                  `json:"delivery_id"`
	NotificationType domain.NotificationType `json:"notification_type"`
}
