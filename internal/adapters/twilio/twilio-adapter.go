package twilio

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/dock-colombia/commons-shared/logging"
)

type TwilioService struct {
	client client.TwilioClientPort
}

func NewTwilioPort(client client.TwilioClientPort) ports.TwilioPort {
	return &TwilioService{
		client: client,
	}
}

func (t *TwilioService) SendWhatsAppMessage(ctx context.Context, to string) error {
	log := logging.FromContext(&ctx)
	err := t.client.SendWhatsAppMessage(ctx, to)
	if err != nil {
		log.LogError("Error sending WhatsApp message", logging.Fields{"error": err})
		return err
	}

	return nil
}
