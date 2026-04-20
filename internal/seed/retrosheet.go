package seed

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// RetrosheetCleanupResult captures which files were considered transient and which were removed.
type RetrosheetCleanupResult struct {
	Candidates []string
	Removed    []string
}

// CleanupRetrosheetArtifacts removes transient Retrosheet artifacts while preserving canonical source files.
func CleanupRetrosheetArtifacts(dataDir string, dryRun bool) (RetrosheetCleanupResult, error) {
	if dataDir == "" {
		dataDir = RetrosheetDir("")
	}

	result := RetrosheetCleanupResult{
		Candidates: make([]string, 0),
		Removed:    make([]string, 0),
	}

	info, err := os.Stat(dataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		return result, fmt.Errorf("failed to inspect %s: %w", dataDir, err)
	}
	if !info.IsDir() {
		return result, fmt.Errorf("retrosheet path is not a directory: %s", dataDir)
	}

	walkErr := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if d.IsDir() {
			return nil
		}

		transient, err := isRetrosheetTransientArtifact(dataDir, path)
		if err != nil {
			return err
		}
		if !transient {
			return nil
		}
		result.Candidates = append(result.Candidates, path)
		return nil
	})
	if walkErr != nil {
		return result, fmt.Errorf("failed to scan Retrosheet artifacts: %w", walkErr)
	}

	slices.Sort(result.Candidates)
	if dryRun {
		return result, nil
	}

	for _, path := range result.Candidates {
		if err := os.Remove(path); err != nil {
			return result, fmt.Errorf("failed to remove %s: %w", path, err)
		}
		result.Removed = append(result.Removed, path)
	}

	return result, nil
}

func isRetrosheetTransientArtifact(dataDir, path string) (bool, error) {
	rel, err := filepath.Rel(dataDir, path)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return false, nil
	}

	lowerRel := filepath.ToSlash(strings.ToLower(rel))
	if shouldKeepRetrosheetArtifact(lowerRel) {
		return false, nil
	}

	if strings.HasSuffix(lowerRel, ".tmp") ||
		strings.HasSuffix(lowerRel, ".partial") ||
		strings.HasSuffix(lowerRel, ".download") ||
		strings.HasSuffix(lowerRel, "~") {
		return true, nil
	}

	if strings.HasPrefix(lowerRel, "gamelogs/") ||
		strings.HasPrefix(lowerRel, "plays/") ||
		strings.HasPrefix(lowerRel, "ejections/") ||
		strings.HasPrefix(lowerRel, "allstar/") ||
		strings.HasPrefix(lowerRel, "negroleagues/") {
		return true, nil
	}

	return strings.HasSuffix(lowerRel, ".csv"), nil
}

func shouldKeepRetrosheetArtifact(lowerRel string) bool {
	if strings.HasSuffix(lowerRel, ".zip") {
		return true
	}

	switch lowerRel {
	case "allplayers.csv", "gameinfo.csv", "negroleagues/gameinfo.csv", "negroleagues/plays.csv", "manifest.json":
		return true
	default:
		return false
	}
}
