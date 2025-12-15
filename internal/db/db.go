package db

import (
	"archive/zip"
	"context"
	"database/sql"
	"embed"
	"encoding/csv"
	"errors"
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

//go:embed sql/*.sql
var migrationFiles embed.FS

//go:embed queries/build_win_expectancy.sql
var buildWinExpectancyQuery string

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

// DatasetRefresh represents the last time a dataset was ingested.
type DatasetRefresh struct {
	Dataset      string
	LastLoadedAt time.Time
	RowCount     int64
}

type Exec interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Connect establishes a connection to the PostgreSQL database.
// If connStr is empty, it falls back to DATABASE_URL environment variable or a default connection string.
func Connect(connStr string) (*DB, error) {
	if connStr == "" {
		connStr = os.Getenv("DATABASE_URL")
		if connStr == "" {
			connStr = "host=localhost port=5432 user=postgres dbname=baseball_dev sslmode=disable"
		}
	}

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
func markApplied(ctx context.Context, exec Exec, name string) error {
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
// The gameType parameter is used to tag all games from this file (e.g., "regular", "allstar", "worldseries").
func (db *DB) LoadRetrosheetGameLog(ctx context.Context, zipPath string, gameType string) (int64, error) {
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
		"acquisition_info", "game_type",
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
	lines := strings.SplitSeq(cleaned, "\n")

	for line := range lines {
		if line == "" {
			continue
		}

		if _, err := outFile.WriteString(line + "," + gameType + "\n"); err != nil {
			outFile.Close()
			return 0, fmt.Errorf("failed to write data: %w", err)
		}
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

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE games_temp (LIKE games INCLUDING DEFAULTS)
		ON COMMIT DROP
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	copySQL := `COPY games_temp (
		date, game_number, day_of_week, visiting_team, visiting_team_league,
		visiting_team_game_number, home_team, home_team_league, home_team_game_number,
		visiting_score, home_score, game_length_outs, day_night, completion_info,
		forfeit_info, protest_info, park_id, attendance, game_time_minutes,
		visiting_line_score, home_line_score, visiting_at_bats, visiting_hits,
		visiting_doubles, visiting_triples, visiting_homeruns, visiting_rbi,
		visiting_sac_hits, visiting_sac_flies, visiting_hit_by_pitch, visiting_walks,
		visiting_int_walks, visiting_strikeouts, visiting_stolen_bases,
		visiting_caught_stealing, visiting_gdp, visiting_interference, visiting_lob,
		visiting_pitchers_used, visiting_ind_er, visiting_team_er, visiting_wild_pitches,
		visiting_balks, visiting_putouts, visiting_assists, visiting_errors,
		visiting_passed_balls, visiting_double_plays, visiting_triple_plays,
		home_at_bats, home_hits, home_doubles, home_triples, home_homeruns,
		home_rbi, home_sac_hits, home_sac_flies, home_hit_by_pitch, home_walks,
		home_int_walks, home_strikeouts, home_stolen_bases, home_caught_stealing,
		home_gdp, home_interference, home_lob, home_pitchers_used, home_ind_er,
		home_team_er, home_wild_pitches, home_balks, home_putouts, home_assists,
		home_errors, home_passed_balls, home_double_plays, home_triple_plays,
		hp_ump_id, hp_ump_name, b1_ump_id, b1_ump_name, b2_ump_id, b2_ump_name,
		b3_ump_id, b3_ump_name, lf_ump_id, lf_ump_name, rf_ump_id, rf_ump_name,
		v_manager_id, v_manager_name, h_manager_id, h_manager_name,
		winning_pitcher_id, winning_pitcher_name, losing_pitcher_id, losing_pitcher_name,
		saving_pitcher_id, saving_pitcher_name, goahead_rbi_id, goahead_rbi_name,
		v_starting_pitcher_id, v_starting_pitcher_name, h_starting_pitcher_id,
		h_starting_pitcher_name, v_player_1_id, v_player_1_name, v_player_1_pos,
		v_player_2_id, v_player_2_name, v_player_2_pos, v_player_3_id,
		v_player_3_name, v_player_3_pos, v_player_4_id, v_player_4_name,
		v_player_4_pos, v_player_5_id, v_player_5_name, v_player_5_pos,
		v_player_6_id, v_player_6_name, v_player_6_pos, v_player_7_id,
		v_player_7_name, v_player_7_pos, v_player_8_id, v_player_8_name,
		v_player_8_pos, v_player_9_id, v_player_9_name, v_player_9_pos,
		h_player_1_id, h_player_1_name, h_player_1_pos, h_player_2_id,
		h_player_2_name, h_player_2_pos, h_player_3_id, h_player_3_name,
		h_player_3_pos, h_player_4_id, h_player_4_name, h_player_4_pos,
		h_player_5_id, h_player_5_name, h_player_5_pos, h_player_6_id,
		h_player_6_name, h_player_6_pos, h_player_7_id, h_player_7_name,
		h_player_7_pos, h_player_8_id, h_player_8_name, h_player_8_pos,
		h_player_9_id, h_player_9_name, h_player_9_pos, additional_info,
		acquisition_info, game_type
	) FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '', QUOTE '"')`
	_, err = conn.PgConn().CopyFrom(ctx, file, copySQL)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	columnList := strings.Join(headers, ", ")
	result, err := tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO games (%s)
		SELECT %s FROM games_temp
		ON CONFLICT (date, home_team, game_number)
		DO UPDATE SET game_type = EXCLUDED.game_type
	`, columnList, columnList))
	if err != nil {
		return 0, fmt.Errorf("failed to insert from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result.RowsAffected(), nil
}

// LoadRetrosheetPlays extracts a retrosheet plays zip file and loads it into the plays table.
// The CSV files already have headers, so this is simpler than game logs.
func (db *DB) LoadRetrosheetPlays(ctx context.Context, zipPath string) (int64, error) {
	tmpDir, err := os.MkdirTemp("", "plays-*")
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
		if strings.HasSuffix(strings.ToLower(f.Name), ".csv") {
			csvFile = f
			break
		}
	}

	if csvFile == nil {
		return 0, fmt.Errorf("no .csv file found in zip archive")
	}

	rc, err := csvFile.Open()
	if err != nil {
		return 0, fmt.Errorf("failed to open file from zip: %w", err)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	tmpCSV := filepath.Join(tmpDir, "plays.csv")
	outFile, err := os.Create(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp CSV: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed to read CSV record: %w", err)
		}

		for i := range record {
			if record[i] == "?" {
				record[i] = ""
			}
		}

		if err := writer.Write(record); err != nil {
			return 0, fmt.Errorf("failed to write cleaned record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return 0, fmt.Errorf("failed to flush CSV writer: %w", err)
	}
	outFile.Close()

	file, err := os.Open(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to open cleaned CSV: %w", err)
	}
	defer file.Close()

	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for COPY: %w", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE plays_temp (LIKE plays INCLUDING DEFAULTS)
		ON COMMIT DROP
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	copySQL := `COPY plays_temp FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '', QUOTE '"')`
	_, err = conn.PgConn().CopyFrom(ctx, file, copySQL)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO plays SELECT * FROM plays_temp
		ON CONFLICT (gid, pn) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to insert from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result.RowsAffected(), nil
}

// LoadRetrosheetEjections extracts a retrosheet ejections zip file and loads it into the ejections table.
// The CSV file has headers: GAMEID,DATE,DH,EJECTEE,EJECTEENAME,TEAM,JOB,UMPIRE,UMPIRENAME,INNING,REASON
func (db *DB) LoadRetrosheetEjections(ctx context.Context, zipPath string) (int64, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	var csvFile *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".csv") {
			csvFile = f
			break
		}
	}

	if csvFile == nil {
		return 0, fmt.Errorf("no .csv file found in zip archive")
	}

	rc, err := csvFile.Open()
	if err != nil {
		return 0, fmt.Errorf("failed to open file from zip: %w", err)
	}
	defer rc.Close()

	tmpDir, err := os.MkdirTemp("", "ejections-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpCSV := filepath.Join(tmpDir, "ejections.csv")
	outFile, err := os.Create(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp CSV: %w", err)
	}

	if _, err := outFile.WriteString(
		"game_id,date,game_number,ejectee_id,ejectee_name,team,role,umpire_id,umpire_name,inning,reason\n",
	); err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to write headers: %w", err)
	}

	data, err := io.ReadAll(rc)
	if err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to read data: %w", err)
	}

	cleaned := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(cleaned, "\n")

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) == 12 {
			line = strings.Join(append(fields[:5], fields[6:]...), ",")
		}

		if _, err := outFile.WriteString(line + "\n"); err != nil {
			outFile.Close()
			return 0, fmt.Errorf("failed to write data: %w", err)
		}
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

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE ejections_temp (LIKE ejections INCLUDING DEFAULTS)
		ON COMMIT DROP
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	copySQL := `COPY ejections_temp FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '', QUOTE '"')`
	_, err = conn.PgConn().CopyFrom(ctx, file, copySQL)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO ejections SELECT * FROM ejections_temp
		ON CONFLICT (game_id, ejectee_id) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to insert from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result.RowsAffected(), nil
}

// RecordDatasetRefresh upserts the refresh timestamp for a dataset after an ETL run.
func (db *DB) RecordDatasetRefresh(ctx context.Context, dataset string, rowCount int64) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO dataset_refreshes (dataset, last_loaded_at, row_count)
		VALUES ($1, NOW(), $2)
		ON CONFLICT (dataset) DO UPDATE
		SET last_loaded_at = EXCLUDED.last_loaded_at,
		    row_count = EXCLUDED.row_count
	`, dataset, rowCount)
	if err != nil {
		return fmt.Errorf("failed to record dataset refresh for %s: %w", dataset, err)
	}
	return nil
}

// DatasetRefreshes returns the last-known refresh metadata for all tracked datasets.
func (db *DB) DatasetRefreshes(ctx context.Context) (map[string]DatasetRefresh, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT dataset, last_loaded_at, row_count
		FROM dataset_refreshes
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query dataset refreshes: %w", err)
	}
	defer rows.Close()

	result := make(map[string]DatasetRefresh)
	for rows.Next() {
		var entry DatasetRefresh
		if err := rows.Scan(&entry.Dataset, &entry.LastLoadedAt, &entry.RowCount); err != nil {
			return nil, fmt.Errorf("failed to scan dataset refresh: %w", err)
		}
		result[entry.Dataset] = entry
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dataset refreshes: %w", err)
	}

	return result, nil
}

