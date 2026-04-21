package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Cache    CacheConfig
}

// ServerConfig contains server settings
type ServerConfig struct {
	Host      string
	Port      int
	BaseURL   string
	DebugMode bool
	CORS      CORSConfig
	RateLimit RateLimitConfig
}

// CORSConfig contains cross-origin request settings.
type CORSConfig struct {
	AllowedOrigins []string
}

// RateLimitConfig contains HTTP rate-limiting settings.
type RateLimitConfig struct {
	PublicPerMinute int
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	URL string
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	URL string
}

// CacheConfig contains caching behavior settings
type CacheConfig struct {
	Enabled bool
	Version string
	TTLs    CacheTTLConfig
}

// CacheTTLConfig defines TTL durations for different cache types (in seconds)
type CacheTTLConfig struct {
	Entity   int // Single resource lookups (e.g., GET /players/:id)
	List     int // Collection queries (e.g., GET /teams?league=AL)
	Search   int // Search results
	Upstream int // Third-party API proxying (MLB Stats API)
	Negative int // "Not found" responses
}

var globalConfig *Config

// Load reads configuration from the specified file or environment variables.
// If configPath is empty, it defaults to "conf.toml" in the current directory.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("conf")
		v.SetConfigType("toml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.baseball")
		v.AddConfigPath("/etc/baseball")
	}

	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.base_url", "http://localhost:8080/v1/")
	v.SetDefault("server.debug_mode", false)
	v.SetDefault("server.cors_allowed_origins", []string{"http://localhost:5173", "http://127.0.0.1:5173"})
	v.SetDefault("server.rate_limit_public_per_minute", 60)
	v.SetDefault("database.url", "postgres://postgres:postgres@localhost:5432/baseball_dev?sslmode=disable")
	v.SetDefault("redis.url", "redis://localhost:6379/0")

	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.version", "v1")
	v.SetDefault("cache.ttls.entity", 1800)
	v.SetDefault("cache.ttls.list", 60)
	v.SetDefault("cache.ttls.search", 45)
	v.SetDefault("cache.ttls.upstream", 120)
	v.SetDefault("cache.ttls.negative", 30)

	v.AutomaticEnv()
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("redis.url", "REDIS_URL")
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.base_url", "SERVER_BASE_URL")
	v.BindEnv("server.debug_mode", "DEBUG_MODE")
	v.BindEnv("server.cors_allowed_origins", "CORS_ALLOWED_ORIGINS")
	v.BindEnv("server.rate_limit_public_per_minute", "RATE_LIMIT_PUBLIC_PER_MINUTE")
	v.BindEnv("cache.enabled", "CACHE_ENABLED")
	v.BindEnv("cache.version", "CACHE_VERSION")

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "No config file found, using defaults and environment variables\n")
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:      v.GetString("server.host"),
			Port:      v.GetInt("server.port"),
			BaseURL:   v.GetString("server.base_url"),
			DebugMode: v.GetBool("server.debug_mode"),
			CORS: CORSConfig{
				AllowedOrigins: normalizedStringList(v.GetStringSlice("server.cors_allowed_origins"), v.GetString("server.cors_allowed_origins")),
			},
			RateLimit: RateLimitConfig{
				PublicPerMinute: v.GetInt("server.rate_limit_public_per_minute"),
			},
		},
		Database: DatabaseConfig{
			URL: v.GetString("database.url"),
		},
		Redis: RedisConfig{
			URL: v.GetString("redis.url"),
		},
		Cache: CacheConfig{
			Enabled: v.GetBool("cache.enabled"),
			Version: v.GetString("cache.version"),
			TTLs: CacheTTLConfig{
				Entity:   v.GetInt("cache.ttls.entity"),
				List:     v.GetInt("cache.ttls.list"),
				Search:   v.GetInt("cache.ttls.search"),
				Upstream: v.GetInt("cache.ttls.upstream"),
				Negative: v.GetInt("cache.ttls.negative"),
			},
		},
	}

	cfg.Server.RateLimit.PublicPerMinute = positiveOrDefault(cfg.Server.RateLimit.PublicPerMinute, 60)

	if shouldPreferContainerServiceNetworking() {
		cfg.Redis.URL = normalizeRedisURLForContainer(cfg.Redis.URL)
	}

	globalConfig = cfg
	return cfg, nil
}

func normalizedStringList(values []string, raw string) []string {
	entries := make([]string, 0, len(values)+1)
	entries = append(entries, values...)
	if strings.TrimSpace(raw) != "" {
		entries = append(entries, raw)
	}

	result := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		for _, part := range strings.Split(entry, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func positiveOrDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func shouldPreferContainerServiceNetworking() bool {
	// Coolify/runtime deployments usually provide at least one of these.
	return strings.TrimSpace(os.Getenv("COOLIFY_RESOURCE_UUID")) != "" ||
		strings.TrimSpace(os.Getenv("POSTGRES_DB")) != "" ||
		strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD")) != ""
}

func normalizeRedisURLForContainer(redisURL string) string {
	trimmed := strings.TrimSpace(redisURL)
	if trimmed == "" {
		return "redis://redis:6379/0"
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return "redis://redis:6379/0"
	default:
		return trimmed
	}
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded; call config.Load() first")
	}
	return globalConfig
}

// MustLoad loads configuration or panics
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
