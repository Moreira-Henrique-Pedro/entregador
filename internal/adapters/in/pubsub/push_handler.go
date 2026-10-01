// Package pubsub is the inbound adapter for Pub/Sub push subscriptions: Pub/Sub POSTs each
// message to the API and retries it while the response is not 2xx.
package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/messages"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	pkgPubsub "github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	"google.golang.org/api/idtoken"
)

const maxPushBodyBytes = 1 << 20

// pushRequest is the body Pub/Sub sends to a push endpoint.
type pushRequest struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
		MessageID  string            `json:"messageId"`
	} `json:"message"`
	Subscription    string `json:"subscription"`
	DeliveryAttempt int    `json:"deliveryAttempt"`
}

// TokenVerifier checks the OIDC token Pub/Sub signs with the push service account.
type TokenVerifier struct {
	audience       string
	serviceAccount string
	validate       func(ctx context.Context, token, audience string) (*idtoken.Payload, error)
}

func NewTokenVerifier(audience, serviceAccount string) (*TokenVerifier, error) {
	if audience == "" || serviceAccount == "" {
		return nil, errors.New("push token verification needs PUBSUB_PUSH_AUDIENCE and PUBSUB_PUSH_SERVICE_ACCOUNT")
	}
	return &TokenVerifier{audience: audience, serviceAccount: serviceAccount, validate: idtoken.Validate}, nil
}

func (v *TokenVerifier) verify(r *http.Request) error {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return errors.New("missing bearer token")
	}

	payload, err := v.validate(r.Context(), token, v.audience)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	if email != v.serviceAccount || !verified {
		return fmt.Errorf("unexpected token email %q", email)
	}
	return nil
}

type PushHandler struct {
	notifyDelivery in.NotifyDelivery
	verifier       *TokenVerifier
}

// NewPushHandler builds the push endpoint; a nil verifier disables the token check (emulator only).
func NewPushHandler(notifyDelivery in.NotifyDelivery, verifier *TokenVerifier) *PushHandler {
	return &PushHandler{
		notifyDelivery: notifyDelivery,
		verifier:       verifier,
	}
}

// ServeHTTP answers 2xx to ack the message. Malformed messages get 400 and, after the
// subscription max delivery attempts, land in its dead-letter topic.
func (h *PushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerFromContext(r.Context())

	if h.verifier != nil {
		if err := h.verifier.verify(r); err != nil {
			log.Warn("Rejected Pub/Sub push request", "error", err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	var request pushRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxPushBodyBytes)).Decode(&request); err != nil {
		log.Error("Invalid Pub/Sub push body", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventType := request.Message.Attributes[pkgPubsub.EventTypeHeader]
	log = log.With(
		"message_id", request.Message.MessageID,
		"event_type", eventType,
		"message_key", request.Message.Attributes[pkgPubsub.KeyHeader],
		"delivery_attempt", request.DeliveryAttempt,
	)
	ctx := log.AddToContext(r.Context(), log)

	if eventType != messages.NotifyDeliveryType {
		log.Warn("Ignoring Pub/Sub message with unknown event type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var command messages.NotifyDelivery
	if err := json.Unmarshal(request.Message.Data, &command); err != nil || command.DeliveryID == "" {
		log.Error("Invalid NotifyDelivery message", "error", fmt.Sprint(err), "data", string(request.Message.Data))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.notifyDelivery.Execute(ctx, command.DeliveryID, command.NotificationType); err != nil {
		log.Error("Failed to notify delivery, Pub/Sub will retry", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info("Pub/Sub message processed successfully")
	w.WriteHeader(http.StatusNoContent)
}
