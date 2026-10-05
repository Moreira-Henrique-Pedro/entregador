package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPendingDelivery(t *testing.T) {
	input := &Delivery{Apartment: "101", ResidentID: "ignored", PackageType: "caixa", Urgency: "alta"}

	assert.Equal(t, &Delivery{
		ID: "d1", DeliveryID: "d1", Apartment: "101", ResidentID: "ana",
		PackageType: "caixa", Urgency: "alta", Status: DeliveryStatusPending,
	}, NewPendingDelivery("d1", "ana", input))
}

func TestDeliveryNotificationSkipReason(t *testing.T) {
	notifiedAt := time.Now()

	tests := []struct {
		name             string
		delivery         Delivery
		notificationType NotificationType
		want             string
	}{
		{name: "pending arrival", delivery: Delivery{Status: DeliveryStatusPending}, notificationType: NotificationTypeDeliveryArrived},
		{name: "arrival already notified", delivery: Delivery{Status: DeliveryStatusPending, ArrivalNotifiedAt: notifiedAt}, notificationType: NotificationTypeDeliveryArrived, want: "arrival already notified"},
		{name: "arrival after pickup", delivery: Delivery{Status: DeliveryStatusDeleted}, notificationType: NotificationTypeDeliveryArrived, want: "delivery already picked up"},
		{name: "picked up", delivery: Delivery{Status: DeliveryStatusDeleted}, notificationType: NotificationTypeDeliveryPickedUp},
		{name: "pickup already notified", delivery: Delivery{Status: DeliveryStatusDeleted, PickupNotifiedAt: notifiedAt}, notificationType: NotificationTypeDeliveryPickedUp, want: "pickup already notified"},
		{name: "pickup before pickup", delivery: Delivery{Status: DeliveryStatusPending}, notificationType: NotificationTypeDeliveryPickedUp, want: "delivery not picked up yet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.delivery.NotificationSkipReason(tt.notificationType))
		})
	}
}

func TestDeliveryValidations(t *testing.T) {
	pending, lost := DeliveryStatusPending, DeliveryStatus("lost")

	assert.NoError(t, (&Delivery{Apartment: "101"}).ValidateForRegister())
	assert.ErrorIs(t, (&Delivery{}).ValidateForRegister(), ErrInvalidDelivery)

	assert.NoError(t, ValidateDeliveryID("d1"))
	assert.ErrorIs(t, ValidateDeliveryID(""), ErrInvalidDelivery)

	assert.NoError(t, ValidateDeliveryFilter("101", nil))
	assert.NoError(t, ValidateDeliveryFilter("101", &pending))
	assert.ErrorIs(t, ValidateDeliveryFilter("", nil), ErrInvalidDelivery)
	assert.ErrorIs(t, ValidateDeliveryFilter("101", &lost), ErrInvalidDelivery)
}
