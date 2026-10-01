package events

import (
	"errors"
	"fmt"
	"testing"
)

func TestPermanent(t *testing.T) {
	failure := errors.New("no resident")

	err := fmt.Errorf("handle command: %w", Permanent(failure))

	if !IsPermanent(err) {
		t.Error("wrapped permanent error must be permanent")
	}
	if !errors.Is(err, failure) {
		t.Error("permanent error must unwrap to the original error")
	}
	if err.Error() != "handle command: no resident" {
		t.Errorf("Error() = %q", err.Error())
	}
	if IsPermanent(failure) || IsPermanent(nil) {
		t.Error("plain errors must not be permanent")
	}
}
