package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
)

type capturedRequest struct {
	method      string
	path        string
	user        string
	password    string
	contentType string
	form        url.Values
}

func newTwilioServer(t *testing.T, status int, body string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.user, captured.password, _ = r.BasicAuth()
		captured.contentType = r.Header.Get("Content-Type")
		captured.form, _ = url.ParseQuery(string(raw))

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, captured
}

func newTestNotifier(t *testing.T, baseURL string, contentSIDs map[entities.NotificationType]string) services.Notifier {
	t.Helper()

	n, err := NewTwilioWhatsApp(TwilioConfig{
		AccountSID:         "AC123",
		AuthToken:          "secret",
		From:               "+14155238886",
		ContentSIDs:        contentSIDs,
		DefaultCountryCode: "55",
		BaseURL:            baseURL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return n
}

var arrivedNotification = entities.Notification{
	Type:      entities.NotificationTypeDeliveryArrived,
	Phone:     "11999999999",
	Body:      "Chegou uma entrega",
	Variables: []string{"Ana", "101", "caixa"},
}

func TestTwilioWhatsApp_SendsFreeFormBody(t *testing.T) {
	server, captured := newTwilioServer(t, http.StatusCreated, `{"sid":"SM1","status":"queued"}`)

	err := newTestNotifier(t, server.URL, nil).Send(context.Background(), arrivedNotification)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.method != http.MethodPost || captured.path != "/2010-04-01/Accounts/AC123/Messages.json" {
		t.Errorf("request = %s %s, want POST to the account messages", captured.method, captured.path)
	}
	if captured.user != "AC123" || captured.password != "secret" {
		t.Errorf("basic auth = %s:%s, want AC123:secret", captured.user, captured.password)
	}
	if captured.contentType != "application/x-www-form-urlencoded" {
		t.Errorf("content type = %q", captured.contentType)
	}

	want := map[string]string{
		"From": "whatsapp:+14155238886",
		"To":   "whatsapp:+5511999999999",
		"Body": "Chegou uma entrega",
	}
	for key, value := range want {
		if got := captured.form.Get(key); got != value {
			t.Errorf("form[%s] = %q, want %q", key, got, value)
		}
	}
	if captured.form.Has("ContentSid") {
		t.Error("free-form message must not send ContentSid")
	}
}

func TestTwilioWhatsApp_SendsContentTemplate(t *testing.T) {
	server, captured := newTwilioServer(t, http.StatusCreated, `{"sid":"SM1","status":"queued"}`)
	contentSIDs := map[entities.NotificationType]string{
		entities.NotificationTypeDeliveryArrived: "HXarrived",
	}

	err := newTestNotifier(t, server.URL, contentSIDs).Send(context.Background(), arrivedNotification)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := captured.form.Get("ContentSid"); got != "HXarrived" {
		t.Errorf("ContentSid = %q, want HXarrived", got)
	}
	if captured.form.Has("Body") {
		t.Error("template message must not send Body")
	}

	var variables map[string]string
	if err := json.Unmarshal([]byte(captured.form.Get("ContentVariables")), &variables); err != nil {
		t.Fatalf("invalid ContentVariables: %v", err)
	}
	want := map[string]string{"1": "Ana", "2": "101", "3": "caixa"}
	for key, value := range want {
		if variables[key] != value {
			t.Errorf("ContentVariables[%s] = %q, want %q", key, variables[key], value)
		}
	}
}

func TestTwilioWhatsApp_TypeWithoutTemplateFallsBackToBody(t *testing.T) {
	server, captured := newTwilioServer(t, http.StatusCreated, `{}`)
	contentSIDs := map[entities.NotificationType]string{
		entities.NotificationTypeDeliveryArrived:  "HXarrived",
		entities.NotificationTypeDeliveryPickedUp: "",
	}
	pickup := arrivedNotification
	pickup.Type = entities.NotificationTypeDeliveryPickedUp
	pickup.Body = "Entrega retirada"

	if err := newTestNotifier(t, server.URL, contentSIDs).Send(context.Background(), pickup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.form.Get("Body") != "Entrega retirada" || captured.form.Has("ContentSid") {
		t.Errorf("form = %v, want free-form body", captured.form)
	}
}

func TestTwilioWhatsApp_Errors(t *testing.T) {
	tests := []struct {
		name                 string
		status               int
		body                 string
		wantInvalidRecipient bool
	}{
		{name: "invalid To number", status: http.StatusBadRequest, body: `{"code":21211,"message":"Invalid 'To' Phone Number"}`, wantInvalidRecipient: true},
		{name: "recipient unsubscribed", status: http.StatusBadRequest, body: `{"code":21610,"message":"unsubscribed"}`, wantInvalidRecipient: true},
		{name: "destination not found on the channel", status: http.StatusBadRequest, body: `{"code":63003,"message":"not found"}`, wantInvalidRecipient: true},
		{name: "other bad request is retried", status: http.StatusBadRequest, body: `{"code":21606,"message":"From is not a valid phone"}`},
		{name: "authentication failure is retried", status: http.StatusUnauthorized, body: `{"code":20003,"message":"Authenticate"}`},
		{name: "throttling is retried", status: http.StatusTooManyRequests, body: `{"code":20429,"message":"Too Many Requests"}`},
		{name: "server error is retried", status: http.StatusInternalServerError, body: `not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := newTwilioServer(t, tt.status, tt.body)

			err := newTestNotifier(t, server.URL, nil).Send(context.Background(), arrivedNotification)

			if err == nil {
				t.Fatal("expected error")
			}
			if got := errors.Is(err, entities.ErrInvalidRecipient); got != tt.wantInvalidRecipient {
				t.Errorf("invalid recipient = %v, want %v (err: %v)", got, tt.wantInvalidRecipient, err)
			}
		})
	}
}

func TestTwilioWhatsApp_InvalidPhoneIsNotSent(t *testing.T) {
	server, captured := newTwilioServer(t, http.StatusCreated, `{}`)
	notification := arrivedNotification
	notification.Phone = "123"

	err := newTestNotifier(t, server.URL, nil).Send(context.Background(), notification)

	if !errors.Is(err, entities.ErrInvalidRecipient) {
		t.Fatalf("err = %v, want ErrInvalidRecipient", err)
	}
	if captured.method != "" {
		t.Error("no request must be sent for an invalid phone")
	}
}

func TestTwilioWhatsApp_NetworkErrorIsRetried(t *testing.T) {
	server, _ := newTwilioServer(t, http.StatusCreated, `{}`)
	server.Close()

	err := newTestNotifier(t, server.URL, nil).Send(context.Background(), arrivedNotification)

	if err == nil || errors.Is(err, entities.ErrInvalidRecipient) {
		t.Fatalf("err = %v, want a retryable error", err)
	}
}

func TestNewTwilioWhatsApp_RequiresCredentials(t *testing.T) {
	tests := []struct {
		name   string
		config TwilioConfig
	}{
		{name: "missing account sid", config: TwilioConfig{AuthToken: "secret", From: "+1"}},
		{name: "missing auth token", config: TwilioConfig{AccountSID: "AC123", From: "+1"}},
		{name: "missing from", config: TwilioConfig{AccountSID: "AC123", AuthToken: "secret"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewTwilioWhatsApp(tt.config); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		want    string
		wantErr bool
	}{
		{name: "mobile without country code", phone: "11999999999", want: "+5511999999999"},
		{name: "landline without country code", phone: "1133334444", want: "+551133334444"},
		{name: "formatted national number", phone: "(11) 99999-9999", want: "+5511999999999"},
		{name: "country code without plus", phone: "5511999999999", want: "+5511999999999"},
		{name: "E.164", phone: "+5511999999999", want: "+5511999999999"},
		{name: "E.164 from another country", phone: "+1 415 523 8886", want: "+14155238886"},
		{name: "empty", phone: "", wantErr: true},
		{name: "too short", phone: "123", wantErr: true},
		{name: "E.164 too short", phone: "+123", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePhone(tt.phone, "55")

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("normalizePhone(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}
