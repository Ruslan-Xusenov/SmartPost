package telegram

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter enforces Telegram API rate limits (30 messages/second).
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a rate limiter capped at 30 requests/second.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(30), 30),
	}
}

// Wait blocks until the rate limiter allows the next request.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}
