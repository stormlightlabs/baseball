package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

// EtlStatusCmd creates the status command
func EtlStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check data freshness and completeness",
		Long:  "Display status of loaded data including freshness and completeness metrics.",
		RunE:  status,
	}
}

func status(cmd *cobra.Command, args []string) error {
	echo.Header("Data Status")
	ctx := cmd.Context()

	archiveChecks := []struct {
		label string
		path  string
		hint  string
	}{
		{
			label: "Lahman CSVs",
			path:  filepath.Join("data", "lahman", "csv"),
			hint:  "Use `baseball etl fetch lahman` to scaffold/download the dataset",
		},
		{
			label: "Retrosheet game logs",
			path:  filepath.Join("data", "retrosheet", "gamelogs"),
			hint:  "Use `baseball etl fetch retrosheet` to download seasonal game logs",
		},
		{
			label: "Retrosheet plays",
			path:  filepath.Join("data", "retrosheet", "plays"),
			hint:  "Use `baseball etl fetch retrosheet` to download parsed play-by-play archives",
		},
	}

	echo.Info("Local archives:")
	for _, check := range archiveChecks {
		exists, fileCount, latestChange, err := dirSnapshot(check.path)
		if err != nil {
			echo.Errorf("  %s: %v", check.label, err)
			continue
		}
		if !exists {
			echo.Infof("  • %s: %s", check.label, echo.ErrorStyle().Render("missing"))
			echo.Infof("    Path: %s", check.path)
			echo.Infof("    Hint: %s", check.hint)
			continue
		}
		if fileCount == 0 {
			echo.Infof("  • %s: directory exists but contains no files (%s)", check.label, check.path)
			continue
		}
		echo.Successf("  ✓ %s: %d files (last change %s)", check.label, fileCount, humanizeModTime(latestChange))
		echo.Infof("    Path: %s", check.path)
	}

	echo.Info("")
	echo.Info("Database:")

	database, err := db.Connect("")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	refreshes, err := database.DatasetRefreshes(ctx)
	if err != nil {
		echo.Infof("  ⚠ Unable to read ETL metadata: %v", err)
		refreshes = map[string]db.DatasetRefresh{}
	}

	lahmanPlayers, lahmanPlayersErr := safeCount(ctx, database, `SELECT COUNT(*) FROM "People"`)
	lahmanTeams, _ := safeCount(ctx, database, `SELECT COUNT(*) FROM "Teams"`)
	lahmanMin, lahmanMax, lahmanRangeErr := seasonRange(ctx, database)

	echo.Info("• Lahman Baseball Database")
	if lahmanPlayersErr != nil {
		echo.Infof("  ⚠ Unable to read player table: %v", lahmanPlayersErr)
	} else if lahmanPlayers == 0 {
		echo.Infof("  • People table is empty. Run `baseball etl load lahman` after downloading CSVs.")
	} else {
		echo.Successf("  ✓ %d players and %d team seasons available", lahmanPlayers, lahmanTeams)
	}

	if lahmanRangeErr == nil && lahmanMin != nil && lahmanMax != nil {
		echo.Infof("    Seasons covered: %d–%d", *lahmanMin, *lahmanMax)
	} else if lahmanRangeErr != nil {
		echo.Infof("    ⚠ Unable to derive season coverage: %v", lahmanRangeErr)
	}

	if entry, ok := refreshes["lahman"]; ok {
		entryCopy := entry
		echo.Infof("    Last ETL run: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Last ETL run: never recorded")
	}

	gamesCount, gamesErr := safeCount(ctx, database, `SELECT COUNT(*) FROM games`)
	playsCount, playsErr := safeCount(ctx, database, `SELECT COUNT(*) FROM plays`)
	gamesStart, gamesEnd, gamesRangeErr := retroDateRange(ctx, database, "games", "date")
	playsStart, playsEnd, playsRangeErr := retroDateRange(ctx, database, "plays", "date")

	echo.Info("")
	echo.Info("• Retrosheet Archives")
	if gamesErr != nil {
		echo.Infof("  ⚠ Unable to read game logs: %v", gamesErr)
	} else if gamesCount == 0 {
		echo.Infof("  • Games table is empty. Run `baseball etl load retrosheet` after downloading archives.")
	} else {
		echo.Successf("  ✓ %d game log rows loaded", gamesCount)
	}

	if gamesRangeErr == nil && gamesStart != nil && gamesEnd != nil {
		echo.Infof("    Game coverage: %s → %s", gamesStart.Format("2006-01-02"), gamesEnd.Format("2006-01-02"))
	} else if gamesRangeErr != nil {
		echo.Infof("    ⚠ Unable to derive game coverage: %v", gamesRangeErr)
	}

	if playsErr != nil {
		echo.Infof("  ⚠ Unable to read play-by-play table: %v", playsErr)
	} else if playsCount > 0 {
		echo.Infof("    Plays ingested: %d rows", playsCount)
	}

	if playsRangeErr == nil && playsStart != nil && playsEnd != nil {
		echo.Infof("    Play coverage: %s → %s", playsStart.Format("2006-01-02"), playsEnd.Format("2006-01-02"))
	}

	if entry, ok := refreshes["retrosheet_games"]; ok {
		entryCopy := entry
		echo.Infof("    Game log ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Game log ETL: never recorded")
	}

	if entry, ok := refreshes["retrosheet_plays"]; ok {
		entryCopy := entry
		echo.Infof("    Plays ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Plays ETL: never recorded")
	}

	echo.Info("")
	echo.Success("✓ Status check completed")
	return nil
}

