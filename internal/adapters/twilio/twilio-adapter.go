package twilio

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
)

type TwilioService struct {
	client client.TwilioClient
}

func NewTwilioPort(client client.TwilioClient) ports.TwilioPort {
	return &TwilioService{
		client: client,
	}
}

func (t *TwilioService) SendWhatsAppMessage(ctx context.Context, to string) error {
	err := t.client.SendWhatsAppMessage(ctx, to)
	if err != nil {
		return err
	}

	return nil
}
