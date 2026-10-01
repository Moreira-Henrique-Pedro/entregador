package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBufferedLogger(t *testing.T, level string) (*LogrusLogger, *bytes.Buffer, *int) {
	t.Helper()
	l, err := NewLogrusLogger("app-test", "test", level)
	require.NoError(t, err)

	logrusLogger, ok := l.(*LogrusLogger)
	require.True(t, ok)

	buf := &bytes.Buffer{}
	exitCode := -1
	logrusLogger.entry.Logger.SetOutput(buf)
	logrusLogger.entry.Logger.ExitFunc = func(code int) { exitCode = code }
	return logrusLogger, buf, &exitCode
}

func decodeLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		entries = append(entries, entry)
	}
	return entries
}

func TestNewLogrusLogger(t *testing.T) {
	l, err := NewLogrusLogger("app", "dev", "debug")
	require.NoError(t, err)
	require.NotNil(t, l)

	l, err = NewLogrusLogger("app", "dev", "loud")
	assert.Nil(t, l)
	assert.ErrorContains(t, err, "parse log level")
}

func TestLogrusLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		log       func(Logger)
		wantLevel string
	}{
		{name: "info", log: func(l Logger) { l.Info("msg", "k", "v") }, wantLevel: "info"},
		{name: "warn", log: func(l Logger) { l.Warn("msg", "k", "v") }, wantLevel: "warning"},
		{name: "error", log: func(l Logger) { l.Error("msg", "k", "v") }, wantLevel: "error"},
		{name: "debug", log: func(l Logger) { l.Debug("msg", "k", "v") }, wantLevel: "debug"},
		{name: "critical", log: func(l Logger) { l.Critical("msg", "k", "v") }, wantLevel: "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, buf, _ := newBufferedLogger(t, "debug")

			tt.log(l)

			entries := decodeLines(t, buf)
			require.Len(t, entries, 1)
			assert.Equal(t, tt.wantLevel, entries[0]["level"])
			assert.Equal(t, "msg", entries[0]["msg"])
			assert.Equal(t, "v", entries[0]["k"])
			assert.Equal(t, "app-test", entries[0]["app"])
			assert.Equal(t, "test", entries[0]["env"])
		})
	}
}

func TestLogrusLoggerFatal(t *testing.T) {
	l, buf, exitCode := newBufferedLogger(t, "info")

	assert.NotPanics(t, func() { l.Fatal("bye", "k", "v") })

	entries := decodeLines(t, buf)
	require.Len(t, entries, 1)
	assert.Equal(t, "fatal", entries[0]["level"])
	assert.Equal(t, 1, *exitCode)
}

func TestLogrusLoggerRespectsLevel(t *testing.T) {
	l, buf, _ := newBufferedLogger(t, "warn")

	l.Debug("hidden")
	l.Info("hidden")
	l.Warn("shown")

	entries := decodeLines(t, buf)
	require.Len(t, entries, 1)
	assert.Equal(t, "shown", entries[0]["msg"])
}

func TestLogrusLoggerWithAndWithFields(t *testing.T) {
	l, buf, _ := newBufferedLogger(t, "info")

	child := l.With("request_id", "r-1").WithFields(map[string]any{"user": "u-1"})
	child.Info("child")
	l.Info("parent")

	entries := decodeLines(t, buf)
	require.Len(t, entries, 2)
	assert.Equal(t, "r-1", entries[0]["request_id"])
	assert.Equal(t, "u-1", entries[0]["user"])
	assert.NotContains(t, entries[1], "request_id")
	assert.NotContains(t, entries[1], "user")
}

func TestLogrusLoggerContextRoundTrip(t *testing.T) {
	l, err := NewLogrusLogger("app", "dev", "info")
	require.NoError(t, err)

	ctx := l.AddToContext(context.Background(), l)

	assert.Same(t, l, GetLoggerFromContext(ctx))
}

func TestNormalizeFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []any
		want   logrus.Fields
	}{
		{name: "empty", fields: nil, want: logrus.Fields{}},
		{name: "key value pairs", fields: []any{"a", 1, "b", "two"}, want: logrus.Fields{"a": 1, "b": "two"}},
		{name: "map", fields: []any{map[string]any{"a": 1, "b": 2}}, want: logrus.Fields{"a": 1, "b": 2}},
		{name: "map and pairs", fields: []any{map[string]any{"a": 1}, "b", 2}, want: logrus.Fields{"a": 1, "b": 2}},
		{name: "odd trailing value", fields: []any{"a", 1, "orphan"}, want: logrus.Fields{"a": 1, "field_0": "orphan"}},
		{name: "non string key", fields: []any{42, "x", 7}, want: logrus.Fields{"field_0": 42, "x": 7}},
		{name: "multiple unnamed", fields: []any{1, 2.5}, want: logrus.Fields{"field_0": 1, "field_1": 2.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeFields(tt.fields...))
		})
	}
}
