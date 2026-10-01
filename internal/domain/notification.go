package domain

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