// LoadFanGraphsWOBA loads wOBA constants from FanGraphs CSV into the database.
// CSV format: Season,wOBA,wOBAScale,wBB,wHBP,w1B,w2B,w3B,wHR,runSB,runCS,R/PA,R/W,cFIP
func (db *DB) LoadFanGraphsWOBA(ctx context.Context, csvPath string) (int64, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open wOBA CSV: %w", err)
	}
	defer file.Close()

	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for wOBA load: %w", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "TRUNCATE TABLE woba_constants"); err != nil {
		return 0, fmt.Errorf("failed to truncate woba_constants: %w", err)
	}

	copyQuery := `
		COPY woba_constants (season, woba, woba_scale, w_bb, w_hbp, w_1b, w_2b, w_3b, w_hr, run_sb, run_cs, r_pa, r_w, c_fip)
		FROM STDIN WITH (FORMAT CSV, HEADER true)
	`

	tag, err := conn.PgConn().CopyFrom(ctx, file, copyQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to copy wOBA data: %w", err)
	}

	return tag.RowsAffected(), nil
}

// LoadFanGraphsParks loads park factors from FanGraphs CSV into the database.
// CSV format: Season,Team,Basic (5yr),3yr,1yr,1B,2B,3B,HR,SO,BB,GB,FB,LD,IFFB,FIP
func (db *DB) LoadFanGraphsParks(ctx context.Context, csvPath string) (int64, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open park factors CSV: %w", err)
	}
	defer file.Close()

	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for park factors load: %w", err)
	}
	defer conn.Close(ctx)
	return db.loadParkFactorsWithTransform(ctx, conn, csvPath)
}

