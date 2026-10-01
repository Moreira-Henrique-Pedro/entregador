package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Moreira-Henrique-Pedro/entregador/config/subscriber"
	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

const (
	DeliveryClusterName = "Delivery"
)

var AppName = "delivery-subscriber"
var Envs *Environment

type Environment struct {
	App struct {
		Env      string `env:"ENVIRONMENT,default=development"`
		LogLevel string `env:"LOG_LEVEL,default=info"`
		Name     string `env:"APP_NAME,default=delivery-subscriber"`
		Version  string `env:"APP_VERSION,default=1.0.0"`
	}
	HTTP struct {
		Port string `env:"HTTP_PORT,default=8081"`
		// PORT is injected by Cloud Run and takes precedence over HTTP_PORT.
		CloudRunPort string `env:"PORT"`
	}
	Notifier struct {
		Provider           string `env:"NOTIFIER_PROVIDER,default=log"`
		DefaultCountryCode string `env:"NOTIFIER_DEFAULT_COUNTRY_CODE,default=55"`
	}
	Twilio struct {
		AccountSID                 string `env:"TWILIO_ACCOUNT_SID"`
		AuthToken                  string `env:"TWILIO_AUTH_TOKEN"`
		WhatsAppFrom               string `env:"TWILIO_WHATSAPP_FROM"`
		ContentSIDDeliveryArrived  string `env:"TWILIO_CONTENT_SID_DELIVERY_ARRIVED"`
		ContentSIDDeliveryPickedUp string `env:"TWILIO_CONTENT_SID_DELIVERY_PICKED_UP"`
		BaseURL                    string `env:"TWILIO_BASE_URL"`
	}
	MongoDB struct {
		URI      string `env:"MONGODB_URI"`
		Database string `env:"MONGODB_DATABASE"`
	}
	Messaging struct {
		// pubsub: Google Cloud Pub/Sub with push delivery to the API (single binary).
		// kafka: Kafka, consumed by cmd/worker.
		Provider string `env:"MESSAGING_PROVIDER,default=pubsub"`
	}
	PubSub struct {
		ProjectID          string `env:"GCP_PROJECT_ID"`
		NotificationsTopic string `env:"PUBSUB_NOTIFICATIONS_TOPIC,default=delivery-notifications"`
		// Push requests carry a Google-signed OIDC token: only disable the check against the emulator.
		VerifyPushToken    bool   `env:"PUBSUB_PUSH_VERIFY_TOKEN,default=true"`
		PushAudience       string `env:"PUBSUB_PUSH_AUDIENCE"`
		PushServiceAccount string `env:"PUBSUB_PUSH_SERVICE_ACCOUNT"`
	}
	Kafka struct {
		DeliveryBrokersHostsRaw string `env:"DELIVERY_BROKER_HOSTS"`
		DeliveryBrokersHosts    []string
		CommandsTopic           string `env:"INTERNAL_COMMANDS_TOPIC,default=delivery-internal.commands"`
		DLQTopic                string `env:"DLQ_TOPIC,default=delivery-subscriber.dlq"`
	}
	Delivery struct {
		URL string `env:"DELIVERY_URL"`
	}
}

type AppConfigs struct {
	Envs              *Environment
	SubscriberConfigs *subscriber.SubscriberConfig
}

func NewConfig() (*AppConfigs, error) {
	Envs, err := ReadEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	subCfg, err := subscriber.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read subscriber config: %w", err)
	}

	return &AppConfigs{
		Envs:              Envs,
		SubscriberConfigs: subCfg,
	}, nil
}

func ReadEnvs() (*Environment, error) {
	if Envs == nil {
		if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}

		Envs = &Environment{}
		if err := envdecode.Decode(Envs); err != nil {
			return nil, fmt.Errorf("error loading environment variables: %w", err)
		}
		Envs.Kafka.DeliveryBrokersHosts = splitHosts(Envs.Kafka.DeliveryBrokersHostsRaw)
		if Envs.HTTP.CloudRunPort != "" {
			Envs.HTTP.Port = Envs.HTTP.CloudRunPort
		}
		AppName = Envs.App.Name
	}

	return Envs, nil
}

func splitHosts(raw string) []string {
	hosts := []string{}
	for _, host := range strings.Split(raw, ",") {
		if host = strings.TrimSpace(host); host != "" {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

const (
	MessagingProviderPubSub = "pubsub"
	MessagingProviderKafka  = "kafka"
)

func (c *Environment) IsProduction() bool {
	return c.App.Env == EnvironmentProduction
}
