package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetEnvs(t *testing.T) {
	oldEnvs, oldAppName := Envs, AppName
	Envs = nil
	t.Cleanup(func() {
		Envs, AppName = oldEnvs, oldAppName
	})
}

func TestReadEnvsDefaults(t *testing.T) {
	resetEnvs(t)
	for _, key := range []string{"ENVIRONMENT", "LOG_LEVEL", "APP_NAME", "APP_VERSION", "HTTP_PORT", "PORT", "NOTIFIER_PROVIDER", "PUBSUB_NOTIFICATIONS_TOPIC", "PUBSUB_PUSH_VERIFY_TOKEN", "AUTH_ENABLED"} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}

	envs, err := ReadEnvs()

	require.NoError(t, err)
	assert.Equal(t, EnvironmentDevelopment, envs.App.Env)
	assert.Equal(t, "info", envs.App.LogLevel)
	assert.Equal(t, "entregador-api", envs.App.Name)
	assert.Equal(t, "1.0.0", envs.App.Version)
	assert.Equal(t, "8081", envs.HTTP.Port)
	assert.Equal(t, "log", envs.Notifier.Provider)
	assert.Equal(t, "55", envs.Notifier.DefaultCountryCode)
	assert.Equal(t, "delivery-notifications", envs.PubSub.NotificationsTopic)
	assert.True(t, envs.PubSub.VerifyPushToken, "push token check must be on by default")
	assert.True(t, envs.Auth.Enabled, "authentication must be on by default")
	assert.Equal(t, "entregador-api", AppName)
	assert.False(t, envs.IsProduction())
}

func TestReadEnvsFromEnvironment(t *testing.T) {
	resetEnvs(t)
	t.Setenv("ENVIRONMENT", EnvironmentProduction)
	t.Setenv("APP_NAME", "custom-app")
	t.Setenv("GCP_PROJECT_ID", "my-project")
	t.Setenv("MONGODB_URI", "mongodb://localhost")
	t.Setenv("HTTP_PORT", "9000")

	envs, err := ReadEnvs()

	require.NoError(t, err)
	assert.Equal(t, "my-project", envs.PubSub.ProjectID)
	assert.Equal(t, "mongodb://localhost", envs.MongoDB.URI)
	assert.Equal(t, "9000", envs.HTTP.Port)
	assert.Equal(t, "custom-app", AppName)
	assert.True(t, envs.IsProduction())
}

func TestReadEnvsIsCached(t *testing.T) {
	resetEnvs(t)
	t.Setenv("APP_NAME", "first")

	first, err := ReadEnvs()
	require.NoError(t, err)

	t.Setenv("APP_NAME", "second")
	second, err := ReadEnvs()
	require.NoError(t, err)

	assert.Same(t, first, second)
	assert.Equal(t, "first", second.App.Name)
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{env: EnvironmentProduction, want: true},
		{env: EnvironmentDevelopment, want: false},
		{env: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			e := &Environment{}
			e.App.Env = tt.env
			assert.Equal(t, tt.want, e.IsProduction())
		})
	}
}

func TestReadEnvsCloudRunPortWins(t *testing.T) {
	resetEnvs(t)
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("PORT", "8080")

	envs, err := ReadEnvs()

	require.NoError(t, err)
	assert.Equal(t, "8080", envs.HTTP.Port)
}

func TestReadEnvsFromEnvFile(t *testing.T) {
	resetEnvs(t)
	envFile := t.TempDir() + "/.env.custom"
	require.NoError(t, os.WriteFile(envFile, []byte("APP_NAME=from-env-file\nNOTIFIER_PROVIDER=log\n"), 0o600))
	t.Setenv("ENV_FILE", envFile)
	t.Setenv("APP_NAME", "")
	require.NoError(t, os.Unsetenv("APP_NAME"))
	t.Cleanup(func() { _ = os.Unsetenv("NOTIFIER_PROVIDER") })

	envs, err := ReadEnvs()

	require.NoError(t, err)
	assert.Equal(t, "from-env-file", envs.App.Name)
}

func TestReadEnvsMissingEnvFileIsIgnored(t *testing.T) {
	resetEnvs(t)
	t.Setenv("ENV_FILE", t.TempDir()+"/missing")

	_, err := ReadEnvs()

	require.NoError(t, err)
}
