package service

import (
	"log/slog"
	"os"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioService struct {
	client *twilio.RestClient
}

func NewTwilioService() *TwilioService {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACC_SID"),
		Password: os.Getenv("TWILIO_TOKEN"),
	})

	return &TwilioService{client: client}
}

func (t *TwilioService) SendWhatsAppMessage(to string) error {

	body := "Sua Entrega chegou"

	params := &api.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(os.Getenv("TWILIO_FROM"))
	params.SetBody(body)

	_, err := t.client.Api.CreateMessage(params)
	if err != nil {
		slog.Error("Error sending WahtsApp message: %v", err)
		return err
	}

	slog.Info("WhatsApp message sent successfully")
	return nil
}
