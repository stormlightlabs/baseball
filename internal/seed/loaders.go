package seed

import (
	"archive/zip"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"stormlightlabs.org/baseball/internal/db"
)

//go:embed sql/build_lg_const.sql
var buildLeagueConstantsSQL string

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

	leagueRows, err := rebuildLeagueConstants(ctx, database)
	if err != nil {
		return 0, fmt.Errorf("failed to rebuild league constants: %w", err)
	}

	totalRows := wobaRows + parkRows + leagueRows
	if err := database.RecordDatasetRefresh(ctx, "fangraphs_constants", totalRows); err != nil {
		return totalRows, fmt.Errorf("failed to record FanGraphs refresh: %w", err)
	}

	return totalRows, nil
}

func rebuildLeagueConstants(ctx context.Context, database *db.DB) (int64, error) {
	result, err := database.ExecContext(ctx, buildLeagueConstantsSQL)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
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
