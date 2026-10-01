package subscriber

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalConfig = `{"app":"a","consumer_group":"g","consumer_name":"n","topic":"t"}`

func TestInitializeSubscriberConfigDefaults(t *testing.T) {
	cfg, err := initializeSubscriberConfig([]byte(minimalConfig))

	require.NoError(t, err)
	assert.Equal(t, "a", cfg.App)
	assert.Equal(t, "g", cfg.ConsumerGroup)
	assert.Equal(t, "n", cfg.ConsumerName)
	assert.Equal(t, "t", cfg.Topic)
	assert.Equal(t, 5*time.Second, cfg.TimeOut.Duration())
	assert.Equal(t, "delivery", cfg.Cluster)
	assert.Equal(t, "delivery-dlq", cfg.DLQCluster)
	assert.Equal(t, &RetryConfig{
		MaxRetries:      5,
		InitialInterval: defaultRetryInitialInterval,
		MaxInterval:     defaultRetryMaxInterval,
		Multiplier:      2,
	}, cfg.RetryConfig)
}

func TestInitializeSubscriberConfigExplicitValues(t *testing.T) {
	data := `{
		"app":"a","consumer_group":"g","consumer_name":"n","topic":"t",
		"timeout":"10s","cluster":"c","dlq_cluster":"d",
		"retry":{"max_retries":3,"initial_interval":"2s","max_interval":"20s","multiplier":1.5}
	}`

	cfg, err := initializeSubscriberConfig([]byte(data))

	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.TimeOut.Duration())
	assert.Equal(t, "c", cfg.Cluster)
	assert.Equal(t, "d", cfg.DLQCluster)
	require.NotNil(t, cfg.RetryConfig)
	assert.Equal(t, 3, cfg.RetryConfig.MaxRetries)
	assert.Equal(t, 2*time.Second, cfg.RetryConfig.InitialInterval.Duration())
	assert.Equal(t, 20*time.Second, cfg.RetryConfig.MaxInterval.Duration())
	assert.InDelta(t, 1.5, cfg.RetryConfig.Multiplier, 0)
}

func TestInitializeSubscriberConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "invalid json", data: `{"app":`},
		{name: "null document", data: `null`},
		{name: "missing app", data: `{"consumer_group":"g","consumer_name":"n","topic":"t"}`},
		{name: "missing consumer group", data: `{"app":"a","consumer_name":"n","topic":"t"}`},
		{name: "missing consumer name", data: `{"app":"a","consumer_group":"g","topic":"t"}`},
		{name: "missing topic", data: `{"app":"a","consumer_group":"g","consumer_name":"n"}`},
		{name: "invalid timeout", data: `{"app":"a","consumer_group":"g","consumer_name":"n","topic":"t","timeout":"x"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := initializeSubscriberConfig([]byte(tt.data))
			assert.Nil(t, cfg)
			assert.Error(t, err)
		})
	}
}

func TestDeploymentConfigs(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("deployments", "*.json"))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file) // #nosec G304 -- test fixture path from Glob
			require.NoError(t, err)

			cfg, err := initializeSubscriberConfig(data)

			require.NoError(t, err)
			assert.Equal(t, "delivery-subscriber", cfg.App)
			assert.NotEmpty(t, cfg.Topic)
			assert.Positive(t, cfg.TimeOut.Duration())
			require.NotNil(t, cfg.RetryConfig)
			assert.Positive(t, cfg.RetryConfig.MaxRetries)
		})
	}
}

func withArgs(t *testing.T, args ...string) {
	oldArgs, oldFlags := os.Args, flag.CommandLine
	t.Cleanup(func() {
		os.Args, flag.CommandLine = oldArgs, oldFlags
	})
	os.Args = append([]string{"subscriber.test"}, args...)
	flag.CommandLine = flag.NewFlagSet("subscriber.test", flag.ContinueOnError)
}

func TestRead(t *testing.T) {
	validPath := filepath.Join(t.TempDir(), "valid.json")
	require.NoError(t, os.WriteFile(validPath, []byte(minimalConfig), 0o600))
	invalidPath := filepath.Join(t.TempDir(), "invalid.json")
	require.NoError(t, os.WriteFile(invalidPath, []byte(`{}`), 0o600))

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "valid file", args: []string{"-config", validPath}},
		{name: "missing flag", wantErr: "config path cannot be empty"},
		{name: "missing file", args: []string{"-config", filepath.Join(t.TempDir(), "nope.json")}, wantErr: "cannot read config file"},
		{name: "invalid file", args: []string{"-config", invalidPath}, wantErr: "cannot initialize config file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withArgs(t, tt.args...)

			cfg, err := Read()

			if tt.wantErr != "" {
				assert.Nil(t, cfg)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "t", cfg.Topic)
		})
	}
}
