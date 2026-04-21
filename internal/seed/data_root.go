package seed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	// DataRootEnvVar allows overriding where ETL source datasets are stored.
	DataRootEnvVar = "BASEBALL_DATA_ROOT"

	// DefaultDataRoot is the preferred repository-local location for dataset snapshots.
	DefaultDataRoot = "data"
)

// ResolveDataRoot applies precedence for ETL dataset roots:
// 1) explicit flag/option
// 2) BASEBALL_DATA_ROOT environment variable
// 3) data (if it exists)
// If data does not exist yet, data is returned as the default target.
func ResolveDataRoot(explicit string) string {
	if value := normalizeRoot(explicit); value != "" {
		return value
	}
	if value := normalizeRoot(os.Getenv(DataRootEnvVar)); value != "" {
		return value
	}
	if isDir(DefaultDataRoot) {
		return DefaultDataRoot
	}
	return DefaultDataRoot
}

func LahmanDir(dataRoot string) string {
	return filepath.Join(resolveRoot(dataRoot), "lahman")
}

func LahmanCSVDir(dataRoot string) string {
	return filepath.Join(LahmanDir(dataRoot), "csv")
}

func RetrosheetDir(dataRoot string) string {
	return filepath.Join(resolveRoot(dataRoot), "retrosheet")
}

func RetrosheetNegroLeaguesDir(dataRoot string) string {
	return filepath.Join(RetrosheetDir(dataRoot), "negroleagues")
}

func RetrosheetGameInfoCSV(dataRoot string) string {
	return filepath.Join(RetrosheetDir(dataRoot), "gameinfo.csv")
}

func RetrosheetAllStarZip(dataRoot string) string {
	return filepath.Join(RetrosheetDir(dataRoot), "allstar", "allstar.zip")
}

func RetrosheetAllPlayersCSV(dataRoot string) string {
	return filepath.Join(RetrosheetDir(dataRoot), "allplayers.csv")
}

func ChadwickDir(dataRoot string) string {
	return filepath.Join(resolveRoot(dataRoot), "chadwick")
}

func FanGraphsDir(dataRoot string) string {
	return filepath.Join(resolveRoot(dataRoot), "fangraphs")
}

func SalariesDir(dataRoot string) string {
	return filepath.Join(resolveRoot(dataRoot), "salaries")
}

func resolveRoot(dataRoot string) string {
	if value := normalizeRoot(dataRoot); value != "" {
		return value
	}
	return ResolveDataRoot("")
}

func normalizeRoot(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.Clean(value)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

type dataRootProvision struct {
	rootPath string
	cleanup  func()
}

func ensurePipelineDataRoot(ctx context.Context, opts PipelineOptions) (dataRootProvision, error) {
	_ = ctx

	for _, dir := range []string{opts.DataRoot, opts.RetrosheetDataDir, opts.ChadwickDataDir} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return dataRootProvision{}, fmt.Errorf("failed to prepare data directory %q: %w", dir, err)
		}
	}

	missing, err := missingPipelineDataArtifacts(opts)
	if err != nil {
		return dataRootProvision{}, err
	}
	if len(missing) > 0 {
		return dataRootProvision{}, fmt.Errorf(
			"data root %q is missing required local source files: %s\n\nRetrosheet archives are worker-fetched by ETL; Lahman/FanGraphs/salary sources are local inputs under the resolved data root. Chadwick `people.csv` is expected at %q and can be refreshed with `baseball-etl fetch chadwick`",
			opts.DataRoot,
			strings.Join(missing, ", "),
			filepath.Join(opts.ChadwickDataDir, "people.csv"),
		)
	}

	return dataRootProvision{rootPath: opts.DataRoot, cleanup: func() {}}, nil
}

func missingPipelineDataArtifacts(opts PipelineOptions) ([]string, error) {
	missing := make([]string, 0)

	requiredFiles := []string{
		filepath.Join(opts.LahmanCSVDir, "People.csv"),
		filepath.Join(opts.LahmanCSVDir, "Teams.csv"),
		filepath.Join(opts.FanGraphsDir, "woba.csv"),
		filepath.Join(opts.SalaryDataDir, "summary.csv"),
	}

	for _, path := range requiredFiles {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			continue
		}
		if errors.Is(err, os.ErrNotExist) || (err == nil && info.IsDir()) {
			missing = append(missing, path)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %w", path, err)
		}
	}

	parkFactorFiles, err := filepath.Glob(filepath.Join(opts.FanGraphsDir, "pf", "*.csv"))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect park factor files: %w", err)
	}
	if len(parkFactorFiles) == 0 {
		missing = append(missing, filepath.Join(opts.FanGraphsDir, "pf", "*.csv"))
	}

	slices.Sort(missing)
	return missing, nil
}
