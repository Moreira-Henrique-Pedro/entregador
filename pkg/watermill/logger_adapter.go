package watermill

import (
	appLogger "github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
	watermillLogs "github.com/ThreeDotsLabs/watermill"
)

type loggerAdapter struct {
	logger appLogger.Logger
	fields watermillLogs.LogFields
}

func NewWatermillLoggerFromLogger(logger appLogger.Logger) watermillLogs.LoggerAdapter {
	return &loggerAdapter{logger: logger, fields: watermillLogs.LogFields{}}
}

func (l *loggerAdapter) Error(msg string, err error, fields watermillLogs.LogFields) {
	l.logger.Error(msg, mergeFields(l.fields, fields, watermillLogs.LogFields{"error": err})...)
}

func (l *loggerAdapter) Info(msg string, fields watermillLogs.LogFields) {
	l.logger.Info(msg, mergeFields(l.fields, fields)...)
}

func (l *loggerAdapter) Debug(msg string, fields watermillLogs.LogFields) {
	l.logger.Debug(msg, mergeFields(l.fields, fields)...)
}

func (l *loggerAdapter) Trace(msg string, fields watermillLogs.LogFields) {
	l.logger.Debug(msg, mergeFields(l.fields, fields)...)
}

func (l *loggerAdapter) With(fields watermillLogs.LogFields) watermillLogs.LoggerAdapter {
	return &loggerAdapter{
		logger: l.logger,
		fields: mergeLogFields(l.fields, fields),
	}
}

func mergeFields(base watermillLogs.LogFields, extra ...watermillLogs.LogFields) []any {
	merged := mergeLogFields(base, extra...)
	result := make([]any, 0, len(merged)*2)
	for key, value := range merged {
		result = append(result, key, value)
	}
	return result
}

func mergeLogFields(base watermillLogs.LogFields, extra ...watermillLogs.LogFields) watermillLogs.LogFields {
	merged := watermillLogs.LogFields{}
	for key, value := range base {
		merged[key] = value
	}
	for _, fields := range extra {
		for key, value := range fields {
			merged[key] = value
		}
	}
	return merged
}
