package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// KeyType represents different categories of cached data.
type KeyType string

const (
	KeyTypeEntity   KeyType = "entity"
	KeyTypeList     KeyType = "list"
	KeyTypeSearch   KeyType = "search"
	KeyTypeUpstream KeyType = "upstream"
)

// EntityKey builds a cache key for a single entity lookup.
// Format: {app}:{env}:{version}:entity:{resource}:{id}
// Example: baseball:prod:v1:entity:player:ruthba01
func (c *Client) EntityKey(resource, id string) string {
	identifier := fmt.Sprintf("%s:%s", resource, id)
	return c.buildKey(string(KeyTypeEntity), identifier)
}

// ListKey builds a cache key for collection queries with normalized parameters.
// Format: {app}:{env}:{version}:list:{resource}:{hash}
// Example: baseball:prod:v1:list:teams:sha256(league=AL&sort=name)
func (c *Client) ListKey(resource string, params map[string]string) string {
	hash := HashParams(params)
	identifier := fmt.Sprintf("%s:%s", resource, hash)
	return c.buildKey(string(KeyTypeList), identifier)
}

// SearchKey builds a cache key for search queries with normalized query string.
// Format: {app}:{env}:{version}:search:{hash}
// Example: baseball:prod:v1:search:sha256(q=ruth&year=1927&type=players)
func (c *Client) SearchKey(params map[string]string) string {
	hash := HashParams(params)
	return c.buildKey(string(KeyTypeSearch), hash)
}

// UpstreamKey builds a cache key for third-party API responses.
// Format: {app}:{env}:{version}:upstream:{method}:{host}:{hash}
// Example: baseball:prod:v1:upstream:GET:statsapi.mlb.com:sha256(path?query)
func (c *Client) UpstreamKey(method, host, pathAndQuery string) string {
	hash := HashParams(map[string]string{"url": pathAndQuery})
	identifier := fmt.Sprintf("%s:%s:%s", method, host, hash)
	return c.buildKey(string(KeyTypeUpstream), identifier)
}

// NormalizeFilterParams converts common filter fields to a normalized parameter map.
// Drops default values to prevent duplicate cache keys.
func NormalizeFilterParams(params map[string]any) map[string]string {
	normalized := make(map[string]string)

	for key, val := range params {
		if val == nil {
			continue
		}

		switch v := val.(type) {
		case string:
			if v != "" {
				normalized[key] = v
			}
		case int:
			if (key == "page" && v == 1) || (key == "per_page" && v == 0) {
				continue
			}
			normalized[key] = strconv.Itoa(v)
		case *int:
			if v != nil {
				normalized[key] = strconv.Itoa(*v)
			}
		case *string:
			if v != nil && *v != "" {
				normalized[key] = *v
			}
		case bool:
			normalized[key] = strconv.FormatBool(v)
		case *bool:
			if v != nil {
				normalized[key] = strconv.FormatBool(*v)
			}
		}
	}

	return normalized
}

// ParsePattern extracts keys matching a glob pattern (e.g., "baseball:prod:v1:entity:player:*")
// Returns matching keys for bulk operations. Use sparingly in production.
func (c *Client) ParsePattern(ctx context.Context, pattern string) ([]string, error) {
	if !c.config.Enabled || c.Redis == nil {
		return nil, nil
	}

	var keys []string
	iter := c.Redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan keys: %w", err)
	}

	return keys, nil
}

// Stats returns cache statistics for a given key pattern.
type Stats struct {
	Keys  []string
	Count int
	TTLs  map[string]time.Duration // Key -> remaining TTL
}

// GetStats retrieves statistics for keys matching a pattern.
// Useful for cache inspection and debugging via CLI.
func (c *Client) GetStats(ctx context.Context, pattern string) (*Stats, error) {
	keys, err := c.ParsePattern(ctx, pattern)
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		Keys:  keys,
		Count: len(keys),
		TTLs:  make(map[string]time.Duration),
	}

	for _, key := range keys {
		ttl, err := c.Redis.TTL(ctx, key).Result()
		if err == nil {
			stats.TTLs[key] = ttl
		}
	}

	return stats, nil
}

// KeyPrefix returns the full prefix for a given key type and resource.
// Useful for building scan patterns.
func (c *Client) KeyPrefix(keyType KeyType, resource string) string {
	if resource == "" {
		return fmt.Sprintf("%s:%s:%s:%s", c.config.App, c.config.Env, c.config.Version, keyType)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", c.config.App, c.config.Env, c.config.Version, keyType, resource)
}

// InvalidateByPrefix deletes all keys matching a prefix pattern.
// Use with caution in production - prefer version bumping for bulk invalidation.
func (c *Client) InvalidateByPrefix(ctx context.Context, prefix string) (int, error) {
	if !c.config.Enabled || c.Redis == nil {
		return 0, nil
	}

	pattern := prefix + "*"
	keys, err := c.ParsePattern(ctx, pattern)
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0, nil
	}

	deleted, err := c.Redis.Del(ctx, keys...).Result()
	return int(deleted), err
}
