package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimiter wraps redis_rate.Limiter for HTTP middleware with first-party bypass rules.
type RateLimiter struct {
	limiter         *redis_rate.Limiter
	enabled         bool
	publicLimit     int
	window          time.Duration
	allowAllOrigins bool
	allowedOrigins  map[string]struct{}
	allow           allowFunc
}

type allowFunc func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error)

const firstPartyClientHeader = "X-BigFly-Client"

// NewRateLimiter creates a new rate limiter with first-party bypass rules.
// If debugMode is true, rate limiting is disabled.
// publicLimit applies to all non-first-party public requests.
func NewRateLimiter(redisClient *redis.Client, debugMode bool, publicLimit int, window time.Duration, corsAllowedOrigins []string) *RateLimiter {
	var limiter *redis_rate.Limiter
	var allow allowFunc
	if !debugMode && redisClient != nil {
		limiter = redis_rate.NewLimiter(redisClient)
		allow = func(ctx context.Context, key string, limit int, _ time.Duration) (*redis_rate.Result, error) {
			return limiter.Allow(ctx, key, redis_rate.PerMinute(limit))
		}
	}

	allowAllOrigins := false
	allowedOrigins := make(map[string]struct{}, len(corsAllowedOrigins))
	for _, origin := range corsAllowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			allowAllOrigins = true
			continue
		}
		allowedOrigins[trimmed] = struct{}{}
	}

	return &RateLimiter{
		limiter:         limiter,
		enabled:         !debugMode && redisClient != nil,
		publicLimit:     publicLimit,
		window:          window,
		allowAllOrigins: allowAllOrigins,
		allowedOrigins:  allowedOrigins,
		allow:           allow,
	}
}

func (rl *RateLimiter) shouldBypass(r *http.Request) bool {
	if r.Method == http.MethodOptions {
		return true
	}

	clientType := strings.ToLower(strings.TrimSpace(r.Header.Get(firstPartyClientHeader)))
	switch clientType {
	case "mobile":
		return true
	case "web":
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			return true
		}
		if rl.allowAllOrigins {
			return true
		}
		_, ok := rl.allowedOrigins[origin]
		return ok
	default:
		return false
	}
}

func extractClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if first != "" {
			return first
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}

	remote := strings.TrimSpace(r.RemoteAddr)
	if remote == "" {
		return "unknown"
	}

	host, _, err := net.SplitHostPort(remote)
	if err == nil && host != "" {
		return host
	}
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) && addrErr.Err == "missing port in address" {
		return remote
	}
	return remote
}

// Middleware returns an HTTP middleware that rate limits requests with a first-party bypass.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.enabled {
			next.ServeHTTP(w, r)
			return
		}
		if rl.shouldBypass(r) {
			next.ServeHTTP(w, r)
			return
		}

		ip := extractClientIP(r)
		rateLimitKey := fmt.Sprintf("rate:ip:%s", ip)
		limit := rl.publicLimit

		if rl.allow == nil {
			next.ServeHTTP(w, r)
			return
		}

		res, err := rl.allow(r.Context(), rateLimitKey, limit, rl.window)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if res == nil {
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
