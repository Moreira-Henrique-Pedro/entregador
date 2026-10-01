package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

const (
	defaultTwilioBaseURL = "https://api.twilio.com"
	twilioRequestTimeout = 10 * time.Second
)

type TwilioConfig struct {
	AccountSID         string
	AuthToken          string
	From               string
	ContentSIDs        map[domain.NotificationType]string
	DefaultCountryCode string
	BaseURL            string
}

type TwilioWhatsApp struct {
	config     TwilioConfig
	httpClient *http.Client
}

func NewTwilioWhatsApp(config TwilioConfig, httpClient *http.Client) (out.Notifier, error) {
	if config.AccountSID == "" || config.AuthToken == "" || config.From == "" {
		return nil, errors.New("twilio account sid, auth token and from are required")
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultTwilioBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: twilioRequestTimeout}
	}

	return &TwilioWhatsApp{
		config:     config,
		httpClient: httpClient,
	}, nil
}

type twilioMessageResponse struct {
	SID    string `json:"sid"`
	Status string `json:"status"`
}

type twilioErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
}

func (n *TwilioWhatsApp) Send(ctx context.Context, notification out.Notification) error {
	to, err := normalizePhone(notification.Phone, n.config.DefaultCountryCode)
	if err != nil {
		return fmt.Errorf("%w: %v", out.ErrInvalidRecipient, err)
	}

	form, err := n.buildForm(to, notification)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", n.config.BaseURL, n.config.AccountSID)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build twilio request: %w", err)
	}
	request.SetBasicAuth(n.config.AccountSID, n.config.AuthToken)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := n.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send twilio request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read twilio response: %w", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return twilioError(response.StatusCode, body)
	}

	var message twilioMessageResponse
	_ = json.Unmarshal(body, &message)
	logger.GetLoggerFromContext(ctx).Info("WhatsApp message sent",
		"type", string(notification.Type),
		"twilio_sid", message.SID,
		"twilio_status", message.Status,
	)

	return nil
}

func (n *TwilioWhatsApp) buildForm(to string, notification out.Notification) (url.Values, error) {
	form := url.Values{}
	form.Set("From", "whatsapp:"+n.config.From)
	form.Set("To", "whatsapp:"+to)

	contentSID, hasTemplate := n.config.ContentSIDs[notification.Type]
	if !hasTemplate || contentSID == "" {
		form.Set("Body", notification.Body)
		return form, nil
	}

	variables := make(map[string]string, len(notification.Variables))
	for i, value := range notification.Variables {
		variables[strconv.Itoa(i+1)] = value
	}
	encoded, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("encode twilio content variables: %w", err)
	}

	form.Set("ContentSid", contentSID)
	form.Set("ContentVariables", string(encoded))
	return form, nil
}

func twilioError(statusCode int, body []byte) error {
	var apiErr twilioErrorResponse
	_ = json.Unmarshal(body, &apiErr)
	err := fmt.Errorf("twilio responded %d: code=%d message=%s", statusCode, apiErr.Code, apiErr.Message)

	if statusCode == http.StatusBadRequest && isRecipientErrorCode(apiErr.Code) {
		return fmt.Errorf("%w: %v", out.ErrInvalidRecipient, err)
	}
	return err
}

func isRecipientErrorCode(code int) bool {
	switch code {
	case 21211, // invalid 'To' phone number
		21408, // permission to send to this region not enabled
		21610, // recipient unsubscribed (STOP)
		21614, // 'To' is not a valid mobile number
		63003: // channel could not find the destination address
		return true
	}
	return false
}

func normalizePhone(phone, defaultCountryCode string) (string, error) {
	if phone == "" {
		return "", errors.New("phone is empty")
	}

	hasPlus := strings.HasPrefix(strings.TrimSpace(phone), "+")
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	if hasPlus {
		if len(digits) < 8 || len(digits) > 15 {
			return "", fmt.Errorf("phone %q is not a valid E.164 number", phone)
		}
		return "+" + digits, nil
	}

	if defaultCountryCode != "" && strings.HasPrefix(digits, defaultCountryCode) && (len(digits) == len(defaultCountryCode)+10 || len(digits) == len(defaultCountryCode)+11) {
		return "+" + digits, nil
	}

	if len(digits) != 10 && len(digits) != 11 {
		return "", fmt.Errorf("phone %q must have the area code and number", phone)
	}
	if defaultCountryCode == "" {
		return "", fmt.Errorf("phone %q has no country code", phone)
	}
	return "+" + defaultCountryCode + digits, nil
}
