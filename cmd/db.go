package cmd

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/config"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/seed"
)

// DbCmd creates the db command group
func DbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long:  "Database migration and management operations.",
	}
	cmd.AddCommand(DbMigrateCmd())
	cmd.AddCommand(DbResetCmd())
	cmd.AddCommand(DbRepopulateCmd())
	cmd.AddCommand(DbRecreateCmd())
	cmd.AddCommand(DbRefreshViewsCmd())
	return cmd
}

// DbMigrateCmd creates the migrate command
func DbMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Create and update database schema for baseball data.",
		RunE:  migrate,
	}
}

// DbResetCmd creates the reset command
func DbResetCmd() *cobra.Command {
	var csvDir string
	var yearsFlag string
	var dataDir string
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Clear Lahman and Retrosheet data before reseeding",
		Long:  "Truncate Lahman and Retrosheet tables, clear refresh metadata, and reseed datasets.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return resetDatabase(cmd, csvDir, dataDir, yearsFlag)
		},
	}
	cmd.Flags().StringVar(&csvDir, "csv-dir", "", "Path to Lahman CSV directory (defaults to data/lahman/csv)")
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years or ranges, e.g. 2022,2023-2025")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Base dir for Retrosheet data (defaults to data/retrosheet)")
	return cmd
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

// DbRepopulateCmd creates the populate command
func DbRepopulateCmd() *cobra.Command {
	var csvDir string
	var yearsFlag string
	var dataDir string
	cmd := &cobra.Command{
		Use:   "populate",
		Short: "Seed the database with Lahman and Retrosheet data",
		Long:  "Seed the database with Lahman CSVs and Retrosheet zip files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPopulateAll(cmd, csvDir, dataDir, yearsFlag)
		},
	}
	cmd.AddCommand(DbRepopulateLahmanCmd())
	cmd.AddCommand(DbRepopulateRetrosheetCmd())
	cmd.AddCommand(DbRepopulateAllCmd())
	cmd.Flags().StringVar(&csvDir, "csv-dir", "", "Path to Lahman CSV directory (defaults to data/lahman/csv)")
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years or ranges, e.g. 2022,2023-2025")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Base dir for Retrosheet data (defaults to data/retrosheet)")
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

func DbRepopulateLahmanCmd() *cobra.Command {
	var csvDir string
	cmd := &cobra.Command{
		Use:   "lahman",
		Short: "Seed Lahman data only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return repopulateLahman(cmd, csvDir)
		},
	}
	cmd.Flags().StringVar(&csvDir, "csv-dir", "", "Path to Lahman CSV directory (defaults to data/lahman/csv)")
	return cmd
}

func DbRepopulateRetrosheetCmd() *cobra.Command {
	var eraFlag string
	var yearsFlag string
	var dataDir string
	var force bool
	cmd := &cobra.Command{
		Use:   "retrosheet",
		Short: "Seed Retrosheet data only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return repopulateRetrosheet(cmd, dataDir, eraFlag, yearsFlag, force)
		},
	}
	cmd.Flags().StringVar(&eraFlag, "era", "", "Load data for a specific era (federal, nlg, 1970s, 1980s, steroid, modern)")
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years, ranges, or 'all', e.g. 2022,2023-2025,all")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Base dir for Retrosheet data (defaults to data/retrosheet)")
	cmd.Flags().BoolVar(&force, "force", false, "Force reload even if data already exists")
	return cmd
}

func DbRepopulateAllCmd() *cobra.Command {
	var csvDir string
	var yearsFlag string
	var dataDir string
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Seed both Lahman and Retrosheet data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPopulateAll(cmd, csvDir, dataDir, yearsFlag)
		},
	}
	cmd.Flags().StringVar(&csvDir, "csv-dir", "", "Path to Lahman CSV directory (defaults to data/lahman/csv)")
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years or ranges, e.g. 2022,2023-2025")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Base dir for Retrosheet data (defaults to data/retrosheet)")
	return cmd
}

func repopulateLahman(cmd *cobra.Command, csvDir string) error {
	echo.Header("Seeding Lahman Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()
	_, err = seed.LoadLahman(ctx, database, seed.LahmanOptions{CSVDir: csvDir})
	return err
}

func repopulateRetrosheet(cmd *cobra.Command, dataDir, eraFlag, yearsFlag string, force bool) error {
	echo.Header("Seeding Retrosheet Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	var years []int
	if eraFlag != "" {
		echo.Infof("Loading data for era: %s", eraFlag)
		years = seed.GetYearsForEras([]string{eraFlag})
		if len(years) == 0 {
			return fmt.Errorf("unknown era: %s", eraFlag)
		}
		era := seed.GetEra(eraFlag)
		if era != nil {
			echo.Infof("Era: %s (%d-%d)", era.Name, era.StartYear, era.EndYear)
		}
	} else {
		years, err = parseYearFlag(yearsFlag)
		if err != nil {
			return err
		}
	}

	ctx := cmd.Context()
	_, err = seed.LoadRetrosheet(ctx, database, seed.RetrosheetOptions{
		DataDir: dataDir,
		Years:   years,
		Force:   force,
	})
	return err
}

func resetDatabase(cmd *cobra.Command, csvDir, dataDir, yearsFlag string) error {
	echo.Header("Database Reset")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	years, err := parseYearFlag(yearsFlag)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	echo.Info("Clearing Lahman tables...")
	if err := seed.ResetLahman(ctx, database, nil); err != nil {
		return err
	}
	echo.Success("✓ Lahman tables cleared")

	echo.Info("Clearing Retrosheet tables...")
	if err := seed.ResetRetrosheet(ctx, database, years); err != nil {
		return err
	}
	echo.Success("✓ Retrosheet tables cleared")

	echo.Info("Reseeding datasets...")
	return runPopulateAll(cmd, csvDir, dataDir, yearsFlag)
}

func recreateDatabase(cmd *cobra.Command, dbURL string) error {
	echo.Header("Recreating Database")

	targetURL, err := resolveDatabaseURL(cmd, dbURL)
	if err != nil {
		return err
	}
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("error: invalid database URL: %w", err)
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		return fmt.Errorf("error: database URL must include a database name: %s", targetURL)
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

	adminURL := *parsed
	adminURL.Path = "/postgres"
	adminURL.RawPath = "/postgres"

	conn, err := sql.Open("pgx", adminURL.String())
	if err != nil {
		return fmt.Errorf("error: failed to connect to server: %w", err)
	}
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
	configPath := findConfigPath(cmd)
	return config.Load(configPath)
}

func findConfigPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}

	if flag := cmd.Flags().Lookup("config"); flag != nil {
		return flag.Value.String()
	}

	return findConfigPath(cmd.Parent())
}

func runPopulateAll(cmd *cobra.Command, csvDir, dataDir, yearsFlag string) error {
	if err := repopulateLahman(cmd, csvDir); err != nil {
		return err
	}

	return repopulateRetrosheet(cmd, dataDir, "", yearsFlag, false)
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

	count, err := database.RefreshMaterializedViews(ctx, args)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Success(fmt.Sprintf("✓ Successfully refreshed %d materialized view(s)", count))
	return nil
}
