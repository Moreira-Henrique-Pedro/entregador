package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/commands"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/usecases/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func pushBody(t *testing.T, eventType string, data any) string {
	t.Helper()

	raw, ok := data.([]byte)
	if !ok {
		var err error
		raw, err = json.Marshal(data)
		require.NoError(t, err)
	}

	body, err := json.Marshal(map[string]any{
		"message": map[string]any{
			"data":         raw,
			"attributes":   map[string]string{"EventType": eventType, "Key": "d1"},
			"messageId":    "m1",
			"message_id":   "m1",
			"publishTime":  "2026-10-04T12:00:00Z",
			"publish_time": "2026-10-04T12:00:00Z",
		},
		"subscription": "projects/p/subscriptions/s",
	})
	require.NoError(t, err)
	return string(body)
}

func TestNotificationsController(t *testing.T) {
	notify := commands.NotifyDeliveryCommand{DeliveryID: "d1", NotificationType: entities.NotificationTypeDeliveryArrived}

	tests := []struct {
		name       string
		body       string
		expect     bool
		useCaseErr error
		wantStatus int
	}{
		{name: "notified", body: pushBody(t, commands.NotifyDeliveryCommandType, notify), expect: true, wantStatus: http.StatusNoContent},
		{name: "use case error is retried", body: pushBody(t, commands.NotifyDeliveryCommandType, notify), expect: true, useCaseErr: errors.New("twilio down"), wantStatus: http.StatusInternalServerError},
		{name: "unknown event type is acked", body: pushBody(t, "Other", notify), wantStatus: http.StatusNoContent},
		{name: "invalid envelope", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "invalid data", body: pushBody(t, commands.NotifyDeliveryCommandType, []byte("not json")), wantStatus: http.StatusBadRequest},
		{name: "missing delivery id", body: pushBody(t, commands.NotifyDeliveryCommandType, commands.NotifyDeliveryCommand{}), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := mocks.NewNotifyDelivery(t)
			if tt.expect {
				useCase.EXPECT().Execute(mock.Anything, "d1", entities.NotificationTypeDeliveryArrived).Return(tt.useCaseErr).Once()
			}

			rec := httptest.NewRecorder()
			serveTo(rec, NewNotificationsController(useCase, nil), httptest.NewRequest(http.MethodPost, "/internal/pubsub/notifications", strings.NewReader(tt.body)))

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestNotificationsController_Auth(t *testing.T) {
	reject := func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusUnauthorized) }

	rec := httptest.NewRecorder()
	body := pushBody(t, commands.NotifyDeliveryCommandType, commands.NotifyDeliveryCommand{DeliveryID: "d1"})
	serveTo(rec, NewNotificationsController(mocks.NewNotifyDelivery(t), reject), httptest.NewRequest(http.MethodPost, PubSubPushPath, strings.NewReader(body)))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
