package bootstrap

import (
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/out/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

const (
	notifierProviderLog    = "log"
	notifierProviderTwilio = "twilio"
)

func newNotifier(env *config.Environment) (out.Notifier, error) {
	switch env.Notifier.Provider {
	case notifierProviderLog:
		return notifier.NewLog(), nil
	case notifierProviderTwilio:
		return notifier.NewTwilioWhatsApp(notifier.TwilioConfig{
			AccountSID: env.Twilio.AccountSID,
			AuthToken:  env.Twilio.AuthToken,
			From:       env.Twilio.WhatsAppFrom,
			ContentSIDs: map[domain.NotificationType]string{
				domain.NotificationTypeDeliveryArrived:  env.Twilio.ContentSIDDeliveryArrived,
				domain.NotificationTypeDeliveryPickedUp: env.Twilio.ContentSIDDeliveryPickedUp,
			},
			DefaultCountryCode: env.Notifier.DefaultCountryCode,
			BaseURL:            env.Twilio.BaseURL,
		}, nil)
	default:
		return nil, fmt.Errorf("unknown notifier provider %q (use %q or %q)", env.Notifier.Provider, notifierProviderLog, notifierProviderTwilio)
	}
}
