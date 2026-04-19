package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/seed"
)

// ETLCmd creates the etl command group
func ETLCmd() *cobra.Command {
	opts := &pipelineCLIOptions{}
	cmd := &cobra.Command{
		Use:   "etl",
		Short: "ETL operations for baseball data",
		Long:  "Extract, Transform, and Load operations for Lahman and Retrosheet data sources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runETLPipeline(cmd, opts)
		},
	}
	addPipelineFlags(cmd, opts)
	cmd.AddCommand(EtlFetchCmd())
	cmd.AddCommand(EtlLoadCmd())
	cmd.AddCommand(EtlStatusCmd())
	cmd.AddCommand(EtlRunCmd())
	cmd.AddCommand(EtlValidateCmd())
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
	cmd.AddCommand(ChadwickFetchCmd())
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
	cmd.AddCommand(SalaryLoadCmd())
	cmd.AddCommand(ParksLoadCmd())
	cmd.AddCommand(AllStarLoadCmd())
	cmd.AddCommand(BiodataLoadCmd())
	cmd.AddCommand(MLBAMCrosswalkLoadCmd())
	return cmd
}

type pipelineCLIOptions struct {
	profile string
	mode    string
	years   string
	era     string
}

func addPipelineFlags(cmd *cobra.Command, opts *pipelineCLIOptions) {
	cmd.Flags().StringVar(&opts.profile, "profile", string(seed.PipelineProfileDev), "Pipeline profile to run (dev|prod)")
	cmd.Flags().StringVar(&opts.mode, "mode", string(seed.PipelineModeIncremental), "Pipeline execution mode (incremental|full)")
	cmd.Flags().StringVar(&opts.years, "years", "", "Comma-separated years, ranges, or 'all', e.g. 2022,2023-2025,all")
	cmd.Flags().StringVar(&opts.era, "era", "", "Comma-separated era names to include (fed,nlg,boomer,pitcher,turf,steroid,moneyball,statcast,modern)")
}

func EtlRunCmd() *cobra.Command {
	opts := &pipelineCLIOptions{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the full ETL pipeline",
		Long:  "Run the full ETL pipeline (extract, transform, load, validate).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runETLPipeline(cmd, opts)
		},
	}
	addPipelineFlags(cmd, opts)
	return cmd
}

func EtlValidateCmd() *cobra.Command {
	var profile string
	var yearsFlag string
	var eraFlag string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate ETL completeness",
		Long:  "Validate core and auxiliary ETL datasets for a profile.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateETLPipeline(cmd, profile, yearsFlag, eraFlag)
		},
	}
	cmd.Flags().StringVar(&profile, "profile", string(seed.PipelineProfileDev), "Pipeline profile to validate (dev|prod)")
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Optional explicit years to validate coverage against")
	cmd.Flags().StringVar(&eraFlag, "era", "", "Optional comma-separated era names to include in coverage validation")
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
		Long: fmt.Sprintf(
			"Download Retrosheet game logs and event files.\nIf --years is omitted, defaults to %s.",
			defaultRetrosheetYears,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchRetrosheet(cmd, yearsFlag, force)
		},
	}
	cmd.Flags().StringVar(&yearsFlag, "years", "", fmt.Sprintf("Comma-separated years, ranges, or 'all', e.g. 2022,2023-2025,all (defaults to %s)", defaultRetrosheetYears))
	cmd.Flags().BoolVar(&force, "force", false, "Force redownload even if files exist")
	return cmd
}

// NegroLeaguesFetchCmd creates the fetch negroleagues command
func NegroLeaguesFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "negroleagues",
		Short: "Download Negro Leagues data from Retrosheet",
		Long:  "Download Negro Leagues event files from Retrosheet.",
		RunE:  fetchNegroLeagues,
	}
}

