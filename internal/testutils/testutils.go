// Package testutils provides testing utilities including testcontainers setup
// and fixture data generation for integration tests.
package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresConfig holds configuration for creating a Postgres testcontainer.
type PostgresConfig struct {
	Database       string
	Username       string
	Password       string
	Image          string
	MigrationsPath string
}

// PostgresOption is a functional option for configuring PostgresConfig.
type PostgresOption func(*PostgresConfig)

// PostgresContainer wraps a testcontainers Postgres instance with helper methods.
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DB        *sql.DB
	ConnStr   string
}

// NewPostgresContainer creates and starts a new Postgres testcontainer.
// It runs the schema migrations and returns a connection to the database.
func NewPostgresContainer(ctx context.Context, opts ...PostgresOption) (*PostgresContainer, error) {
	config := &PostgresConfig{
		Database: "baseball_test",
		Username: "postgres",
		Password: "postgres",
		Image:    "postgres:16-alpine",
	}

	for _, opt := range opts {
		opt(config)
	}

	container, err := postgres.Run(ctx,
		config.Image,
		postgres.WithDatabase(config.Database),
		postgres.WithUsername(config.Username),
		postgres.WithPassword(config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pc := &PostgresContainer{
		Container: container,
		DB:        db,
		ConnStr:   connStr,
	}

	if config.MigrationsPath != "" {
		if err := pc.RunMigrations(ctx, config.MigrationsPath); err != nil {
			pc.Terminate(ctx)
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return pc, nil
}

// Terminate stops and removes the container.
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// RunMigrations executes SQL migration files from the given directory.
func (c *PostgresContainer) RunMigrations(ctx context.Context, migrationsPath string) error {
	migrations, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	for _, migration := range migrations {
		content, err := os.ReadFile(migration)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migration, err)
		}

		if _, err := c.DB.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration, err)
		}
	}

	return nil
}

// Seed loads fixture data from a SQL file.
func (c *PostgresContainer) Seed(ctx context.Context, fixturePath string) error {
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture file: %w", err)
	}

	if _, err := c.DB.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}

// Truncate removes all data from the specified tables.
func (c *PostgresContainer) Truncate(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := c.DB.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}

// WithDatabase sets the database name.
func WithDatabase(database string) PostgresOption {
	return func(c *PostgresConfig) {
		c.Database = database
	}
}

// WithUsername sets the username.
func WithUsername(username string) PostgresOption {
	return func(c *PostgresConfig) {
		c.Username = username
	}
}

// WithPassword sets the password.
func WithPassword(password string) PostgresOption {
	return func(c *PostgresConfig) {
		c.Password = password
	}
}

// WithImage sets the Postgres image.
func WithImage(image string) PostgresOption {
	return func(c *PostgresConfig) {
		c.Image = image
	}
}

// WithMigrations sets the path to migration files and enables automatic migration.
func WithMigrations(path string) PostgresOption {
	return func(c *PostgresConfig) {
		c.MigrationsPath = path
	}
}

// SetupTestDB creates a test database with migrations and returns cleanup function.
func SetupTestDB(t *testing.T, opts ...PostgresOption) (*sql.DB, string, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := NewPostgresContainer(ctx, opts...)
	if err != nil {
		t.Fatalf("failed to create postgres container: %v", err)
	}

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}

	return container.DB, container.ConnStr, cleanup
}

// GetProjectRoot returns the project root directory by walking up from the current file.
func GetProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}

	dir := filepath.Dir(filename)
	for range 10 {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("could not find project root (go.mod)")
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
