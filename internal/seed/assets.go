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

// FetchRetrosheetData downloads Retrosheet archives needed by the ETL pipeline.
func FetchRetrosheetData(_ context.Context, dataDir string, years []int, force bool) error {
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet")
	}
	if len(years) == 0 {
		years = defaultRetrosheetYears()
	}

	gameLogsDir := filepath.Join(dataDir, "gamelogs")
	playsDir := filepath.Join(dataDir, "plays")
	ejectionsDir := filepath.Join(dataDir, "ejections")
	allStarDir := filepath.Join(dataDir, "allstar")

	for _, dir := range []string{dataDir, gameLogsDir, playsDir, ejectionsDir, allStarDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	for _, year := range years {
		gamelogFile := filepath.Join(gameLogsDir, fmt.Sprintf("GL%d.zip", year))
		gamelogURL := fmt.Sprintf("https://www.retrosheet.org/gamelogs/gl%d.zip", year)
		if err := downloadIfNeeded(gamelogURL, gamelogFile, force); err != nil {
			echo.Infof("  ⚠ Retrosheet game log download failed for %d: %v", year, err)
		}

		playsFile := filepath.Join(playsDir, fmt.Sprintf("%dplays.zip", year))
		playsURL := fmt.Sprintf("https://www.retrosheet.org/downloads/plays/%dplays.zip", year)
		if err := downloadIfNeeded(playsURL, playsFile, force); err != nil {
			echo.Infof("  ⚠ Retrosheet plays download failed for %d: %v", year, err)
		}
	}

	downloads := []struct {
		url      string
		destPath string
	}{
		{"https://www.retrosheet.org/ejections.zip", filepath.Join(ejectionsDir, "ejections.zip")},
		{"https://www.retrosheet.org/downloads/allplayers.zip", filepath.Join(dataDir, "allplayers.zip")},
		{"https://www.retrosheet.org/downloads/biodata.zip", filepath.Join(dataDir, "biodata.zip")},
		{"https://www.retrosheet.org/downloads/allstar.zip", filepath.Join(allStarDir, "allstar.zip")},
		{"https://www.retrosheet.org/gameinfo.zip", filepath.Join(dataDir, "gameinfo.zip")},
	}

	for _, file := range downloads {
		if err := downloadIfNeeded(file.url, file.destPath, force); err != nil {
			echo.Infof("  ⚠ Download failed for %s: %v", filepath.Base(file.destPath), err)
		}
	}

	gameInfoCSV := filepath.Join(dataDir, "gameinfo.csv")
	if force {
		_ = os.Remove(gameInfoCSV)
	}
	if _, err := os.Stat(gameInfoCSV); errors.Is(err, os.ErrNotExist) {
		if err := extractZipEntry(filepath.Join(dataDir, "gameinfo.zip"), "gameinfo.csv", gameInfoCSV); err != nil {
			echo.Infof("  ⚠ Failed to extract gameinfo.csv: %v", err)
		}
	}

	return nil
}

// FetchNegroLeaguesData downloads and extracts the Retrosheet Negro Leagues archive.
func FetchNegroLeaguesData(_ context.Context, dataDir string, force bool) error {
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet", "negroleagues")
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", dataDir, err)
	}

	zipPath := filepath.Join(dataDir, "negroleagues.zip")
	if err := downloadIfNeeded("https://www.retrosheet.org/downloads/negroleagues.zip", zipPath, force); err != nil {
		return err
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", zipPath, err)
	}
	defer reader.Close()

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}

		dest := filepath.Join(dataDir, entry.Name)
		if !force {
			if _, err := os.Stat(dest); err == nil {
				continue
			}
		}

		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed opening %s in archive: %w", entry.Name, err)
		}

		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed creating %s: %w", dest, err)
		}

		if _, err = io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("failed extracting %s: %w", entry.Name, err)
		}

		out.Close()
		rc.Close()
	}

	return nil
}

