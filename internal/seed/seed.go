package seed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

	echo.Info("Loading game logs...")
	for _, year := range years {
		zipFile := filepath.Join(gameLogsDir, fmt.Sprintf("GL%d.zip", year))
		gamesKey := fmt.Sprintf("retrosheet_games_%d", year)

		if _, ok := refreshes[gamesKey]; ok {
			echo.Infof("  Skipping %d game logs (already loaded)", year)
			continue
		}

		if _, err := os.Stat(zipFile); errors.Is(err, os.ErrNotExist) {
			echo.Infof("  Downloading %d game logs...", year)
			if err := downloadRetrosheetGameLog(year, zipFile); err != nil {
				return result, fmt.Errorf("error: failed to download %d game logs: %w", year, err)
			}
		} else if err != nil {
			return result, fmt.Errorf("error: unable to stat %s: %w", zipFile, err)
		} else {
			echo.Infof("  ✓ Using cached %d game logs", year)
		}

		echo.Infof("  Loading %d game logs...", year)

		rows, err := database.LoadRetrosheetGameLog(ctx, zipFile)
		if err != nil {
			return result, fmt.Errorf("error: failed to load %d: %w", year, err)
		}

		result.GameRows += rows
		echo.Successf("  ✓ Loaded %d (%d rows)", year, rows)

		if err := database.RecordDatasetRefresh(ctx, gamesKey, rows); err != nil {
			return result, fmt.Errorf("error: failed to record %s refresh: %w", gamesKey, err)
		}
		refreshes[gamesKey] = db.DatasetRefresh{}
	}

	echo.Info("")
	echo.Info("Loading play-by-play data...")
	for _, year := range years {
		zipFile := filepath.Join(playsDir, fmt.Sprintf("%dplays.zip", year))
		playsKey := fmt.Sprintf("retrosheet_plays_%d", year)

		if _, ok := refreshes[playsKey]; ok {
			echo.Infof("  Skipping %d plays (already loaded)", year)
			continue
		}

		if _, err := os.Stat(zipFile); errors.Is(err, os.ErrNotExist) {
			echo.Infof("  Downloading %d plays...", year)
			if err := downloadRetrosheetPlays(year, zipFile); err != nil {
				return result, fmt.Errorf("error: failed to download %d plays: %w", year, err)
			}
		} else if err != nil {
			return result, fmt.Errorf("error: unable to stat %s: %w", zipFile, err)
		} else {
			echo.Infof("  ✓ Using cached %d plays", year)
		}

		echo.Infof("  Loading %d plays...", year)

		rows, err := database.LoadRetrosheetPlays(ctx, zipFile)
		if err != nil {
			return result, fmt.Errorf("error: failed to load %d plays: %w", year, err)
		}

		result.PlayRows += rows
		echo.Successf("  ✓ Loaded %d (%d rows)", year, rows)

		if err := database.RecordDatasetRefresh(ctx, playsKey, rows); err != nil {
			return result, fmt.Errorf("error: failed to record %s refresh: %w", playsKey, err)
		}
		refreshes[playsKey] = db.DatasetRefresh{}
	}

	totalRows := result.GameRows + result.PlayRows

	echo.Info("")
	echo.Success("✓ All Retrosheet data loaded successfully")
	echo.Infof("  Total rows: %d", totalRows)
	echo.Infof("  Play-by-play rows: %d", result.PlayRows)

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
