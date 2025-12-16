package cmd

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/config"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/seed"
)

// ETLCmd creates the etl command group
func ETLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "etl",
		Short: "ETL operations for baseball data",
		Long:  "Extract, Transform, and Load operations for Lahman and Retrosheet data sources.",
	}
	cmd.AddCommand(EtlFetchCmd())
	cmd.AddCommand(EtlLoadCmd())
	cmd.AddCommand(EtlStatusCmd())
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
	cmd.AddCommand(DbResetCmd())
	cmd.AddCommand(DbRepopulateCmd())
	cmd.AddCommand(DbRecreateCmd())
	cmd.AddCommand(DbRefreshViewsCmd())
	return cmd
}

// EtlFetchCmd creates the fetch command group under etl
func EtlFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Download baseball data sources",
		Long:  "Download data from Lahman and Retrosheet sources.",
	}
	cmd.AddCommand(LahmanFetchCmd())
	cmd.AddCommand(RetrosheetFetchCmd())
	cmd.AddCommand(NegroLeaguesFetchCmd())
	return cmd
}

// EtlLoadCmd creates the load command group under etl
func EtlLoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load data into database",
		Long:  "Load downloaded data into PostgreSQL database.",
	}
	cmd.AddCommand(LahmanLoadCmd())
	cmd.AddCommand(RetrosheetLoadCmd())
	cmd.AddCommand(NegroLeaguesLoadCmd())
	cmd.AddCommand(FanGraphsLoadCmd())
	cmd.AddCommand(WeatherLoadCmd())
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
	var yearsFlag string
	var force bool
	cmd := &cobra.Command{
		Use:   "retrosheet",
		Short: "Download Retrosheet data",
		Long:  "Download Retrosheet game logs and event files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchRetrosheet(cmd, yearsFlag, force)
		},
	}
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years, ranges, or 'all', e.g. 2022,2023-2025,all")
	cmd.Flags().BoolVar(&force, "force", false, "Force redownload even if files exist")
	return cmd
}

// NegroLeaguesFetchCmd creates the fetch negroleagues command
func NegroLeaguesFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "negroleagues",
		Short: "Get instructions to download Negro Leagues data",
		Long:  "Provides instructions for downloading Negro Leagues event files from Retrosheet.",
		RunE:  fetchNegroLeagues,
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
	var eraFlag string
	var yearsFlag string
	cmd := &cobra.Command{
		Use:   "retrosheet",
		Short: "Load Retrosheet data into database",
		Long:  "Load Retrosheet CSV files into PostgreSQL database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return loadRetrosheet(cmd, eraFlag, yearsFlag)
		},
	}
	cmd.Flags().StringVar(&eraFlag, "era", "", "Load data for a specific era (federal, nlg, 1970s, 1980s, steroid, modern)")
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years or ranges, e.g. 2022,2023-2025")
	return cmd
}

// NegroLeaguesLoadCmd creates the load negroleagues command
func NegroLeaguesLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "negroleagues",
		Short: "Load Negro Leagues data into database",
		Long:  "Load Negro Leagues gameinfo and plays data from CSV files into separate tables.",
		RunE:  loadNegroLeagues,
	}
}

// FanGraphsLoadCmd creates the load fangraphs command
func FanGraphsLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fangraphs",
		Short: "Load FanGraphs constants into database",
		Long:  "Load FanGraphs wOBA constants and park factors from CSV files.",
		RunE:  loadFanGraphs,
	}
}

// WeatherLoadCmd creates the load weather command
func WeatherLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "weather",
		Short: "Load weather data into database",
		Long:  "Updates existing games with weather and game metadata from Retrosheet's master gameinfo.csv file.",
		RunE:  loadWeatherData,
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

