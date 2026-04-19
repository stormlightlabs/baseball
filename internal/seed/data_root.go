package seed

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// DataRootEnvVar allows overriding where ETL source datasets are stored.
	DataRootEnvVar = "BASEBALL_DATA_ROOT"

	// DefaultDataRoot is the preferred repository-local location for dataset snapshots.
	DefaultDataRoot = "tools/data"

	// LegacyDataRoot is the historical in-repo dataset location.
	LegacyDataRoot = "data"
)

// ResolveDataRoot applies precedence for ETL dataset roots:
// 1) explicit flag/option
// 2) BASEBALL_DATA_ROOT environment variable
// 3) tools/data (if it exists)
// 4) legacy data (if it exists)
// If neither location exists, tools/data is returned as the default target.
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
	if isDir(LegacyDataRoot) {
		return LegacyDataRoot
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
