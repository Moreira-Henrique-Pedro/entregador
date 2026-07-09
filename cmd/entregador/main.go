package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastrucuture/providers"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	appWatermill "github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill"
	watermillKafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
)

var version = "dev"

const internalCommandsTopic = "delivery-internal.commands"

type Application struct {
	Configs              *config.AppConfigs
	Logger               appLogger.Logger
	ServiceProviders     *providers.ServiceProviders
	TransporterProviders *providers.TransporterProviders
	WriterProviders      *providers.WriterProviders
	EventBus             *pkgEvents.EventBus
	Registry             *pkgEvents.EventHandlerRegistry
	Subscriber           *watermillKafka.Subscriber
}

func processMessage(ctx context.Context, app *Application, kafkaMessage *watermillMessage.Message) error {
	messageCtx, cancel := context.WithTimeout(ctx, app.Configs.SubscriberConfigs.TimeOut.Duration())
	defer cancel()

	pubsubMessage, err := appWatermill.ConvertWatermillToPubsub(kafkaMessage, nil)
	if err != nil {
		return fmt.Errorf("convert kafka message: %w", err)
	}

	messageLogger := app.Logger.With(
		"message_uuid", kafkaMessage.UUID,
		"event_type", pubsubMessage.Headers.EventType,
		"message_key", pubsubMessage.Headers.Key,
		"topic", app.Configs.SubscriberConfigs.Topic,
	)
	messageCtx = messageLogger.AddToContext(messageCtx, messageLogger)

	messageLogger.Info("Processing Kafka message")

	if err := app.EventBus.Handle(messageCtx, pubsubMessage); err != nil {
		return fmt.Errorf("handle event bus message: %w", err)
	}

	messageLogger.Info("Kafka message processed successfully")
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := initializeApplication(ctx)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	app.Logger.Info("Application initialized",
		"version", version,
		"topic", app.Configs.SubscriberConfigs.Topic,
		"consumer_group", app.Configs.SubscriberConfigs.ConsumerGroup,
		"registered_handlers", app.Registry.GetAllEventTypes(),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- runApplication(ctx, app)
	}()

	select {
	case <-ctx.Done():
		app.Logger.Info("Shutdown signal received")
	case err := <-errCh:
		if err != nil {
			app.Logger.Error("Application stopped with error", "error", err.Error())
		}
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := shutdownApplication(shutdownCtx, app); err != nil {
		app.Logger.Error("Application shutdown finished with errors", "error", err.Error())
		return
	}

	app.Logger.Info("Application shutdown completed")
}

func initializeApplication(ctx context.Context) (*Application, error) {
	appConfigs, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("load configs: %w", err)
	}

	appConfigs.Envs.App.Version = version

	logger, err := appLogger.NewLogrusLogger(
		appConfigs.Envs.App.Name,
		appConfigs.Envs.App.Env,
		appConfigs.Envs.App.LogLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	ctx = logger.AddToContext(ctx, logger)

	serviceProviders, err := providers.NewServiceProviders(appConfigs.Envs, logger)
	if err != nil {
		return nil, fmt.Errorf("create service providers: %w", err)
	}

	transporterProviders, err := providers.NewTransporterProviders(appConfigs.Envs, appConfigs.SubscriberConfigs, serviceProviders)
	if err != nil {
		return nil, fmt.Errorf("create transporter providers: %w", err)
	}

	var (
		writerProviders *providers.WriterProviders
		registry        *pkgEvents.EventHandlerRegistry
	)

	if appConfigs.SubscriberConfigs.Topic == internalCommandsTopic {
		writerProviders, err = providers.NewWriterProviders(appConfigs.Envs, serviceProviders)
		if err != nil {
			return nil, fmt.Errorf("create writer providers: %w", err)
		}
		registry = writerProviders.Registry
	} else {
		registry = transporterProviders.Registry
	}

	eventBus := pkgEvents.NewEventBus(pkgEvents.EventBusDependencies{
		EventHandlerRegistry: registry,
	})

	subscriber, err := createKafkaSubscriber(appConfigs, logger)
	if err != nil {
		return nil, fmt.Errorf("create kafka subscriber: %w", err)
	}

	logger.Info("Bootstrap completed",
		"cluster", appConfigs.SubscriberConfigs.Cluster,
		"dlq_cluster", appConfigs.SubscriberConfigs.DLQCluster,
		"brokers", appConfigs.Envs.Pubsub.DeliveryBrokersHosts,
		"mode_topic", appConfigs.SubscriberConfigs.Topic,
	)

	return &Application{
		Configs:              appConfigs,
		Logger:               logger,
		ServiceProviders:     serviceProviders,
		TransporterProviders: transporterProviders,
		WriterProviders:      writerProviders,
		EventBus:             eventBus,
		Registry:             registry,
		Subscriber:           subscriber,
	}, nil
}

func createKafkaSubscriber(appConfigs *config.AppConfigs, logger appLogger.Logger) (*watermillKafka.Subscriber, error) {
	saramaConfig := watermillKafka.DefaultSaramaSubscriberConfig()
	saramaConfig.ClientID = appConfigs.SubscriberConfigs.ConsumerName
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return watermillKafka.NewSubscriber(
		watermillKafka.SubscriberConfig{
			Brokers:               appConfigs.Envs.Pubsub.DeliveryBrokersHosts,
			ConsumerGroup:         appConfigs.SubscriberConfigs.ConsumerGroup,
			OverwriteSaramaConfig: saramaConfig,
			Unmarshaler:           watermillKafka.DefaultMarshaler{},
		},
		appWatermill.NewWatermillLoggerFromLogger(logger),
	)
}

func runApplication(ctx context.Context, app *Application) error {
	messages, err := app.Subscriber.Subscribe(ctx, app.Configs.SubscriberConfigs.Topic)
	if err != nil {
		return fmt.Errorf("subscribe to topic %s: %w", app.Configs.SubscriberConfigs.Topic, err)
	}

	app.Logger.Info("Kafka consumer started",
		"topic", app.Configs.SubscriberConfigs.Topic,
		"consumer_group", app.Configs.SubscriberConfigs.ConsumerGroup,
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-messages:
			if !ok {
				return nil
			}

			if err := processMessage(ctx, app, msg); err != nil {
				app.Logger.Error("Failed to process Kafka message",
					"error", err.Error(),
					"message_uuid", msg.UUID,
				)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func shutdownApplication(ctx context.Context, app *Application) error {
	if app != nil && app.Subscriber != nil {
		if err := app.Subscriber.Close(); err != nil {
			return fmt.Errorf("close subscriber: %w", err)
		}
	}

	if app != nil && app.WriterProviders != nil {
		if err := app.WriterProviders.Close(ctx); err != nil {
			return fmt.Errorf("close writer providers: %w", err)
		}
	}

	if app == nil || app.ServiceProviders == nil || app.ServiceProviders.MessagePublisher == nil {
		return nil
	}

	if err := app.ServiceProviders.MessagePublisher.Close(ctx); err != nil {
		return fmt.Errorf("close message publisher: %w", err)
	}

	return nil
}
