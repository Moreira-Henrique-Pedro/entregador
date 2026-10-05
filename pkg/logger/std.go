package logger

import (
	"log"
	"strings"
)

func NewStdLogger(l Logger) *log.Logger {
	return log.New(stdWriter{logger: l}, "", 0)
}

type stdWriter struct {
	logger Logger
}

func (w stdWriter) Write(p []byte) (int, error) {
	w.logger.Error(strings.TrimSpace(string(p)))
	return len(p), nil
}
