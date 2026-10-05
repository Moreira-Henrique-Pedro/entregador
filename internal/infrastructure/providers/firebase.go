package providers

import (
	"context"
	"errors"
	"fmt"

	firebasesdk "firebase.google.com/go/v4"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/application/controllers"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/middlewares"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/services/firebase"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

func NewIdentityProvider(ctx context.Context, env *config.Environment) (*firebase.IdentityProvider, error) {
	if env.IsProduction() && env.Firebase.AuthEmulatorHost != "" {
		return nil, errors.New("FIREBASE_AUTH_EMULATOR_HOST must not be set in production: the emulator accepts unsigned tokens")
	}

	app, err := firebasesdk.NewApp(ctx, &firebasesdk.Config{ProjectID: firebaseProjectID(env)})
	if err != nil {
		return nil, fmt.Errorf("create firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("create firebase auth client: %w", err)
	}

	return firebase.NewIdentityProvider(client), nil
}

func firebaseProjectID(env *config.Environment) string {
	if env.Firebase.ProjectID != "" {
		return env.Firebase.ProjectID
	}
	return env.PubSub.ProjectID
}

func newAuthorizer(env *config.Environment, log logger.Logger, identityProvider *firebase.IdentityProvider) (controllers.Authorizer, error) {
	if env.Auth.Enabled {
		return middlewares.NewAuthorizer(identityProvider), nil
	}
	if env.IsProduction() {
		return nil, errors.New("AUTH_ENABLED=false is not allowed in production")
	}
	log.Warn("Authentication is DISABLED: every /v1 request is accepted. Use it only for local development")
	return middlewares.NewOpenAuthorizer(), nil
}
