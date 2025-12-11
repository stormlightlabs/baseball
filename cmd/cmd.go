package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/api"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/repository"
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

// ServerCmd creates the server command group
func ServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server operations",
		Long:  "Start and manage the baseball API server.",
	}

	cmd.AddCommand(ServerStartCmd())
	cmd.AddCommand(ServerFetchCmd())
	cmd.AddCommand(ServerHealthCmd())
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

// StatusCmd creates the status command
func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check data freshness and completeness",
		Long:  "Display status of loaded data including freshness and completeness metrics.",
		RunE:  status,
	}
}

// ServerStartCmd creates the start command
func ServerStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the API server",
		Long:  "Start the baseball API HTTP server.",
		RunE:  startServer,
	}
}

// ServerFetchCmd creates the server fetch command
func ServerFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch [path]",
		Short: "Test API endpoints",
		Long:  "cURL-like tool for testing API endpoints with formatted output. Path should be relative to /v1/ (e.g., 'players?name=ruth' or 'teams/BOS?year=2023').",
		Args:  cobra.ExactArgs(1),
		RunE:  fetchEndpoint,
	}

	cmd.Flags().StringP("format", "f", "json", "Output format (json|table)")
	return cmd
}

// ServerHealthCmd creates the health command
func ServerHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check server health",
		Long:  "Perform health check on the running API server.",
		RunE:  checkHealth,
	}
}

// Command handler implementations
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

func status(cmd *cobra.Command, args []string) error {
	echo.Header("Data Status")
	ctx := cmd.Context()

	archiveChecks := []struct {
		label string
		path  string
		hint  string
	}{
		{
			label: "Lahman CSVs",
			path:  filepath.Join("data", "lahman", "csv"),
			hint:  "Use `baseball etl fetch lahman` to scaffold/download the dataset",
		},
		{
			label: "Retrosheet game logs",
			path:  filepath.Join("data", "retrosheet", "gamelogs"),
			hint:  "Use `baseball etl fetch retrosheet` to download seasonal game logs",
		},
		{
			label: "Retrosheet plays",
			path:  filepath.Join("data", "retrosheet", "plays"),
			hint:  "Use `baseball etl fetch retrosheet` to download parsed play-by-play archives",
		},
	}

	echo.Info("Local archives:")
	for _, check := range archiveChecks {
		exists, fileCount, latestChange, err := dirSnapshot(check.path)
		if err != nil {
			echo.Errorf("  %s: %v", check.label, err)
			continue
		}
		if !exists {
			echo.Infof("  • %s: %s", check.label, echo.ErrorStyle().Render("missing"))
			echo.Infof("    Path: %s", check.path)
			echo.Infof("    Hint: %s", check.hint)
			continue
		}
		if fileCount == 0 {
			echo.Infof("  • %s: directory exists but contains no files (%s)", check.label, check.path)
			continue
		}
		echo.Successf("  ✓ %s: %d files (last change %s)", check.label, fileCount, humanizeModTime(latestChange))
		echo.Infof("    Path: %s", check.path)
	}

	echo.Info("")
	echo.Info("Database:")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	refreshes, err := database.DatasetRefreshes(ctx)
	if err != nil {
		echo.Infof("  ⚠ Unable to read ETL metadata: %v", err)
		refreshes = map[string]db.DatasetRefresh{}
	}

	lahmanPlayers, lahmanPlayersErr := safeCount(ctx, database, `SELECT COUNT(*) FROM "People"`)
	lahmanTeams, _ := safeCount(ctx, database, `SELECT COUNT(*) FROM "Teams"`)
	lahmanMin, lahmanMax, lahmanRangeErr := seasonRange(ctx, database)

	echo.Info("• Lahman Baseball Database")
	if lahmanPlayersErr != nil {
		echo.Infof("  ⚠ Unable to read player table: %v", lahmanPlayersErr)
	} else if lahmanPlayers == 0 {
		echo.Infof("  • People table is empty. Run `baseball etl load lahman` after downloading CSVs.")
	} else {
		echo.Successf("  ✓ %d players and %d team seasons available", lahmanPlayers, lahmanTeams)
	}
	if lahmanRangeErr == nil && lahmanMin != nil && lahmanMax != nil {
		echo.Infof("    Seasons covered: %d–%d", *lahmanMin, *lahmanMax)
	} else if lahmanRangeErr != nil {
		echo.Infof("    ⚠ Unable to derive season coverage: %v", lahmanRangeErr)
	}
	if entry, ok := refreshes["lahman"]; ok {
		entryCopy := entry
		echo.Infof("    Last ETL run: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Last ETL run: never recorded")
	}

	gamesCount, gamesErr := safeCount(ctx, database, `SELECT COUNT(*) FROM games`)
	playsCount, playsErr := safeCount(ctx, database, `SELECT COUNT(*) FROM plays`)
	gamesStart, gamesEnd, gamesRangeErr := retroDateRange(ctx, database, "games", "date")
	playsStart, playsEnd, playsRangeErr := retroDateRange(ctx, database, "plays", "date")

	echo.Info("")
	echo.Info("• Retrosheet Archives")
	if gamesErr != nil {
		echo.Infof("  ⚠ Unable to read game logs: %v", gamesErr)
	} else if gamesCount == 0 {
		echo.Infof("  • Games table is empty. Run `baseball etl load retrosheet` after downloading archives.")
	} else {
		echo.Successf("  ✓ %d game log rows loaded", gamesCount)
	}
	if gamesRangeErr == nil && gamesStart != nil && gamesEnd != nil {
		echo.Infof("    Game coverage: %s → %s", gamesStart.Format("2006-01-02"), gamesEnd.Format("2006-01-02"))
	} else if gamesRangeErr != nil {
		echo.Infof("    ⚠ Unable to derive game coverage: %v", gamesRangeErr)
	}
	if playsErr != nil {
		echo.Infof("  ⚠ Unable to read play-by-play table: %v", playsErr)
	} else if playsCount > 0 {
		echo.Infof("    Plays ingested: %d rows", playsCount)
	}
	if playsRangeErr == nil && playsStart != nil && playsEnd != nil {
		echo.Infof("    Play coverage: %s → %s", playsStart.Format("2006-01-02"), playsEnd.Format("2006-01-02"))
	}
	if entry, ok := refreshes["retrosheet_games"]; ok {
		entryCopy := entry
		echo.Infof("    Game log ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Game log ETL: never recorded")
	}
	if entry, ok := refreshes["retrosheet_plays"]; ok {
		entryCopy := entry
		echo.Infof("    Plays ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Plays ETL: never recorded")
	}

	echo.Info("")
	echo.Success("✓ Status check completed")
	return nil
}

