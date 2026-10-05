package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Moreira-Henrique-Pedro/entregador/config"
	"github.com/Moreira-Henrique-Pedro/entregador/internal/infrastructure/middlewares"
	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

func TestNewIdentityProvider_RejectsEmulatorInProduction(t *testing.T) {
	env := &config.Environment{}
	env.App.Env = config.EnvironmentProduction
	env.Firebase.AuthEmulatorHost = "localhost:9099"

	_, err := NewIdentityProvider(context.Background(), env)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "FIREBASE_AUTH_EMULATOR_HOST")
}

func TestNewIdentityProvider_WithEmulator(t *testing.T) {
	t.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")
	env := &config.Environment{}
	env.Firebase.ProjectID = "demo-entregador"
	env.Firebase.AuthEmulatorHost = "localhost:9099"

	provider, err := NewIdentityProvider(context.Background(), env)

	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestFirebaseProjectID(t *testing.T) {
	env := &config.Environment{}
	env.PubSub.ProjectID = "gcp-project"
	assert.Equal(t, "gcp-project", firebaseProjectID(env))

	env.Firebase.ProjectID = "firebase-project"
	assert.Equal(t, "firebase-project", firebaseProjectID(env))
}

func TestNewAuthorizer(t *testing.T) {
	log := logger.NewNoopLogger()

	enabled := &config.Environment{}
	enabled.Auth.Enabled = true
	authorizer, err := newAuthorizer(enabled, log, nil)
	require.NoError(t, err)
	assert.IsType(t, &middlewares.Authorizer{}, authorizer)

	disabled := &config.Environment{}
	authorizer, err = newAuthorizer(disabled, log, nil)
	require.NoError(t, err)
	assert.IsType(t, &middlewares.OpenAuthorizer{}, authorizer)

	disabledInProduction := &config.Environment{}
	disabledInProduction.App.Env = config.EnvironmentProduction
	_, err = newAuthorizer(disabledInProduction, log, nil)
	assert.ErrorContains(t, err, "AUTH_ENABLED=false")
}
