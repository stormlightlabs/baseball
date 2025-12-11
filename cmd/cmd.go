package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

// ETLCmd creates the etl command group
func ETLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "etl",
		Short: "ETL operations for baseball data",
		Long:  "Extract, Transform, and Load operations for Lahman and Retrosheet data sources.",
	}

	cmd.AddCommand(FetchCmd())
	cmd.AddCommand(LoadCmd())
	cmd.AddCommand(StatusCmd())
	return cmd
}

// DbCmd creates the db command group
func DbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long:  "Database migration and management operations.",
	}

	cmd.AddCommand(DbMigrateCmd())
	return cmd
}

// FetchCmd creates the fetch command group under etl
func FetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Download baseball data sources",
		Long:  "Download data from Lahman and Retrosheet sources.",
	}

	cmd.AddCommand(LahmanFetchCmd())
	cmd.AddCommand(RetrosheetFetchCmd())
	return cmd
}

// LoadCmd creates the load command group under etl
func LoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load data into database",
		Long:  "Load downloaded data into PostgreSQL database.",
	}

	cmd.AddCommand(LahmanLoadCmd())
	cmd.AddCommand(RetrosheetLoadCmd())
	return cmd
}

// LahmanFetchCmd creates the fetch lahman command
func LahmanFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lahman",
		Short: "Get instructions to download Lahman baseball database",
		Long:  "Provides instructions and creates directories for downloading the Lahman baseball database from SABR.",
		RunE:  fetchLahman,
	}
}

// RetrosheetFetchCmd creates the fetch retrosheet command
func RetrosheetFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "retrosheet",
		Short: "Download Retrosheet data",
		Long:  "Download Retrosheet game logs and event files.",
		RunE:  fetchRetrosheet,
	}
}

// LahmanLoadCmd creates the load lahman command
func LahmanLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lahman",
		Short: "Load Lahman data into database",
		Long:  "Load Lahman CSV files into PostgreSQL database.",
		RunE:  loadLahman,
	}
}

// RetrosheetLoadCmd creates the load retrosheet command
func RetrosheetLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "retrosheet",
		Short: "Load Retrosheet data into database",
		Long:  "Load Retrosheet CSV files into PostgreSQL database.",
		RunE:  loadRetrosheet,
	}
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

func fetchLahman(cmd *cobra.Command, args []string) error {
	echo.Header("Lahman Database Download Instructions")

	dataDir := "data/lahman"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create data directory: %w", err)
	}

	echo.Info("The Lahman database must be downloaded manually from SABR:")
	echo.Info("")
	echo.Info("Download Instructions:")
	echo.Info("  1. Visit: https://sabr.org/lahman-database/")
	echo.Info("  2. Look for 'Download Database' section")
	echo.Info("  3. Download the CSV format (recommended)")
	echo.Infof("  4. Extract files to: %s", filepath.Join(dataDir, "csv"))
	echo.Info("")
	echo.Info("Alternative sources:")
	echo.Info("  • GitHub: https://github.com/cdalzell/Lahman")
	echo.Info("  • Direct CSV: Individual tables from SABR site")
	echo.Info("")
	echo.Success("✓ Data directory created successfully")
	echo.Infof("  Directory: %s", dataDir)
	echo.Info("")
	echo.Info("After downloading, use: baseball etl load lahman")

	return nil
}

func fetchRetrosheet(cmd *cobra.Command, args []string) error {
	echo.Header("Fetching Retrosheet Data")
	echo.Info("Downloading Retrosheet game logs and play-by-play data...")

	dataDir := "data/retrosheet"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create data directory: %w", err)
	}

	gameLogsDir := filepath.Join(dataDir, "gamelogs")
	playsDir := filepath.Join(dataDir, "plays")
	if err := os.MkdirAll(gameLogsDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create gamelogs directory: %w", err)
	}

	if err := os.MkdirAll(playsDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create plays directory: %w", err)
	}

	gameLogFiles := map[string]string{
		"GL2025.zip": "https://www.retrosheet.org/gamelogs/gl2025.zip",
		"GL2024.zip": "https://www.retrosheet.org/gamelogs/gl2024.zip",
		"GL2023.zip": "https://www.retrosheet.org/gamelogs/gl2023.zip",
	}

	playFiles := map[string]string{
		"2025plays.zip": "https://www.retrosheet.org/downloads/plays/2025plays.zip",
		"2024plays.zip": "https://www.retrosheet.org/downloads/plays/2024plays.zip",
		"2023plays.zip": "https://www.retrosheet.org/downloads/plays/2023plays.zip",
	}

	echo.Info("Downloading game logs...")
	for filename, url := range gameLogFiles {
		echo.Infof("  Downloading %s...", filename)

		resp, err := http.Get(url)
		if err != nil {
			echo.Infof("  ⚠ Failed to download %s: %v", filename, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			echo.Infof("  ⚠ %s not available (HTTP %d)", filename, resp.StatusCode)
			continue
		}

		outputPath := filepath.Join(gameLogsDir, filename)
		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error: failed to create %s: %w", filename, err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("error: failed to save %s: %w", filename, err)
		}

		echo.Successf("  ✓ %s downloaded", filename)
	}

	echo.Info("")
	echo.Info("Downloading play-by-play data...")
	for filename, url := range playFiles {
		echo.Infof("  Downloading %s...", filename)

		resp, err := http.Get(url)
		if err != nil {
			echo.Infof("  ⚠ Failed to download %s: %v", filename, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			echo.Infof("  ⚠ %s not available (HTTP %d)", filename, resp.StatusCode)
			continue
		}

		outputPath := filepath.Join(playsDir, filename)
		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error: failed to create %s: %w", filename, err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("error: failed to save %s: %w", filename, err)
		}

		echo.Successf("  ✓ %s downloaded", filename)
	}

	echo.Info("")
	echo.Success("✓ Retrosheet data downloaded successfully")
	echo.Infof("  Game logs: %s", gameLogsDir)
	echo.Infof("  Play-by-play: %s", playsDir)
	return nil
}

