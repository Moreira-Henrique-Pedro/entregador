package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	inmocks "github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in/mocks"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
)

const serviceAccount = "pubsub-push@project.iam.gserviceaccount.com"

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
			"data":       raw, // []byte is encoded as base64, like Pub/Sub does
			"attributes": map[string]string{"EventType": eventType, "Key": "d1"},
			"messageId":  "m1",
		},
		"subscription": "projects/p/subscriptions/s",
	})
	require.NoError(t, err)
	return string(body)
}

func TestPushHandler(t *testing.T) {
	notify := messages.NotifyDelivery{DeliveryID: "d1", NotificationType: domain.NotificationTypeDeliveryArrived}

	tests := []struct {
		name       string
		body       string
		expect     bool
		useCaseErr error
		wantStatus int
	}{
		{name: "notified", body: pushBody(t, messages.NotifyDeliveryType, notify), expect: true, wantStatus: http.StatusNoContent},
		{name: "use case error is retried", body: pushBody(t, messages.NotifyDeliveryType, notify), expect: true, useCaseErr: errors.New("twilio down"), wantStatus: http.StatusInternalServerError},
		{name: "unknown event type is acked", body: pushBody(t, "Other", notify), wantStatus: http.StatusNoContent},
		{name: "invalid envelope", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "invalid data", body: pushBody(t, messages.NotifyDeliveryType, []byte("not json")), wantStatus: http.StatusBadRequest},
		{name: "missing delivery id", body: pushBody(t, messages.NotifyDeliveryType, messages.NotifyDelivery{}), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := inmocks.NewNotifyDelivery(t)
			if tt.expect {
				useCase.EXPECT().Execute(mock.Anything, "d1", domain.NotificationTypeDeliveryArrived).Return(tt.useCaseErr).Once()
			}

			rec := httptest.NewRecorder()
			NewPushHandler(useCase, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/internal/pubsub/notifications", strings.NewReader(tt.body)))

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestPushHandler_TokenVerification(t *testing.T) {
	validClaims := map[string]any{"email": serviceAccount, "email_verified": true}

	tests := []struct {
		name       string
		header     string
		claims     map[string]any
		validErr   error
		wantStatus int
	}{
		{name: "valid token", header: "Bearer good", claims: validClaims, wantStatus: http.StatusNoContent},
		{name: "missing token", header: "", wantStatus: http.StatusUnauthorized},
		{name: "not a bearer token", header: "Basic abc", wantStatus: http.StatusUnauthorized},
		{name: "invalid signature or audience", header: "Bearer bad", validErr: errors.New("invalid"), wantStatus: http.StatusUnauthorized},
		{name: "other service account", header: "Bearer good", claims: map[string]any{"email": "attacker@evil.com", "email_verified": true}, wantStatus: http.StatusUnauthorized},
		{name: "unverified email", header: "Bearer good", claims: map[string]any{"email": serviceAccount, "email_verified": false}, wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := NewTokenVerifier("https://api.example.com/internal/pubsub/notifications", serviceAccount)
			require.NoError(t, err)
			verifier.validate = func(_ context.Context, token, audience string) (*idtoken.Payload, error) {
				assert.Equal(t, "https://api.example.com/internal/pubsub/notifications", audience)
				if tt.validErr != nil {
					return nil, tt.validErr
				}
				return &idtoken.Payload{Claims: tt.claims}, nil
			}

			useCase := inmocks.NewNotifyDelivery(t)
			if tt.wantStatus == http.StatusNoContent {
				useCase.EXPECT().Execute(mock.Anything, "d1", domain.NotificationTypeDeliveryArrived).Return(nil).Once()
			}

			body := pushBody(t, messages.NotifyDeliveryType, messages.NotifyDelivery{DeliveryID: "d1", NotificationType: domain.NotificationTypeDeliveryArrived})
			req := httptest.NewRequest(http.MethodPost, "/internal/pubsub/notifications", strings.NewReader(body))
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			NewPushHandler(useCase, verifier).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestNewTokenVerifier_RequiresConfig(t *testing.T) {
	_, err := NewTokenVerifier("", serviceAccount)
	require.Error(t, err)

	_, err = NewTokenVerifier("https://api.example.com", "")
	require.Error(t, err)
}
