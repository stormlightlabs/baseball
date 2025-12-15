package seed

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

// LahmanOptions controls how Lahman data is ingested.
type LahmanOptions struct {
	CSVDir string
	Tables []string
}

// RetrosheetOptions controls Retrosheet ingestion.
type RetrosheetOptions struct {
	DataDir string
	Years   []int
	Force   bool
}

// RetrosheetResult contains counters for Retrosheet loads.
type RetrosheetResult struct {
	GameRows int64
	PlayRows int64
}

// LoadLahman loads Lahman CSVs into the database, truncating the target tables first.
func LoadLahman(ctx context.Context, database *db.DB, opts LahmanOptions) (int64, error) {
	csvDir := opts.CSVDir
	if csvDir == "" {
		csvDir = filepath.Join("data", "lahman", "csv")
	}

	tables := opts.Tables
	if len(tables) == 0 {
		tables = defaultLahmanTables()
	}

	var totalRows int64

	for _, table := range tables {
		csvFile := filepath.Join(csvDir, table+".csv")

		if _, err := os.Stat(csvFile); errors.Is(err, os.ErrNotExist) {
			echo.Infof("Skipping %s (file not found)", table)
			continue
		} else if err != nil {
			return 0, fmt.Errorf("error: failed to stat %s: %w", csvFile, err)
		}

		if err := truncateTable(ctx, database, table); err != nil {
			return 0, err
		}

		echo.Infof("Loading %s...", table)

		rows, err := database.CopyCSV(ctx, table, csvFile)
		if err != nil {
			return 0, fmt.Errorf("error: failed to load %s: %w", table, err)
		}

		totalRows += rows
		echo.Successf("✓ Loaded %s (%d rows)", table, rows)
	}

	echo.Success(fmt.Sprintf("✓ All Lahman data loaded successfully (%d total rows)", totalRows))
	if err := database.RecordDatasetRefresh(ctx, "lahman", totalRows); err != nil {
		return totalRows, fmt.Errorf("error: failed to record Lahman refresh: %w", err)
	}

	return totalRows, nil
}

