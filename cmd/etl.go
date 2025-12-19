package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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
	cmd.AddCommand(SalaryLoadCmd())
	cmd.AddCommand(ParksLoadCmd())
	cmd.AddCommand(AllStarLoadCmd())
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
	echo.Info("Downloading player data...")
	allplayersURL := "https://www.retrosheet.org/downloads/allplayers.zip"
	echo.Infof("  Downloading allplayers.zip...")

	resp2, err := http.Get(allplayersURL)
	if err != nil {
		echo.Infof("  ⚠ Failed to download allplayers: %v", err)
	} else {
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			echo.Infof("  ⚠ allplayers.zip not available (HTTP %d)", resp2.StatusCode)
		} else {
			outputPath := filepath.Join(dataDir, "allplayers.zip")
			out, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("error: failed to create allplayers.zip: %w", err)
			}
			defer out.Close()

			if _, err = io.Copy(out, resp2.Body); err != nil {
				return fmt.Errorf("error: failed to save allplayers.zip: %w", err)
			}

			echo.Successf("  ✓ allplayers.zip downloaded")
		}
	}

	allstarDir := filepath.Join(dataDir, "allstar")
	if err := os.MkdirAll(allstarDir, 0755); err != nil {
		return fmt.Errorf("error: failed to create allstar directory: %w", err)
	}

	echo.Info("")
	echo.Info("Downloading all-star data...")
	allstarURL := "https://www.retrosheet.org/downloads/allstar.zip"
	echo.Infof("  Downloading allstar.zip...")

	resp3, err := http.Get(allstarURL)
	if err != nil {
		echo.Infof("  ⚠ Failed to download allstar: %v", err)
	} else {
		defer resp3.Body.Close()

		if resp3.StatusCode != http.StatusOK {
			echo.Infof("  ⚠ allstar.zip not available (HTTP %d)", resp3.StatusCode)
		} else {
			outputPath := filepath.Join(allstarDir, "allstar.zip")
			out, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("error: failed to create allstar.zip: %w", err)
			}
			defer out.Close()

			if _, err = io.Copy(out, resp3.Body); err != nil {
				return fmt.Errorf("error: failed to save allstar.zip: %w", err)
			}

			echo.Successf("  ✓ allstar.zip downloaded")
		}
	}

	echo.Info("")
	echo.Success("✓ Retrosheet data downloaded successfully")
	echo.Infof("  Game logs: %s", gameLogsDir)
	echo.Infof("  Play-by-play: %s", playsDir)
	echo.Infof("  Ejections: %s", ejectionsDir)
	echo.Infof("  Players: %s/allplayers.zip", dataDir)
	echo.Infof("  All-Star: %s", allstarDir)
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

func loadRetrosheetPlayers(cmd *cobra.Command, args []string) error {
	echo.Header("Loading Retrosheet Player Data")
	echo.Info("Connecting to database...")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")

	ctx := cmd.Context()
	csvPath := "data/retrosheet/allplayers.csv"
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		zipPath := "data/retrosheet/allplayers.zip"
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			return fmt.Errorf(`error: allplayers data not found

Run this command first to download the data:
  ./tmp/baseball etl fetch retrosheet`)
		}

		echo.Info("Extracting allplayers.zip...")
		r, err := zip.OpenReader(zipPath)
		if err != nil {
			return fmt.Errorf("error: failed to open allplayers.zip: %w", err)
		}
		defer r.Close()

		for _, f := range r.File {
			if f.Name == "allplayers.csv" {
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("error: failed to read allplayers.csv from zip: %w", err)
				}
				defer rc.Close()

				out, err := os.Create(csvPath)
				if err != nil {
					return fmt.Errorf("error: failed to create allplayers.csv: %w", err)
				}
				defer out.Close()

				if _, err = io.Copy(out, rc); err != nil {
					return fmt.Errorf("error: failed to extract allplayers.csv: %w", err)
				}

				echo.Success("✓ Extracted allplayers.csv")
				break
			}
		}
	}

	rowCount, err := seed.LoadRetrosheetPlayers(ctx, database, csvPath)
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

	ctx := cmd.Context()

	echo.Info("Filling missing park metadata...")
	rows, err := database.LoadMissingParks(ctx)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	echo.Info("")
	echo.Success("✓ Missing parks data loaded successfully")
	echo.Infof("  Parks processed: %d", rows)
	echo.Info("  Refreshed park_map materialized view")

	if err := database.RecordDatasetRefresh(ctx, "parks_metadata", rows); err != nil {
		return fmt.Errorf("error: failed to record parks refresh: %w", err)
	}

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

	ctx := cmd.Context()
	dataDir := "data/retrosheet/allstar"
	zipFile := filepath.Join(dataDir, "allstar.zip")

	if _, err := os.Stat(zipFile); os.IsNotExist(err) {
		return fmt.Errorf(`error: allstar.zip not found at %s

Run this command first to download the data:
  ./tmp/baseball etl fetch retrosheet`, zipFile)
	}

	echo.Info("Extracting all-star data...")
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("error: failed to open allstar.zip: %w", err)
	}
	defer r.Close()

	tmpDir, err := os.MkdirTemp("", "allstar-*")
	if err != nil {
		return fmt.Errorf("error: failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, f := range r.File {
		if f.Name == "gameinfo.csv" || f.Name == "plays.csv" {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("error: failed to read %s from zip: %w", f.Name, err)
			}

			outPath := filepath.Join(tmpDir, f.Name)
			out, err := os.Create(outPath)
			if err != nil {
				rc.Close()
				return fmt.Errorf("error: failed to create %s: %w", f.Name, err)
			}

			_, err = io.Copy(out, rc)
			out.Close()
			rc.Close()

			if err != nil {
				return fmt.Errorf("error: failed to extract %s: %w", f.Name, err)
			}

			echo.Infof("  ✓ Extracted %s", f.Name)
		}
	}

	gameinfoFile := filepath.Join(tmpDir, "gameinfo.csv")
	playsFile := filepath.Join(tmpDir, "plays.csv")

	echo.Info("")
	echo.Info("Loading all-star games...")
	gameRows, err := database.LoadAllStarGameInfo(ctx, gameinfoFile)
	if err != nil {
		return fmt.Errorf("error: failed to load gameinfo: %w", err)
	}
	echo.Successf("✓ Loaded all-star games (%d rows)", gameRows)

	if err := database.RecordDatasetRefresh(ctx, "allstar_games", gameRows); err != nil {
		return fmt.Errorf("error: failed to record all-star games refresh: %w", err)
	}

	echo.Info("")
	echo.Info("Loading all-star plays...")
	playRows, err := database.LoadAllStarPlays(ctx, playsFile)
	if err != nil {
		return fmt.Errorf("error: failed to load plays: %w", err)
	}
	echo.Successf("✓ Loaded all-star plays (%d rows)", playRows)

	if err := database.RecordDatasetRefresh(ctx, "allstar_plays", playRows); err != nil {
		return fmt.Errorf("error: failed to record all-star plays refresh: %w", err)
	}

	echo.Info("")
	echo.Success("✓ All-star data loaded successfully")
	echo.Infof("  Games: %d", gameRows)
	echo.Infof("  Plays: %d", playRows)
	return nil
}
