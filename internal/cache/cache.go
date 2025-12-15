package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Client wraps Redis operations with caching patterns including cache-aside,
// singleflight for stampede protection, and TTL jitter to prevent thundering herd.
type Client struct {
	Redis  *redis.Client // Exported for direct access (e.g., CLI operations)
	sf     singleflight.Group
	config Config
}

// Config defines cache behavior and TTL durations for different cache types.
type Config struct {
	App     string // Application namespace (e.g., "baseball")
	Env     string
	Version string    // Cache schema version for invalidation via version bumping
	Enabled bool      // Controls whether caching is active
	TTLs    TTLConfig // TTLs for different cache types
}

// TTLConfig defines time-to-live durations for different cache categories.
// All durations use jitter to prevent simultaneous expiration (thundering herd).
type TTLConfig struct {
	Entity   time.Duration // Caches for single resource lookups (e.g., GET /players/:id)
	List     time.Duration // Caches for collection queries (e.g., GET /teams?league=AL)
	Search   time.Duration // Search result caches for high-cardinality queries
	Upstream time.Duration // Caches for third-party API proxying (e.g., MLB Stats API)
	Negative time.Duration // Caches for "not found" responses to reduce DB load
}

// DefaultTTLConfig returns the recommended TTL values from the caching strategy doc.
func DefaultTTLConfig() TTLConfig {
	return TTLConfig{
		Entity:   30 * time.Minute,
		List:     60 * time.Second,
		Search:   45 * time.Second,
		Upstream: 120 * time.Second,
		Negative: 30 * time.Second,
	}
}

// NewClient creates a cache client with singleflight support for stampede protection.
func NewClient(redisClient *redis.Client, config Config) *Client {
	return &Client{
		Redis:  redisClient,
		config: config,
	}
}

// buildKey constructs a cache key following the pattern:
// {app}:{env}:{version}:{type}:{identifier}
//
// Example: baseball:prod:v1:entity:player:ruthba01
func (c *Client) buildKey(keyType, identifier string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s",
		c.config.App,
		c.config.Env,
		c.config.Version,
		keyType,
		identifier,
	)
}

// HashParams generates a stable SHA-256 hash from query parameters for cache keys.
// It normalizes parameters by sorting keys lexicographically to ensure consistent hashing.
func HashParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		if v := params[k]; v != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}

	normalized := strings.Join(parts, "&")
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// addJitter adds random jitter (±10%) to TTL to prevent simultaneous expiration.
// This implements the stampede protection strategy from the caching doc.
func addJitter(ttl time.Duration) time.Duration {
	jitterPercent := 0.1
	jitter := time.Duration(float64(ttl) * jitterPercent * (rand.Float64()*2 - 1))
	return ttl + jitter
}

// Get retrieves a value from cache and unmarshals it into dest.
// Returns true if found, false if miss or error (cache failures are non-fatal).
func (c *Client) Get(ctx context.Context, key string, dest any) bool {
	if !c.config.Enabled || c.Redis == nil {
		return false
	}

	data, err := c.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false
	}

	return true
}

// Set stores a value in cache with the specified TTL.
// Adds jitter to TTL to prevent thundering herd.
// Cache write failures are logged but non-fatal.
func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !c.config.Enabled || c.Redis == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}

	ttlWithJitter := addJitter(ttl)
	return c.Redis.Set(ctx, key, data, ttlWithJitter).Err()
}

// Delete removes a key from cache (used for explicit invalidation).
func (c *Client) Delete(ctx context.Context, key string) error {
	if !c.config.Enabled || c.Redis == nil {
		return nil
	}
	return c.Redis.Del(ctx, key).Err()
}

// GetOrCompute implements cache-aside pattern with singleflight.
// If cache miss, calls compute function and stores result.
// Singleflight ensures only one concurrent request computes the value.
func (c *Client) GetOrCompute(ctx context.Context, key string, ttl time.Duration, compute func() (any, error)) (any, error) {
	if !c.config.Enabled || c.Redis == nil {
		return compute()
	}

	var result any
	if c.Get(ctx, key, &result) {
		return result, nil
	}

	val, err, _ := c.sf.Do(key, func() (any, error) {
		if c.Get(ctx, key, &result) {
			return result, nil
		}

		computed, err := compute()
		if err != nil {
			return nil, err
		}

		_ = c.Set(ctx, key, computed, ttl)

		return computed, nil
	})

	return val, err
}
