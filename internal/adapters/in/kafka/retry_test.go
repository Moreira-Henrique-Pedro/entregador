package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
)

func TestRetryPolicy(t *testing.T) {
	policy := RetryPolicy{MaxRetries: 3, InitialInterval: time.Millisecond, MaxInterval: 2 * time.Millisecond, Multiplier: 2}
	failure := errors.New("boom")

	t.Run("succeeds after transient failures", func(t *testing.T) {
		calls := 0
		err := policy.run(context.Background(), func() error {
			calls++
			if calls < 3 {
				return failure
			}
			return nil
		}, nil)
		if err != nil || calls != 3 {
			t.Fatalf("err = %v, calls = %d, want nil and 3", err, calls)
		}
	})

	t.Run("gives up after max retries", func(t *testing.T) {
		calls := 0
		var waits []time.Duration
		err := policy.run(context.Background(), func() error {
			calls++
			return failure
		}, func(_ int, _ error, wait time.Duration) { waits = append(waits, wait) })
		if !errors.Is(err, failure) || calls != 4 {
			t.Fatalf("err = %v, calls = %d, want boom and 4", err, calls)
		}
		want := []time.Duration{time.Millisecond, 2 * time.Millisecond, 2 * time.Millisecond}
		for i := range want {
			if waits[i] != want[i] {
				t.Errorf("wait[%d] = %v, want %v", i, waits[i], want[i])
			}
		}
	})

	t.Run("does not retry permanent errors", func(t *testing.T) {
		calls := 0
		err := policy.run(context.Background(), func() error {
			calls++
			return pkgEvents.Permanent(failure)
		}, nil)
		if !pkgEvents.IsPermanent(err) || !errors.Is(err, failure) || calls != 1 {
			t.Fatalf("err = %v, calls = %d, want permanent boom and 1", err, calls)
		}
	})

	t.Run("stops when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		slow := RetryPolicy{MaxRetries: 3, InitialInterval: time.Hour, Multiplier: 2}
		err := slow.run(ctx, func() error { return failure }, nil)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
	})
}
