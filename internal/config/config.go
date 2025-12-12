package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	OAuth    OAuthConfig
}

// ServerConfig contains server settings
type ServerConfig struct {
	Host      string
	Port      int
	DebugMode bool
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	URL string
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	URL string
}

// OAuthConfig contains OAuth provider settings
type OAuthConfig struct {
	GitHub   OAuthProvider
	Codeberg OAuthProvider
}

// OAuthProvider represents an OAuth provider configuration
type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	Enabled      bool
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
	v.SetDefault("server.debug_mode", false)
	v.SetDefault("database.url", "postgres://postgres:postgres@localhost:5432/baseball_dev?sslmode=disable")
	v.SetDefault("redis.url", "redis://localhost:6379/0")

	v.AutomaticEnv()
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("redis.url", "REDIS_URL")
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.debug_mode", "DEBUG_MODE")
	v.BindEnv("oauth.github.client_id", "GITHUB_CLIENT_ID")
	v.BindEnv("oauth.github.client_secret", "GITHUB_CLIENT_SECRET")
	v.BindEnv("oauth.codeberg.client_id", "CODEBERG_CLIENT_ID")
	v.BindEnv("oauth.codeberg.client_secret", "CODEBERG_CLIENT_SECRET")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "No config file found, using defaults and environment variables\n")
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:      v.GetString("server.host"),
			Port:      v.GetInt("server.port"),
			DebugMode: v.GetBool("server.debug_mode"),
		},
		Database: DatabaseConfig{
			URL: v.GetString("database.url"),
		},
		Redis: RedisConfig{
			URL: v.GetString("redis.url"),
		},
		OAuth: OAuthConfig{
			GitHub: OAuthProvider{
				ClientID:     v.GetString("oauth.github.client_id"),
				ClientSecret: v.GetString("oauth.github.client_secret"),
				CallbackURL:  v.GetString("oauth.github.callback_url"),
				Enabled:      v.GetString("oauth.github.client_id") != "",
			},
			Codeberg: OAuthProvider{
				ClientID:     v.GetString("oauth.codeberg.client_id"),
				ClientSecret: v.GetString("oauth.codeberg.client_secret"),
				CallbackURL:  v.GetString("oauth.codeberg.callback_url"),
				Enabled:      v.GetString("oauth.codeberg.client_id") != "",
			},
		},
	}

	globalConfig = cfg
	return cfg, nil
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
