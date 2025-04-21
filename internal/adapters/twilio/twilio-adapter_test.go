package twilio_test

import (
	"context"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/twilio/errors"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	message = "sua entrega chegou"
)

func TestTwilioServiceSendWhatsAppMessageSuccess(t *testing.T) {
	// Arrange
	mockClient := new(mocks.TwilioClientPort)
	mockClient.On("SendWhatsAppMessage", mock.Anything, "+5511999999999", message).Return(nil)

	service := twilio.NewTwilioPort(mockClient)

	// Act
	err := service.SendWhatsAppMessage(context.Background(), "+5511999999999", message)

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestTwilioServiceSendWhatsAppMessageError(t *testing.T) {
	// Arrange
	mockClient := new(mocks.TwilioClientPort)
	mockClient.On("SendWhatsAppMessage", mock.Anything, "+5511999999999", message).Return(errors.ErrFailedToSendWhatsApp)

	service := twilio.NewTwilioPort(mockClient)

	// Act
	err := service.SendWhatsAppMessage(context.Background(), "+5511999999999", message)

	// Assert
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to send WhatsApp message")
	mockClient.AssertExpectations(t)
}
