package notifier

import "testing"

func TestNotificationTypeIsValid(t *testing.T) {
	tests := []struct {
		name string
		typ  NotificationType
		want bool
	}{
		{name: "delivery arrived", typ: NotificationTypeDeliveryArrived, want: true},
		{name: "delivery picked up", typ: NotificationTypeDeliveryPickedUp, want: true},
		{name: "empty", typ: "", want: false},
		{name: "unknown", typ: "delivery_lost", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.IsValid(); got != tt.want {
				t.Errorf("NotificationType(%q).IsValid() = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}
