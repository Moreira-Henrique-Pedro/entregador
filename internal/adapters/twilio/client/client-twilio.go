package client

import (
	"context"
	"log/slog"
	"os"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioClientPort é a interface que define os métodos do cliente Twilio
type TwilioClientPort interface {
	SendWhatsAppMessage(ctx context.Context, to string, message string) error
}

// TwilioClient é a estrutura que implementa o cliente Twilio
type TwilioClient struct {
	client twilio.RestClient
}

// NewTwilioClient cria uma nova instância do cliente Twilio
func NewTwilioClient() TwilioClient {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACC_SID"),
		Password: os.Getenv("TWILIO_TOKEN"),
	})

	return TwilioClient{client: *client}
}

// SendWhatsAppMessage envia uma mensagem via WhatsApp usando o Twilio
func (t *TwilioClient) SendWhatsAppMessage(ctx context.Context, to string, message string) error {

	params := &openapi.CreateMessageParams{}
	params.SetTo("whatsapp:" + to)
	params.SetFrom("whatsapp:" + os.Getenv("TWILIO_FROM"))
	params.SetBody(message)

	_, err := t.client.Api.CreateMessage(params)
	if err != nil {
		slog.Error("Error sending WhatsApp message", "error", err)
		return err
	}

	slog.Info("WhatsApp message sent successfully")
	return nil
}
