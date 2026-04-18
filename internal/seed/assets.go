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
		{"https://www.retrosheet.org/downloads/gameinfo.zip", filepath.Join(dataDir, "gameinfo.zip")},
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
