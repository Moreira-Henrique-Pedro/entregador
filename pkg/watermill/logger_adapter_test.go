package watermill

import (
	"context"
	"errors"
	"testing"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	watermillLogs "github.com/ThreeDotsLabs/watermill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logCall struct {
	level   string
	message string
	fields  map[string]any
}

type recordingLogger struct {
	calls []logCall
}

func (r *recordingLogger) record(level, msg string, fields []any) {
	m := map[string]any{}
	for i := 0; i+1 < len(fields); i += 2 {
		m[fields[i].(string)] = fields[i+1]
	}
	r.calls = append(r.calls, logCall{level: level, message: msg, fields: m})
}

func (r *recordingLogger) With(fields ...any) logger.Logger               { return r }
func (r *recordingLogger) Info(msg string, fields ...any)                 { r.record("info", msg, fields) }
func (r *recordingLogger) Warn(msg string, fields ...any)                 { r.record("warn", msg, fields) }
func (r *recordingLogger) Error(msg string, fields ...any)                { r.record("error", msg, fields) }
func (r *recordingLogger) Debug(msg string, fields ...any)                { r.record("debug", msg, fields) }
func (r *recordingLogger) Critical(msg string, fields ...any)             { r.record("critical", msg, fields) }
func (r *recordingLogger) Fatal(msg string, fields ...any)                { r.record("fatal", msg, fields) }
func (r *recordingLogger) WithFields(fields map[string]any) logger.Logger { return r }
func (r *recordingLogger) AddToContext(ctx context.Context, l logger.Logger) context.Context {
	return ctx
}

func TestLoggerAdapterForwardsCalls(t *testing.T) {
	failure := errors.New("boom")

	tests := []struct {
		name      string
		log       func(watermillLogs.LoggerAdapter)
		wantLevel string
		want      map[string]any
	}{
		{
			name:      "info",
			log:       func(a watermillLogs.LoggerAdapter) { a.Info("msg", watermillLogs.LogFields{"k": "v"}) },
			wantLevel: "info",
			want:      map[string]any{"k": "v"},
		},
		{
			name:      "debug",
			log:       func(a watermillLogs.LoggerAdapter) { a.Debug("msg", watermillLogs.LogFields{"k": 1}) },
			wantLevel: "debug",
			want:      map[string]any{"k": 1},
		},
		{
			name:      "trace maps to debug",
			log:       func(a watermillLogs.LoggerAdapter) { a.Trace("msg", nil) },
			wantLevel: "debug",
			want:      map[string]any{},
		},
		{
			name:      "error includes error field",
			log:       func(a watermillLogs.LoggerAdapter) { a.Error("msg", failure, watermillLogs.LogFields{"k": "v"}) },
			wantLevel: "error",
			want:      map[string]any{"k": "v", "error": failure},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recordingLogger{}

			tt.log(NewWatermillLoggerFromLogger(rec))

			require.Len(t, rec.calls, 1)
			assert.Equal(t, tt.wantLevel, rec.calls[0].level)
			assert.Equal(t, "msg", rec.calls[0].message)
			assert.Equal(t, tt.want, rec.calls[0].fields)
		})
	}
}

func TestLoggerAdapterWithMergesFields(t *testing.T) {
	rec := &recordingLogger{}
	base := NewWatermillLoggerFromLogger(rec)

	child := base.With(watermillLogs.LogFields{"a": 1, "b": 1}).With(watermillLogs.LogFields{"b": 2})
	child.Info("child", watermillLogs.LogFields{"c": 3})
	base.Info("base", nil)

	require.Len(t, rec.calls, 2)
	assert.Equal(t, map[string]any{"a": 1, "b": 2, "c": 3}, rec.calls[0].fields)
	assert.Empty(t, rec.calls[1].fields)
}

func TestLoggerAdapterWithNoopLogger(t *testing.T) {
	a := NewWatermillLoggerFromLogger(logger.NewNoopLogger())

	assert.NotPanics(t, func() {
		a.Info("msg", nil)
		a.Error("msg", errors.New("x"), nil)
		a.With(watermillLogs.LogFields{"k": "v"}).Debug("msg", nil)
	})
}