func fetchNegroLeagues(cmd *cobra.Command, args []string) error {
	echo.Header("Fetching Negro Leagues Data")
	dataDir := "data/retrosheet/negroleagues"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create data directory: %w", err)
	}

	zipURL := "https://www.retrosheet.org/downloads/negroleagues.zip"
	zipFile := filepath.Join(dataDir, "negroleagues.zip")

	if _, err := os.Stat(zipFile); err == nil {
		echo.Info("Negro Leagues zip already downloaded")
	} else {
		echo.Infof("Downloading Negro Leagues data...")

		resp, err := http.Get(zipURL)
		if err != nil {
			return fmt.Errorf("error: failed to download: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error: download failed (HTTP %d)", resp.StatusCode)
		}

		out, err := os.Create(zipFile)
		if err != nil {
			return fmt.Errorf("error: failed to create file: %w", err)
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err != nil {
			return fmt.Errorf("error: failed to save file: %w", err)
		}

		echo.Success("✓ Downloaded Negro Leagues data")
	}

	echo.Info("Extracting files...")
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("error: failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		outPath := filepath.Join(dataDir, f.Name)

		if _, err := os.Stat(outPath); err == nil {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("error: failed to open %s in zip: %w", f.Name, err)
		}

		outFile, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("error: failed to create %s: %w", outPath, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("error: failed to extract %s: %w", f.Name, err)
		}
	}

	echo.Success("✓ Extracted all files")
	echo.Info("")
	echo.Info("Extracted files:")
	echo.Info("  • gameinfo.csv - Game metadata")
	echo.Info("  • plays.csv - Play-by-play data")
	echo.Info("  • batting.csv, pitching.csv, fielding.csv - Statistics")
	echo.Info("")
	echo.Infof("  Directory: %s", dataDir)
	echo.Info("")
	echo.Info("Next step: baseball etl load negroleagues")
	return nil
}

func fetchRetrosheet(_ *cobra.Command, yearsFlag string, force bool) error {
	echo.Header("Fetching Retrosheet Data")
	years, err := parseYearFlag(yearsFlag)
	if err != nil {
		return err
	}

	if len(years) == 0 {
		years = []int{2023, 2024, 2025}
	}

	echo.Infof("Downloading Retrosheet data for %d years...", len(years))

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

	gameLogFiles := make(map[string]string)
	playFiles := make(map[string]string)

	for _, year := range years {
		gameLogFiles[fmt.Sprintf("GL%d.zip", year)] = fmt.Sprintf("https://www.retrosheet.org/gamelogs/gl%d.zip", year)
		playFiles[fmt.Sprintf("%dplays.zip", year)] = fmt.Sprintf("https://www.retrosheet.org/downloads/plays/%dplays.zip", year)
	}

	ejectionsDir := filepath.Join(dataDir, "ejections")
	if err := os.MkdirAll(ejectionsDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create ejections directory: %w", err)
	}

	echo.Info("Downloading game logs...")
	for filename, url := range gameLogFiles {
		outputPath := filepath.Join(gameLogsDir, filename)

		if !force {
			if _, err := os.Stat(outputPath); err == nil {
				echo.Infof("  ✓ Using cached %s", filename)
				continue
			}
		}

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

		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error: failed to create %s: %w", filename, err)
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err != nil {
			return fmt.Errorf("error: failed to save %s: %w", filename, err)
		}

		echo.Successf("  ✓ %s downloaded", filename)
	}

	echo.Info("")
	echo.Info("Downloading play-by-play data...")
	for filename, url := range playFiles {
		outputPath := filepath.Join(playsDir, filename)

		if !force {
			if _, err := os.Stat(outputPath); err == nil {
				echo.Infof("  ✓ Using cached %s", filename)
				continue
			}
		}

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

		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error: failed to create %s: %w", filename, err)
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err != nil {
			return fmt.Errorf("error: failed to save %s: %w", filename, err)
		}

		echo.Successf("  ✓ %s downloaded", filename)
	}

	echo.Info("")
	echo.Info("Downloading ejections data...")
	ejectionsURL := "https://www.retrosheet.org/ejections.zip"
	echo.Infof("  Downloading ejections.zip...")

	resp, err := http.Get(ejectionsURL)
	if err != nil {
		echo.Infof("  ⚠ Failed to download ejections: %v", err)
	} else {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			echo.Infof("  ⚠ ejections.zip not available (HTTP %d)", resp.StatusCode)
		} else {
			outputPath := filepath.Join(ejectionsDir, "ejections.zip")
			out, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("error: failed to create ejections.zip: %w", err)
			}
			defer out.Close()

			if _, err = io.Copy(out, resp.Body); err != nil {
				return fmt.Errorf("error: failed to save ejections.zip: %w", err)
			}

			echo.Successf("  ✓ ejections.zip downloaded")
		}
	}

	echo.Info("")
	echo.Success("✓ Retrosheet data downloaded successfully")
	echo.Infof("  Game logs: %s", gameLogsDir)
	echo.Infof("  Play-by-play: %s", playsDir)
	echo.Infof("  Ejections: %s", ejectionsDir)
	return nil
}

