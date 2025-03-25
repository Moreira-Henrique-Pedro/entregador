// Package logging contains the logger that includes trace and span IDs from the context.
package logging

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type contextDelivery string

const loggerDelivery contextDelivery = "logger"
const traceNameDefault = "logging"
const spanNameDefault = "logging"

// Logger is a logger that includes trace and span IDs from the context.
type Logger struct {
	*logrus.Entry
}

// Fields is an alias for logrus.Fields to simplify usage.
type Fields = logrus.Fields

// Level is an alias for logrus.Level to simplify usage.
type Level = logrus.Level

// Log levels.
const (
	ErrorLevel Level = logrus.ErrorLevel
	WarnLevel  Level = logrus.WarnLevel
	InfoLevel  Level = logrus.InfoLevel
)

// FromContext retrieves the Logger from a context.
// Searches for the context log. If it does not exist, it creates a log using the *newContextWithLoggerNoPropagateTrace* function, i.e., without propagating the log to the current context.
// The function only returns the log and does not propagate the trace, nor is it possible to add logs to OpenTelemetry.
func FromContext(ctx *context.Context) *Logger {
	logger, ok := (*ctx).Value(loggerDelivery).(*Logger)
	if !ok {
		newCtx, span := newContextWithLoggerNoPropagateTrace(*ctx)
		defer span.End()
		logger = newCtx.Value(loggerDelivery).(*Logger)
	}
	return logger
}

// FromContextWithSpan retrieves the Logger and Span from a context.
// Searches for the context log. If it does not exist, it creates a log using the *newContextWithLoggerNoPropagateTrace* function, that is, without propagating the log to the current context.
// The function returns the log and the span, but without propagating the trace.
func FromContextWithSpan(ctx context.Context) (*Logger, trace.Span) {
	span := trace.SpanFromContext(ctx)
	logger, ok := ctx.Value(loggerDelivery).(*Logger)
	if !ok {
		newCtx, newSpan := newContextWithLoggerNoPropagateTrace(ctx)
		logger = newCtx.Value(loggerDelivery).(*Logger)
		span = newSpan
	}
	return logger, span
}

// FromContextWithSpanName retrieves the Logger from a context.
// Fetches the context log.
// Creates a new context from the current context, appending the OpenTelemetry span, propagating the trace id from the current context to the new context, if it exists; if it does not exist, it creates a new trace id.
// The function returns the new context, the log, and the span that can be used to append logs to OpenTelemetry.
func FromContextWithSpanName(ctx context.Context, traceName string, spanName string) (context.Context, *Logger, trace.Span) {
	newCtx, span := newContextWithNewSpan(ctx, traceName, spanName)
	logger := newCtx.Value(loggerDelivery).(*Logger)
	return newCtx, logger, span
}

// FromContextWithTrace retrieves the Logger from a context and sets the trace and span IDs.
// Fetch the context log.
// Creates a new context from the current context, adding the OpenTelemetry span.
// If the trace id is provided, it will be propagated to the new context.
// If the trace id provided is empty, it propagates the trace id of the current context, if it exists; if it does not exist, it creates a new trace id.
// The function returns the new context, the log and the span that can be used to add logs to OpenTelemetry.
func FromContextWithTrace(ctx context.Context, traceName string, spanName string, traceID string, spanID string) (context.Context, *Logger, trace.Span) {
	newCtx, span := newContextWithTrace(ctx, traceName, spanName, traceID, spanID)
	logger := newCtx.Value(loggerDelivery).(*Logger)
	return newCtx, logger, span
}

// GetTraceIDFromContext retrieves the trace ID from the logger.
func (l *Logger) GetTraceIDFromContext() (string, bool) {
	traceID, ok := l.Data["trace_id"].(string)
	return traceID, ok
}

// Log logs a message at the Info level.
func (l *Logger) Log(message string) {
	l.Info(message)
}

// LogWithFields logs a message at the Info level with additional fields.
// Deprecated: Use LogInfo instead.
func (l *Logger) LogWithFields(message string, fields Fields) {
	fields["deprecated"] = "use LogInfo instead"
	maskedFields := maskFields(fields)
	marshalStructs(&maskedFields)
	l.WithFields(maskedFields).Info(message)
}

// LogInfo logs a message at the Info level.
func (l *Logger) LogInfo(message string, fields Fields) {
	allFields := addLogCallerInfo(fields)

	maskedFields := maskFields(allFields)
	marshalStructs(&maskedFields)
	l.WithFields(maskedFields).Info(message)
}

// LogError logs a message at the Error level.
func (l *Logger) LogError(message string, fields Fields) {
	allFields := addLogCallerInfo(fields)

	maskedFields := maskFields(allFields)
	marshalStructs(&maskedFields)
	l.WithFields(maskedFields).Error(message)
}

// LogWithLevel logs a message at the specified log level with additional fields.
func (l *Logger) LogWithLevel(level Level, message string, fields Fields) {
	allFields := addLogCallerInfo(fields)

	switch level {
	case logrus.ErrorLevel:
		l.WithFields(allFields).Error(message)
	case logrus.WarnLevel:
		l.WithFields(allFields).Warn(message)
	default:
		l.WithFields(allFields).Info(message)
	}
}

// RetrieveTraceAndSpanIdFromContext retrieves the trace and span IDs from the span context.
func RetrieveTraceAndSpanIdFromContext(ctx context.Context) (string, string) {
	spanCurrent := trace.SpanFromContext(ctx)
	if spanCurrent.SpanContext().IsValid() {
		return spanCurrent.SpanContext().TraceID().String(), spanCurrent.SpanContext().SpanID().String()
	}
	return "", ""
}

