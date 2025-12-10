package db

import (
	"context"
	"database/sql"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type csvSource struct {
	reader  *csv.Reader
	current []string
	err     error
}

func (c *csvSource) Next() bool {
	c.current, c.err = c.reader.Read()
	return c.err == nil
}

func (c *csvSource) Values() ([]any, error) {
	if c.err != nil {
		return nil, c.err
	}
	values := make([]any, len(c.current))
	for i, v := range c.current {
		if v == "" {
			values[i] = nil
		} else {
			values[i] = v
		}
	}
	return values, nil
}

func (c *csvSource) Err() error {
	if c.err == io.EOF {
		return nil
	}
	return c.err
}

//go:embed sql/*.sql
var migrationFiles embed.FS

// Migration represents a single database migration.
type Migration struct {
	Name    string
	Content string
}

// DB wraps a database connection with migration capabilities.
type DB struct {
	*sql.DB
	connStr string
}

// Connect establishes a connection to the PostgreSQL database.
// FIXME: remove default connection string for local postgres.app (eventually)
func Connect() (*DB, error) {
	connStr := "host=localhost port=5432 user=postgres dbname=baseball_dev sslmode=disable"
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: sqlDB, connStr: connStr}, nil
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist.
func (db *DB) ensureMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// isApplied checks if a migration has already been applied.
func (db *DB) isApplied(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`
	err := db.QueryRowContext(ctx, query, name).Scan(&exists)
	return exists, err
}

// markApplied marks a migration as applied in the migrations table.
// Can be called on either *DB or *Tx (both implement ExecContext).
func markApplied(ctx context.Context, exec interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}, name string) error {
	query := `INSERT INTO schema_migrations (name, applied_at) VALUES ($1, $2)`
	_, err := exec.ExecContext(ctx, query, name, time.Now())
	return err
}

// loadMigrations reads all SQL files from the embedded filesystem and returns them sorted by name.
func (db *DB) loadMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir("sql")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		content, err := migrationFiles.ReadFile("sql/" + name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		migrations = append(migrations, Migration{
			Name:    name,
			Content: string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

// Migrate runs all pending database migrations.
// It creates the migrations table if needed, checks which migrations have been applied, and executes any new migrations in order.
func (db *DB) Migrate(ctx context.Context) error {
	if err := db.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := db.loadMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no migration files found")
	}

	for _, migration := range migrations {
		applied, err := db.isApplied(ctx, migration.Name)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Name, err)
		}

		if applied {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", migration.Name, err)
		}

		if _, err := tx.ExecContext(ctx, migration.Content); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.Name, err)
		}

		if err := markApplied(ctx, tx, migration.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to mark migration %s as applied: %w", migration.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Name, err)
		}
	}

	return nil
}

// CopyCSV efficiently loads CSV data into a PostgreSQL table using COPY.
// The CSV file must have a header row that matches the table columns.
func (db *DB) CopyCSV(ctx context.Context, tableName, csvPath string) (int64, error) {
	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for COPY: %w", err)
	}
	defer conn.Close(ctx)

	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	copySQL := fmt.Sprintf(`COPY "%s" FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '')`, tableName)

	tag, err := conn.PgConn().CopyFrom(ctx, file, copySQL)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	return tag.RowsAffected(), nil
}
