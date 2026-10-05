package entities

import "fmt"

const defaultPackageLabel = "encomenda"

type Notification struct {
	Type      NotificationType
	Phone     string
	Body      string
	Variables []string
}

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

func NewDeliveryNotification(notificationType NotificationType, resident *Resident, delivery *Delivery) Notification {
	return Notification{
		Type:      notificationType,
		Phone:     resident.Phone,
		Body:      deliveryMessageBody(notificationType, resident, delivery),
		Variables: []string{resident.Name, delivery.Apartment, packageLabel(delivery)},
	}
}

func packageLabel(delivery *Delivery) string {
	if delivery.PackageType == "" {
		return defaultPackageLabel
	}
	return delivery.PackageType
}

func deliveryMessageBody(notificationType NotificationType, resident *Resident, delivery *Delivery) string {
	if notificationType == NotificationTypeDeliveryPickedUp {
		return fmt.Sprintf("Olá, %s! A entrega%s do apartamento %s foi retirada na portaria.", resident.Name, packageSuffix(delivery), delivery.Apartment)
	}
	return fmt.Sprintf("Olá, %s! Chegou uma entrega para o apartamento %s%s. Retire na portaria.%s", resident.Name, delivery.Apartment, packageSuffix(delivery), urgencySuffix(delivery))
}

func packageSuffix(delivery *Delivery) string {
	if delivery.PackageType == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", delivery.PackageType)
}

func urgencySuffix(delivery *Delivery) string {
	if delivery.Urgency == "" {
		return ""
	}
	return fmt.Sprintf(" Urgência: %s.", delivery.Urgency)
}