// loadParkFactorsWithTransform reads CSV and transforms data before inserting
func (db *DB) loadParkFactorsWithTransform(ctx context.Context, conn *pgx.Conn, csvPath string) (int64, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csvReader(file)
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, name := range header {
		colIdx[name] = i
	}

	query := `
		INSERT INTO park_factors (
			park_id, season, team_id,
			basic_5yr, basic_3yr, basic_1yr,
			factor_1b, factor_2b, factor_3b, factor_hr,
			factor_so, factor_bb, factor_gb, factor_fb,
			factor_ld, factor_iffb, factor_fip
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (park_id, season) DO UPDATE SET
			team_id = EXCLUDED.team_id,
			basic_5yr = EXCLUDED.basic_5yr,
			basic_3yr = EXCLUDED.basic_3yr,
			basic_1yr = EXCLUDED.basic_1yr,
			factor_1b = EXCLUDED.factor_1b,
			factor_2b = EXCLUDED.factor_2b,
			factor_3b = EXCLUDED.factor_3b,
			factor_hr = EXCLUDED.factor_hr,
			factor_so = EXCLUDED.factor_so,
			factor_bb = EXCLUDED.factor_bb,
			factor_gb = EXCLUDED.factor_gb,
			factor_fb = EXCLUDED.factor_fb,
			factor_ld = EXCLUDED.factor_ld,
			factor_iffb = EXCLUDED.factor_iffb,
			factor_fip = EXCLUDED.factor_fip
	`

	var rowCount int64
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rowCount, fmt.Errorf("failed to read row: %w", err)
		}

		season := record[colIdx["Season"]]
		fangraphsTeam := record[colIdx["Team"]]

		var parkID, teamID string
		lookupQuery := `
			SELECT primary_park_id, retrosheet_team_id
			FROM fangraphs_team_park_map
			WHERE fangraphs_team = $1
			  AND (start_year IS NULL OR start_year <= $2::int)
			  AND (end_year IS NULL OR end_year >= $2::int)
		`
		err = conn.QueryRow(ctx, lookupQuery, fangraphsTeam, season).Scan(&parkID, &teamID)
		if err != nil {
			return rowCount, fmt.Errorf("failed to map team %s: %w (check fangraphs_team_park_map table)", fangraphsTeam, err)
		}

		_, err = conn.Exec(ctx, query,
			parkID, season, teamID,
			record[colIdx["Basic (5yr)"]], record[colIdx["3yr"]], record[colIdx["1yr"]],
			record[colIdx["1B"]], record[colIdx["2B"]], record[colIdx["3B"]],
			record[colIdx["HR"]], record[colIdx["SO"]], record[colIdx["BB"]],
			record[colIdx["GB"]], record[colIdx["FB"]], record[colIdx["LD"]],
			record[colIdx["IFFB"]], record[colIdx["FIP"]],
		)

		if err != nil {
			return rowCount, fmt.Errorf("failed to insert park factor for %s/%s: %w", teamID, season, err)
		}

		rowCount++
	}

	return rowCount, nil
}