// ChadwickFetchCmd creates the fetch chadwick command
func ChadwickFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chadwick",
		Short: "Download Chadwick register data",
		Long:  "Download Chadwick register people.csv used for MLBAM player crosswalks.",
		RunE:  fetchChadwick,
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
		Long: fmt.Sprintf(
			"Load Retrosheet CSV files into PostgreSQL database.\nIf both --era and --years are omitted, defaults to %s.",
			defaultRetrosheetYears,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			return loadRetrosheet(cmd, eraFlag, yearsFlag)
		},
	}
	cmd.Flags().StringVar(&eraFlag, "era", "", retrosheetEraHelp())
	cmd.Flags().StringVar(&yearsFlag, "years", "", fmt.Sprintf("Comma-separated years, ranges, or 'all', e.g. 2022,2023-2025,all (defaults to %s when --era is omitted)", defaultRetrosheetYears))
	cmd.AddCommand(RetrosheetPlayersLoadCmd())
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

// SalaryLoadCmd creates the load salary command
func SalaryLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "salary",
		Short: "Load salary data into database",
		Long:  "Enriches the Salaries table with additional salary data by matching player names to Lahman IDs. Also loads salary summary statistics.",
		RunE:  loadSalaryData,
	}
}

// ParksLoadCmd creates the load parks command
func ParksLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parks",
		Short: "Load missing parks data into database",
		Long:  "Fills gaps in the Parks table for high-usage Negro Leagues parks and modern parks lacking metadata. Also refreshes the park_map materialized view.",
		RunE:  loadParksData,
	}
}

// AllStarLoadCmd creates the load allstar command
func AllStarLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "allstar",
		Short: "Load all-star game data into database",
		Long:  "Load all-star game metadata and play-by-play data from Retrosheet allstar.zip into the games and plays tables.",
		RunE:  loadAllStar,
	}
}

// RetrosheetPlayersLoadCmd creates the load retrosheet players command
func RetrosheetPlayersLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "players",
		Short: "Load Retrosheet player data into database",
		Long:  "Load Retrosheet allplayers.csv with per-team-season appearances, pitcher roles, and exact game dates.",
		RunE:  loadRetrosheetPlayers,
	}
}

func runETLPipeline(cmd *cobra.Command, opts *pipelineCLIOptions) error {
	years, err := parseYearFlag(opts.years)
	if err != nil {
		return err
	}

	eras, err := parseEraFlagList(opts.era)
	if err != nil {
		return err
	}

	echo.Header("ETL Pipeline")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	result, err := seed.RunPipeline(cmd.Context(), database, seed.PipelineOptions{
		Profile:  seed.PipelineProfile(strings.ToLower(strings.TrimSpace(opts.profile))),
		Mode:     seed.PipelineMode(strings.ToLower(strings.TrimSpace(opts.mode))),
		Years:    years,
		EraNames: eras,
	})
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	warnings := result.Validation.Warnings()
	if len(warnings) > 0 {
		echo.Info("")
		echo.Info("Validation warnings:")
		for _, warning := range warnings {
			echo.Infof("  • [%s] %s", warning.Dataset, warning.Message)
		}
	}

	echo.Info("")
	echo.Success("✓ ETL pipeline completed successfully")
	echo.Infof("  Run ID: %d", result.RunID)
	echo.Infof("  Years: %s", describeYears(result.Years))
	return nil
}

func validateETLPipeline(cmd *cobra.Command, profile, yearsFlag, eraFlag string) error {
	years, err := parseYearFlag(yearsFlag)
	if err != nil {
		return err
	}

	eras, err := parseEraFlagList(eraFlag)
	if err != nil {
		return err
	}

	opts, err := seed.NormalizePipelineOptions(seed.PipelineOptions{
		Profile:  seed.PipelineProfile(strings.ToLower(strings.TrimSpace(profile))),
		Mode:     seed.PipelineModeIncremental,
		Years:    years,
		EraNames: eras,
	})
	if err != nil {
		return err
	}

	echo.Header("ETL Validation")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	result, err := seed.ValidatePipeline(cmd.Context(), database, opts.Profile, opts.Years)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	if len(result.Issues) == 0 {
		echo.Success("✓ Validation passed")
		return nil
	}

	for _, issue := range result.Issues {
		if issue.Severity == "warning" {
			echo.Infof("⚠ [%s] %s", issue.Dataset, issue.Message)
			continue
		}
		echo.Infof("✗ [%s] %s", issue.Dataset, issue.Message)
	}

	if !result.OK() {
		return fmt.Errorf("validation failed with %d error(s)", len(result.Errors()))
	}

	echo.Success("✓ Validation passed with warnings")
	return nil
}

