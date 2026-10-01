package providers

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	kafkaIn "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/in/kafka"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/usecases"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill"
	watermillKafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
)

// Worker wires the Kafka adapter that sends the delivery notifications.
type Worker struct {
	Consumer     *kafkaIn.Consumer
	repositories *repositories
	publisher    pubsub.MessagePublisher[any]
}

func NewWorker(cfg *config.AppConfigs, log logger.Logger) (*Worker, error) {
	deliveryNotifier, err := newNotifier(cfg.Envs)
	if err != nil {
		return nil, fmt.Errorf("create notifier: %w", err)
	}

	repos, err := newRepositories(cfg.Envs)
	if err != nil {
		return nil, err
	}

	publisher, err := watermill.NewWatermillPublisher[any](cfg.Envs.Kafka.DeliveryBrokersHosts, log)
	if err != nil {
		_ = repos.close(context.Background())
		return nil, fmt.Errorf("create kafka publisher: %w", err)
	}

	subscriber, err := newKafkaSubscriber(cfg, log)
	if err != nil {
		_ = publisher.Close(context.Background())
		_ = repos.close(context.Background())
		return nil, fmt.Errorf("create kafka subscriber: %w", err)
	}

	registry := kafkaIn.NewHandlerRegistry(kafkaIn.UseCases{
		NotifyDelivery: usecases.NewNotifyDelivery(repos.deliveryRepository, repos.residentRepository, deliveryNotifier),
	})

	sub := cfg.SubscriberConfigs
	consumer := kafkaIn.NewConsumer(subscriber, publisher, registry, kafkaIn.ConsumerConfig{
		Topic:    sub.Topic,
		DLQTopic: cfg.Envs.Kafka.DLQTopic,
		Timeout:  sub.TimeOut.Duration(),
		Retry: kafkaIn.RetryPolicy{
			MaxRetries:      sub.RetryConfig.MaxRetries,
			InitialInterval: sub.RetryConfig.InitialInterval.Duration(),
			MaxInterval:     sub.RetryConfig.MaxInterval.Duration(),
			Multiplier:      sub.RetryConfig.Multiplier,
		},
	}, log)

	return &Worker{
		Consumer:     consumer,
		repositories: repos,
		publisher:    publisher,
	}, nil
}

// Close stops consuming first, so no message is processed without its dependencies.
func (w *Worker) Close(ctx context.Context) error {
	return errors.Join(
		w.Consumer.Close(),
		w.publisher.Close(ctx),
		w.repositories.close(ctx),
	)
}

func newKafkaSubscriber(cfg *config.AppConfigs, log logger.Logger) (*watermillKafka.Subscriber, error) {
	saramaConfig := watermillKafka.DefaultSaramaSubscriberConfig()
	saramaConfig.ClientID = cfg.SubscriberConfigs.ConsumerName
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return watermillKafka.NewSubscriber(
		watermillKafka.SubscriberConfig{
			Brokers:               cfg.Envs.Kafka.DeliveryBrokersHosts,
			ConsumerGroup:         cfg.SubscriberConfigs.ConsumerGroup,
			OverwriteSaramaConfig: saramaConfig,
			Unmarshaler:           watermillKafka.DefaultMarshaler{},
		},
		watermill.NewWatermillLoggerFromLogger(log),
	)
}
