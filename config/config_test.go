package config

import (
	"flag"
	"os"
	"path/filepath"
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
	for _, key := range []string{"ENVIRONMENT", "LOG_LEVEL", "APP_NAME", "APP_VERSION", "HTTP_PORT", "DLQ_TOPIC", "DELIVERY_BROKER_HOSTS", "NOTIFIER_PROVIDER"} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}

	envs, err := ReadEnvs()

	require.NoError(t, err)
	assert.Equal(t, EnvironmentDevelopment, envs.App.Env)
	assert.Equal(t, "info", envs.App.LogLevel)
	assert.Equal(t, "delivery-subscriber", envs.App.Name)
	assert.Equal(t, "1.0.0", envs.App.Version)
	assert.Equal(t, "8081", envs.HTTP.Port)
	assert.Equal(t, "log", envs.Notifier.Provider)
	assert.Equal(t, "55", envs.Notifier.DefaultCountryCode)
	assert.Equal(t, "delivery-subscriber.dlq", envs.Kafka.DLQTopic)
	assert.Empty(t, envs.Kafka.DeliveryBrokersHosts)
	assert.Equal(t, "delivery-subscriber", AppName)
	assert.False(t, envs.IsProduction())
}

func TestReadEnvsFromEnvironment(t *testing.T) {
	resetEnvs(t)
	t.Setenv("ENVIRONMENT", EnvironmentProduction)
	t.Setenv("APP_NAME", "custom-app")
	t.Setenv("DELIVERY_BROKER_HOSTS", "kafka-1:9092,kafka-2:9092")
	t.Setenv("MONGODB_URI", "mongodb://localhost")
	t.Setenv("HTTP_PORT", "9000")

	envs, err := ReadEnvs()

	require.NoError(t, err)
	assert.Equal(t, []string{"kafka-1:9092", "kafka-2:9092"}, envs.Kafka.DeliveryBrokersHosts)
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

func TestNewConfig(t *testing.T) {
	resetEnvs(t)
	oldArgs, oldFlags := os.Args, flag.CommandLine
	t.Cleanup(func() { os.Args, flag.CommandLine = oldArgs, oldFlags })

	path := filepath.Join(t.TempDir(), "sub.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"app":"a","consumer_group":"g","consumer_name":"n","topic":"t"}`), 0o600))

	t.Run("valid", func(t *testing.T) {
		os.Args = []string{"config.test", "-config", path}
		flag.CommandLine = flag.NewFlagSet("config.test", flag.ContinueOnError)

		cfg, err := NewConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg.Envs)
		require.NotNil(t, cfg.SubscriberConfigs)
		assert.Equal(t, "t", cfg.SubscriberConfigs.Topic)
	})

	t.Run("missing subscriber config", func(t *testing.T) {
		os.Args = []string{"config.test"}
		flag.CommandLine = flag.NewFlagSet("config.test", flag.ContinueOnError)

		cfg, err := NewConfig()

		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "failed to read subscriber config")
	})
}

func TestSplitHosts(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "empty", raw: "", want: []string{}},
		{name: "single host", raw: "kafka:9092", want: []string{"kafka:9092"}},
		{name: "trims spaces", raw: " kafka-1:9092 , kafka-2:9092 ", want: []string{"kafka-1:9092", "kafka-2:9092"}},
		{name: "drops empty items", raw: "kafka-1:9092,,kafka-2:9092,", want: []string{"kafka-1:9092", "kafka-2:9092"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitHosts(tt.raw))
		})
	}
}
