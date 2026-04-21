package seed

import (
	"archive/zip"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/echo"
)

// FetchRetrosheetData downloads Retrosheet archives needed by the ETL pipeline.
func FetchRetrosheetData(ctx context.Context, dataDir string, years []int, force bool) error {
	if dataDir == "" {
		dataDir = RetrosheetDir("")
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

	totalYears := len(years)
	echo.Infof("Preparing Retrosheet archives for %d year(s)...", totalYears)
	for i, year := range years {
		gameStatus := "cached"
		playsStatus := "cached"

		gamelogFile := filepath.Join(gameLogsDir, fmt.Sprintf("GL%d.zip", year))
		gamelogURL := fmt.Sprintf("https://www.retrosheet.org/gamelogs/gl%d.zip", year)
		downloaded, err := downloadIfNeeded(ctx, gamelogURL, gamelogFile, force)
		if err != nil {
			gameStatus = "failed"
			echo.Infof("  ⚠ Retrosheet game log download failed for %d: %v", year, err)
		} else if downloaded {
			gameStatus = "downloaded"
		}

		playsFile := filepath.Join(playsDir, fmt.Sprintf("%dplays.zip", year))
		playsURL := fmt.Sprintf("https://www.retrosheet.org/downloads/plays/%dplays.zip", year)
		downloaded, err = downloadIfNeeded(ctx, playsURL, playsFile, force)
		if err != nil {
			playsStatus = "failed"
			echo.Infof("  ⚠ Retrosheet plays download failed for %d: %v", year, err)
		} else if downloaded {
			playsStatus = "downloaded"
		}

		echo.Infof("  [%d/%d] %d: gamelog=%s plays=%s", i+1, totalYears, year, gameStatus, playsStatus)
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
		downloaded, err := downloadIfNeeded(ctx, file.url, file.destPath, force)
		if err != nil {
			echo.Infof("  ⚠ Download failed for %s: %v", filepath.Base(file.destPath), err)
			continue
		}
		if downloaded {
			echo.Infof("  ✓ Downloaded %s", filepath.Base(file.destPath))
		} else {
			echo.Infof("  ✓ Using cached %s", filepath.Base(file.destPath))
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
func FetchNegroLeaguesData(ctx context.Context, dataDir string, force bool) error {
	if dataDir == "" {
		dataDir = RetrosheetNegroLeaguesDir("")
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", dataDir, err)
	}

	zipPath := filepath.Join(dataDir, "negroleagues.zip")
	if _, err := downloadIfNeeded(ctx, "https://www.retrosheet.org/downloads/negroleagues.zip", zipPath, force); err != nil {
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
		dataDir = RetrosheetDir("")
	}

	csvPath := filepath.Join(dataDir, "allplayers.csv")
	if hasRows, err := csvHasDataRows(csvPath); err == nil && hasRows {
		return csvPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to inspect allplayers.csv at %s: %w", csvPath, err)
	}

	zipPath := filepath.Join(dataDir, "allplayers.zip")
	if _, err := os.Stat(zipPath); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("allplayers.zip not found at %s", zipPath)
	}

	if err := extractZipEntry(zipPath, "allplayers.csv", csvPath); err != nil {
		return "", err
	}
	hasRows, err := csvHasDataRows(csvPath)
	if err != nil {
		return "", fmt.Errorf("failed to inspect extracted allplayers.csv at %s: %w", csvPath, err)
	}
	if !hasRows {
		return "", fmt.Errorf("allplayers.csv at %s has no data rows", csvPath)
	}

	return csvPath, nil
}

func csvHasDataRows(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}

	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return false, nil
}

// ExtractBiodataArchive extracts biodata CSV files into a temporary directory.
func ExtractBiodataArchive(dataDir string) (string, func(), error) {
	if dataDir == "" {
		dataDir = RetrosheetDir("")
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

func downloadIfNeeded(ctx context.Context, url, destPath string, force bool) (bool, error) {
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			return false, nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return false, err
	}

	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := downloadToFile(ctx, url, destPath); err == nil {
			return true, nil
		} else {
			lastErr = err
		}

		if attempt == maxAttempts {
			break
		}

		backoff := time.Duration(attempt*attempt) * time.Second
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(backoff):
		}
	}

	return false, fmt.Errorf("download failed after %d attempts for %s: %w", maxAttempts, url, lastErr)
}

func downloadToFile(ctx context.Context, url, destPath string) error {
	if err := waitForRetrosheetDownloadSlot(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
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

	if _, err = io.Copy(out, resp.Body); err != nil {
		_ = os.Remove(destPath)
		return err
	}
	return nil
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