func parseEraFlagList(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	eras := make([]string, 0, len(parts))

	for _, part := range parts {
		era := normalizeEraFlag(part)
		if era == "" {
			continue
		}
		if seed.GetEra(era) == nil {
			return nil, unknownEraError(era)
		}
		eras = append(eras, era)
	}

	if len(eras) == 0 {
		return nil, nil
	}

	slices.Sort(eras)
	eras = slices.Compact(eras)
	return eras, nil
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
	dataDir := filepath.Join("data", "retrosheet", "negroleagues")
	if err := seed.FetchNegroLeaguesData(cmd.Context(), dataDir, false); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	echo.Success("✓ Negro Leagues data downloaded and extracted")
	echo.Infof("  Directory: %s", dataDir)
	return nil
}

func fetchChadwick(cmd *cobra.Command, args []string) error {
	echo.Header("Fetching Chadwick Register Data")
	dataDir := filepath.Join("data", "chadwick")
	if err := seed.FetchChadwickRegisterData(cmd.Context(), dataDir, false); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	echo.Success("✓ Chadwick register downloaded successfully")
	echo.Infof("  File: %s", filepath.Join(dataDir, "people.csv"))
	return nil
}

func fetchRetrosheet(cmd *cobra.Command, yearsFlag string, force bool) error {
	echo.Header("Fetching Retrosheet Data")
	years, err := parseYearFlag(yearsFlag)
	if err != nil {
		return err
	}

	if len(years) == 0 {
		years = []int{2023, 2024, 2025}
	}

	dataDir := filepath.Join("data", "retrosheet")
	if err := seed.FetchRetrosheetData(cmd.Context(), dataDir, years, force); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Success("✓ Retrosheet data downloaded successfully")
	echo.Infof("  Years: %s", describeYears(years))
	echo.Infof("  Directory: %s", dataDir)
	return nil
}

