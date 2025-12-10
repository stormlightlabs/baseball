package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
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
		Use:   "fetch [url]",
		Short: "Test API endpoints",
		Long:  "cURL-like tool for testing API endpoints with formatted output.",
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

// TODO: Implement database loading
func loadLahman(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Lahman Data")
	echo.Info("Loading Lahman CSV files into database...")
	echo.Success("✓ Lahman data loaded (placeholder)")
	return nil
}

// TODO: Implement database loading
func loadRetrosheet(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Retrosheet Data")
	echo.Info("Loading Retrosheet CSV files into database...")
	echo.Success("✓ Retrosheet data loaded (placeholder)")
	return nil
}

// TODO: Implement migration logic
func migrate(cmd *cobra.Command, args []string) error {
	echo.Header("Database Migration")
	echo.Info("Running database migrations...")
	echo.Success("✓ Database migrated (placeholder)")
	return nil
}

// TODO: Implement status checking
func status(cmd *cobra.Command, args []string) error {
	echo.Header("Data Status")
	echo.Info("Checking data freshness and completeness...")
	echo.Success("✓ All data is current (placeholder)")
	return nil
}

// TODO: Implement server startup
func startServer(cmd *cobra.Command, args []string) error {
	echo.Header("Starting Server")
	echo.Info("Starting baseball API server...")
	echo.Success("✓ Server started on :8080 (placeholder)")
	return nil
}

// TODO: Implement API fetching with formatting
func fetchEndpoint(cmd *cobra.Command, args []string) error {
	url := args[0]
	format, _ := cmd.Flags().GetString("format")

	echo.Header("API Test")
	echo.Infof("Fetching: %s", url)
	echo.Infof("Format: %s", format)

	echo.Success("✓ Request completed (placeholder)")
	return nil
}

// TODO: Implement health check
func checkHealth(cmd *cobra.Command, args []string) error {
	echo.Header("Health Check")
	echo.Info("Checking server health...")
	echo.Success("✓ Server is healthy (placeholder)")
	return nil
}
