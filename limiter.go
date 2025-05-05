package kyro

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter is a wrapper around the golang.org/x/time/rate.Limiter
// to provide a simple interface for rate limiting.
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a new RateLimiter with the given rate and burst.
// r is the number of events per second. b is the burst size.
func NewRateLimiter(r int, b int) *RateLimiter {
	return &RateLimiter{limiter: rate.NewLimiter(rate.Limit(r), b)}
}

// Wait waits for the rate limiter to allow an event. It blocks until the limiter allows the event
// or the context is cancelled. This function uses context.Background() for simplicity.
func (rl *RateLimiter) Wait() error {
	return rl.limiter.Wait(context.Background())
}
