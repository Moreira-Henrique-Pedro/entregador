package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

const (
	serviceAccount = "pubsub-push@project.iam.gserviceaccount.com"
	audience       = "https://api.example.com/internal/pubsub/notifications"
)

func TestPubSubPushAuth(t *testing.T) {
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
			auth, err := NewPubSubPushAuth(audience, serviceAccount)
			require.NoError(t, err)
			auth.validate = func(_ context.Context, token, gotAudience string) (*idtoken.Payload, error) {
				assert.Equal(t, audience, gotAudience)
				if tt.validErr != nil {
					return nil, tt.validErr
				}
				return &idtoken.Payload{Claims: tt.claims}, nil
			}

			nextCalled := false
			engine := gin.New()
			engine.POST("/internal/pubsub/notifications", auth.Middleware, func(ctx *gin.Context) {
				nextCalled = true
				ctx.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodPost, "/internal/pubsub/notifications", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantStatus == http.StatusNoContent, nextCalled)
		})
	}
}

func TestNewPubSubPushAuth_RequiresConfig(t *testing.T) {
	_, err := NewPubSubPushAuth("", serviceAccount)
	require.Error(t, err)

	_, err = NewPubSubPushAuth(audience, "")
	require.Error(t, err)
}