// EnsureRetrosheetPlayersCSV returns an extracted allplayers CSV path.
func EnsureRetrosheetPlayersCSV(dataDir string) (string, error) {
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet")
	}

	csvPath := filepath.Join(dataDir, "allplayers.csv")
	if _, err := os.Stat(csvPath); err == nil {
		return csvPath, nil
	}

	zipPath := filepath.Join(dataDir, "allplayers.zip")
	if _, err := os.Stat(zipPath); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("allplayers.zip not found at %s", zipPath)
	}

	if err := extractZipEntry(zipPath, "allplayers.csv", csvPath); err != nil {
		return "", err
	}

	return csvPath, nil
}

// ExtractBiodataArchive extracts biodata CSV files into a temporary directory.
func ExtractBiodataArchive(dataDir string) (string, func(), error) {
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet")
	}

	zipPath := filepath.Join(dataDir, "biodata.zip")
	if _, err := os.Stat(zipPath); errors.Is(err, os.ErrNotExist) {
		return "", nil, fmt.Errorf("biodata.zip not found at %s", zipPath)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed opening biodata.zip: %w", err)
	}
	defer reader.Close()

	tmpDir, err := os.MkdirTemp("", "biodata-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed creating temp dir: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name), ".csv") {
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			cleanup()
			return "", nil, fmt.Errorf("failed opening %s: %w", entry.Name, err)
		}

		dest := filepath.Join(tmpDir, filepath.Base(entry.Name))
		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			cleanup()
			return "", nil, fmt.Errorf("failed creating %s: %w", dest, err)
		}

		if _, err = io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			cleanup()
			return "", nil, fmt.Errorf("failed extracting %s: %w", entry.Name, err)
		}

		out.Close()
		rc.Close()
	}

	return tmpDir, cleanup, nil
}

// LoadFanGraphsData loads FanGraphs constants and park factors.
func LoadFanGraphsData(ctx context.Context, database *db.DB, dataDir string) (int64, error) {
	if dataDir == "" {
		dataDir = filepath.Join("data", "fangraphs")
	}

	wobaFile := filepath.Join(dataDir, "woba.csv")
	if _, err := os.Stat(wobaFile); errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("wOBA constants file not found at %s", wobaFile)
	}

	wobaRows, err := database.LoadFanGraphsWOBA(ctx, wobaFile)
	if err != nil {
		return 0, fmt.Errorf("failed to load wOBA constants: %w", err)
	}

	parkDir := filepath.Join(dataDir, "pf")
	parkFiles, err := filepath.Glob(filepath.Join(parkDir, "*.csv"))
	if err != nil {
		return 0, fmt.Errorf("failed listing park factor files: %w", err)
	}

	var parkRows int64
	for _, file := range parkFiles {
		rows, err := database.LoadFanGraphsParks(ctx, file)
		if err != nil {
			return 0, fmt.Errorf("failed to load %s: %w", filepath.Base(file), err)
		}
		parkRows += rows
	}

	totalRows := wobaRows + parkRows
	if err := database.RecordDatasetRefresh(ctx, "fangraphs_constants", totalRows); err != nil {
		return totalRows, fmt.Errorf("failed to record FanGraphs refresh: %w", err)
	}

	return totalRows, nil
}

// LoadNegroLeagues loads Negro Leagues gameinfo and plays files.
func LoadNegroLeagues(ctx context.Context, database *db.DB, dataDir string) (int64, error) {
	if dataDir == "" {
		dataDir = filepath.Join("data", "retrosheet", "negroleagues")
	}

	gameinfoPath := filepath.Join(dataDir, "gameinfo.csv")
	if _, err := os.Stat(gameinfoPath); errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("gameinfo.csv not found at %s", gameinfoPath)
	}

	gameRows, err := database.LoadNegroLeaguesGameInfo(ctx, gameinfoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load Negro Leagues gameinfo: %w", err)
	}
	if err := database.RecordDatasetRefresh(ctx, "negroleagues_games", gameRows); err != nil {
		return gameRows, fmt.Errorf("failed to record Negro Leagues game refresh: %w", err)
	}

	total := gameRows
	playsPath := filepath.Join(dataDir, "plays.csv")
	if _, err := os.Stat(playsPath); err == nil {
		playRows, err := database.LoadNegroLeaguesPlays(ctx, playsPath)
		if err != nil {
			return total, fmt.Errorf("failed to load Negro Leagues plays: %w", err)
		}
		total += playRows

		if err := database.RecordDatasetRefresh(ctx, "negroleagues_plays", playRows); err != nil {
			return total, fmt.Errorf("failed to record Negro Leagues play refresh: %w", err)
		}
	}

	return total, nil
}