// LoadRetrosheet loads Retrosheet game logs and plays for the requested years.
func LoadRetrosheet(ctx context.Context, database *db.DB, opts RetrosheetOptions) (RetrosheetResult, error) {
	dataDir := opts.DataDir
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return RetrosheetResult{}, fmt.Errorf("error: failed to create retrosheet data dir: %w", err)
	}

	years := opts.Years
	if len(years) == 0 {
		years = defaultRetrosheetYears()
	}

	result := RetrosheetResult{}
	gameLogsDir := filepath.Join(dataDir, "gamelogs")
	playsDir := filepath.Join(dataDir, "plays")

	if err := os.MkdirAll(gameLogsDir, 0755); err != nil {
		return result, fmt.Errorf("error: failed to create gamelogs dir: %w", err)
	}
	if err := os.MkdirAll(playsDir, 0755); err != nil {
		return result, fmt.Errorf("error: failed to create plays dir: %w", err)
	}

	refreshes, err := database.DatasetRefreshes(ctx)
	if err != nil {
		return result, fmt.Errorf("error: failed to read dataset refreshes: %w", err)
	}

	if opts.Force {
		echo.Info("Force mode enabled - clearing data and metadata for specified years...")
		for _, year := range years {
			gamesKey := fmt.Sprintf("retrosheet_games_%d", year)
			playsKey := fmt.Sprintf("retrosheet_plays_%d", year)

			yearStr := fmt.Sprintf("%d", year)
			_, err := database.ExecContext(ctx, `DELETE FROM games WHERE date LIKE $1 || '%'`, yearStr)
			if err != nil {
				return result, fmt.Errorf("error: failed to delete games for %d: %w", year, err)
			}

			if _, err = database.ExecContext(ctx, `DELETE FROM plays WHERE SUBSTRING(gid, 4, 4) = $1`, yearStr); err != nil {
				return result, fmt.Errorf("error: failed to delete plays for %d: %w", year, err)
			}

			if err := clearDatasetRefresh(ctx, database, gamesKey); err != nil {
				return result, fmt.Errorf("error: failed to clear %s: %w", gamesKey, err)
			}
			if err := clearDatasetRefresh(ctx, database, playsKey); err != nil {
				return result, fmt.Errorf("error: failed to clear %s: %w", playsKey, err)
			}
			delete(refreshes, gamesKey)
			delete(refreshes, playsKey)
		}
	}

	var yearsNeedingGameLogs []int
	for _, year := range years {
		gamesKey := fmt.Sprintf("retrosheet_games_%d", year)
		if _, ok := refreshes[gamesKey]; !ok {
			yearsNeedingGameLogs = append(yearsNeedingGameLogs, year)
		}
	}

	batchFile := filepath.Join(gameLogsDir, "gl1871_2025.zip")
	if len(yearsNeedingGameLogs) > 0 {
		if _, err := os.Stat(batchFile); errors.Is(err, os.ErrNotExist) {
			echo.Info("Downloading game logs batch file (gl1871_2025.zip - 33MB)...")
			if err := downloadGameLogBatch(batchFile); err != nil {
				echo.Infof("  ⚠ Failed to download batch, falling back to individual downloads: %v", err)
				batchFile = ""
			} else {
				echo.Success("  ✓ Game logs batch downloaded")
			}
		} else if err == nil {
			echo.Info("  ✓ Using cached game logs batch")
		}
	}

	echo.Info("Loading game logs...")
	totalYears := len(years)
	loadedCount := 0
	skippedCount := 0

	for i, year := range years {
		zipFile := filepath.Join(gameLogsDir, fmt.Sprintf("GL%d.zip", year))
		gamesKey := fmt.Sprintf("retrosheet_games_%d", year)

		if _, ok := refreshes[gamesKey]; ok {
			skippedCount++
			continue
		}

		if _, err := os.Stat(zipFile); errors.Is(err, os.ErrNotExist) {
			if batchFile != "" {
				if err := ExtractGameLogFromBatch(batchFile, year, zipFile); err != nil {
					if err := downloadRetrosheetGameLog(year, zipFile); err != nil {
						return result, fmt.Errorf("error: failed to download %d game logs: %w", year, err)
					}
				}
			} else {
				if err := downloadRetrosheetGameLog(year, zipFile); err != nil {
					return result, fmt.Errorf("error: failed to download %d game logs: %w", year, err)
				}
			}
		} else if err != nil {
			return result, fmt.Errorf("error: unable to stat %s: %w", zipFile, err)
		}

		rows, err := database.LoadRetrosheetGameLog(ctx, zipFile, "regular")
		if err != nil {
			return result, fmt.Errorf("error: failed to load %d: %w", year, err)
		}

		result.GameRows += rows
		loadedCount++
		echo.Infof("  [%d/%d] %d: %s rows", i+1, totalYears, year, formatNumber(rows))

		if err := database.RecordDatasetRefresh(ctx, gamesKey, rows); err != nil {
			return result, fmt.Errorf("error: failed to record %s refresh: %w", gamesKey, err)
		}
		refreshes[gamesKey] = db.DatasetRefresh{}
	}

	if skippedCount > 0 {
		echo.Infof("  Skipped %d already-loaded years", skippedCount)
	}

	echo.Info("")
	echo.Info("Loading play-by-play data...")

	var yearsNeedingPlays []int
	for _, year := range years {
		playsKey := fmt.Sprintf("retrosheet_plays_%d", year)
		if _, ok := refreshes[playsKey]; !ok {
			zipFile := filepath.Join(playsDir, fmt.Sprintf("%dplays.zip", year))
			if _, err := os.Stat(zipFile); errors.Is(err, os.ErrNotExist) {
				yearsNeedingPlays = append(yearsNeedingPlays, year)
			}
		}
	}

	if len(yearsNeedingPlays) > 0 {
		echo.Infof("  Downloading %d years (parallel)...", len(yearsNeedingPlays))
		if err := downloadPlaysParallel(playsDir, yearsNeedingPlays); err != nil {
			echo.Infof("  ⚠ Some downloads failed, retrying individually")
		}
	}

	playsLoadedCount := 0
	playsSkippedCount := 0
	var emptyPlayYears []int

	for i, year := range years {
		zipFile := filepath.Join(playsDir, fmt.Sprintf("%dplays.zip", year))
		playsKey := fmt.Sprintf("retrosheet_plays_%d", year)

		if _, ok := refreshes[playsKey]; ok {
			playsSkippedCount++
			continue
		}

		if _, err := os.Stat(zipFile); errors.Is(err, os.ErrNotExist) {
			if err := downloadRetrosheetPlays(year, zipFile); err != nil {
				return result, fmt.Errorf("error: failed to download %d plays: %w", year, err)
			}
		}

		rows, err := database.LoadRetrosheetPlays(ctx, zipFile)
		if err != nil {
			return result, fmt.Errorf("error: failed to load %d plays: %w", year, err)
		}

		result.PlayRows += rows
		playsLoadedCount++
		if rows == 0 {
			emptyPlayYears = append(emptyPlayYears, year)
			echo.Infof("  [%d/%d] %d: no plays found (file empty)", i+1, totalYears, year)
		} else {
			echo.Infof("  [%d/%d] %d: %s rows", i+1, totalYears, year, formatNumber(rows))
		}

		if err := database.RecordDatasetRefresh(ctx, playsKey, rows); err != nil {
			return result, fmt.Errorf("error: failed to record %s refresh: %w", playsKey, err)
		}
		refreshes[playsKey] = db.DatasetRefresh{}
	}

	if playsSkippedCount > 0 {
		echo.Infof("  Skipped %d already-loaded years", playsSkippedCount)
	}
	if len(emptyPlayYears) > 0 {
		echo.Infof("  No play-by-play rows found for: %v", emptyPlayYears)
	}

	specialGameTypes := []struct {
		name         string
		file         string
		gameType     string
		downloadFunc func(string) error
		refreshKey   string
	}{
		{"All-Star games", filepath.Join(gameLogsDir, "glas.zip"), "allstar", downloadRetrosheetAllStarGames, "retrosheet_allstar_games"},
		{"World Series games", filepath.Join(gameLogsDir, "glws.zip"), "worldseries", downloadRetrosheetWorldSeriesGames, "retrosheet_worldseries_games"},
		{"Division Series games", filepath.Join(gameLogsDir, "gldv.zip"), "divisionseries", downloadRetrosheetDivisionSeriesGames, "retrosheet_divisional_games"},
		{"Championship Series games", filepath.Join(gameLogsDir, "gllc.zip"), "lcs", downloadRetrosheetChampionshipSeriesGames, "retrosheet_championship_games"},
		{"Wild Card games", filepath.Join(gameLogsDir, "glwc.zip"), "wildcard", downloadRetrosheetWildCardGames, "retrosheet_wildcard_games"},
	}

	echo.Info("")
	echo.Info("Loading special game types...")
	specialLoadedCount := 0
	specialSkippedCount := 0

	for _, gameType := range specialGameTypes {
		if _, ok := refreshes[gameType.refreshKey]; ok {
			specialSkippedCount++
			continue
		}

		if _, err := os.Stat(gameType.file); errors.Is(err, os.ErrNotExist) {
			if err := gameType.downloadFunc(gameType.file); err != nil {
				return result, fmt.Errorf("error: failed to download %s: %w", gameType.name, err)
			}
		} else if err != nil {
			return result, fmt.Errorf("error: unable to stat %s: %w", gameType.file, err)
		}

		rows, err := database.LoadRetrosheetGameLog(ctx, gameType.file, gameType.gameType)
		if err != nil {
			return result, fmt.Errorf("error: failed to load %s: %w", gameType.name, err)
		}

		result.GameRows += rows
		specialLoadedCount++
		echo.Infof("  %s: %s rows", gameType.name, formatNumber(rows))

		if err := database.RecordDatasetRefresh(ctx, gameType.refreshKey, rows); err != nil {
			return result, fmt.Errorf("error: failed to record %s refresh: %w", gameType.name, err)
		}
		refreshes[gameType.refreshKey] = db.DatasetRefresh{}
	}

	if specialSkippedCount > 0 {
		echo.Infof("  Skipped %d already-loaded types", specialSkippedCount)
	}

	echo.Info("")
	echo.Info("Loading Negro Leagues data (if available)...")
	negroDir := filepath.Join(dataDir, "negroleagues")
	negroGameRows, negroPlayRows, err := database.LoadNegroLeaguesData(ctx, negroDir)
	if err != nil {
		return result, fmt.Errorf("error: failed to load Negro Leagues data: %w", err)
	}

	if negroGameRows == 0 && negroPlayRows == 0 {
		echo.Info("  Negro Leagues files not found (expected gameinfo.csv and plays.csv)")
	} else {
		result.GameRows += negroGameRows
		result.PlayRows += negroPlayRows

		if negroGameRows > 0 {
			echo.Successf("  ✓ Loaded Negro Leagues games (%d rows)", negroGameRows)
			if err := database.RecordDatasetRefresh(ctx, "negroleagues_games", negroGameRows); err != nil {
				return result, fmt.Errorf("error: failed to record Negro Leagues games refresh: %w", err)
			}
		}
		if negroPlayRows > 0 {
			echo.Successf("  ✓ Loaded Negro Leagues plays (%d rows)", negroPlayRows)
			if err := database.RecordDatasetRefresh(ctx, "negroleagues_plays", negroPlayRows); err != nil {
				return result, fmt.Errorf("error: failed to record Negro Leagues plays refresh: %w", err)
			}
		}
	}

	totalRows := result.GameRows + result.PlayRows

	echo.Info("")
	echo.Success("✓ Retrosheet data loaded successfully")
	echo.Infof("  %s total rows (%s games, %s plays)",
		formatNumber(totalRows),
		formatNumber(result.GameRows),
		formatNumber(result.PlayRows))

	if result.GameRows > 0 {
		if err := database.RecordDatasetRefresh(ctx, "retrosheet_games", result.GameRows); err != nil {
			return result, fmt.Errorf("error: failed to record Retrosheet games refresh: %w", err)
		}
	}
	if result.PlayRows > 0 {
		if err := database.RecordDatasetRefresh(ctx, "retrosheet_plays", result.PlayRows); err != nil {
			return result, fmt.Errorf("error: failed to record Retrosheet plays refresh: %w", err)
		}
	}

	return result, nil
}