func dirSnapshot(path string) (bool, int, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, time.Time{}, nil
		}
		return false, 0, time.Time{}, err
	}
	if !info.IsDir() {
		return false, 0, time.Time{}, fmt.Errorf("path is not a directory: %s", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	var latest time.Time
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		if entryInfo.ModTime().After(latest) {
			latest = entryInfo.ModTime()
		}
	}

	return true, len(entries), latest, nil
}

func safeCount(ctx context.Context, database *db.DB, query string) (int64, error) {
	var count int64
	if err := database.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func seasonRange(ctx context.Context, database *db.DB) (*int, *int, error) {
	var minYear, maxYear sql.NullInt64
	err := database.QueryRowContext(ctx, `SELECT MIN("yearID"), MAX("yearID") FROM "Teams"`).Scan(&minYear, &maxYear)
	if err != nil {
		return nil, nil, err
	}

	var minPtr, maxPtr *int
	if minYear.Valid {
		v := int(minYear.Int64)
		minPtr = &v
	}
	if maxYear.Valid {
		v := int(maxYear.Int64)
		maxPtr = &v
	}
	return minPtr, maxPtr, nil
}

func formatRefresh(entry *db.DatasetRefresh) string {
	if entry == nil || entry.LastLoadedAt.IsZero() {
		return "not yet recorded"
	}

	return fmt.Sprintf("%s (%s ago, %d rows)",
		entry.LastLoadedAt.Format(time.RFC1123),
		time.Since(entry.LastLoadedAt).Round(time.Minute),
		entry.RowCount,
	)
}

func humanizeModTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	ago := time.Since(t)
	return fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04"), ago.Round(time.Minute))
}

func retroDateRange(ctx context.Context, database *db.DB, table, column string) (*time.Time, *time.Time, error) {
	query := fmt.Sprintf(`SELECT MIN(%s), MAX(%s) FROM %s`, column, column, table)
	var minVal, maxVal sql.NullString
	if err := database.QueryRowContext(ctx, query).Scan(&minVal, &maxVal); err != nil {
		return nil, nil, err
	}

	start, err := parseRetroDate(minVal)
	if err != nil {
		return nil, nil, err
	}
	end, err := parseRetroDate(maxVal)
	if err != nil {
		return nil, nil, err
	}
	return start, end, nil
}

func parseRetroDate(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}
	t, err := time.Parse("20060102", value.String)
	if err != nil {
		return nil, fmt.Errorf("invalid Retrosheet date %q: %w", value.String, err)
	}
	return &t, nil
}
