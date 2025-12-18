package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimiter wraps redis_rate.Limiter for HTTP middleware with two-tier limits.
type RateLimiter struct {
	limiter              *redis_rate.Limiter
	enabled              bool
	authenticatedLimit   int
	unauthenticatedLimit int
	window               time.Duration
}

// NewRateLimiter creates a new rate limiter with two-tier limits.
// If debugMode is true, rate limiting is disabled.
// authenticatedLimit applies to requests with API keys.
// unauthenticatedLimit applies to anonymous requests (for docs/sampling).
func NewRateLimiter(redisClient *redis.Client, debugMode bool, authenticatedLimit, unauthenticatedLimit int, window time.Duration) *RateLimiter {
	var limiter *redis_rate.Limiter
	if !debugMode && redisClient != nil {
		limiter = redis_rate.NewLimiter(redisClient)
	}

	return &RateLimiter{
		limiter:              limiter,
		enabled:              !debugMode && redisClient != nil,
		authenticatedLimit:   authenticatedLimit,
		unauthenticatedLimit: unauthenticatedLimit,
		window:               window,
	}
}

// Middleware returns an HTTP middleware that rate limits requests with two-tier limits.
// Authenticated requests (with API key) get higher limits per key.
// Unauthenticated requests get lower limits per IP (for docs/sampling).
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.enabled {
			next.ServeHTTP(w, r)
			return
		}

		apiKeyVal := r.Context().Value("api_key")
		var rateLimitKey string
		var limit int

		if apiKeyVal != nil {
			rateLimitKey = fmt.Sprintf("rate:key:%v", apiKeyVal)
			limit = rl.authenticatedLimit
		} else {
			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = xff
			}
			rateLimitKey = fmt.Sprintf("rate:ip:%s", ip)
			limit = rl.unauthenticatedLimit
		}

		ctx := context.Background()
		res, err := rl.limiter.Allow(ctx, rateLimitKey, redis_rate.PerMinute(limit))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
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
