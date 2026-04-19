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
	"stormlightlabs.org/baseball/internal/seed"
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
	dataRoot := resolveDataRoot(cmd)

	archiveChecks := []struct {
		label string
		path  string
		hint  string
	}{
		{
			label: "Lahman CSVs",
			path:  seed.LahmanCSVDir(dataRoot),
			hint:  "Use `baseball etl fetch lahman` to scaffold/download the dataset",
		},
		{
			label: "Retrosheet game logs",
			path:  filepath.Join(seed.RetrosheetDir(dataRoot), "gamelogs"),
			hint:  "Use `baseball etl fetch retrosheet` to download seasonal game logs",
		},
		{
			label: "Retrosheet plays",
			path:  filepath.Join(seed.RetrosheetDir(dataRoot), "plays"),
			hint:  "Use `baseball etl fetch retrosheet` to download parsed play-by-play archives",
		},
	}

	echo.Info("Local archives:")
	echo.Infof("  Data root: %s", dataRoot)
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
	echo.Info("• Supplemental Datasets")

	wobaConstants, wobaErr := safeCount(ctx, database, `SELECT COUNT(*) FROM woba_constants`)
	leagueConstants, leagueErr := safeCount(ctx, database, `SELECT COUNT(*) FROM league_constants`)
	parkFactors, pfErr := safeCount(ctx, database, `SELECT COUNT(*) FROM park_factors`)
	if wobaErr != nil || leagueErr != nil || pfErr != nil {
		echo.Infof("  ⚠ FanGraphs constants unavailable: %v %v %v", wobaErr, leagueErr, pfErr)
	} else if wobaConstants > 0 && leagueConstants > 0 && parkFactors > 0 {
		echo.Successf("  ✓ FanGraphs constants loaded (wOBA=%d, league=%d, park factors=%d)", wobaConstants, leagueConstants, parkFactors)
	} else {
		echo.Infof("  • FanGraphs constants incomplete (wOBA=%d, league=%d, park factors=%d)", wobaConstants, leagueConstants, parkFactors)
	}
	if entry, ok := refreshes["fangraphs_constants"]; ok {
		entryCopy := entry
		echo.Infof("    FanGraphs ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    FanGraphs ETL: never recorded")
	}

	retroPlayersCount, retroPlayersErr := safeCount(ctx, database, `SELECT COUNT(*) FROM retrosheet_players`)
	if retroPlayersErr != nil {
		echo.Infof("  ⚠ Retrosheet players unavailable: %v", retroPlayersErr)
	} else if retroPlayersCount > 0 {
		echo.Successf("  ✓ Retrosheet player crosswalk loaded (%d rows)", retroPlayersCount)
	} else {
		echo.Infof("  • Retrosheet player crosswalk empty")
	}
	if entry, ok := refreshes["retrosheet_players"]; ok {
		entryCopy := entry
		echo.Infof("    Retrosheet players ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Retrosheet players ETL: never recorded")
	}

	bioExtendedCount, bioExtendedErr := safeCount(ctx, database, `SELECT COUNT(*) FROM player_bio_extended`)
	relativesCount, relativesErr := safeCount(ctx, database, `SELECT COUNT(*) FROM player_relatives`)
	coachesCount, coachesErr := safeCount(ctx, database, `SELECT COUNT(*) FROM coaches`)
	umpiresCount, umpiresErr := safeCount(ctx, database, `SELECT COUNT(*) FROM umpires`)
	if bioExtendedErr != nil || relativesErr != nil || coachesErr != nil || umpiresErr != nil {
		echo.Infof("  ⚠ Biodata tables unavailable: %v %v %v %v", bioExtendedErr, relativesErr, coachesErr, umpiresErr)
	} else if bioExtendedCount > 0 || relativesCount > 0 || coachesCount > 0 || umpiresCount > 0 {
		echo.Successf("  ✓ Biodata loaded (bio=%d, relatives=%d, coaches=%d, umpires=%d)", bioExtendedCount, relativesCount, coachesCount, umpiresCount)
	} else {
		echo.Infof("  • Biodata tables are empty")
	}
	if entry, ok := refreshes["biodata"]; ok {
		entryCopy := entry
		echo.Infof("    Biodata ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Biodata ETL: never recorded")
	}

	salarySummaryCount, salarySummaryErr := safeCount(ctx, database, `SELECT COUNT(*) FROM salary_summary`)
	if salarySummaryErr != nil {
		echo.Infof("  ⚠ Salary summary unavailable: %v", salarySummaryErr)
	} else if salarySummaryCount > 0 {
		echo.Successf("  ✓ Salary summary loaded (%d seasons)", salarySummaryCount)
	} else {
		echo.Infof("  • Salary summary table is empty")
	}
	if entry, ok := refreshes["salaries"]; ok {
		entryCopy := entry
		echo.Infof("    Salary ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Salary ETL: never recorded")
	}

	weatherGamesCount, weatherGamesErr := safeCount(ctx, database, `SELECT COUNT(*) FROM games WHERE temp_f IS NOT NULL OR wind_speed_mph IS NOT NULL OR sky IS NOT NULL OR precip IS NOT NULL OR field_condition IS NOT NULL`)
	if weatherGamesErr != nil {
		echo.Infof("  ⚠ Weather metadata check failed: %v", weatherGamesErr)
	} else if weatherGamesCount > 0 {
		echo.Successf("  ✓ Weather/gameinfo metadata applied to %d games", weatherGamesCount)
	} else {
		echo.Infof("  • No games currently have weather metadata")
	}
	if entry, ok := refreshes["retrosheet_gameinfo"]; ok {
		entryCopy := entry
		echo.Infof("    Weather ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Weather ETL: never recorded")
	}

	allStarGamesCount, allStarGamesErr := safeCount(ctx, database, `SELECT COUNT(*) FROM games WHERE game_type = 'allstar'`)
	allStarPlaysCount, allStarPlaysErr := safeCount(ctx, database, `SELECT COUNT(*) FROM plays WHERE gid LIKE 'ALS%'`)
	if allStarGamesErr != nil || allStarPlaysErr != nil {
		echo.Infof("  ⚠ All-Star data check failed: %v %v", allStarGamesErr, allStarPlaysErr)
	} else if allStarGamesCount > 0 {
		echo.Successf("  ✓ All-Star data loaded (games=%d, plays=%d)", allStarGamesCount, allStarPlaysCount)
	} else {
		echo.Infof("  • All-Star game data not detected in games table")
	}
	if entry, ok := refreshes["allstar_games"]; ok {
		entryCopy := entry
		echo.Infof("    All-Star games ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    All-Star games ETL: never recorded")
	}
	if entry, ok := refreshes["allstar_plays"]; ok {
		entryCopy := entry
		echo.Infof("    All-Star plays ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    All-Star plays ETL: never recorded")
	}

	ejectionsCount, ejectionsErr := safeCount(ctx, database, `SELECT COUNT(*) FROM ejections`)
	if ejectionsErr != nil {
		echo.Infof("  ⚠ Ejections check failed: %v", ejectionsErr)
	} else if ejectionsCount > 0 {
		echo.Successf("  ✓ Ejections loaded (%d rows)", ejectionsCount)
	} else {
		echo.Infof("  • Ejections table is empty")
	}
	if entry, ok := refreshes["retrosheet_ejections"]; ok {
		entryCopy := entry
		echo.Infof("    Ejections ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Ejections ETL: never recorded")
	}

	parkMapCount, parkMapErr := safeCount(ctx, database, `SELECT COUNT(*) FROM park_map`)
	if parkMapErr != nil {
		echo.Infof("  ⚠ Park map check failed: %v", parkMapErr)
	} else if parkMapCount > 0 {
		echo.Successf("  ✓ Park map available (%d rows)", parkMapCount)
	} else {
		echo.Infof("  • Park map view is empty")
	}
	if entry, ok := refreshes["parks_metadata"]; ok {
		entryCopy := entry
		echo.Infof("    Parks metadata ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Parks metadata ETL: never recorded")
	}

	if entry, ok := refreshes["negroleagues_games"]; ok {
		entryCopy := entry
		echo.Infof("    Negro Leagues games ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Negro Leagues games ETL: never recorded")
	}
	if entry, ok := refreshes["negroleagues_plays"]; ok {
		entryCopy := entry
		echo.Infof("    Negro Leagues plays ETL: %s", formatRefresh(&entryCopy))
	} else {
		echo.Infof("    Negro Leagues plays ETL: never recorded")
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