// LoadParksData fills missing park metadata and records refresh metadata.
func LoadParksData(ctx context.Context, database *db.DB) (int64, error) {
	rows, err := database.LoadMissingParks(ctx)
	if err != nil {
		return 0, err
	}
	if err := database.RecordDatasetRefresh(ctx, "parks_metadata", rows); err != nil {
		return rows, fmt.Errorf("failed to record parks metadata refresh: %w", err)
	}
	return rows, nil
}

// LoadAllStarData loads all-star gameinfo and plays from allstar.zip.
func LoadAllStarData(ctx context.Context, database *db.DB, zipPath string) (int64, error) {
	if zipPath == "" {
		zipPath = filepath.Join("data", "retrosheet", "allstar", "allstar.zip")
	}
	if _, err := os.Stat(zipPath); errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("allstar.zip not found at %s", zipPath)
	}

	gameInfoCSV, playsCSV, cleanup, err := extractAllStarArchive(zipPath)
	if err != nil {
		return 0, err
	}
	defer cleanup()

	gameRows, err := database.LoadAllStarGameInfo(ctx, gameInfoCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to load all-star games: %w", err)
	}
	if err := database.RecordDatasetRefresh(ctx, "allstar_games", gameRows); err != nil {
		return gameRows, fmt.Errorf("failed to record all-star games refresh: %w", err)
	}

	playRows, err := database.LoadAllStarPlays(ctx, playsCSV)
	if err != nil {
		return gameRows, fmt.Errorf("failed to load all-star plays: %w", err)
	}
	if err := database.RecordDatasetRefresh(ctx, "allstar_plays", playRows); err != nil {
		return gameRows + playRows, fmt.Errorf("failed to record all-star plays refresh: %w", err)
	}

	return gameRows + playRows, nil
}

func downloadIfNeeded(url, destPath string, force bool) error {
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractZipEntry(zipPath, entryName, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip %s: %w", zipPath, err)
	}
	defer reader.Close()

	for _, entry := range reader.File {
		if filepath.Base(entry.Name) != entryName {
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s in %s: %w", entryName, zipPath, err)
		}
		defer rc.Close()

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err = io.Copy(out, rc); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("%s not found in %s", entryName, zipPath)
}

func extractAllStarArchive(zipPath string) (string, string, func(), error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to open %s: %w", zipPath, err)
	}
	defer reader.Close()

	tmpDir, err := os.MkdirTemp("", "allstar-*")
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create all-star temp dir: %w", err)
	}

	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	var gameInfoCSV, playsCSV string

	for _, entry := range reader.File {
		base := strings.ToLower(filepath.Base(entry.Name))
		if base != "gameinfo.csv" && base != "plays.csv" {
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			cleanup()
			return "", "", nil, fmt.Errorf("failed to open %s in allstar archive: %w", entry.Name, err)
		}

		dest := filepath.Join(tmpDir, base)
		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			cleanup()
			return "", "", nil, fmt.Errorf("failed to create %s: %w", dest, err)
		}

		if _, err = io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			cleanup()
			return "", "", nil, fmt.Errorf("failed to extract %s: %w", entry.Name, err)
		}

		out.Close()
		rc.Close()

		if base == "gameinfo.csv" {
			gameInfoCSV = dest
		} else {
			playsCSV = dest
		}
	}

	if gameInfoCSV == "" || playsCSV == "" {
		cleanup()
		return "", "", nil, fmt.Errorf("allstar archive missing gameinfo.csv or plays.csv")
	}

	return gameInfoCSV, playsCSV, cleanup, nil
}
