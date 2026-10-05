package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

const defaultEnvFile = ".env"

var AppName = "entregador-api"
var Envs *Environment

type Environment struct {
	App struct {
		Env      string `env:"ENVIRONMENT,default=development"`
		LogLevel string `env:"LOG_LEVEL,default=info"`
		Name     string `env:"APP_NAME,default=entregador-api"`
		Version  string `env:"APP_VERSION,default=1.0.0"`
	}
	HTTP struct {
		Port         string `env:"HTTP_PORT,default=8081"`
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
	PubSub struct {
		ProjectID          string `env:"GCP_PROJECT_ID"`
		NotificationsTopic string `env:"PUBSUB_NOTIFICATIONS_TOPIC,default=delivery-notifications"`
		VerifyPushToken    bool   `env:"PUBSUB_PUSH_VERIFY_TOKEN,default=true"`
		PushAudience       string `env:"PUBSUB_PUSH_AUDIENCE"`
		PushServiceAccount string `env:"PUBSUB_PUSH_SERVICE_ACCOUNT"`
	}
	Auth struct {
		Enabled bool `env:"AUTH_ENABLED,default=true"`
	}
	Firebase struct {
		ProjectID        string `env:"FIREBASE_PROJECT_ID"`
		AuthEmulatorHost string `env:"FIREBASE_AUTH_EMULATOR_HOST"`
	}
}

func ReadEnvs() (*Environment, error) {
	if Envs == nil {
		if err := loadEnvFile(); err != nil {
			return nil, err
		}

		Envs = &Environment{}
		if err := envdecode.Decode(Envs); err != nil {
			return nil, fmt.Errorf("error loading environment variables: %w", err)
		}
		if Envs.HTTP.CloudRunPort != "" {
			Envs.HTTP.Port = Envs.HTTP.CloudRunPort
		}
		AppName = Envs.App.Name
	}

	return Envs, nil
}

func (c *Environment) IsProduction() bool {
	return c.App.Env == EnvironmentProduction
}

func loadEnvFile() error {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = defaultEnvFile
	}
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error loading env file %s: %w", envFile, err)
	}
	return nil
}
