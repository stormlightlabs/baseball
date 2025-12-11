package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
		return fmt.Errorf("%s failed to create data directory: %w",
			echo.ErrorStyle().Render("Error:"), err)
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
	echo.Info("Downloading Retrosheet game logs and events...")

	dataDir := "data/retrosheet"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create data directory: %w", err)
	}

	retrosheetFiles := map[string]string{
		"GL2025.zip": "https://www.retrosheet.org/gamelogs/gl2025.zip",
		"GL2024.zip": "https://www.retrosheet.org/gamelogs/gl2024.zip",
		"GL2023.zip": "https://www.retrosheet.org/gamelogs/gl2023.zip",
	}

	for filename, url := range retrosheetFiles {
		echo.Infof("Downloading %s...", filename)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error: failed to download %s: %w", filename, err)
		}
		defer resp.Body.Close()

		outputPath := filepath.Join(dataDir, filename)
		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error: failed to create %s: %w", filename, err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("error: failed to save %s: %w", filename, err)
		}

		echo.Successf("✓ %s downloaded", filename)
	}

	echo.Success("✓ Retrosheet data downloaded successfully")
	echo.Infof("  Saved to: %s", dataDir)
	return nil
}

func loadLahman(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Lahman Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("%s %w", echo.ErrorStyle().Render("Error:"), err)
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
			return fmt.Errorf("%s failed to load %s: %w",
				echo.ErrorStyle().Render("Error:"), table, err)
		}

		totalRows += rows
		echo.Successf("✓ Loaded %s (%d rows)", table, rows)
	}

	echo.Success(fmt.Sprintf("✓ All Lahman data loaded successfully (%d total rows)", totalRows))
	return nil
}

func migrate(cmd *cobra.Command, args []string) error {
	echo.Header("Database Migration")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("%s %w", echo.ErrorStyle().Render("Error:"), err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")
	echo.Info("Running migrations...")

	ctx := cmd.Context()
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("%s %w", echo.ErrorStyle().Render("Error:"), err)
	}

	echo.Success("✓ All migrations applied successfully")
	return nil
}

func loadRetrosheet(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Retrosheet Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("%s %w", echo.ErrorStyle().Render("Error:"), err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	dataDir := "data/retrosheet"
	years := []string{"2023", "2024", "2025"}

	ctx := cmd.Context()
	totalRows := int64(0)

	for _, year := range years {
		zipFile := filepath.Join(dataDir, fmt.Sprintf("GL%s.zip", year))

		if _, err := os.Stat(zipFile); os.IsNotExist(err) {
			echo.Infof("Skipping %s (file not found)", year)
			continue
		}

		echo.Infof("Loading %s game logs...", year)

		rows, err := database.LoadRetrosheetGameLog(ctx, zipFile)
		if err != nil {
			return fmt.Errorf("%s failed to load %s: %w",
				echo.ErrorStyle().Render("Error:"), year, err)
		}

		totalRows += rows
		echo.Successf("✓ Loaded %s (%d rows)", year, rows)
	}

	echo.Success(fmt.Sprintf("✓ All Retrosheet data loaded successfully (%d total rows)", totalRows))
	return nil
}

// TODO: Implement status checking
func status(cmd *cobra.Command, args []string) error {
	echo.Header("Data Status")
	echo.Info("Checking data freshness and completeness...")
	echo.Success("✓ All data is current (placeholder)")
	return nil
}

func startServer(cmd *cobra.Command, args []string) error {
	echo.Header("Starting Server")
	echo.Info("Connecting to database...")

	database, err := db.Connect()
	if err != nil {
		return fmt.Errorf("%s %w", echo.ErrorStyle().Render("Error:"), err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")
	echo.Info("Initializing repositories...")

	playerRepo := repository.NewPlayerRepository(database.DB)
	teamRepo := repository.NewTeamRepository(database.DB)
	statsRepo := repository.NewStatsRepository(database.DB)
	awardRepo := repository.NewAwardRepository(database.DB)
	gameRepo := repository.NewGameRepository(database.DB)

	echo.Info("Registering routes...")

	server := api.NewServer(
		api.NewPlayerRoutes(playerRepo, awardRepo),
		api.NewTeamRoutes(teamRepo),
		api.NewStatsRoutes(statsRepo),
		api.NewGameRoutes(gameRepo),
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
	echo.Info("  GET  /v1/seasons/{year}/schedule")
	echo.Info("  GET  /v1/seasons/{year}/dates/{date}/games")
	echo.Info("  GET  /v1/seasons/{year}/teams/{team_id}/games")
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
		return fmt.Errorf("%s %w", echo.ErrorStyle().Render("Error:"), err)
	}
	defer resp.Body.Close()

	echo.Infof("Status: %s", resp.Status)
	echo.Info("")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s failed to read response: %w", echo.ErrorStyle().Render("Error:"), err)
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

	return fmt.Errorf("Error: server returned status: %s", resp.Status)
}
