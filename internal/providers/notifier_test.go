package bootstrap

import (
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
)

func TestNewNotifier(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		twilio   bool
		wantErr  bool
	}{
		{name: "log provider", provider: "log"},
		{name: "twilio provider with credentials", provider: "twilio", twilio: true},
		{name: "twilio provider without credentials", provider: "twilio", wantErr: true},
		{name: "unknown provider", provider: "carrier-pigeon", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &config.Environment{}
			env.Notifier.Provider = tt.provider
			if tt.twilio {
				env.Twilio.AccountSID = "AC123"
				env.Twilio.AuthToken = "secret"
				env.Twilio.WhatsAppFrom = "+14155238886"
			}

			got, err := newNotifier(env)

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Error("expected a notifier")
			}
		})
	}
}
