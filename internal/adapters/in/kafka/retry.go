package kafka

import (
	"context"
	"errors"
	"time"

	pkgEvents "github.com/Moreira-Henrique-Pedro/entregador/pkg/events"
)

// RetryPolicy retries a failed message with exponential backoff; permanent errors are not retried.
type RetryPolicy struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

func (p RetryPolicy) run(ctx context.Context, fn func() error, onRetry func(attempt int, err error, wait time.Duration)) error {
	wait := p.InitialInterval

	for attempt := 0; ; attempt++ {
		err := fn()
		if err == nil || pkgEvents.IsPermanent(err) || attempt >= p.MaxRetries {
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

		wait = time.Duration(float64(wait) * p.Multiplier)
		if p.MaxInterval > 0 && wait > p.MaxInterval {
			wait = p.MaxInterval
		}
	}
}
