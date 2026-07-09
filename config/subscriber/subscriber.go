package subscriber

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/duration"
	validator "github.com/go-playground/validator/v10"
)

const (
	defaultTimeOut              duration.Duration = duration.Duration(5 * time.Second)
	defaultCluster              string            = "delivery"
	defaultDlqCluster           string            = "delivery-dlq"
	defaultMaxRetries           int               = 5
	defaultRetryInitialInterval duration.Duration = duration.Duration(time.Second)
	defaultRetryMaxInterval     duration.Duration = duration.Duration(time.Second * 30)
	defaultRetryMultiplier      float64           = 2.0
)

type RetryConfig struct {
	MaxRetries      int               `json:"max_retries"`
	InitialInterval duration.Duration `json:"initial_interval"`
	MaxInterval     duration.Duration `json:"max_interval"`
	Multiplier      float64           `json:"multiplier"`
}

type SubscriberConfig struct {
	App           string            `json:"app" validate:"required"`
	ConsumerGroup string            `json:"consumer_group" validate:"required"`
	ConsumerName  string            `json:"consumer_name" validate:"required"`
	Topic         string            `json:"topic" validate:"required"`
	TimeOut       duration.Duration `json:"timeout"`
	Cluster       string            `json:"cluster"`
	DLQCluster    string            `json:"dlq_cluster"`
	RetryConfig   *RetryConfig      `json:"retry"`
}

func initializeSubscriberConfig(dat []byte) (*SubscriberConfig, error) {
	var subscriberCfg *SubscriberConfig
	if err := json.Unmarshal(dat, &subscriberCfg); err != nil {
		return nil, err
	}

	if err := validator.New().Struct(subscriberCfg); err != nil {
		return nil, err
	}

	if subscriberCfg.TimeOut == 0 {
		subscriberCfg.TimeOut = defaultTimeOut
	}

	if subscriberCfg.Cluster == "" {
		subscriberCfg.Cluster = defaultCluster
	}

	if subscriberCfg.DLQCluster == "" {
		subscriberCfg.DLQCluster = defaultDlqCluster
	}

	if subscriberCfg.RetryConfig == nil {
		subscriberCfg.RetryConfig = &RetryConfig{
			MaxRetries:      defaultMaxRetries,
			InitialInterval: defaultRetryInitialInterval,
			MaxInterval:     defaultRetryMaxInterval,
			Multiplier:      defaultRetryMultiplier,
		}
	}

	return subscriberCfg, nil
}

func Read() (*SubscriberConfig, error) {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()
	if *configPath == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	f, err := os.ReadFile(*configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	cfg, err := initializeSubscriberConfig(f)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize config file: %w", err)
	}

	return cfg, nil
}
