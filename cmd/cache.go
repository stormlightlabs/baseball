package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/config"
	"stormlightlabs.org/baseball/internal/echo"
)

// CacheCmd creates the cache command group
func CacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache inspection and management",
		Long:  "Inspect and manage the Redis cache used by the Baseball API.",
	}

	cmd.AddCommand(CacheStatsCmd())
	cmd.AddCommand(CacheKeysCmd())
	cmd.AddCommand(CacheGetCmd())
	cmd.AddCommand(CacheDeleteCmd())
	cmd.AddCommand(CacheClearCmd())
	return cmd
}

// CacheStatsCmd shows cache statistics for a given pattern
func CacheStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats [pattern]",
		Short: "Show cache statistics",
		Long:  "Display statistics for cache keys matching a pattern (e.g., 'baseball:*:entity:player:*')",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showCacheStats,
	}

	cmd.Flags().StringP("type", "t", "", "Filter by cache type (entity, list, search, upstream)")
	cmd.Flags().StringP("resource", "r", "", "Filter by resource (player, team, game, etc.)")
	return cmd
}

// CacheKeysCmd lists cache keys matching a pattern
func CacheKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys [pattern]",
		Short: "List cache keys",
		Long:  "List all cache keys matching a pattern",
		Args:  cobra.MaximumNArgs(1),
		RunE:  listCacheKeys,
	}

	cmd.Flags().StringP("type", "t", "", "Filter by cache type (entity, list, search, upstream)")
	cmd.Flags().StringP("resource", "r", "", "Filter by resource (player, team, game, etc.)")
	cmd.Flags().IntP("limit", "l", 100, "Maximum number of keys to display")
	return cmd
}

// CacheGetCmd retrieves and displays a cached value
func CacheGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a cached value",
		Long:  "Retrieve and display the value for a specific cache key",
		Args:  cobra.ExactArgs(1),
		RunE:  getCacheValue,
	}
}

// CacheDeleteCmd deletes cache keys
func CacheDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <pattern>",
		Short: "Delete cache keys",
		Long:  "Delete all cache keys matching a pattern. Use with caution!",
		Args:  cobra.ExactArgs(1),
		RunE:  deleteCacheKeys,
	}

	cmd.Flags().BoolP("confirm", "y", false, "Skip confirmation prompt")
	return cmd
}

// CacheClearCmd clears all cache keys
func CacheClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all cache",
		Long:  "Delete ALL cache keys for the application. Use with extreme caution!",
		RunE:  clearCache,
	}

	cmd.Flags().BoolP("confirm", "y", false, "Skip confirmation prompt")
	return cmd
}

func connectToCache(cmd *cobra.Command) (*cache.Client, error) {
	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	redisClient := redis.NewClient(redisOpts)
	ctx := context.Background()

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	env := "dev"
	cacheConfig := cache.Config{
		App:     "baseball",
		Env:     env,
		Version: cfg.Cache.Version,
		Enabled: true, // Always enabled for CLI inspection
		TTLs: cache.TTLConfig{
			Entity:   time.Duration(cfg.Cache.TTLs.Entity) * time.Second,
			List:     time.Duration(cfg.Cache.TTLs.List) * time.Second,
			Search:   time.Duration(cfg.Cache.TTLs.Search) * time.Second,
			Upstream: time.Duration(cfg.Cache.TTLs.Upstream) * time.Second,
			Negative: time.Duration(cfg.Cache.TTLs.Negative) * time.Second,
		},
	}

	return cache.NewClient(redisClient, cacheConfig), nil
}

func buildPattern(cmd *cobra.Command, providedPattern string) string {
	if providedPattern != "" {
		return providedPattern
	}

	cacheType, _ := cmd.Flags().GetString("type")
	resource, _ := cmd.Flags().GetString("resource")

	pattern := "baseball:*:*"
	if cacheType != "" {
		pattern += ":" + cacheType
		if resource != "" {
			pattern += ":" + resource + ":*"
		} else {
			pattern += ":*"
		}
	} else {
		pattern += ":*"
	}

	return pattern
}