func loadLahman(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Lahman Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	dataDir := "data/lahman"
	csvDir := filepath.Join(dataDir, "csv")

	tables := []string{
		"AllstarFull",
		"Appearances",
		"AwardsManagers", "AwardsPlayers", "AwardsShareManagers", "AwardsSharePlayers",
		"Batting", "BattingPost",
		"CollegePlaying",
		"Fielding", "FieldingOF", "FieldingOFsplit", "FieldingPost",
		"HomeGames",
		"HallOfFame",
		"Managers", "ManagersHalf",
		"Parks",
		"People",
		"Pitching", "PitchingPost",
		"Salaries",
		"Schools",
		"SeriesPost",
		"Teams", "TeamsFranchises", "TeamsHalf",
	}

	ctx := cmd.Context()
	totalRows := int64(0)

	for _, table := range tables {
		csvFile := filepath.Join(csvDir, table+".csv")

		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			echo.Infof("Skipping %s (file not found)", table)
			continue
		}

		echo.Infof("Loading %s...", table)

		rows, err := database.CopyCSV(ctx, table, csvFile)
		if err != nil {
			return fmt.Errorf("error: failed to load %s: %w", table, err)
		}

		totalRows += rows
		echo.Successf("✓ Loaded %s (%d rows)", table, rows)
	}

	echo.Success(fmt.Sprintf("✓ All Lahman data loaded successfully (%d total rows)", totalRows))
	if err := database.RecordDatasetRefresh(ctx, "lahman", totalRows); err != nil {
		return fmt.Errorf("error: failed to record Lahman refresh: %w", err)
	}
	return nil
}

func migrate(cmd *cobra.Command, args []string) error {
	echo.Header("Database Migration")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
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

func loadRetrosheet(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Retrosheet Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	dataDir := "data/retrosheet"
	gameLogsDir := filepath.Join(dataDir, "gamelogs")
	years := []string{"2023", "2024", "2025"}

	ctx := cmd.Context()
	totalRows := int64(0)
	gamesLoaded := int64(0)

	echo.Info("Loading game logs...")
	for _, year := range years {
		zipFile := filepath.Join(gameLogsDir, fmt.Sprintf("GL%s.zip", year))

		if _, err := os.Stat(zipFile); os.IsNotExist(err) {
			echo.Infof("  Skipping %s (file not found)", year)
			continue
		}

		echo.Infof("  Loading %s game logs...", year)

		rows, err := database.LoadRetrosheetGameLog(ctx, zipFile)
		if err != nil {
			return fmt.Errorf("error: failed to load %s: %w", year, err)
		}

		totalRows += rows
		gamesLoaded += rows
		echo.Successf("  ✓ Loaded %s (%d rows)", year, rows)
	}

	echo.Info("")
	echo.Info("Loading play-by-play data...")
	playsDir := filepath.Join(dataDir, "plays")
	playsLoaded := int64(0)
	for _, year := range years {
		zipFile := filepath.Join(playsDir, fmt.Sprintf("%splays.zip", year))

		if _, err := os.Stat(zipFile); os.IsNotExist(err) {
			echo.Infof("  Skipping %s (file not found)", year)
			continue
		}

		echo.Infof("  Loading %s plays...", year)

		rows, err := database.LoadRetrosheetPlays(ctx, zipFile)
		if err != nil {
			return fmt.Errorf("error: failed to load %s plays: %w",
				year, err)
		}

		playsLoaded += rows
		totalRows += rows
		echo.Successf("  ✓ Loaded %s (%d rows)", year, rows)
	}

	echo.Info("")
	echo.Success("✓ All Retrosheet data loaded successfully")
	echo.Infof("  Total rows: %d", totalRows)
	echo.Infof("  Play-by-play rows: %d", playsLoaded)
	if err := database.RecordDatasetRefresh(ctx, "retrosheet_games", gamesLoaded); err != nil {
		return fmt.Errorf("error: failed to record Retrosheet games refresh: %w", err)
	}
	if err := database.RecordDatasetRefresh(ctx, "retrosheet_plays", playsLoaded); err != nil {
		return fmt.Errorf("error: failed to record Retrosheet plays refresh: %w", err)
	}
	return nil
}
