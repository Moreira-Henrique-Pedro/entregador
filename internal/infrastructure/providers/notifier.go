package providers

import (
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/services/notifier"
)

const (
	notifierProviderLog    = "log"
	notifierProviderTwilio = "twilio"
)

func newNotifier(env *config.Environment) (services.Notifier, error) {
	switch env.Notifier.Provider {
	case notifierProviderLog:
		return notifier.NewLog(), nil
	case notifierProviderTwilio:
		return notifier.NewTwilioWhatsApp(notifier.TwilioConfig{
			AccountSID: env.Twilio.AccountSID,
			AuthToken:  env.Twilio.AuthToken,
			From:       env.Twilio.WhatsAppFrom,
			ContentSIDs: map[entities.NotificationType]string{
				entities.NotificationTypeDeliveryArrived:  env.Twilio.ContentSIDDeliveryArrived,
				entities.NotificationTypeDeliveryPickedUp: env.Twilio.ContentSIDDeliveryPickedUp,
			},
			DefaultCountryCode: env.Notifier.DefaultCountryCode,
			BaseURL:            env.Twilio.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unknown notifier provider %q (use %q or %q)", env.Notifier.Provider, notifierProviderLog, notifierProviderTwilio)
	}
}