func startServer(cmd *cobra.Command, args []string) error {
	echo.Header("Starting Server")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")
	echo.Info("Initializing repositories...")

	playerRepo := repository.NewPlayerRepository(database.DB)
	teamRepo := repository.NewTeamRepository(database.DB)
	statsRepo := repository.NewStatsRepository(database.DB)
	awardRepo := repository.NewAwardRepository(database.DB)
	gameRepo := repository.NewGameRepository(database.DB)
	playRepo := repository.NewPlayRepository(database.DB)
	metaRepo := repository.NewMetaRepository(database.DB)

	echo.Info("Registering routes...")

	server := api.NewServer(
		api.NewPlayerRoutes(playerRepo, awardRepo),
		api.NewTeamRoutes(teamRepo),
		api.NewStatsRoutes(statsRepo),
		api.NewGameRoutes(gameRepo),
		api.NewPlayRoutes(playRepo, playerRepo),
		api.NewMetaRoutes(metaRepo),
	)

	addr := ":8080"
	echo.Success(fmt.Sprintf("✓ Server starting on %s", addr))
	echo.Info("Press Ctrl+C to stop")
	echo.Info("")
	echo.Info("Available endpoints:")
	echo.Info("  GET  /v1/health")
	echo.Info("  GET  /v1/players?name={name}&page={page}&per_page={per_page}")
	echo.Info("  GET  /v1/players/{id}")
	echo.Info("  GET  /v1/players/{id}/seasons")
	echo.Info("  GET  /v1/players/{id}/awards")
	echo.Info("  GET  /v1/players/{id}/hall-of-fame")
	echo.Info("  GET  /v1/players/{id}/game-logs?season={year}&page={page}&per_page={per_page}")
	echo.Info("  GET  /v1/players/{id}/appearances")
	echo.Info("  GET  /v1/players/{id}/teams")
	echo.Info("  GET  /v1/players/{id}/salaries")
	echo.Info("  GET  /v1/players/{id}/plays?page={page}&per_page={per_page}")
	echo.Info("  GET  /v1/players/{id}/plate-appearances?season={year}&pitcher={retro_id}")
	echo.Info("  GET  /v1/teams?year={year}&league={league}")
	echo.Info("  GET  /v1/teams/{id}?year={year}")
	echo.Info("  GET  /v1/seasons/{year}/teams?league={league}")
	echo.Info("  GET  /v1/franchises?active={true|false}")
	echo.Info("  GET  /v1/franchises/{id}")
	echo.Info("  GET  /v1/seasons/{year}/leaders/batting?stat={stat}&league={league}&limit={limit}")
	echo.Info("  GET  /v1/seasons/{year}/leaders/pitching?stat={stat}&league={league}&limit={limit}")
	echo.Info("  GET  /v1/stats/batting?player_id={id}&year={year}&team_id={id}")
	echo.Info("  GET  /v1/stats/pitching?player_id={id}&year={year}&team_id={id}")
	echo.Info("  GET  /v1/games?season={year}&team_id={id}&date={date}")
	echo.Info("  GET  /v1/games/{id}")
	echo.Info("  GET  /v1/games/{id}/boxscore")
	echo.Info("  GET  /v1/games/{id}/plays?page={page}&per_page={per_page}")
	echo.Info("  GET  /v1/seasons/{year}/schedule")
	echo.Info("  GET  /v1/seasons/{year}/dates/{date}/games")
	echo.Info("  GET  /v1/seasons/{year}/teams/{team_id}/games")
	echo.Info("  GET  /v1/plays?batter={id}&pitcher={id}&date={YYYYMMDD}")
	echo.Info("  GET  /v1/meta")
	echo.Info("  GET  /v1/meta/datasets")
	echo.Info("")
	return http.ListenAndServe(addr, server)
}

