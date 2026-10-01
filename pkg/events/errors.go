package events

import "errors"

// permanentError marks failures that will not succeed on retry, such as an
// unparseable payload or a command that breaks a business rule.
type permanentError struct {
	err error
}

func (e *permanentError) Error() string { return e.err.Error() }
func (e *permanentError) Unwrap() error { return e.err }

// Permanent wraps err so the consumer sends the message to the DLQ without retrying.
func Permanent(err error) error {
	return &permanentError{err: err}
}

func IsPermanent(err error) bool {
	var target *permanentError
	return errors.As(err, &target)
}
