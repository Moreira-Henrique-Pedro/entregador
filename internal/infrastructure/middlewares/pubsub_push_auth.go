package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type PubSubPushAuth struct {
	audience       string
	serviceAccount string
	validate       func(ctx context.Context, token, audience string) (*idtoken.Payload, error)
}

func NewPubSubPushAuth(audience, serviceAccount string) (*PubSubPushAuth, error) {
	if audience == "" || serviceAccount == "" {
		return nil, errors.New("push token verification needs PUBSUB_PUSH_AUDIENCE and PUBSUB_PUSH_SERVICE_ACCOUNT")
	}
	return &PubSubPushAuth{audience: audience, serviceAccount: serviceAccount, validate: idtoken.Validate}, nil
}

func (a *PubSubPushAuth) Middleware(ctx *gin.Context) {
	if err := a.verify(ctx.Request); err != nil {
		logger.GetLoggerFromContext(ctx.Request.Context()).Warn("Rejected Pub/Sub push request", "error", err.Error())
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.Next()
}

func (a *PubSubPushAuth) verify(r *http.Request) error {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return errors.New("missing bearer token")
	}

	payload, err := a.validate(r.Context(), token, a.audience)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	if email != a.serviceAccount || !verified {
		return fmt.Errorf("unexpected token email %q", email)
	}
	return nil
}
