package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	entry *logrus.Entry
}

func NewLogrusLogger(appName, env, level string) (Logger, error) {
	baseLogger := logrus.New()
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	baseLogger.SetLevel(parsedLevel)
	baseLogger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})

	entry := logrus.NewEntry(baseLogger).WithFields(logrus.Fields{
		"app": appName,
		"env": env,
	})

	return &LogrusLogger{entry: entry}, nil
}

func (l *LogrusLogger) With(fields ...interface{}) Logger {
	return &LogrusLogger{entry: l.entry.WithFields(normalizeFields(fields...))}
}

func (l *LogrusLogger) Info(message string, fields ...interface{}) {
	l.entry.WithFields(normalizeFields(fields...)).Info(message)
}

func (l *LogrusLogger) Warn(message string, fields ...interface{}) {
	l.entry.WithFields(normalizeFields(fields...)).Warn(message)
}

func (l *LogrusLogger) Error(message string, fields ...interface{}) {
	l.entry.WithFields(normalizeFields(fields...)).Error(message)
}

func (l *LogrusLogger) Debug(message string, fields ...interface{}) {
	l.entry.WithFields(normalizeFields(fields...)).Debug(message)
}

func (l *LogrusLogger) Critical(message string, fields ...interface{}) {
	l.entry.WithFields(normalizeFields(fields...)).Error(message)
}

func (l *LogrusLogger) Fatal(message string, fields ...interface{}) {
	l.entry.WithFields(normalizeFields(fields...)).Fatal(message)
}

func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{entry: l.entry.WithFields(logrus.Fields(fields))}
}

func (l *LogrusLogger) AddToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

func normalizeFields(fields ...interface{}) logrus.Fields {
	normalized := logrus.Fields{}
	unnamedIndex := 0

	for index := 0; index < len(fields); index++ {
		switch value := fields[index].(type) {
		case map[string]any:
			for key, fieldValue := range value {
				normalized[key] = fieldValue
			}
		default:
			if key, ok := value.(string); ok && index+1 < len(fields) {
				normalized[key] = fields[index+1]
				index++
				continue
			}

			normalized[fmt.Sprintf("field_%d", unnamedIndex)] = value
			unnamedIndex++
		}
	}

	return normalized
}