func describeYears(years []int) string {
	if len(years) == 0 {
		return "0 years"
	}
	sorted := slices.Clone(years)
	slices.Sort(sorted)
	return fmt.Sprintf("%d years: %d-%d", len(sorted), sorted[0], sorted[len(sorted)-1])
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

	if _, err := seed.LoadLahman(cmd.Context(), database, seed.LahmanOptions{CSVDir: csvDir}); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

func loadRetrosheet(cmd *cobra.Command, eraFlag, yearsFlag string) error {
	echo.Header("Loading Retrosheet Data")

	var yearInts []int
	var err error

	if eraFlag != "" {
		eraFlag = normalizeEraFlag(eraFlag)
		echo.Infof("Loading data for era: %s", eraFlag)
		yearInts = seed.GetYearsForEras([]string{eraFlag})
		if len(yearInts) == 0 {
			return unknownEraError(eraFlag)
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

func loadFanGraphs(cmd *cobra.Command, args []string) error {
	echo.Header("Loading FanGraphs Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	rows, err := seed.LoadFanGraphsData(cmd.Context(), database, filepath.Join("data", "fangraphs"))
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ FanGraphs data loaded successfully")
	echo.Infof("  Rows loaded: %d", rows)
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

	totalRows, err := seed.LoadNegroLeagues(cmd.Context(), database, filepath.Join("data", "retrosheet", "negroleagues"))
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ Negro Leagues data loaded successfully")
	echo.Infof("  Rows loaded: %d", totalRows)
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

func loadRetrosheetPlayers(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Retrosheet Player Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	csvPath, err := seed.EnsureRetrosheetPlayersCSV(filepath.Join("data", "retrosheet"))
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	rowCount, err := seed.LoadRetrosheetPlayers(cmd.Context(), database, csvPath)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ Retrosheet player data loaded successfully")
	echo.Infof("  Rows loaded: %d", rowCount)
	echo.Infof("  Coverage: per-team-season appearances (1898-2025)")
	return nil
}

func loadSalaryData(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Salary Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()
	dataDir := "data/salaries"

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return fmt.Errorf(`error: salary data directory not found: %s

The salary data directory should contain:
  - Individual year CSV files (2000.csv, 2001.csv, etc.)
  - summary.csv with yearly aggregate statistics

Expected format:
  Year,Player,Pos,Salary`, dataDir)
	}

	_, err = seed.LoadSalaryData(ctx, database, dataDir)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ Salary data loaded successfully")
	echo.Infof("  Data enriches Lahman Salaries table with player name matching")
	return nil
}

func loadParksData(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Missing Parks Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	echo.Info("Filling missing park metadata...")
	rows, err := seed.LoadParksData(cmd.Context(), database)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ Missing parks data loaded successfully")
	echo.Infof("  Parks processed: %d", rows)
	echo.Info("  Refreshed park_map materialized view")

	return nil
}

func loadAllStar(cmd *cobra.Command, args []string) error {
	echo.Header("Loading All-Star Game Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	totalRows, err := seed.LoadAllStarData(cmd.Context(), database, filepath.Join("data", "retrosheet", "allstar", "allstar.zip"))
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ All-star data loaded successfully")
	echo.Infof("  Rows loaded: %d", totalRows)
	return nil
}

// BiodataLoadCmd creates the load biodata command
func BiodataLoadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "biodata",
		Short: "Load Retrosheet biodata into database",
		Long:  "Load Retrosheet biodata including player biographical info, relatives, coaches, and umpires.",
		RunE:  loadBiodata,
	}
}

func loadBiodata(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Retrosheet Biodata")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	tmpDir, cleanup, err := seed.ExtractBiodataArchive(filepath.Join("data", "retrosheet"))
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer cleanup()

	totalRows, err := seed.LoadBiodata(cmd.Context(), database, tmpDir)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ All biodata loaded successfully")
	echo.Infof("  Total rows: %d", totalRows)
	return nil
}

func loadMLBAMCrosswalk(cmd *cobra.Command, yearsFlag string) error {
	echo.Header("Loading MLBAM Crosswalk Data")
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
	if len(years) == 0 {
		years = []int{time.Now().Year()}
	}

	playerRows, err := seed.LoadPlayerMLBAMMappings(cmd.Context(), database, filepath.Join("data", "chadwick"))
	if err != nil {
		return fmt.Errorf("error loading player MLBAM mappings: %w", err)
	}

	teamRows, err := seed.LoadTeamMLBAMMappings(cmd.Context(), database, years)
	if err != nil {
		return fmt.Errorf("error loading team MLBAM mappings: %w", err)
	}

	echo.Info("")
	echo.Success("✓ MLBAM crosswalk loaded successfully")
	echo.Infof("  Player mappings: %d", playerRows)
	echo.Infof("  Team mappings: %d", teamRows)
	echo.Infof("  Team years: %s", describeYears(years))
	return nil
}

// MLBAMCrosswalkLoadCmd creates the load crosswalk command.
func MLBAMCrosswalkLoadCmd() *cobra.Command {
	var yearsFlag string
	cmd := &cobra.Command{
		Use:   "crosswalk",
		Short: "Load persisted MLBAM crosswalk data",
		Long:  "Load MLBAM player/team crosswalk mappings into player_mlbam_map and team_mlbam_map.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return loadMLBAMCrosswalk(cmd, yearsFlag)
		},
	}
	cmd.Flags().StringVar(&yearsFlag, "years", "", "Comma-separated years, ranges, or 'all' for team MLBAM mappings")
	return cmd
}
