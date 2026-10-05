package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/entities"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/services"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/httpclient"
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
	ContentSIDs        map[entities.NotificationType]string
	DefaultCountryCode string
	BaseURL            string
}

type TwilioWhatsApp struct {
	config TwilioConfig
	client *httpclient.Client
}

func NewTwilioWhatsApp(config TwilioConfig) (services.Notifier, error) {
	if config.AccountSID == "" || config.AuthToken == "" || config.From == "" {
		return nil, errors.New("twilio account sid, auth token and from are required")
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultTwilioBaseURL
	}

	return &TwilioWhatsApp{
		config: config,
		client: httpclient.New(
			httpclient.WithBaseURL(config.BaseURL),
			httpclient.WithBasicAuth(config.AccountSID, config.AuthToken),
			httpclient.WithTimeout(twilioRequestTimeout),
		),
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

func (n *TwilioWhatsApp) Send(ctx context.Context, notification entities.Notification) error {
	to, err := normalizePhone(notification.Phone, n.config.DefaultCountryCode)
	if err != nil {
		return fmt.Errorf("%w: %v", entities.ErrInvalidRecipient, err)
	}

	form, err := n.buildForm(to, notification)
	if err != nil {
		return err
	}

	response, err := n.client.PostForm(ctx, "/2010-04-01/Accounts/"+n.config.AccountSID+"/Messages.json", form)
	if err != nil {
		return fmt.Errorf("send twilio request: %w", err)
	}
	if !response.IsSuccess() {
		return twilioError(response.StatusCode, response.Body)
	}

	var message twilioMessageResponse
	_ = response.DecodeJSON(&message)
	logger.GetLoggerFromContext(ctx).Info("WhatsApp message sent",
		"type", string(notification.Type),
		"twilio_sid", message.SID,
		"twilio_status", message.Status,
	)

	return nil
}

func (n *TwilioWhatsApp) buildForm(to string, notification entities.Notification) (url.Values, error) {
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
		return fmt.Errorf("%w: %v", entities.ErrInvalidRecipient, err)
	}
	return err
}

const (
	twilioErrInvalidToNumber         = 21211
	twilioErrRegionNotEnabled        = 21408
	twilioErrRecipientUnsubscribed   = 21610
	twilioErrNotMobileNumber         = 21614
	twilioErrDestinationNotReachable = 63003
)

func isRecipientErrorCode(code int) bool {
	switch code {
	case twilioErrInvalidToNumber,
		twilioErrRegionNotEnabled,
		twilioErrRecipientUnsubscribed,
		twilioErrNotMobileNumber,
		twilioErrDestinationNotReachable:
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
