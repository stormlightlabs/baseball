package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

func newTestRateLimiter(t *testing.T, allow allowFunc) *RateLimiter {
	t.Helper()
	return &RateLimiter{
		enabled:     true,
		publicLimit: 60,
		window:      time.Minute,
		allowedOrigins: map[string]struct{}{
			"https://bigfly.tech": {},
		},
		allow: allow,
	}
}

func TestRateLimiter_BypassesAllowlistedWebOrigin(t *testing.T) {
	called := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		called = true
		return &redis_rate.Result{Allowed: 1}, nil
	})
	nextCalled := false
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set(firstPartyClientHeader, "web")
	req.Header.Set("Origin", "https://bigfly.tech")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !nextCalled {
		t.Fatal("expected downstream handler to be called")
	}
	if called {
		t.Fatal("expected rate limiter to be bypassed for allowlisted web origin")
	}
}

func TestRateLimiter_DoesNotBypassDisallowedWebOrigin(t *testing.T) {
	called := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		called = true
		return &redis_rate.Result{Allowed: 1}, nil
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set(firstPartyClientHeader, "web")
	req.Header.Set("Origin", "https://evil.example")
	req.RemoteAddr = "203.0.113.4:1234"
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !called {
		t.Fatal("expected rate limiter to run for disallowed web origin")
	}
}

func TestRateLimiter_BypassesWebWithoutOrigin(t *testing.T) {
	called := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		called = true
		return &redis_rate.Result{Allowed: 1}, nil
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set(firstPartyClientHeader, "web")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if called {
		t.Fatal("expected rate limiter to be bypassed for web without origin")
	}
}

func TestRateLimiter_BypassesMobileClient(t *testing.T) {
	called := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		called = true
		return &redis_rate.Result{Allowed: 1}, nil
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set(firstPartyClientHeader, "mobile")
	req.Header.Set("Origin", "https://evil.example")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if called {
		t.Fatal("expected rate limiter to be bypassed for mobile marker")
	}
}

func TestRateLimiter_SkipsOptions(t *testing.T) {
	called := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		called = true
		return &redis_rate.Result{Allowed: 1}, nil
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/v1/meta", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if called {
		t.Fatal("expected OPTIONS to bypass rate limiter")
	}
}

func TestRateLimiter_PublicClientAddsHeaders(t *testing.T) {
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		if key != "rate:ip:203.0.113.4" {
			t.Fatalf("unexpected key: %s", key)
		}
		if limit != 60 {
			t.Fatalf("unexpected limit: %d", limit)
		}
		return &redis_rate.Result{
			Allowed:    1,
			Remaining:  59,
			ResetAfter: 30 * time.Second,
		}, nil
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.RemoteAddr = "203.0.113.4:1234"
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if got := res.Header().Get("X-RateLimit-Limit"); got != "60" {
		t.Fatalf("expected X-RateLimit-Limit=60, got %q", got)
	}
	if got := res.Header().Get("X-RateLimit-Remaining"); got != "59" {
		t.Fatalf("expected X-RateLimit-Remaining=59, got %q", got)
	}
}

func TestRateLimiter_Returns429WhenExceeded(t *testing.T) {
	nextCalled := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		return &redis_rate.Result{
			Allowed:    0,
			Remaining:  0,
			RetryAfter: 12 * time.Second,
			ResetAfter: 12 * time.Second,
		}, nil
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.RemoteAddr = "203.0.113.4:1234"
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if nextCalled {
		t.Fatal("expected request to stop at 429")
	}
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res.Code)
	}
	if got := res.Header().Get("Retry-After"); got != "12" {
		t.Fatalf("expected Retry-After=12, got %q", got)
	}
}

func TestRateLimiter_FailOpenOnAllowError(t *testing.T) {
	nextCalled := false
	rl := newTestRateLimiter(t, func(ctx context.Context, key string, limit int, window time.Duration) (*redis_rate.Result, error) {
		return &redis_rate.Result{}, errors.New("redis unavailable")
	})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !nextCalled {
		t.Fatal("expected fail-open when limiter errors")
	}
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestExtractClientIP_UsesFirstXForwardedForEntry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.8, 203.0.113.9")
	req.Header.Set("X-Real-IP", "203.0.113.5")
	req.RemoteAddr = "203.0.113.4:1234"

	if got := extractClientIP(req); got != "198.51.100.8" {
		t.Fatalf("expected first X-Forwarded-For IP, got %q", got)
	}
}

func TestExtractClientIP_FallsBackToXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("X-Real-IP", "203.0.113.5")
	req.RemoteAddr = "203.0.113.4:1234"

	if got := extractClientIP(req); got != "203.0.113.5" {
		t.Fatalf("expected X-Real-IP, got %q", got)
	}
}

func TestExtractClientIP_FallsBackToRemoteAddrHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.RemoteAddr = "203.0.113.4:1234"

	if got := extractClientIP(req); got != "203.0.113.4" {
		t.Fatalf("expected remote host, got %q", got)
	}
}