// TODO: configurable baseURL
func fetchEndpoint(cmd *cobra.Command, args []string) error {
	path := args[0]
	format, _ := cmd.Flags().GetString("format")

	baseURL := "http://localhost:8080/v1/"
	url := baseURL + path

	echo.Header("API Test")
	echo.Infof("Fetching: %s", url)
	echo.Info("")

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer resp.Body.Close()

	echo.Infof("Status: %s", resp.Status)
	echo.Info("")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error: failed to read response: %w", err)
	}

	if format == "table" {
		echo.Info("Table format not yet implemented, showing JSON:")
		echo.Info("")
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		echo.Info(string(body))
	} else {
		echo.Info(prettyJSON.String())
	}

	echo.Info("")
	echo.Successf("✓ Request completed (%d bytes)", len(body))
	return nil
}

func checkHealth(cmd *cobra.Command, args []string) error {
	echo.Header("Health Check")

	serverURL := "http://localhost:8080/v1/health"
	echo.Infof("Checking: %s", serverURL)
	echo.Info("")

	resp, err := http.Get(serverURL)
	if err != nil {
		return fmt.Errorf("error: server is not running or unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		echo.Successf("✓ Server is healthy (Status: %s)", resp.Status)

		body, err := io.ReadAll(resp.Body)
		if err == nil && len(body) > 0 {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
				echo.Info("")
				echo.Info(prettyJSON.String())
			}
		}
		return nil
	}

	return fmt.Errorf("error: server returned status: %s", resp.Status)
}

func dirSnapshot(path string) (bool, int, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, time.Time{}, nil
		}
		return false, 0, time.Time{}, err
	}
	if !info.IsDir() {
		return false, 0, time.Time{}, fmt.Errorf("path is not a directory: %s", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	var latest time.Time
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		if entryInfo.ModTime().After(latest) {
			latest = entryInfo.ModTime()
		}
	}

	return true, len(entries), latest, nil
}

func humanizeModTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	ago := time.Since(t)
	return fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04"), ago.Round(time.Minute))
}

func safeCount(ctx context.Context, database *db.DB, query string) (int64, error) {
	var count int64
	if err := database.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func seasonRange(ctx context.Context, database *db.DB) (*int, *int, error) {
	var minYear, maxYear sql.NullInt64
	err := database.QueryRowContext(ctx, `SELECT MIN("yearID"), MAX("yearID") FROM "Teams"`).Scan(&minYear, &maxYear)
	if err != nil {
		return nil, nil, err
	}

	var minPtr, maxPtr *int
	if minYear.Valid {
		v := int(minYear.Int64)
		minPtr = &v
	}
	if maxYear.Valid {
		v := int(maxYear.Int64)
		maxPtr = &v
	}
	return minPtr, maxPtr, nil
}

func retroDateRange(ctx context.Context, database *db.DB, table, column string) (*time.Time, *time.Time, error) {
	query := fmt.Sprintf(`SELECT MIN(%s), MAX(%s) FROM %s`, column, column, table)
	var minVal, maxVal sql.NullString
	if err := database.QueryRowContext(ctx, query).Scan(&minVal, &maxVal); err != nil {
		return nil, nil, err
	}

	start, err := parseRetroDate(minVal)
	if err != nil {
		return nil, nil, err
	}
	end, err := parseRetroDate(maxVal)
	if err != nil {
		return nil, nil, err
	}
	return start, end, nil
}

func parseRetroDate(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}
	t, err := time.Parse("20060102", value.String)
	if err != nil {
		return nil, fmt.Errorf("invalid Retrosheet date %q: %w", value.String, err)
	}
	return &t, nil
}

func formatRefresh(entry *db.DatasetRefresh) string {
	if entry == nil || entry.LastLoadedAt.IsZero() {
		return "not yet recorded"
	}

	return fmt.Sprintf("%s (%s ago, %d rows)",
		entry.LastLoadedAt.Format(time.RFC1123),
		time.Since(entry.LastLoadedAt).Round(time.Minute),
		entry.RowCount,
	)
}
