// package shared contains command interfaces common to all built binaries,
// namely database management commands.
package shared

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/config"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

// DbCmd creates the db command group
func DbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long: `Database migration and management operations.

Canonical data workflow is owned by baseball-etl (baseball-etl worker/run/validate/status).`,
	}
	cmd.PersistentFlags().String("data-root", "", "Base dataset root (default resolution: --data-root, BASEBALL_DATA_ROOT, data)")
	cmd.AddCommand(DbMigrateCmd())
	cmd.AddCommand(DbRecreateCmd())
	cmd.AddCommand(DbRefreshViewsCmd())
	return cmd
}

// DbMigrateCmd creates the migrate command
func DbMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Create and update structural database schema objects. Materialized view population is handled separately (etl/db refresh-views).",
		RunE:  migrate,
	}
}

// DbRecreateCmd creates the recreate command
func DbRecreateCmd() *cobra.Command {
	var dbURL string
	cmd := &cobra.Command{
		Use:   "recreate",
		Short: "Drop and recreate the configured PostgreSQL database",
		Long:  "Drops the database referenced by --url (or DATABASE_URL) and creates it again. Useful before re-running migrations from scratch.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return recreateDatabase(cmd, dbURL)
		},
	}
	cmd.Flags().StringVar(&dbURL, "url", "", "Database URL to recreate (defaults to DATABASE_URL or local dev)")
	return cmd
}

func migrate(cmd *cobra.Command, args []string) error {
	echo.Header("Database Migration")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")
	echo.Info("Running migrations...")

	ctx := cmd.Context()
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Success("✓ All migrations applied successfully")
	return nil
}

func recreateDatabase(cmd *cobra.Command, dbURL string) error {
	echo.Header("Recreating Database")

	targetURL, err := resolveDatabaseURL(cmd, dbURL)
	if err != nil {
		return err
	}

	dbName, adminConnConfig, err := parseRecreateConnection(targetURL)
	if err != nil {
		return err
	}

	echo.Error(fmt.Sprintf("⚠ WARNING: This will drop and recreate database %s (all data will be lost).", dbName))
	ctx := cmd.Context()

	for i := 5; i > 0; i-- {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			echo.Infof("  Continuing in %d seconds... (Ctrl-C to cancel)", i)
			time.Sleep(time.Second)
		}
	}

	conn := stdlib.OpenDB(*adminConnConfig)
	defer conn.Close()

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("error: failed to ping server: %w", err)
	}

	echo.Info("Terminating active connections...")
	if _, err := conn.ExecContext(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, dbName); err != nil {
		return fmt.Errorf("error: failed to terminate sessions: %w", err)
	}

	echo.Info("Dropping database...")
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", quoteIdentifier(dbName))); err != nil {
		return fmt.Errorf("error: failed to drop database: %w", err)
	}

	echo.Info("Creating database...")
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dbName))); err != nil {
		return fmt.Errorf("error: failed to create database: %w", err)
	}

	echo.Successf("✓ Recreated database %s", dbName)
	return nil
}

func parseRecreateConnection(connString string) (string, *pgx.ConnConfig, error) {
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return "", nil, fmt.Errorf("error: invalid database connection string: %w", err)
	}

	dbName := strings.TrimSpace(cfg.Database)
	if dbName == "" {
		return "", nil, fmt.Errorf("error: database connection string must include a database name")
	}

	adminCfg := cfg.Copy()
	adminCfg.Database = "postgres"

	return dbName, adminCfg, nil
}

func resolveDatabaseURL(cmd *cobra.Command, flagValue string) (string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return flagValue, nil
	}

	cfg, err := loadConfigForCmd(cmd)
	if err == nil && cfg != nil && strings.TrimSpace(cfg.Database.URL) != "" {
		return cfg.Database.URL, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	if env := os.Getenv("DATABASE_URL"); env != "" {
		return env, nil
	}

	return "postgres://postgres:postgres@localhost:5432/baseball_dev?sslmode=disable", nil
}

func quoteIdentifier(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

func loadConfigForCmd(cmd *cobra.Command) (*config.Config, error) {
	configPath := FindConfigPath(cmd)
	return config.Load(configPath)
}

func FindConfigPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}

	if flag := cmd.Flags().Lookup("config"); flag != nil {
		return flag.Value.String()
	}

	return FindConfigPath(cmd.Parent())
}

// DbRefreshViewsCmd creates the refresh-views command
func DbRefreshViewsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh-views [view-names...]",
		Short: "Refresh materialized views",
		Long: `Refresh one or more materialized views. If no view names are provided, refreshes all materialized views.

Available materialized views:
  • player_game_batting_stats - Per-game batting statistics
  • player_game_pitching_stats - Per-game pitching statistics
  • player_game_fielding_stats - Per-game fielding statistics
  • team_game_stats - Per-game team statistics
  • player_id_crosswalk - Player ID mappings (Lahman/Retrosheet)
  • team_franchise_crosswalk - Team and franchise mappings
  • park_map - Park ID crosswalk and metadata
  • no_hitters - No-hitter achievements
  • cycles - Hitting for the cycle achievements
  • multi_hr_games - Multi-home run games
  • triple_plays - Triple play achievements
  • extra_inning_games - Extra inning games
  • win_expectancy_historical - Win expectancy probabilities by game state
  • season_batting_leaders - Season batting statistics and leaderboards
  • season_pitching_leaders - Season pitching statistics and leaderboards
  • career_batting_leaders - Career batting statistics and leaderboards
  • career_pitching_leaders - Career pitching statistics and leaderboards

Examples:
  baseball db refresh-views                            # Refresh all views
  baseball db refresh-views season_batting_leaders     # Refresh one view
  baseball db refresh-views park_map no_hitters        # Refresh multiple views
`,
		RunE: refreshViews,
	}
}

func refreshViews(cmd *cobra.Command, args []string) error {
	echo.Header("Refreshing Materialized Views")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()

	if len(args) == 0 {
		echo.Info("Refreshing all materialized views...")
	} else {
		echo.Infof("Refreshing %d view(s): %v", len(args), args)
	}

	count, err := database.RefreshMaterializedViewsWithOptions(ctx, args, db.MaterializedViewRefreshOptions{
		Step:  "db.refresh-views",
		Group: "manual",
		OnAttempt: func(attempt db.MaterializedViewRefreshAttempt) {
			echo.Infof(
				"  MV view=%s pass=%d attempt=%d mode=%s status=%s duration=%s",
				attempt.ViewName,
				attempt.Pass,
				attempt.Attempt,
				attempt.Mode,
				attempt.Status,
				attempt.Duration.Round(time.Millisecond),
			)
			if attempt.Error != "" {
				echo.Infof("    error=%s", attempt.Error)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Success(fmt.Sprintf("✓ Successfully refreshed %d materialized view(s)", count))
	return nil
}
