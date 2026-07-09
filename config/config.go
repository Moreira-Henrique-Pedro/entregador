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
	MongoDB struct {
		URI      string `env:"MONGODB_URI"`
		Database string `env:"MONGODB_DATABASE"`
	}
	Pubsub struct {
		DeliveryBrokersHostsRaw string `env:"DELIVERY_BROKER_HOSTS"`
		DeliveryBrokersHosts    []string
		DLQTopic                string `env:"DLQ_TOPIC,default=delivery-subscriber.dlq"`
		BrokerHosts             []string
	}
	Delivery struct {
		URL string `env:"DELIVERY_URL,required"`
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
		Envs.Pubsub.DeliveryBrokersHosts = strings.Split(Envs.Pubsub.DeliveryBrokersHostsRaw, ",")
		AppName = Envs.App.Name
	}

	return Envs, nil
}

func (c *Environment) IsProduction() bool {
	return c.App.Env == EnvironmentProduction
}
