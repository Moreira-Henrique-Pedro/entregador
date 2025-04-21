// package twilio contem a implementação do serviço Twilio, que é responsável por enviar mensagens via WhatsApp.
package twilio

import (
	"context"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/client"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports"
	"github.com/sirupsen/logrus"
)

// TwilioService é a estrutura que representa o serviço Twilio
type TwilioService struct {
	client client.TwilioClientPort
}

// NewTwilioService cria uma nova instância do serviço Twilio
func NewTwilioPort(client client.TwilioClientPort) ports.TwilioPort {
	return &TwilioService{
		client: client,
	}
}

// SendWhatsAppMessage envia uma mensagem via WhatsApp
func (t *TwilioService) SendWhatsAppMessage(ctx context.Context, to string, message string) error {
	logger := logrus.New()
	logger.Info("Sending WhatsApp message to: ", to)
	// Call the Twilio client to send the message
	err := t.client.SendWhatsAppMessage(ctx, to, message)
	if err != nil {
		logger.Error("Failed to send WhatsApp message: ", err)
		return err
	}

	return nil
}
