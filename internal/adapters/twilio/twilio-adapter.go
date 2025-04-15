package twilio

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
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
	logger := logrus.New()
	logger.Info("Sending WhatsApp message to: ", to)
	// Call the Twilio client to send the message
	err := t.client.SendWhatsAppMessage(ctx, to)
	if err != nil {
		logger.Error("Failed to send WhatsApp message: ", err)
		return err
	}

	return nil
}
