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

// ErrInvalidRecipient is returned when the provider rejects the recipient (invalid
// or unreachable phone); retrying the same notification will not succeed.
var ErrInvalidRecipient = errors.New("invalid notification recipient")

type Notification struct {
	Type  NotificationType
	Phone string
	// Body is the free-form text, used when the provider has no template for Type.
	Body string
	// Variables fill the provider template placeholders, in order ({{1}}, {{2}}, ...).
	Variables []string
}

type NotifierPort interface {
	Send(ctx context.Context, notification Notification) error
}
