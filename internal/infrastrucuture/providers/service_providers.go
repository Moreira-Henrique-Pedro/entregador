package providers

import (
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/domain/interfaces/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill"
)

type ServiceProviders struct {
	Logger           logger.Logger
	MessagePublisher pubsub.MessagePublisher[any]
}

func NewServiceProviders(envs *config.Environment, logger logger.Logger) (*ServiceProviders, error) {

	messagePublisher, err := createMessagePublisher(envs, logger)
	if err != nil {
		return nil, err
	}

	return &ServiceProviders{
		Logger:           logger,
		MessagePublisher: messagePublisher,
	}, nil
}

func createMessagePublisher(cfg *config.Environment, logger logger.Logger) (pubsub.MessagePublisher[any], error) {
	publisher, err := watermill.NewWatermillPublisher[any](cfg.Pubsub.DeliveryBrokersHosts, logger)
	if err != nil {
		logger.Error("Failed to create Watermill publisher", "error", err.Error())
		return nil, err
	}
	return publisher, nil
}
