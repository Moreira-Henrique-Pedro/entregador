package providers

import (
	"fmt"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/notifier"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/notifiers"
)

const (
	notifierProviderLog    = "log"
	notifierProviderTwilio = "twilio"
)

func newNotifier(env *config.Environment) (notifier.NotifierPort, error) {
	switch env.Notifier.Provider {
	case notifierProviderLog:
		return notifiers.NewLogNotifier(), nil
	case notifierProviderTwilio:
		return notifiers.NewTwilioWhatsAppNotifier(notifiers.TwilioConfig{
			AccountSID: env.Twilio.AccountSID,
			AuthToken:  env.Twilio.AuthToken,
			From:       env.Twilio.WhatsAppFrom,
			ContentSIDs: map[notifier.NotificationType]string{
				notifier.NotificationTypeDeliveryArrived:  env.Twilio.ContentSIDDeliveryArrived,
				notifier.NotificationTypeDeliveryPickedUp: env.Twilio.ContentSIDDeliveryPickedUp,
			},
			DefaultCountryCode: env.Notifier.DefaultCountryCode,
			BaseURL:            env.Twilio.BaseURL,
		}, nil)
	default:
		return nil, fmt.Errorf("unknown notifier provider %q (use %q or %q)", env.Notifier.Provider, notifierProviderLog, notifierProviderTwilio)
	}
}
