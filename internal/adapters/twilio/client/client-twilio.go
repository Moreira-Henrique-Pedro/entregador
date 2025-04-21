package client

import (
	"context"
	"log/slog"
	"os"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioClientPort interface {
	SendWhatsAppMessage(ctx context.Context, to string, message string) error
}

type TwilioClient struct {
	client twilio.RestClient
}

func NewTwilioClient() TwilioClient {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACC_SID"),
		Password: os.Getenv("TWILIO_TOKEN"),
	})

	return TwilioClient{client: *client}
}

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
