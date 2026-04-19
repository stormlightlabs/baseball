package config

import (
	"path/filepath"
	"testing"
)

func TestLoad_MissingExplicitConfigPathFallsBackToEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://dbuser:dbpass@dbhost:5432/baseball?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://redis-host:6379/0")

	missingPath := filepath.Join(t.TempDir(), "conf.toml")
	cfg, err := Load(missingPath)
	if err != nil {
		t.Fatalf("Load returned error for missing explicit config path: %v", err)
	}

	if cfg.Database.URL != "postgres://dbuser:dbpass@dbhost:5432/baseball?sslmode=disable" {
		t.Fatalf("expected DATABASE_URL env override, got %q", cfg.Database.URL)
	}
	if cfg.Redis.URL != "redis://redis-host:6379/0" {
		t.Fatalf("expected REDIS_URL env override, got %q", cfg.Redis.URL)
	}
}

func TestLoad_MissingExplicitConfigPathFallsBackToDefaults(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "conf.toml")
	cfg, err := Load(missingPath)
	if err != nil {
		t.Fatalf("Load returned error for missing explicit config path: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default server port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL == "" {
		t.Fatal("expected default database URL to be set")
	}
	if cfg.Redis.URL == "" {
		t.Fatal("expected default redis URL to be set")
	}
}

func TestLoad_ContainerModeRewritesLocalhostRedisURL(t *testing.T) {
	t.Setenv("POSTGRES_DB", "baseball")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")

	missingPath := filepath.Join(t.TempDir(), "conf.toml")
	cfg, err := Load(missingPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Redis.URL != "redis://redis:6379/0" {
		t.Fatalf("expected localhost redis URL to be rewritten, got %q", cfg.Redis.URL)
	}
}

func TestLoad_ContainerModeKeepsNonLocalRedisURL(t *testing.T) {
	t.Setenv("POSTGRES_DB", "baseball")
	t.Setenv("REDIS_URL", "redis://cache.example.internal:6379/1")

	missingPath := filepath.Join(t.TempDir(), "conf.toml")
	cfg, err := Load(missingPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Redis.URL != "redis://cache.example.internal:6379/1" {
		t.Fatalf("expected external redis URL to be kept, got %q", cfg.Redis.URL)
	}
}
