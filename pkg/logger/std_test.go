package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorSpy struct {
	NoopLogger
	messages []string
}

func (s *errorSpy) Error(message string, _ ...interface{}) {
	s.messages = append(s.messages, message)
}

func TestNewStdLogger(t *testing.T) {
	spy := &errorSpy{}

	NewStdLogger(spy).Printf("http: TLS handshake error from %s", "1.2.3.4")

	assert.Equal(t, []string{"http: TLS handshake error from 1.2.3.4"}, spy.messages)
}