// csvReader creates a CSV reader with common settings
func csvReader(r io.Reader) *csv.Reader {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	return reader
}

// BuildWinExpectancy computes win expectancy data from historical play-by-play data and populates the win_expectancy_historical table.
// This analyzes all plays in the plays table, joins with game outcomes, and calculates win probabilities for each unique game state.
func (db *DB) BuildWinExpectancy(ctx context.Context, minSampleSize int) (int64, error) {
	if minSampleSize < 1 {
		minSampleSize = 50
	}

	result, err := db.ExecContext(ctx, buildWinExpectancyQuery, minSampleSize)
	if err != nil {
		return 0, fmt.Errorf("failed to build win expectancy: %w", err)
	}

	return result.RowsAffected()
}

// loadNegroLeaguesTeamMapping reads the team-to-league mapping CSV
func loadNegroLeaguesTeamMapping() (map[string]string, error) {
	mappingPath := "internal/db/negro_leagues_teams.csv"

	file, err := os.Open(mappingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open team mapping file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	teamMap := make(map[string]string)

	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read mapping row: %w", err)
		}

		if len(record) < 2 {
			continue
		}

		teamID := strings.TrimSpace(record[0])
		league := strings.TrimSpace(record[1])

		if teamID != "" && league != "" {
			teamMap[teamID] = league
		}
	}

	return teamMap, nil
}

