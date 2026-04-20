package seed

import (
	"context"
	"fmt"
	"sort"
	"strings"
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

const defaultMaterializedViewSlowThreshold = 30 * time.Second

// RefreshRetrosheetMaterializedViews refreshes materialized views derived from Retrosheet games/plays.
func RefreshRetrosheetMaterializedViews(ctx context.Context, database *db.DB) (int, error) {
	return refreshNamedMaterializedViews(
		ctx,
		database,
		"Retrosheet-derived",
		retrosheetMaterializedViews,
		db.MaterializedViewRefreshOptions{
			Step:               "etl.load.retrosheet.refresh_views",
			ForceNonConcurrent: true,
		},
	)
}

// RefreshPipelineMaterializedViews refreshes all materialized views required by the ETL pipeline.
func RefreshPipelineMaterializedViews(ctx context.Context, database *db.DB, opts db.MaterializedViewRefreshOptions) (int, error) {
	total := 0

	count, err := refreshNamedMaterializedViews(ctx, database, "Retrosheet-derived", retrosheetMaterializedViews, opts)
	if err != nil {
		return total, err
	}
	total += count

	count, err = refreshNamedMaterializedViews(ctx, database, "crosswalk/metadata", supplementalPipelineMaterializedViews, opts)
	if err != nil {
		return total, err
	}
	total += count

	return total, nil
}

func refreshNamedMaterializedViews(
	ctx context.Context,
	database *db.DB,
	group string,
	views []string,
	opts db.MaterializedViewRefreshOptions,
) (int, error) {
	if len(views) == 0 {
		return 0, nil
	}

	slowThreshold := opts.SlowThreshold
	if slowThreshold <= 0 {
		slowThreshold = defaultMaterializedViewSlowThreshold
	}

	var attempts []db.MaterializedViewRefreshAttempt
	userCallback := opts.OnAttempt
	opts.OnAttempt = func(attempt db.MaterializedViewRefreshAttempt) {
		attempts = append(attempts, attempt)
		echo.Infof(
			"  MV view=%s pass=%d attempt=%d mode=%s status=%s duration=%s",
			attempt.ViewName,
			attempt.Pass,
			attempt.Attempt,
			attempt.Mode,
			attempt.Status,
			attempt.Duration.Round(time.Millisecond),
		)
		if attempt.Error != "" {
			echo.Infof("    error=%s", attempt.Error)
		}
		if userCallback != nil {
			userCallback(attempt)
		}
	}
	opts.Group = group

	start := time.Now()
	echo.Infof("Refreshing %s materialized views (%d)...", group, len(views))

	count, err := database.RefreshMaterializedViewsWithOptions(ctx, views, opts)
	logMaterializedViewSummary(group, attempts, slowThreshold)
	summary := materializedViewAttemptSummary(attempts, slowThreshold)

	if err != nil {
		recordETLPhaseEvent(
			ctx,
			database,
			"refresh.materialized_views",
			materializedViewPhaseName(group),
			"failed",
			int64(count),
			start,
			map[string]any{
				"group":       group,
				"view_count":  len(views),
				"attempts":    summary["attempts"],
				"retries":     summary["retries"],
				"deferred":    summary["deferred"],
				"failed":      summary["failed"],
				"slow_count":  summary["slow_count"],
				"slow_thresh": slowThreshold.Milliseconds(),
			},
			err,
		)
		return 0, fmt.Errorf("failed to refresh %s materialized views: %w", group, err)
	}

	recordETLPhaseEvent(
		ctx,
		database,
		"refresh.materialized_views",
		materializedViewPhaseName(group),
		"completed",
		int64(count),
		start,
		map[string]any{
			"group":       group,
			"view_count":  len(views),
			"attempts":    summary["attempts"],
			"retries":     summary["retries"],
			"deferred":    summary["deferred"],
			"failed":      summary["failed"],
			"slow_count":  summary["slow_count"],
			"slow_thresh": slowThreshold.Milliseconds(),
		},
		nil,
	)

	echo.Successf("  ✓ Refreshed %d %s views (%s)", count, group, time.Since(start).Round(time.Second))
	return count, nil
}

func logMaterializedViewSummary(group string, attempts []db.MaterializedViewRefreshAttempt, slowThreshold time.Duration) {
	if len(attempts) == 0 {
		return
	}

	retries := 0
	failed := 0
	deferred := 0
	var slow []db.MaterializedViewRefreshAttempt

	for _, attempt := range attempts {
		if attempt.Attempt > 1 {
			retries++
		}
		switch attempt.Status {
		case "failed":
			failed++
		case "deferred_dependency":
			deferred++
		case "completed":
			if attempt.Duration >= slowThreshold {
				slow = append(slow, attempt)
			}
		}
	}

	echo.Infof(
		"  MV summary group=%s attempts=%d retries=%d deferred=%d failed=%d",
		group,
		len(attempts),
		retries,
		deferred,
		failed,
	)

	if len(slow) == 0 {
		return
	}

	sort.Slice(slow, func(i, j int) bool {
		return slow[i].Duration > slow[j].Duration
	})

	limit := 5
	if len(slow) < limit {
		limit = len(slow)
	}
	echo.Infof("  MV slowest (threshold=%s):", slowThreshold.Round(time.Second))
	for i := 0; i < limit; i++ {
		echo.Infof("    %s (%s)", slow[i].ViewName, slow[i].Duration.Round(time.Millisecond))
	}
}

func materializedViewAttemptSummary(attempts []db.MaterializedViewRefreshAttempt, slowThreshold time.Duration) map[string]int {
	summary := map[string]int{
		"attempts":   len(attempts),
		"retries":    0,
		"deferred":   0,
		"failed":     0,
		"slow_count": 0,
	}

	for _, attempt := range attempts {
		if attempt.Attempt > 1 {
			summary["retries"]++
		}
		switch attempt.Status {
		case "failed":
			summary["failed"]++
		case "deferred_dependency":
			summary["deferred"]++
		case "completed":
			if attempt.Duration >= slowThreshold {
				summary["slow_count"]++
			}
		}
	}
	return summary
}

func materializedViewPhaseName(group string) string {
	if group == "" {
		return "materialized_views.refresh"
	}
	sanitized := strings.NewReplacer("/", "_", " ", "_").Replace(group)
	return "materialized_views." + sanitized
}
