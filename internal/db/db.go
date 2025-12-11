package db

import (
	"archive/zip"
	"context"
	"database/sql"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	ExecContext(context.Context, string, ...any) (sql.Result, error)
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

// LoadRetrosheetGameLog extracts a retrosheet game log zip file and loads it into the games table.
// Retrosheet CSVs don't have headers, so this function adds them before loading.
func (db *DB) LoadRetrosheetGameLog(ctx context.Context, zipPath string) (int64, error) {
	tmpDir, err := os.MkdirTemp("", "retrosheet-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	var csvFile *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".txt") {
			csvFile = f
			break
		}
	}

	if csvFile == nil {
		return 0, fmt.Errorf("no .txt file found in zip archive")
	}

	rc, err := csvFile.Open()
	if err != nil {
		return 0, fmt.Errorf("failed to open file from zip: %w", err)
	}
	defer rc.Close()

	tmpCSV := filepath.Join(tmpDir, "gamelog.csv")
	outFile, err := os.Create(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp CSV: %w", err)
	}

	headers := []string{
		"date", "game_number", "day_of_week", "visiting_team", "visiting_team_league",
		"visiting_team_game_number", "home_team", "home_team_league", "home_team_game_number",
		"visiting_score", "home_score", "game_length_outs", "day_night", "completion_info",
		"forfeit_info", "protest_info", "park_id", "attendance", "game_time_minutes",
		"visiting_line_score", "home_line_score", "visiting_at_bats", "visiting_hits",
		"visiting_doubles", "visiting_triples", "visiting_homeruns", "visiting_rbi",
		"visiting_sac_hits", "visiting_sac_flies", "visiting_hit_by_pitch", "visiting_walks",
		"visiting_int_walks", "visiting_strikeouts", "visiting_stolen_bases",
		"visiting_caught_stealing", "visiting_gdp", "visiting_interference", "visiting_lob",
		"visiting_pitchers_used", "visiting_ind_er", "visiting_team_er", "visiting_wild_pitches",
		"visiting_balks", "visiting_putouts", "visiting_assists", "visiting_errors",
		"visiting_passed_balls", "visiting_double_plays", "visiting_triple_plays",
		"home_at_bats", "home_hits", "home_doubles", "home_triples", "home_homeruns",
		"home_rbi", "home_sac_hits", "home_sac_flies", "home_hit_by_pitch", "home_walks",
		"home_int_walks", "home_strikeouts", "home_stolen_bases", "home_caught_stealing",
		"home_gdp", "home_interference", "home_lob", "home_pitchers_used", "home_ind_er",
		"home_team_er", "home_wild_pitches", "home_balks", "home_putouts", "home_assists",
		"home_errors", "home_passed_balls", "home_double_plays", "home_triple_plays",
		"hp_ump_id", "hp_ump_name", "b1_ump_id", "b1_ump_name", "b2_ump_id", "b2_ump_name",
		"b3_ump_id", "b3_ump_name", "lf_ump_id", "lf_ump_name", "rf_ump_id", "rf_ump_name",
		"v_manager_id", "v_manager_name", "h_manager_id", "h_manager_name",
		"winning_pitcher_id", "winning_pitcher_name", "losing_pitcher_id", "losing_pitcher_name",
		"saving_pitcher_id", "saving_pitcher_name", "goahead_rbi_id", "goahead_rbi_name",
		"v_starting_pitcher_id", "v_starting_pitcher_name", "h_starting_pitcher_id",
		"h_starting_pitcher_name", "v_player_1_id", "v_player_1_name", "v_player_1_pos",
		"v_player_2_id", "v_player_2_name", "v_player_2_pos", "v_player_3_id",
		"v_player_3_name", "v_player_3_pos", "v_player_4_id", "v_player_4_name",
		"v_player_4_pos", "v_player_5_id", "v_player_5_name", "v_player_5_pos",
		"v_player_6_id", "v_player_6_name", "v_player_6_pos", "v_player_7_id",
		"v_player_7_name", "v_player_7_pos", "v_player_8_id", "v_player_8_name",
		"v_player_8_pos", "v_player_9_id", "v_player_9_name", "v_player_9_pos",
		"h_player_1_id", "h_player_1_name", "h_player_1_pos", "h_player_2_id",
		"h_player_2_name", "h_player_2_pos", "h_player_3_id", "h_player_3_name",
		"h_player_3_pos", "h_player_4_id", "h_player_4_name", "h_player_4_pos",
		"h_player_5_id", "h_player_5_name", "h_player_5_pos", "h_player_6_id",
		"h_player_6_name", "h_player_6_pos", "h_player_7_id", "h_player_7_name",
		"h_player_7_pos", "h_player_8_id", "h_player_8_name", "h_player_8_pos",
		"h_player_9_id", "h_player_9_name", "h_player_9_pos", "additional_info",
		"acquisition_info",
	}

	if _, err := outFile.WriteString(strings.Join(headers, ",") + "\n"); err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to write headers: %w", err)
	}

	data, err := io.ReadAll(rc)
	if err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to read data: %w", err)
	}

	cleaned := strings.ReplaceAll(string(data), "\r\n", "\n")
	if _, err := outFile.WriteString(cleaned); err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to write data: %w", err)
	}

	if err := outFile.Close(); err != nil {
		return 0, fmt.Errorf("failed to close temp file: %w", err)
	}

	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for COPY: %w", err)
	}
	defer conn.Close(ctx)

	file, err := os.Open(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	copySQL := `COPY games FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '', QUOTE '"')`

	tag, err := conn.PgConn().CopyFrom(ctx, file, copySQL)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	return tag.RowsAffected(), nil
}