// LoadNegroLeaguesGameInfo loads Negro Leagues game data from gameinfo.csv into games table
// The gameinfo.csv has a different schema than standard Retrosheet gamelogs, so we map what's available
func (db *DB) LoadNegroLeaguesGameInfo(ctx context.Context, csvPath string) (int64, error) {
	teamLeagueMap, err := loadNegroLeaguesTeamMapping()
	if err != nil {
		return 0, fmt.Errorf("failed to load team mapping: %w", err)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open gameinfo CSV: %w", err)
	}
	defer file.Close()

	tmpDir, err := os.MkdirTemp("", "negroleagues-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpCSV := filepath.Join(tmpDir, "gameinfo_transformed.csv")
	outFile, err := os.Create(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp CSV: %w", err)
	}

	headers := []string{
		"date", "game_number", "home_team", "visiting_team", "park_id",
		"day_night", "attendance", "game_time_minutes",
		"visiting_score", "home_score",
		"winning_pitcher_id", "losing_pitcher_id", "saving_pitcher_id",
		"hp_ump_id", "b1_ump_id", "b2_ump_id", "b3_ump_id",
		"lf_ump_id", "rf_ump_id",
		"game_type", "home_team_league", "visiting_team_league",
	}

	csvWriter := csv.NewWriter(outFile)
	if err := csvWriter.Write(headers); err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to write headers: %w", err)
	}

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	_, err = reader.Read()
	if err != nil {
		outFile.Close()
		return 0, fmt.Errorf("failed to read source header: %w", err)
	}

	cleanNumeric := func(val string) string {
		if val == "" {
			return ""
		}

		if strings.HasSuffix(val, "?") || strings.HasPrefix(val, "<") || strings.HasPrefix(val, ">") {
			return ""
		}
		return val
	}

	rowCount := int64(0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			outFile.Close()
			return 0, fmt.Errorf("failed to read CSV row: %w", err)
		}

		if len(record) < 35 {
			continue
		}

		gameType := record[32]
		if gameType == "" {
			gameType = "regular"
		}

		homeTeam := record[2]
		visitingTeam := record[1]

		homeLeague := teamLeagueMap[homeTeam]
		visitingLeague := teamLeagueMap[visitingTeam]

		row := []string{
			record[4], record[5], // date, game_number
			homeTeam, visitingTeam, // home_team, visiting_team
			record[3], record[7], // park_id, day_night
			cleanNumeric(record[13]), cleanNumeric(record[12]), cleanNumeric(record[33]), cleanNumeric(record[34]), // attendance, game_time_minutes, visiting_score, home_score
			record[29], record[30], record[31], // winning_pitcher_id, losing_pitcher_id, saving_pitcher_id
			record[23], record[24], record[25], record[26], // hp_ump_id, b1_ump_id, b2_ump_id, b3_ump_id
			record[27], record[28], // lf_ump_id, rf_ump_id
			gameType, homeLeague, visitingLeague, // game_type, home_team_league, visiting_team_league
		}

		if err := csvWriter.Write(row); err != nil {
			outFile.Close()
			return 0, fmt.Errorf("failed to write CSV row: %w", err)
		}
		rowCount++
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		outFile.Close()
		return 0, fmt.Errorf("CSV writer error: %w", err)
	}
	outFile.Close()

	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for COPY: %w", err)
	}
	defer conn.Close(ctx)

	csvFile, err := os.Open(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer csvFile.Close()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE games_temp (LIKE games INCLUDING DEFAULTS)
		ON COMMIT DROP
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	copySQL := fmt.Sprintf(`COPY games_temp (%s) FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '', QUOTE '"')`,
		strings.Join(headers, ", "))

	_, err = conn.PgConn().CopyFrom(ctx, csvFile, copySQL)
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	columnList := strings.Join(headers, ", ")
	result, err := tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO games (%s)
		SELECT DISTINCT ON (date, home_team, game_number) %s
		FROM games_temp
		ON CONFLICT (date, home_team, game_number)
		DO UPDATE SET
			game_type = EXCLUDED.game_type,
			home_team_league = EXCLUDED.home_team_league,
			visiting_team_league = EXCLUDED.visiting_team_league
	`, columnList, columnList))
	if err != nil {
		return 0, fmt.Errorf("failed to insert from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result.RowsAffected(), nil
}

// LoadNegroLeaguesPlays loads Negro Leagues play-by-play data from plays.csv into plays table
func (db *DB) LoadNegroLeaguesPlays(ctx context.Context, csvPath string) (int64, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open plays CSV: %w", err)
	}
	defer file.Close()

	tmpDir, err := os.MkdirTemp("", "negroleagues-plays-*")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpCSV := filepath.Join(tmpDir, "plays_cleaned.csv")
	outFile, err := os.Create(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp CSV: %w", err)
	}

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			outFile.Close()
			return 0, fmt.Errorf("failed to read CSV row: %w", err)
		}

		for i := range record {
			if record[i] == "?" || record[i] == "??" {
				record[i] = ""
			}
		}

		if err := writer.Write(record); err != nil {
			outFile.Close()
			return 0, fmt.Errorf("failed to write cleaned CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		outFile.Close()
		return 0, fmt.Errorf("CSV writer error: %w", err)
	}
	outFile.Close()

	cleanedFile, err := os.Open(tmpCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to open cleaned CSV: %w", err)
	}
	defer cleanedFile.Close()

	conn, err := pgx.Connect(ctx, db.connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect for COPY: %w", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE plays_temp (LIKE plays INCLUDING DEFAULTS)
		ON COMMIT DROP
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp table: %w", err)
	}

	copySQL := `COPY plays_temp FROM STDIN WITH (FORMAT CSV, HEADER true, NULL '', QUOTE '"')`
	_, err = conn.PgConn().CopyFrom(ctx, cleanedFile, copySQL)
	if err != nil {
		return 0, fmt.Errorf("COPY failed: %w", err)
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO plays
		SELECT * FROM plays_temp
		ON CONFLICT (gid, pn) DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to insert from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result.RowsAffected(), nil
}

// LoadNegroLeaguesData loads Negro Leagues gameinfo and plays files if they exist.
// The provided directory should contain gameinfo.csv and plays.csv extracted from Retrosheet.
func (db *DB) LoadNegroLeaguesData(ctx context.Context, dataDir string) (int64, int64, error) {
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet", "negroleagues")
	}

	var gameRows, playRows int64

	gameinfoFile := filepath.Join(dataDir, "gameinfo.csv")
	if info, err := os.Stat(gameinfoFile); err == nil && !info.IsDir() {
		rows, err := db.LoadNegroLeaguesGameInfo(ctx, gameinfoFile)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to load Negro Leagues gameinfo: %w", err)
		}
		gameRows = rows
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, 0, fmt.Errorf("failed to access %s: %w", gameinfoFile, err)
	}

	playsFile := filepath.Join(dataDir, "plays.csv")
	if info, err := os.Stat(playsFile); err == nil && !info.IsDir() {
		rows, err := db.LoadNegroLeaguesPlays(ctx, playsFile)
		if err != nil {
			return gameRows, 0, fmt.Errorf("failed to load Negro Leagues plays: %w", err)
		}
		playRows = rows
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return gameRows, 0, fmt.Errorf("failed to access %s: %w", playsFile, err)
	}

	return gameRows, playRows, nil
}
