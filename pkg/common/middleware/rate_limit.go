package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter is a Redis-backed fixed-window limiter.
type RateLimiter struct {
	client *redis.Client
	limit  int
	prefix string
	window time.Duration
}

// NewRateLimiter constructs a Redis-backed limiter with per-window limit (uses 1s window).
func NewRateLimiter(client *redis.Client, rps, burst int) *RateLimiter {
	limit := burst
	if limit <= 0 {
		limit = 1
	}
	if client == nil {
		return nil
	}
	return &RateLimiter{
		client: client,
		limit:  limit,
		prefix: "rate",
		window: time.Second,
	}
}

// Allow increments the request counter for the key and tells whether the call is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	if rl == nil || rl.client == nil {
		return true
	}

	ctx := context.Background()
	cacheKey := rl.prefix + ":" + key
	pipe := rl.client.TxPipeline()
	incr := pipe.Incr(ctx, cacheKey)
	pipe.Expire(ctx, cacheKey, rl.window)
	if _, err := pipe.Exec(ctx); err != nil {
		// Fail-open on Redis errors to avoid breaking the API in demos.
		return true
	}
	return incr.Val() <= int64(rl.limit)
}
