package entities

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

func TestNewDeliveryNotification(t *testing.T) {
	ana := &Resident{Name: "Ana", Phone: "111"}

	tests := []struct {
		name             string
		notificationType NotificationType
		delivery         *Delivery
		wantBody         string
		wantPackage      string
	}{
		{
			name:             "arrival with package and urgency",
			notificationType: NotificationTypeDeliveryArrived,
			delivery:         &Delivery{Apartment: "101", PackageType: "caixa", Urgency: "alta"},
			wantBody:         "Olá, Ana! Chegou uma entrega para o apartamento 101 (caixa). Retire na portaria. Urgência: alta.",
			wantPackage:      "caixa",
		},
		{
			name:             "arrival without package",
			notificationType: NotificationTypeDeliveryArrived,
			delivery:         &Delivery{Apartment: "101"},
			wantBody:         "Olá, Ana! Chegou uma entrega para o apartamento 101. Retire na portaria.",
			wantPackage:      "encomenda",
		},
		{
			name:             "pickup",
			notificationType: NotificationTypeDeliveryPickedUp,
			delivery:         &Delivery{Apartment: "101", PackageType: "caixa"},
			wantBody:         "Olá, Ana! A entrega (caixa) do apartamento 101 foi retirada na portaria.",
			wantPackage:      "caixa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDeliveryNotification(tt.notificationType, ana, tt.delivery)

			if got.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", got.Body, tt.wantBody)
			}
			want := []string{"Ana", "101", tt.wantPackage}
			for i := range want {
				if got.Variables[i] != want[i] {
					t.Errorf("Variables = %v, want %v", got.Variables, want)
				}
			}
			if got.Phone != "111" || got.Type != tt.notificationType {
				t.Errorf("unexpected notification %+v", got)
			}
		})
	}
}