func truncateTable(ctx context.Context, database *db.DB, table string) error {
	query := fmt.Sprintf(`TRUNCATE TABLE "%s"`, table)
	if _, err := database.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("error: failed to truncate %s: %w", table, err)
	}
	return nil
}

func defaultLahmanTables() []string {
	return []string{
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
}

func defaultRetrosheetYears() []int {
	return []int{2023, 2024, 2025}
}

// ResetLahman truncates Lahman tables and clears refresh metadata.
func ResetLahman(ctx context.Context, database *db.DB, tables []string) error {
	if len(tables) == 0 {
		tables = defaultLahmanTables()
	}

	for _, table := range tables {
		if err := truncateTable(ctx, database, table); err != nil {
			return err
		}
	}

	if err := clearDatasetRefresh(ctx, database, "lahman"); err != nil {
		return err
	}
	return nil
}

// ResetRetrosheet truncates Retrosheet tables and clears refresh metadata for the requested years.
func ResetRetrosheet(ctx context.Context, database *db.DB, years []int) error {
	if len(years) == 0 {
		years = defaultRetrosheetYears()
	}

	if err := truncateTable(ctx, database, "games"); err != nil {
		return err
	}
	if err := truncateTable(ctx, database, "plays"); err != nil {
		return err
	}

	if err := clearDatasetRefresh(ctx, database, "retrosheet_games"); err != nil {
		return err
	}
	if err := clearDatasetRefresh(ctx, database, "retrosheet_plays"); err != nil {
		return err
	}

	for _, year := range years {
		if err := clearDatasetRefresh(ctx, database, fmt.Sprintf("retrosheet_games_%d", year)); err != nil {
			return err
		}
		if err := clearDatasetRefresh(ctx, database, fmt.Sprintf("retrosheet_plays_%d", year)); err != nil {
			return err
		}
	}

	return nil
}

func clearDatasetRefresh(ctx context.Context, database *db.DB, dataset string) error {
	if dataset == "" {
		return nil
	}

	if _, err := database.ExecContext(ctx, `DELETE FROM dataset_refreshes WHERE dataset = $1`, dataset); err != nil {
		return fmt.Errorf("error: failed to clear dataset refresh for %s: %w", dataset, err)
	}

	return nil
}

func downloadRetrosheetGameLog(year int, dest string) error {
	url := fmt.Sprintf("https://www.retrosheet.org/gamelogs/gl%d.zip", year)
	return downloadFile(url, dest)
}

func downloadRetrosheetPlays(year int, dest string) error {
	url := fmt.Sprintf("https://www.retrosheet.org/downloads/plays/%dplays.zip", year)
	return downloadFile(url, dest)
}

func downloadRetrosheetAllStarGames(dest string) error {
	url := "https://www.retrosheet.org/gamelogs/glas.zip"
	return downloadFile(url, dest)
}

func downloadRetrosheetWorldSeriesGames(dest string) error {
	url := "https://www.retrosheet.org/gamelogs/glws.zip"
	return downloadFile(url, dest)
}

func downloadRetrosheetDivisionSeriesGames(dest string) error {
	url := "https://www.retrosheet.org/gamelogs/gldv.zip"
	return downloadFile(url, dest)
}

func downloadRetrosheetChampionshipSeriesGames(dest string) error {
	url := "https://www.retrosheet.org/gamelogs/gllc.zip"
	return downloadFile(url, dest)
}

func downloadRetrosheetWildCardGames(dest string) error {
	url := "https://www.retrosheet.org/gamelogs/glwc.zip"
	return downloadFile(url, dest)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

func downloadGameLogBatch(dest string) error {
	url := "https://www.retrosheet.org/gamelogs/gl1871_2025.zip"
	return downloadFile(url, dest)
}

// downloadPlaysParallel downloads multiple years of plays in parallel with rate limiting.
func downloadPlaysParallel(playsDir string, years []int) error {
	const maxConcurrent = 3
	semaphore := make(chan struct{}, maxConcurrent)
	errChan := make(chan error, len(years))

	for _, year := range years {
		semaphore <- struct{}{}
		go func(y int) {
			defer func() { <-semaphore }()

			zipFile := filepath.Join(playsDir, fmt.Sprintf("%dplays.zip", y))
			if err := downloadRetrosheetPlays(y, zipFile); err != nil {
				errChan <- fmt.Errorf("year %d: %w", y, err)
			} else {
				errChan <- nil
			}
		}(year)
	}

	var errors []error
	for range len(years) {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to download %d years", len(errors))
	}
	return nil
}

// formatNumber adds thousand separators to numbers for better readability.
func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// ExtractGameLogFromBatch extracts a specific year's game log from the batch zip.
// batchZip is the path to gl1871_2025.zip, year is the year to extract, destZip is where to save the extracted file.
func ExtractGameLogFromBatch(batchZip string, year int, destZip string) error {
	r, err := zip.OpenReader(batchZip)
	if err != nil {
		return fmt.Errorf("failed to open batch zip: %w", err)
	}
	defer r.Close()

	targetFile := fmt.Sprintf("gl%d.txt", year)

	for _, f := range r.File {
		if strings.EqualFold(f.Name, targetFile) {
			return extractAndZipFile(f, destZip)
		}
	}

	return fmt.Errorf("year %d not found in batch", year)
}

// extractAndZipFile extracts a file from a zip and creates a new zip with just that file.
func extractAndZipFile(zipFile *zip.File, destZip string) error {
	rc, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer rc.Close()

	outFile, err := os.Create(destZip)
	if err != nil {
		return fmt.Errorf("failed to create output zip: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	writer, err := zipWriter.Create(zipFile.Name)
	if err != nil {
		return fmt.Errorf("failed to create file in output zip: %w", err)
	}

	if _, err := io.Copy(writer, rc); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	return nil
}
