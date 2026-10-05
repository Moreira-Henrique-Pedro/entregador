package entities

import "testing"

func TestDeliveryStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status DeliveryStatus
		want   bool
	}{
		{name: "pending", status: DeliveryStatusPending, want: true},
		{name: "deleted", status: DeliveryStatusDeleted, want: true},
		{name: "empty", status: "", want: false},
		{name: "unknown", status: "lost", want: false},
		{name: "case sensitive", status: "Pending", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("DeliveryStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
