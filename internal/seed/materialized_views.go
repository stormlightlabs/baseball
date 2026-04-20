package seed

import (
	"context"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

var retrosheetMaterializedViews = []string{
	"player_game_batting_stats",
	"player_game_pitching_stats",
	"player_game_fielding_stats",
	"team_game_stats",
	"win_expectancy_historical",
	"no_hitters",
	"cycles",
	"multi_hr_games",
	"triple_plays",
	"extra_inning_games",
	"season_batting_leaders",
	"season_pitching_leaders",
	"career_batting_leaders",
	"career_pitching_leaders",
}

var supplementalPipelineMaterializedViews = []string{
	"player_id_map",
	"team_franchise_map",
	"park_map",
}

// RefreshRetrosheetMaterializedViews refreshes materialized views derived from Retrosheet games/plays.
func RefreshRetrosheetMaterializedViews(ctx context.Context, database *db.DB) (int, error) {
	return refreshNamedMaterializedViews(ctx, database, "Retrosheet-derived", retrosheetMaterializedViews)
}

// RefreshPipelineMaterializedViews refreshes all materialized views required by the ETL pipeline.
func RefreshPipelineMaterializedViews(ctx context.Context, database *db.DB) (int, error) {
	total := 0

	count, err := refreshNamedMaterializedViews(ctx, database, "Retrosheet-derived", retrosheetMaterializedViews)
	if err != nil {
		return total, err
	}
	total += count

	count, err = refreshNamedMaterializedViews(ctx, database, "crosswalk/metadata", supplementalPipelineMaterializedViews)
	if err != nil {
		return total, err
	}
	total += count

	return total, nil
}

func refreshNamedMaterializedViews(ctx context.Context, database *db.DB, group string, views []string) (int, error) {
	if len(views) == 0 {
		return 0, nil
	}

	start := time.Now()
	echo.Infof("Refreshing %s materialized views (%d)...", group, len(views))

	count, err := database.RefreshMaterializedViews(ctx, views)
	if err != nil {
		return 0, fmt.Errorf("failed to refresh %s materialized views: %w", group, err)
	}

	echo.Successf("  ✓ Refreshed %d %s views (%s)", count, group, time.Since(start).Round(time.Second))
	return count, nil
}
