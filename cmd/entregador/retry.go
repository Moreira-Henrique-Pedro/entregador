package main

import (
	"context"
	"errors"
	"time"

	subscriberConfig "github.com/Moreira-Henrique-Pedro/entregador/config/subscriber"
	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
)

type retryPolicy struct {
	maxRetries      int
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
}

func newRetryPolicy(cfg *subscriberConfig.RetryConfig) retryPolicy {
	return retryPolicy{
		maxRetries:      cfg.MaxRetries,
		initialInterval: cfg.InitialInterval.Duration(),
		maxInterval:     cfg.MaxInterval.Duration(),
		multiplier:      cfg.Multiplier,
	}
}

// run calls fn until it succeeds, returns a permanent error, the retries are
// exhausted or ctx is done, waiting with exponential backoff between attempts.
// It returns the last error.
func (p retryPolicy) run(ctx context.Context, fn func() error, onRetry func(attempt int, err error, wait time.Duration)) error {
	wait := p.initialInterval

	for attempt := 0; ; attempt++ {
		err := fn()
		if err == nil || pkgEvents.IsPermanent(err) || attempt >= p.maxRetries {
			return err
		}

		if onRetry != nil {
			onRetry(attempt+1, err, wait)
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(err, ctx.Err())
		case <-timer.C:
		}

		wait = time.Duration(float64(wait) * p.multiplier)
		if p.maxInterval > 0 && wait > p.maxInterval {
			wait = p.maxInterval
		}
	}
}
