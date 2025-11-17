package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
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
func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	if rl == nil || rl.client == nil {
		return true
	}

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

// EchoRateLimit returns Echo middleware that applies the provided rate limiter keyed by client IP.
func EchoRateLimit(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil || rl.Allow(c.Request().Context(), c.RealIP()) {
				return next(c)
			}
			return c.NoContent(http.StatusTooManyRequests)
		}
	}
}