func loadLahman(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Lahman Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	dataDir := "data/lahman"
	csvDir := filepath.Join(dataDir, "csv")

	tables := []string{
		"AllstarFull", "Appearances",
		"AwardsManagers", "AwardsPlayers", "AwardsShareManagers", "AwardsSharePlayers",
		"Batting", "BattingPost", "CollegePlaying",
		"Fielding", "FieldingOF", "FieldingOFsplit", "FieldingPost",
		"HomeGames", "HallOfFame", "Managers", "ManagersHalf",
		"Parks", "People", "Pitching", "PitchingPost",
		"Salaries", "Schools", "SeriesPost",
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

func loadRetrosheet(cmd *cobra.Command, eraFlag, yearsFlag string) error {
	echo.Header("Loading Retrosheet Data")

	var yearInts []int
	var err error

	if eraFlag != "" {
		echo.Infof("Loading data for era: %s", eraFlag)
		yearInts = seed.GetYearsForEras([]string{eraFlag})
		if len(yearInts) == 0 {
			return fmt.Errorf("unknown era: %s", eraFlag)
		}
		era := seed.GetEra(eraFlag)
		if era != nil {
			echo.Infof("Era: %s (%d-%d)", era.Name, era.StartYear, era.EndYear)
		}
	} else if yearsFlag != "" {
		yearInts, err = parseYearFlag(yearsFlag)
		if err != nil {
			return err
		}
	} else {
		yearInts = []int{2023, 2024, 2025}
	}

	years := make([]string, len(yearInts))
	for i, y := range yearInts {
		years[i] = fmt.Sprintf("%d", y)
	}

	echo.Infof("Loading data for %d years: %v", len(years), years)

	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	dataDir := "data/retrosheet"
	gameLogsDir := filepath.Join(dataDir, "gamelogs")

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

		rows, err := database.LoadRetrosheetGameLog(ctx, zipFile, "regular")
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
	var emptyPlayYears []string
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

		if rows == 0 {
			emptyPlayYears = append(emptyPlayYears, year)
			echo.Infof("  No plays found for %s (file empty)", year)
		} else {
			echo.Successf("  ✓ Loaded %s (%d rows)", year, rows)
		}
		playsLoaded += rows
		totalRows += rows
	}

	if len(emptyPlayYears) > 0 {
		if eraFlag == "nlg" {
			echo.Info("  Retrosheet annual play-by-play zips for Negro Leagues are empty; plays are loaded from data/retrosheet/negroleagues/plays.csv.")
		} else {
			echo.Infof("  No play-by-play rows found for: %s", strings.Join(emptyPlayYears, ", "))
		}
	}

	echo.Info("")
	echo.Info("Loading ejections data...")
	ejectionsDir := filepath.Join(dataDir, "ejections")
	ejectionsZip := filepath.Join(ejectionsDir, "ejections.zip")
	ejectionsLoaded := int64(0)

	if _, err := os.Stat(ejectionsZip); os.IsNotExist(err) {
		echo.Info("  Skipping ejections (file not found)")
	} else {
		echo.Info("  Loading ejections...")

		rows, err := database.LoadRetrosheetEjections(ctx, ejectionsZip)
		if err != nil {
			return fmt.Errorf("error: failed to load ejections: %w", err)
		}

		ejectionsLoaded = rows
		totalRows += rows
		echo.Successf("  ✓ Loaded ejections (%d rows)", rows)
	}

	echo.Info("")
	echo.Info("Loading Negro Leagues data (if available)...")
	negroLeagueDir := filepath.Join(dataDir, "negroleagues")
	negroLgGameRows, negroLgPlayRows, err := database.LoadNegroLeaguesData(ctx, negroLeagueDir)
	if err != nil {
		return fmt.Errorf("error: failed to load Negro Leagues data: %w", err)
	}

	if negroLgGameRows == 0 && negroLgPlayRows == 0 {
		echo.Info("  Negro Leagues files not found (expected gameinfo.csv and plays.csv)")
	} else {
		totalRows += negroLgGameRows + negroLgPlayRows
		gamesLoaded += negroLgGameRows
		playsLoaded += negroLgPlayRows

		if negroLgGameRows > 0 {
			echo.Successf("  ✓ Loaded Negro Leagues games (%d rows)", negroLgGameRows)
			if err := database.RecordDatasetRefresh(ctx, "negroleagues_games", negroLgGameRows); err != nil {
				return fmt.Errorf("error: failed to record Negro Leagues games refresh: %w", err)
			}
		}
		if negroLgPlayRows > 0 {
			echo.Successf("  ✓ Loaded Negro Leagues plays (%d rows)", negroLgPlayRows)
			if err := database.RecordDatasetRefresh(ctx, "negroleagues_plays", negroLgPlayRows); err != nil {
				return fmt.Errorf("error: failed to record Negro Leagues plays refresh: %w", err)
			}
		}
	}

	echo.Info("")
	echo.Success("✓ All Retrosheet data loaded successfully")
	echo.Infof("  Total rows: %d", totalRows)
	echo.Infof("  Game logs: %d", gamesLoaded)
	echo.Infof("  Play-by-play rows: %d", playsLoaded)
	echo.Infof("  Ejections: %d", ejectionsLoaded)
	if err := database.RecordDatasetRefresh(ctx, "retrosheet_games", gamesLoaded); err != nil {
		return fmt.Errorf("error: failed to record Retrosheet games refresh: %w", err)
	}
	if err := database.RecordDatasetRefresh(ctx, "retrosheet_plays", playsLoaded); err != nil {
		return fmt.Errorf("error: failed to record Retrosheet plays refresh: %w", err)
	}
	if ejectionsLoaded > 0 {
		if err := database.RecordDatasetRefresh(ctx, "retrosheet_ejections", ejectionsLoaded); err != nil {
			return fmt.Errorf("error: failed to record Retrosheet ejections refresh: %w", err)
		}
	}
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

func parseYearFlag(flagValue string) ([]int, error) {
	if strings.TrimSpace(flagValue) == "" {
		return nil, nil
	}

	var years []int
	tokens := strings.SplitSeq(flagValue, ",")
	for token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if token == "all" {
			currentYear := time.Now().Year()
			for year := 1910; year <= currentYear; year++ {
				years = append(years, year)
			}
			continue
		}

		if strings.Contains(token, "-") {
			parts := strings.SplitN(token, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid year in range: %s", parts[0])
			}
			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid year in range: %s", parts[1])
			}
			if end < start {
				return nil, fmt.Errorf("invalid range %s: end before start", token)
			}
			for year := start; year <= end; year++ {
				years = append(years, year)
			}
			continue
		}

		year, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid year: %s", token)
		}
		years = append(years, year)
	}

	if len(years) == 0 {
		return nil, nil
	}

	sort.Ints(years)
	years = uniqueInts(years)
	return years, nil
}

