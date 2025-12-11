package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimiter wraps redis_rate.Limiter for HTTP middleware.
type RateLimiter struct {
	limiter *redis_rate.Limiter
	enabled bool
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a new rate limiter. If debugMode is true, rate limiting is disabled.
func NewRateLimiter(redisClient *redis.Client, debugMode bool, limit int, window time.Duration) *RateLimiter {
	var limiter *redis_rate.Limiter
	if !debugMode && redisClient != nil {
		limiter = redis_rate.NewLimiter(redisClient)
	}

	return &RateLimiter{
		limiter: limiter,
		enabled: !debugMode && redisClient != nil,
		limit:   limit,
		window:  window,
	}
}

// Middleware returns an HTTP middleware that rate limits requests by IP address.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: improve for prod
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		ctx := context.Background()
		res, err := rl.limiter.Allow(ctx, fmt.Sprintf("rate:%s", ip), redis_rate.PerMinute(rl.limit))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(res.ResetAfter).Unix()))

		if res.Allowed == 0 {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(res.RetryAfter.Seconds())))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
