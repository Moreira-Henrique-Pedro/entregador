package providers

import (
	"context"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"github.com/gin-gonic/gin"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/middlewares"
	pubsubAdapter "github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/services/pubsub"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

type pubSub struct {
	client    *gcppubsub.Client
	scheduler *pubsubAdapter.NotificationScheduler
}

func newPubSub(env *config.Environment) (*pubSub, error) {
	if env.PubSub.ProjectID == "" {
		return nil, fmt.Errorf("GCP_PROJECT_ID is required")
	}

	client, err := gcppubsub.NewClient(context.Background(), env.PubSub.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	return &pubSub{
		client:    client,
		scheduler: pubsubAdapter.NewNotificationScheduler(client, env.PubSub.NotificationsTopic),
	}, nil
}

func (p *pubSub) close() error {
	p.scheduler.Stop()
	return p.client.Close()
}

func newPubSubPushAuth(env *config.Environment, log logger.Logger) (gin.HandlerFunc, error) {
	if !env.PubSub.VerifyPushToken {
		log.Warn("Pub/Sub push token verification is DISABLED: use it only with the emulator")
		return nil, nil
	}
	auth, err := middlewares.NewPubSubPushAuth(env.PubSub.PushAudience, env.PubSub.PushServiceAccount)
	if err != nil {
		return nil, err
	}
	return auth.Middleware, nil
}
