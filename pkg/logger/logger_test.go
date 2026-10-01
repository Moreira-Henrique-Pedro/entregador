package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLoggerFromContextFallsBackToNoop(t *testing.T) {
	assert.IsType(t, &NoopLogger{}, GetLoggerFromContext(context.Background()))

	ctx := context.WithValue(context.Background(), LoggerContextKey, "not a logger")
	assert.IsType(t, &NoopLogger{}, GetLoggerFromContext(ctx))
}

func TestNoopLogger(t *testing.T) {
	l := NewNoopLogger()

	assert.NotPanics(t, func() {
		l.Info("msg", "k", "v")
		l.Warn("msg")
		l.Error("msg", map[string]any{"k": "v"})
		l.Debug("msg")
		l.Critical("msg")
		l.Fatal("msg")
	})
	assert.Same(t, l, l.With("k", "v"))
	assert.Same(t, l, l.WithFields(map[string]any{"k": "v"}))

	ctx := context.Background()
	assert.Equal(t, ctx, l.AddToContext(ctx, l))
}