func uniqueInts(values []int) []int {
	if len(values) == 0 {
		return values
	}

	result := make([]int, 0, len(values))
	prev := values[0]
	result = append(result, prev)

	for _, v := range values[1:] {
		if v == prev {
			continue
		}
		result = append(result, v)
		prev = v
	}

	return result
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

func loadFanGraphs(cmd *cobra.Command, args []string) error {
	echo.Header("Loading FanGraphs Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()

	wobaFile := "data/fangraphs/woba.csv"
	echo.Info("Loading wOBA constants...")

	if _, err := os.Stat(wobaFile); os.IsNotExist(err) {
		return fmt.Errorf("error: wOBA constants file not found: %s", wobaFile)
	}

	wobaRows, err := database.LoadFanGraphsWOBA(ctx, wobaFile)
	if err != nil {
		return fmt.Errorf("error: failed to load wOBA constants: %w", err)
	}

	echo.Successf("✓ Loaded wOBA constants (%d rows)", wobaRows)

	parkFactorDir := "data/fangraphs/pf"
	echo.Info("Loading park factors...")

	files, err := filepath.Glob(filepath.Join(parkFactorDir, "*.csv"))
	if err != nil {
		return fmt.Errorf("error: failed to list park factor files: %w", err)
	}

	if len(files) == 0 {
		echo.Info("  No park factor files found")
		return nil
	}

	totalParkRows := int64(0)
	for _, file := range files {
		basename := filepath.Base(file)
		echo.Infof("  Loading %s...", basename)

		rows, err := database.LoadFanGraphsParks(ctx, file)
		if err != nil {
			return fmt.Errorf("error: failed to load %s: %w", basename, err)
		}

		totalParkRows += rows
		echo.Successf("  ✓ Loaded %s (%d rows)", basename, rows)
	}

	echo.Info("")
	echo.Success("✓ All FanGraphs data loaded successfully")
	echo.Infof("  wOBA constants: %d rows", wobaRows)
	echo.Infof("  Park factors: %d rows", totalParkRows)
	return nil
}

func loadNegroLeagues(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Negro Leagues Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()
	dataDir := "data/retrosheet/negroleagues"

	gameinfoFile := filepath.Join(dataDir, "gameinfo.csv")
	echo.Info("Loading Negro Leagues games from gameinfo.csv...")

	if _, err := os.Stat(gameinfoFile); os.IsNotExist(err) {
		return fmt.Errorf("error: gameinfo.csv not found at %s", gameinfoFile)
	}

	gameRows, err := database.LoadNegroLeaguesGameInfo(ctx, gameinfoFile)
	if err != nil {
		return fmt.Errorf("error: failed to load gameinfo: %w", err)
	}

	echo.Successf("✓ Loaded Negro Leagues games (%d rows)", gameRows)
	if err := database.RecordDatasetRefresh(ctx, "negroleagues_games", gameRows); err != nil {
		return fmt.Errorf("error: failed to record Negro Leagues games refresh: %w", err)
	}

	playsFile := filepath.Join(dataDir, "plays.csv")
	echo.Info("Loading Negro Leagues plays from plays.csv...")

	if _, err := os.Stat(playsFile); os.IsNotExist(err) {
		echo.Info("  Skipping plays (file not found)")
	} else {
		playRows, err := database.LoadNegroLeaguesPlays(ctx, playsFile)
		if err != nil {
			return fmt.Errorf("error: failed to load plays: %w", err)
		}
		echo.Successf("✓ Loaded Negro Leagues plays (%d rows)", playRows)
		if err := database.RecordDatasetRefresh(ctx, "negroleagues_plays", playRows); err != nil {
			return fmt.Errorf("error: failed to record Negro Leagues plays refresh: %w", err)
		}
	}

	echo.Info("")
	echo.Success("✓ All Negro Leagues data loaded successfully")
	return nil
}

func loadWeatherData(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Game Weather Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()

	csvPath := "data/retrosheet/gameinfo.csv"
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf(`error: gameinfo.csv not found at %s

The gameinfo.csv file should be downloaded as part of the Retrosheet data.
It contains weather and game metadata for 224K games (1898-2025).`, csvPath)
	}

	_, err = seed.LoadWeatherData(ctx, database, csvPath)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ Game weather data loaded successfully")
	echo.Infof("  Coverage: 1898-2025 (weather details from 2015+)")
	return nil
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
