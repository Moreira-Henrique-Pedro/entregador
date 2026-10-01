package logger

import (
	"context"
)

type contextKey string

const LoggerContextKey contextKey = "logger"

type Logger interface {
	With(fields ...interface{}) Logger
	Info(message string, fields ...interface{})
	Warn(message string, fields ...interface{})
	Error(message string, fields ...interface{})
	Debug(message string, fields ...interface{})
	Critical(message string, fields ...interface{})
	Fatal(message string, fields ...interface{})
	WithFields(fields map[string]interface{}) Logger
	AddToContext(ctx context.Context, l Logger) context.Context
}

func GetLoggerFromContext(ctx context.Context) Logger {
	logger := getBaseLoggerFromContext(ctx)
	return logger
}

func getBaseLoggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(Logger); ok {
		return logger
	}
	return NewNoopLogger()
}

type NoopLogger struct{}

func NewNoopLogger() Logger {
	return &NoopLogger{}
}

func (n *NoopLogger) With(fields ...any) Logger {
	return n
}

func (n *NoopLogger) Info(message string, fields ...any) {
}

func (n *NoopLogger) Warn(message string, fields ...any) {
}

func (n *NoopLogger) Error(message string, fields ...any) {
}

func (n *NoopLogger) Debug(message string, fields ...any) {
}

func (n *NoopLogger) Critical(message string, fields ...any) {
}

func (n *NoopLogger) Fatal(message string, fields ...any) {
}

func (n *NoopLogger) WithFields(fields map[string]any) Logger {
	return n
}

func (n *NoopLogger) AddToContext(ctx context.Context, l Logger) context.Context {
	return ctx
}
