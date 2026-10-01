package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	inmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

func TestNewHandlerRegistry_NotifyDelivery(t *testing.T) {
	failure := errors.New("twilio down")

	for _, wantErr := range []error{nil, failure} {
		notify := inmocks.NewNotifyDelivery(t)
		notify.EXPECT().Execute(mock.Anything, "d1", domain.NotificationTypeDeliveryArrived).Return(wantErr).Once()

		handler, err := NewHandlerRegistry(UseCases{NotifyDelivery: notify}).GetEventHandlerByEventType(messages.NotifyDeliveryType)
		require.NoError(t, err)

		err = handler.Handler(context.Background(), &messages.NotifyDelivery{DeliveryID: "d1", NotificationType: domain.NotificationTypeDeliveryArrived})
		require.ErrorIs(t, err, wantErr)
	}
}
