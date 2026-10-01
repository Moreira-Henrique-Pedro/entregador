package notifier

import (
	"context"
	"errors"
)

type NotificationType string

const (
	NotificationTypeDeliveryArrived  NotificationType = "delivery_arrived"
	NotificationTypeDeliveryPickedUp NotificationType = "delivery_picked_up"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeDeliveryArrived, NotificationTypeDeliveryPickedUp:
		return true
	}
	return false
}

var ErrInvalidRecipient = errors.New("invalid notification recipient")

type Notification struct {
	Type      NotificationType
	Phone     string
	Body      string
	Variables []string
}

type NotifierPort interface {
	Send(ctx context.Context, notification Notification) error
}
