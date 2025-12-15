package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPCacheEntry stores cached HTTP responses with headers for conditional revalidation.
// Implements RFC 9111 HTTP caching semantics including ETag and Last-Modified support.
type HTTPCacheEntry struct {
	Body         []byte // Response body
	Status       int    // HTTP status code
	ETag         string // Headers for conditional revalidation
	LastModified string
	CacheControl string
	Vary         string
	CachedAt     time.Time // Timestamp when cached
}

// CacheHTTPResponse stores an HTTP response in cache with headers for conditional revalidation.
// Extracts ETag, Last-Modified, and Cache-Control headers for RFC 9111 compliance.
func (c *Client) CacheHTTPResponse(ctx context.Context, key string, resp *http.Response, body []byte, ttl time.Duration) error {
	entry := HTTPCacheEntry{
		Body:         body,
		Status:       resp.StatusCode,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		CacheControl: resp.Header.Get("Cache-Control"),
		Vary:         resp.Header.Get("Vary"),
		CachedAt:     time.Now(),
	}

	return c.Set(ctx, key, entry, ttl)
}

// GetHTTPCache retrieves a cached HTTP response.
// Returns the cache entry and a boolean indicating cache hit.
func (c *Client) GetHTTPCache(ctx context.Context, key string) (*HTTPCacheEntry, bool) {
	var entry HTTPCacheEntry
	if c.Get(ctx, key, &entry) {
		return &entry, true
	}
	return nil, false
}

// AddConditionalHeaders adds If-None-Match and If-Modified-Since headers to a request for conditional revalidation (RFC 9111 section 4.3.2).
// Returns true if conditional headers were added.
func (c *Client) AddConditionalHeaders(ctx context.Context, key string, req *http.Request) bool {
	entry, ok := c.GetHTTPCache(ctx, key)
	if !ok {
		return false
	}

	if entry.ETag != "" {
		req.Header.Set("If-None-Match", entry.ETag)
	}

	if entry.LastModified != "" {
		req.Header.Set("If-Modified-Since", entry.LastModified)
	}

	return entry.ETag != "" || entry.LastModified != ""
}

// RefreshHTTPCache updates the TTL of a cached response without re-fetching the body.
// Used when receiving 304 Not Modified responses to extend cache lifetime.
func (c *Client) RefreshHTTPCache(ctx context.Context, key string, ttl time.Duration) error {
	entry, ok := c.GetHTTPCache(ctx, key)
	if !ok {
		return fmt.Errorf("cache entry not found for key: %s", key)
	}

	return c.Set(ctx, key, entry, ttl)
}

// NegativeCacheEntry stores negative responses (404, 429, 5xx) with short TTLs
// to prevent repeated upstream requests for non-existent resources or during outages.
type NegativeCacheEntry struct {
	Status     int
	Message    string
	RetryAfter string // From 429 Retry-After header
	CachedAt   time.Time
}

// CacheNegativeResponse stores a negative HTTP response (4xx/5xx) with a short TTL.
// Helps reduce load on upstream APIs during outages or rate limiting.
func (c *Client) CacheNegativeResponse(ctx context.Context, key string, status int, message string, retryAfter string) error {
	entry := NegativeCacheEntry{
		Status:     status,
		Message:    message,
		RetryAfter: retryAfter,
		CachedAt:   time.Now(),
	}

	return c.Set(ctx, key, entry, c.config.TTLs.Negative)
}

// GetNegativeCache retrieves a cached negative response.
func (c *Client) GetNegativeCache(ctx context.Context, key string) (*NegativeCacheEntry, bool) {
	var entry NegativeCacheEntry
	if c.Get(ctx, key, &entry) {
		return &entry, true
	}
	return nil, false
}

// ParseCacheControlMaxAge extracts max-age directive from Cache-Control header.
// Returns 0 if not present or invalid. Use as upper bound for cache TTL.
func ParseCacheControlMaxAge(cacheControl string) time.Duration {
	var maxAge int
	_, err := fmt.Sscanf(cacheControl, "max-age=%d", &maxAge)
	if err != nil {
		for _, directive := range []string{
			"max-age=%d",
			"public, max-age=%d",
			"private, max-age=%d",
			"s-maxage=%d",
		} {
			if _, err := fmt.Sscanf(cacheControl, directive, &maxAge); err == nil {
				break
			}
		}
	}

	if maxAge <= 0 {
		return 0
	}
	return time.Duration(maxAge) * time.Second
}

// UpstreamCacheConfig defines caching behavior for third-party API proxying.
type UpstreamCacheConfig struct {
	RespectCacheControl           bool          // Enables RFC 9111 Cache-Control directive parsing
	MaxTTL                        time.Duration // The upper bound for cache TTL (safety cap)
	DefaultTTL                    time.Duration // Used when upstream doesn't provide caching directives
	EnableConditionalRevalidation bool          // Enables 304 Not Modified support
	CacheNegativeResponses        bool          // Caches 4xx/5xx with short TTL
}

// DefaultUpstreamConfig returns recommended settings for third-party API proxying.
func DefaultUpstreamConfig() UpstreamCacheConfig {
	return UpstreamCacheConfig{
		RespectCacheControl:           true,
		MaxTTL:                        5 * time.Minute,
		DefaultTTL:                    2 * time.Minute,
		EnableConditionalRevalidation: true,
		CacheNegativeResponses:        true,
	}
}

// DetermineTTL calculates the TTL for an upstream response based on Cache-Control headers
// and configuration policy. Returns the minimum of max-age and configured MaxTTL.
func (cfg UpstreamCacheConfig) DetermineTTL(resp *http.Response) time.Duration {
	if !cfg.RespectCacheControl {
		return cfg.DefaultTTL
	}

	cacheControl := resp.Header.Get("Cache-Control")
	if maxAge := ParseCacheControlMaxAge(cacheControl); maxAge > 0 {
		if maxAge > cfg.MaxTTL {
			return cfg.MaxTTL
		}
		return maxAge
	}

	return cfg.DefaultTTL
}

// UpstreamCacheMetrics tracks cache performance for monitoring.
type UpstreamCacheMetrics struct {
	Hits                 int64
	Misses               int64
	ConditionalRefreshes int64 // 304 responses
	Errors               int64
}

// MarshalJSON implements json.Marshaler for metrics export.
func (m *UpstreamCacheMetrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]int64{
		"hits":                  m.Hits,
		"misses":                m.Misses,
		"conditional_refreshes": m.ConditionalRefreshes,
		"errors":                m.Errors,
		"hit_rate":              m.HitRate(),
	})
}

// HitRate calculates cache hit ratio as percentage.
func (m *UpstreamCacheMetrics) HitRate() int64 {
	total := m.Hits + m.Misses
	if total == 0 {
		return 0
	}
	return (m.Hits * 100) / total
}
