package providers

import (
	"context"
	"fmt"
	"net/http"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"github.com/Moreira-Henrique-Pedro/entregador/config"
	pubsubIn "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/in/pubsub"
	kafkaOut "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/out/kafka"
	pubsubOut "github.com/Moreira-Henrique-Pedro/entregador/internal/adapters/out/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/in"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/ports/out"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/watermill"
)

// messaging is the NotificationScheduler adapter chosen by MESSAGING_PROVIDER.
type messaging struct {
	scheduler out.NotificationScheduler
	close     func(ctx context.Context) error
}

func newMessaging(env *config.Environment, log logger.Logger) (*messaging, error) {
	switch env.Messaging.Provider {
	case config.MessagingProviderPubSub:
		return newPubSubMessaging(env)
	case config.MessagingProviderKafka:
		return newKafkaMessaging(env, log)
	default:
		return nil, fmt.Errorf("unknown messaging provider %q (use %q or %q)",
			env.Messaging.Provider, config.MessagingProviderPubSub, config.MessagingProviderKafka)
	}
}

// newPubSubMessaging uses Application Default Credentials (the Cloud Run service account), or
// the emulator when PUBSUB_EMULATOR_HOST is set.
func newPubSubMessaging(env *config.Environment) (*messaging, error) {
	if env.PubSub.ProjectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID is required for the pubsub messaging provider")
	}

	client, err := gcppubsub.NewClient(context.Background(), env.PubSub.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	scheduler := pubsubOut.NewNotificationScheduler(client, env.PubSub.NotificationsTopic)
	return &messaging{
		scheduler: scheduler,
		close: func(context.Context) error {
			scheduler.Stop()
			return client.Close()
		},
	}, nil
}

func newKafkaMessaging(env *config.Environment, log logger.Logger) (*messaging, error) {
	publisher, err := watermill.NewWatermillPublisher[any](env.Kafka.DeliveryBrokersHosts, log)
	if err != nil {
		return nil, fmt.Errorf("create kafka publisher: %w", err)
	}

	return &messaging{
		scheduler: kafkaOut.NewNotificationScheduler(publisher, env.Kafka.CommandsTopic),
		close:     publisher.Close,
	}, nil
}

func newPushHandler(env *config.Environment, log logger.Logger, notifyDelivery in.NotifyDelivery) (http.Handler, error) {
	if !env.PubSub.VerifyPushToken {
		log.Warn("Pub/Sub push token verification is DISABLED: use it only with the emulator")
		return pubsubIn.NewPushHandler(notifyDelivery, nil), nil
	}

	verifier, err := pubsubIn.NewTokenVerifier(env.PubSub.PushAudience, env.PubSub.PushServiceAccount)
	if err != nil {
		return nil, err
	}
	return pubsubIn.NewPushHandler(notifyDelivery, verifier), nil
}