// CaptureStackTrace captures the current stack trace.
func CaptureStackTrace() string {
	stack := make([]byte, 1024)
	n := runtime.Stack(stack, true)
	return string(stack[:n])
}

// createSpanContext creates a span context from trace and span IDs.
func createSpanContext(traceID string, spanID string) trace.SpanContext {
	var spanContext trace.SpanContext
	if traceID != "" && spanID != "" {
		traceIDHex, errTrace := trace.TraceIDFromHex(traceID)
		spanIDHex, errSpan := trace.SpanIDFromHex(spanID)
		if errTrace == nil && errSpan == nil {
			spanContext = trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: traceIDHex,
				SpanID:  spanIDHex,
			})
		}
	}
	return spanContext
}

// copySpanContext copies a span context.
func copySpanContext(spanContext trace.SpanContext) trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    spanContext.TraceID(),
		SpanID:     spanContext.SpanID(),
		TraceFlags: spanContext.TraceFlags(),
		TraceState: spanContext.TraceState(),
	})
}

func createTrace(ctx context.Context, traceName string, spanName string, traceID string, spanID string) (context.Context, trace.Span) {
	spanCurrent := trace.SpanFromContext(ctx)
	tracer := otel.Tracer(traceName)
	var spanContext trace.SpanContext
	if spanCurrent.SpanContext().IsValid() {
		spanContext = copySpanContext(spanCurrent.SpanContext())
	} else {
		spanContext = createSpanContext(traceID, spanID)
	}
	if spanContext.IsValid() {
		ctx = trace.ContextWithRemoteSpanContext(ctx, spanContext)
		return tracer.Start(ctx, spanName)
	}
	return tracer.Start(ctx, spanName, trace.WithNewRoot())
}

// createLoggerWithContext creates a logger with trace and span IDs and returns a new context with the logger.
func createLoggerWithContext(ctx context.Context, traceID, spanID string) context.Context {
	entry := logrus.New().WithFields(logrus.Fields{
		"trace_id": traceID,
		"span_id":  spanID,
	})
	entry.Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	logger := &Logger{entry}

	return context.WithValue(ctx, loggerDelivery, logger)
}

// newContextWithLoggerNoPropagateTrace creates a new context with a logger but no created new trace.
func newContextWithLoggerNoPropagateTrace(ctx context.Context) (context.Context, trace.Span) {
	spanCurrent := trace.SpanFromContext(ctx)
	traceID := spanCurrent.SpanContext().TraceID().String()
	spanID := spanCurrent.SpanContext().SpanID().String()

	return createLoggerWithContext(ctx, traceID, spanID), spanCurrent
}

// newContextWithNewSpan creates a new context with a logger that includes trace and span IDs.
func newContextWithNewSpan(ctx context.Context, traceName string, spanName string) (context.Context, trace.Span) {
	if traceName == "" || spanName == "" {
		return newContextWithLoggerNoPropagateTrace(ctx)
	}

	ctx, spanCurrent := createTrace(ctx, traceName, spanName, "", "")
	traceID := spanCurrent.SpanContext().TraceID().String()
	spanID := spanCurrent.SpanContext().SpanID().String()

	return createLoggerWithContext(ctx, traceID, spanID), spanCurrent
}

// newContextWithTrace creates a new context with a logger that includes trace and span IDs.
func newContextWithTrace(ctx context.Context, traceName string, spanName string, traceID string, spanID string) (context.Context, trace.Span) {
	if traceName == "" || spanName == "" || traceID == "" || spanID == "" {
		return newContextWithLoggerNoPropagateTrace(ctx)
	}

	if traceName == "" {
		traceName = traceNameDefault
	}

	if spanName == "" {
		spanName = spanNameDefault
	}

	ctx, spanCurrent := createTrace(ctx, traceName, spanName, traceID, spanID)
	traceID = spanCurrent.SpanContext().TraceID().String()
	spanID = spanCurrent.SpanContext().SpanID().String()

	return createLoggerWithContext(ctx, traceID, spanID), spanCurrent
}

func maskText(text string) string {
	if utf8.RuneCountInString(text) <= 2 {
		return text
	}
	first := string([]rune(text)[0])
	last := string([]rune(text)[utf8.RuneCountInString(text)-1])
	return first + strings.Repeat("*", utf8.RuneCountInString(text)-2) + last
}

func maskFields(fields Fields) Fields {
	maskedFields := Fields{}
	for key, value := range fields {
		if key == "name" || key == "email" {
			if strVal, ok := value.(string); ok {
				maskedFields[key] = maskText(strVal)
			} else {
				maskedFields[key] = value
			}
		} else if nestedFields, ok := value.(Fields); ok {
			// If value is a nested struct, calls recursively
			maskedFields[key] = maskFields(nestedFields)
		} else {
			maskedFields[key] = value
		}
	}
	return maskedFields
}

func marshalStructs(fields *Fields) {
	for key, value := range *fields {
		if value == nil {
			continue
		}
		if reflect.TypeOf(value).Kind() == reflect.Struct {
			jsonBytes, err := json.Marshal(value)
			if err == nil {
				(*fields)[key] = string(jsonBytes)
			} else {
				log.Printf("Logging, error serializing struct: '%s': %v", key, err)
			}
		}
	}
}

func addLogCallerInfo(fields Fields) Fields {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return fields
	}

	originFile := file[strings.LastIndex(file, "/")+1:] + ":" + strconv.Itoa(line)
	originFunction := runtime.FuncForPC(pc).Name()
	fmtOriginFunction := originFunction[strings.LastIndex(originFunction, ".")+1:]

	originFields := Fields{
		"file":     originFile,
		"function": fmtOriginFunction,
	}

	for k, v := range fields {
		originFields[k] = v
	}

	return originFields
}