func showCacheStats(cmd *cobra.Command, args []string) error {
	echo.Header("Cache Statistics")

	cacheClient, err := connectToCache(cmd)
	if err != nil {
		return err
	}

	pattern := buildPattern(cmd, "")
	if len(args) > 0 {
		pattern = args[0]
	}

	ctx := context.Background()
	stats, err := cacheClient.GetStats(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	echo.Infof("Pattern: %s", pattern)
	echo.Infof("Total Keys: %d", stats.Count)
	echo.Info("")

	if stats.Count == 0 {
		echo.Info("No cache keys found matching pattern")
		return nil
	}

	ttlRanges := map[string]int{
		"< 1m":      0,
		"1m - 5m":   0,
		"5m - 15m":  0,
		"15m - 30m": 0,
		"30m - 1h":  0,
		"> 1h":      0,
		"No expiry": 0,
	}

	for _, ttl := range stats.TTLs {
		switch {
		case ttl < 0:
			ttlRanges["No expiry"]++
		case ttl < time.Minute:
			ttlRanges["< 1m"]++
		case ttl < 5*time.Minute:
			ttlRanges["1m - 5m"]++
		case ttl < 15*time.Minute:
			ttlRanges["5m - 15m"]++
		case ttl < 30*time.Minute:
			ttlRanges["15m - 30m"]++
		case ttl < time.Hour:
			ttlRanges["30m - 1h"]++
		default:
			ttlRanges["> 1h"]++
		}
	}

	echo.Info("TTL Distribution:")
	ranges := []string{"< 1m", "1m - 5m", "5m - 15m", "15m - 30m", "30m - 1h", "> 1h", "No expiry"}
	for _, r := range ranges {
		if count := ttlRanges[r]; count > 0 {
			echo.Infof("  %s: %d keys (%.1f%%)", r, count, float64(count)/float64(stats.Count)*100)
		}
	}

	return nil
}

func listCacheKeys(cmd *cobra.Command, args []string) error {
	echo.Header("Cache Keys")

	cacheClient, err := connectToCache(cmd)
	if err != nil {
		return err
	}

	pattern := buildPattern(cmd, "")
	if len(args) > 0 {
		pattern = args[0]
	}

	limit, _ := cmd.Flags().GetInt("limit")

	ctx := context.Background()
	keys, err := cacheClient.ParsePattern(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	echo.Infof("Pattern: %s", pattern)
	echo.Infof("Found: %d keys", len(keys))
	echo.Info("")

	if len(keys) == 0 {
		echo.Info("No keys found")
		return nil
	}

	sort.Strings(keys)

	displayed := 0
	for _, key := range keys {
		if displayed >= limit {
			echo.Infof("\n... and %d more (use --limit to show more)", len(keys)-displayed)
			break
		}

		ttl, _ := cacheClient.Redis.TTL(ctx, key).Result()
		ttlStr := formatTTL(ttl)

		parts := strings.Split(key, ":")
		if len(parts) >= 4 {
			keyType := parts[3]
			echo.Infof("  [%s] %s (TTL: %s)", keyType, key, ttlStr)
		} else {
			echo.Infof("  %s (TTL: %s)", key, ttlStr)
		}

		displayed++
	}

	return nil
}

func getCacheValue(cmd *cobra.Command, args []string) error {
	key := args[0]

	echo.Header("Cache Value")
	echo.Infof("Key: %s", key)
	echo.Info("")

	cacheClient, err := connectToCache(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()

	val, err := cacheClient.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		echo.Info("Key not found in cache")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get value: %w", err)
	}

	ttl, _ := cacheClient.Redis.TTL(ctx, key).Result()

	echo.Infof("TTL: %s", formatTTL(ttl))
	echo.Info("")
	echo.Info("Value:")
	echo.Info(val)

	return nil
}

func deleteCacheKeys(cmd *cobra.Command, args []string) error {
	pattern := args[0]
	confirm, _ := cmd.Flags().GetBool("confirm")

	echo.Header("Delete Cache Keys")
	echo.Infof("Pattern: %s", pattern)
	echo.Info("")

	cacheClient, err := connectToCache(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()

	keys, err := cacheClient.ParsePattern(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to find keys: %w", err)
	}

	if len(keys) == 0 {
		echo.Info("No keys found matching pattern")
		return nil
	}

	echo.Infof("Found %d keys to delete", len(keys))

	if !confirm {
		echo.Info("")
		echo.Info("Keys to be deleted:")
		for i, key := range keys {
			if i >= 10 {
				echo.Infof("  ... and %d more", len(keys)-i)
				break
			}
			echo.Infof("  - %s", key)
		}
		echo.Info("")
		echo.Info("Run with --confirm to proceed with deletion")
		return nil
	}

	deleted, err := cacheClient.Redis.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	echo.Successf("✓ Deleted %d keys", deleted)
	return nil
}

func clearCache(cmd *cobra.Command, args []string) error {
	confirm, _ := cmd.Flags().GetBool("confirm")

	echo.Header("Clear All Cache")
	echo.Info("")

	if !confirm {
		echo.Info("This will delete ALL cache keys for the baseball application.")
		echo.Info("Run with --confirm to proceed")
		return nil
	}

	cacheClient, err := connectToCache(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()
	pattern := "baseball:*"

	deleted, err := cacheClient.InvalidateByPrefix(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	echo.Successf("✓ Cleared cache: %d keys deleted", deleted)
	return nil
}
