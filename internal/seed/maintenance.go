package seed

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"stormlightlabs.org/baseball/internal/db"
)

const (
	maintenanceTypeGameStats      = "mv.batch.recompute.game_stats"
	maintenanceTypeAchievements   = "mv.batch.recompute.achievements"
	maintenanceTypeSeasonLeaders  = "mv.batch.recompute.season_leaders"
	maintenanceTypeCareerLeaders  = "mv.batch.recompute.career_leaders"
	maintenanceTypeCrosswalkViews = "mv.batch.recompute.crosswalk_views"
	maintenanceTypeWinExpCounts   = "mv.batch.recompute.win_expectancy_counts"
	maintenanceTypeWinExpPublish  = "mv.publish.win_expectancy"
)

type MVRefreshMode string

const (
	MVRefreshModeAuto          MVRefreshMode = "auto"
	MVRefreshModeNonConcurrent MVRefreshMode = "non_concurrent"
)

func ParseMVRefreshMode(raw string) (MVRefreshMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	if normalized == "" {
		return MVRefreshModeAuto, nil
	}

	switch MVRefreshMode(normalized) {
	case MVRefreshModeAuto:
		return MVRefreshModeAuto, nil
	case MVRefreshModeNonConcurrent:
		return MVRefreshModeNonConcurrent, nil
	default:
		return "", fmt.Errorf("invalid mv refresh mode %q (valid: auto, non_concurrent)", raw)
	}
}

func (m MVRefreshMode) forceNonConcurrent() bool {
	return m == MVRefreshModeNonConcurrent
}

func EnqueueMaintenanceJob(
	ctx context.Context,
	database *db.DB,
	opts PipelineOptions,
	workerOpts JobWorkerOptions,
	refreshMode MVRefreshMode,
) (int64, error) {
	opts, err := NormalizePipelineOptions(opts)
	if err != nil {
		return 0, err
	}
	workerOpts = NormalizeJobWorkerOptions(workerOpts)
	if refreshMode == "" {
		refreshMode = MVRefreshModeAuto
	}

	scope := map[string]any{
		"profile":   opts.Profile,
		"mode":      opts.Mode,
		"years":     opts.Years,
		"era_names": opts.EraNames,
		"data_root": opts.DataRoot,
	}
	jobID, err := database.EnqueueETLJob(ctx, db.ETLJobSpec{
		JobType:    db.ETLJobTypeMaintenance,
		Priority:   workerOpts.Priority,
		Profile:    string(opts.Profile),
		Mode:       string(opts.Mode),
		Scope:      scope,
		Options:    map[string]any{"mv_refresh_mode": string(refreshMode)},
		MaxRetries: workerOpts.MaxJobRetries,
	}, workerOpts.MaxQueuedJobs)
	if err != nil {
		return 0, err
	}

	return jobID, nil
}

func RunMaintenanceWindow(
	ctx context.Context,
	database *db.DB,
	runID int64,
	years []int,
	refreshMode MVRefreshMode,
) (int64, error) {
	years = dedupeSortedYears(years)
	if len(years) == 0 {
		return 0, nil
	}

	refreshOpts := db.MaterializedViewRefreshOptions{
		RunID:              optionalRunIDPtr(runID),
		ForceNonConcurrent: refreshMode.forceNonConcurrent(),
	}

	var rows int64

	if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.game_stats", []string{
		"player_game_batting_stats",
		"player_game_pitching_stats",
		"player_game_fielding_stats",
		"team_game_stats",
	}, withRefreshStep(refreshOpts, "maintenance.game_stats.refresh")); err != nil {
		return rows, err
	}

	count, err := syncServingTableBySeasons(ctx, database, "serving_player_game_batting_stats", "player_game_batting_stats", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_player_game_pitching_stats", "player_game_pitching_stats", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_player_game_fielding_stats", "player_game_fielding_stats", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_team_game_stats", "team_game_stats", years)
	if err != nil {
		return rows, err
	}
	rows += count

	if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.achievements", []string{
		"no_hitters",
		"cycles",
		"multi_hr_games",
		"triple_plays",
		"extra_inning_games",
	}, withRefreshStep(refreshOpts, "maintenance.achievements.refresh")); err != nil {
		return rows, err
	}

	count, err = syncServingTableBySeasons(ctx, database, "serving_no_hitters", "no_hitters", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_cycles", "cycles", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_multi_hr_games", "multi_hr_games", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_triple_plays", "triple_plays", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_extra_inning_games", "extra_inning_games", years)
	if err != nil {
		return rows, err
	}
	rows += count

	if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.season_leaders", []string{
		"season_batting_leaders",
		"season_pitching_leaders",
	}, withRefreshStep(refreshOpts, "maintenance.season_leaders.refresh")); err != nil {
		return rows, err
	}

	count, err = syncServingTableBySeasons(ctx, database, "serving_season_batting_leaders", "season_batting_leaders", years)
	if err != nil {
		return rows, err
	}
	rows += count
	count, err = syncServingTableBySeasons(ctx, database, "serving_season_pitching_leaders", "season_pitching_leaders", years)
	if err != nil {
		return rows, err
	}
	rows += count

	if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.career_leaders", []string{
		"career_batting_leaders",
		"career_pitching_leaders",
	}, withRefreshStep(refreshOpts, "maintenance.career_leaders.refresh")); err != nil {
		return rows, err
	}

	players, err := changedPlayersForYearScope(ctx, database, years)
	if err != nil {
		return rows, err
	}
	if len(players) > 0 {
		count, err = syncServingTableByPlayers(ctx, database, "serving_career_batting_leaders", "career_batting_leaders", players)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableByPlayers(ctx, database, "serving_career_pitching_leaders", "career_pitching_leaders", players)
		if err != nil {
			return rows, err
		}
		rows += count
	}

	if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.crosswalk_views", []string{
		"player_id_map",
		"team_franchise_map",
		"park_map",
	}, withRefreshStep(refreshOpts, "maintenance.crosswalk_views.refresh")); err != nil {
		return rows, err
	}

	count, err = rebuildWinExpectancyStateCounts(ctx, database, years)
	if err != nil {
		return rows, err
	}
	rows += count

	count, err = publishWinExpectancyServingTable(ctx, database)
	if err != nil {
		return rows, err
	}
	rows += count

	return rows, nil
}

func withRefreshStep(opts db.MaterializedViewRefreshOptions, step string) db.MaterializedViewRefreshOptions {
	out := opts
	out.Step = step
	return out
}

func captureDeltaScopeForRun(ctx context.Context, database *db.DB, runID int64, years []int) error {
	years = dedupeSortedYears(years)
	if runID <= 0 || len(years) == 0 {
		return nil
	}

	if err := database.UpsertDeltaSeasonsForRun(ctx, runID, years); err != nil {
		return err
	}
	if err := database.UpsertDeltaGamesForRun(ctx, runID, years); err != nil {
		return err
	}
	if err := database.UpsertDeltaPlayersForRun(ctx, runID, years); err != nil {
		return err
	}
	return nil
}

func enqueueMaintenanceJobsForRun(
	ctx context.Context,
	database *db.DB,
	runID int64,
	opts PipelineOptions,
	workerOpts JobWorkerOptions,
) ([]int64, error) {
	years := dedupeSortedYears(opts.Years)
	batches := splitYearsIntoBatches(years, workerOpts.YearBatchSize)
	jobIDs := make([]int64, 0, len(batches)*4+3)
	queueLimit := workerOpts.MaxQueuedJobs
	plannedJobs := len(batches)*4 + 3
	if queueLimit > 0 && queueLimit < plannedJobs+8 {
		queueLimit = plannedJobs + 8
	}

	enqueue := func(maintenanceType string, scopeYears []int) error {
		scope := map[string]any{
			"run_id":           runID,
			"maintenance_type": maintenanceType,
			"years":            scopeYears,
			"profile":          opts.Profile,
			"mode":             opts.Mode,
			"data_root":        opts.DataRoot,
		}
		jobID, err := database.EnqueueETLJob(ctx, db.ETLJobSpec{
			JobType:    db.ETLJobTypeMaintenance,
			Priority:   workerOpts.Priority,
			Profile:    string(opts.Profile),
			Mode:       string(opts.Mode),
			Scope:      scope,
			Options:    map[string]any{},
			MaxRetries: workerOpts.MaxJobRetries,
		}, queueLimit)
		if err != nil {
			return err
		}
		jobIDs = append(jobIDs, jobID)
		return nil
	}

	for _, batchYears := range batches {
		if err := enqueue(maintenanceTypeGameStats, batchYears); err != nil {
			return nil, err
		}
		if err := enqueue(maintenanceTypeAchievements, batchYears); err != nil {
			return nil, err
		}
		if err := enqueue(maintenanceTypeSeasonLeaders, batchYears); err != nil {
			return nil, err
		}
		if err := enqueue(maintenanceTypeWinExpCounts, batchYears); err != nil {
			return nil, err
		}
	}

	if err := enqueue(maintenanceTypeCareerLeaders, nil); err != nil {
		return nil, err
	}
	if err := enqueue(maintenanceTypeCrosswalkViews, nil); err != nil {
		return nil, err
	}
	if err := enqueue(maintenanceTypeWinExpPublish, nil); err != nil {
		return nil, err
	}

	return jobIDs, nil
}

func executeMaintenanceJob(ctx context.Context, database *db.DB, job *db.ETLJob) (int64, error) {
	maintenanceType := strings.TrimSpace(stringFromAny(job.Scope["maintenance_type"]))
	years, err := intSliceFromAny(job.Scope["years"])
	if err != nil {
		return 0, fmt.Errorf("invalid years in maintenance scope: %w", err)
	}
	years = dedupeSortedYears(years)
	runID := int64FromAny(job.Scope["run_id"])
	refreshMode, err := ParseMVRefreshMode(stringFromAny(job.Options["mv_refresh_mode"]))
	if err != nil {
		return 0, err
	}

	if maintenanceType == "" {
		return RunMaintenanceWindow(ctx, database, runID, years, refreshMode)
	}

	switch maintenanceType {
	case maintenanceTypeGameStats:
		if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.game_stats", []string{
			"player_game_batting_stats",
			"player_game_pitching_stats",
			"player_game_fielding_stats",
			"team_game_stats",
		}, db.MaterializedViewRefreshOptions{
			RunID:              optionalRunIDPtr(runID),
			Step:               "maintenance.game_stats.refresh",
			ForceNonConcurrent: true,
		}); err != nil {
			return 0, err
		}

		var rows int64
		count, err := syncServingTableBySeasons(ctx, database, "serving_player_game_batting_stats", "player_game_batting_stats", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_player_game_pitching_stats", "player_game_pitching_stats", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_player_game_fielding_stats", "player_game_fielding_stats", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_team_game_stats", "team_game_stats", years)
		if err != nil {
			return rows, err
		}
		rows += count
		return rows, nil

	case maintenanceTypeAchievements:
		if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.achievements", []string{
			"no_hitters",
			"cycles",
			"multi_hr_games",
			"triple_plays",
			"extra_inning_games",
		}, db.MaterializedViewRefreshOptions{
			RunID:              optionalRunIDPtr(runID),
			Step:               "maintenance.achievements.refresh",
			ForceNonConcurrent: true,
		}); err != nil {
			return 0, err
		}

		var rows int64
		count, err := syncServingTableBySeasons(ctx, database, "serving_no_hitters", "no_hitters", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_cycles", "cycles", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_multi_hr_games", "multi_hr_games", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_triple_plays", "triple_plays", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_extra_inning_games", "extra_inning_games", years)
		if err != nil {
			return rows, err
		}
		rows += count
		return rows, nil

	case maintenanceTypeSeasonLeaders:
		if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.season_leaders", []string{
			"season_batting_leaders",
			"season_pitching_leaders",
		}, db.MaterializedViewRefreshOptions{
			RunID:              optionalRunIDPtr(runID),
			Step:               "maintenance.season_leaders.refresh",
			ForceNonConcurrent: true,
		}); err != nil {
			return 0, err
		}

		var rows int64
		count, err := syncServingTableBySeasons(ctx, database, "serving_season_batting_leaders", "season_batting_leaders", years)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableBySeasons(ctx, database, "serving_season_pitching_leaders", "season_pitching_leaders", years)
		if err != nil {
			return rows, err
		}
		rows += count
		return rows, nil

	case maintenanceTypeCareerLeaders:
		if _, err := refreshNamedMaterializedViews(ctx, database, "maintenance.career_leaders", []string{
			"career_batting_leaders",
			"career_pitching_leaders",
		}, db.MaterializedViewRefreshOptions{
			RunID:              optionalRunIDPtr(runID),
			Step:               "maintenance.career_leaders.refresh",
			ForceNonConcurrent: true,
		}); err != nil {
			return 0, err
		}

		players, err := changedPlayersForRunScope(ctx, database, runID)
		if err != nil {
			return 0, err
		}
		if len(players) == 0 {
			return 0, nil
		}

		var rows int64
		count, err := syncServingTableByPlayers(ctx, database, "serving_career_batting_leaders", "career_batting_leaders", players)
		if err != nil {
			return rows, err
		}
		rows += count
		count, err = syncServingTableByPlayers(ctx, database, "serving_career_pitching_leaders", "career_pitching_leaders", players)
		if err != nil {
			return rows, err
		}
		rows += count
		return rows, nil

	case maintenanceTypeCrosswalkViews:
		count, err := refreshNamedMaterializedViews(ctx, database, "maintenance.crosswalk_views", []string{
			"player_id_map",
			"team_franchise_map",
			"park_map",
		}, db.MaterializedViewRefreshOptions{
			RunID:              optionalRunIDPtr(runID),
			Step:               "maintenance.crosswalk_views.refresh",
			ForceNonConcurrent: true,
		})
		if err != nil {
			return 0, err
		}
		return int64(count), nil

	case maintenanceTypeWinExpCounts:
		return rebuildWinExpectancyStateCounts(ctx, database, years)

	case maintenanceTypeWinExpPublish:
		return publishWinExpectancyServingTable(ctx, database)

	default:
		count, err := RefreshPipelineMaterializedViews(ctx, database, db.MaterializedViewRefreshOptions{
			RunID:              optionalRunIDPtr(runID),
			Step:               "maintenance.legacy_refresh",
			ForceNonConcurrent: true,
		})
		if err != nil {
			return 0, err
		}
		return int64(count), nil
	}
}

func changedPlayersForRunScope(ctx context.Context, database *db.DB, runID int64) ([]string, error) {
	if runID <= 0 {
		return nil, nil
	}
	players, err := database.DeltaPlayersForRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return players, nil
}

func changedPlayersForYearScope(ctx context.Context, database *db.DB, years []int) ([]string, error) {
	years = dedupeSortedYears(years)
	if len(years) == 0 {
		return nil, nil
	}

	players, err := database.DeltaPlayersForYears(ctx, years)
	if err != nil {
		return nil, err
	}
	return players, nil
}

func syncServingTableBySeasons(ctx context.Context, database *db.DB, servingTable, sourceTable string, years []int) (int64, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin serving sync tx for %s: %w", servingTable, err)
	}
	defer tx.Rollback()

	if len(years) == 0 {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`TRUNCATE TABLE %s`, servingTable)); err != nil {
			return 0, fmt.Errorf("failed to truncate %s: %w", servingTable, err)
		}
		result, err := tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %s
			SELECT * FROM %s
		`, servingTable, sourceTable))
		if err != nil {
			return 0, fmt.Errorf("failed to sync %s from %s: %w", servingTable, sourceTable, err)
		}
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("failed to commit serving sync for %s: %w", servingTable, err)
		}
		rows, _ := result.RowsAffected()
		return rows, nil
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE season = ANY($1::int[])
	`, servingTable), years); err != nil {
		return 0, fmt.Errorf("failed to clear scoped rows in %s: %w", servingTable, err)
	}

	result, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
		SELECT *
		FROM %s
		WHERE season = ANY($1::int[])
	`, servingTable, sourceTable), years)
	if err != nil {
		return 0, fmt.Errorf("failed to sync scoped rows for %s from %s: %w", servingTable, sourceTable, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit serving sync for %s: %w", servingTable, err)
	}

	rows, _ := result.RowsAffected()
	return rows, nil
}

func syncServingTableByPlayers(ctx context.Context, database *db.DB, servingTable, sourceTable string, players []string) (int64, error) {
	if len(players) == 0 {
		return 0, nil
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin serving player sync tx for %s: %w", servingTable, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE player_id = ANY($1::text[])
	`, servingTable), players); err != nil {
		return 0, fmt.Errorf("failed to clear scoped players in %s: %w", servingTable, err)
	}

	result, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
		SELECT *
		FROM %s
		WHERE player_id = ANY($1::text[])
	`, servingTable, sourceTable), players)
	if err != nil {
		return 0, fmt.Errorf("failed to sync scoped players for %s from %s: %w", servingTable, sourceTable, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit serving player sync for %s: %w", servingTable, err)
	}

	rows, _ := result.RowsAffected()
	return rows, nil
}

func rebuildWinExpectancyStateCounts(ctx context.Context, database *db.DB, years []int) (int64, error) {
	years = dedupeSortedYears(years)
	if len(years) == 0 {
		return 0, nil
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin win expectancy counts tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM serving_win_expectancy_state_counts
		WHERE season = ANY($1::int[])
	`, years); err != nil {
		return 0, fmt.Errorf("failed to clear scoped win expectancy counts: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		WITH game_outcomes AS (
			SELECT
				date,
				home_team,
				game_number,
				CAST(SUBSTRING(date, 1, 4) AS INTEGER) AS season,
				CASE
					WHEN home_score > visiting_score THEN TRUE
					WHEN home_score < visiting_score THEN FALSE
					ELSE NULL
				END AS home_won
			FROM games
			WHERE home_score IS NOT NULL
			  AND visiting_score IS NOT NULL
			  AND home_score <> visiting_score
			  AND CAST(SUBSTRING(date, 1, 4) AS INTEGER) = ANY($1::int[])
		),
		game_states AS (
			SELECT
				go.season,
				LEAST(p.inning, 9) AS inning,
				(p.top_bot = 1)::boolean AS is_bottom,
				p.outs_pre AS outs,
				CONCAT(
					CASE WHEN p.br1_pre IS NOT NULL AND p.br1_pre <> '' THEN '1' ELSE '_' END,
					CASE WHEN p.br2_pre IS NOT NULL AND p.br2_pre <> '' THEN '2' ELSE '_' END,
					CASE WHEN p.br3_pre IS NOT NULL AND p.br3_pre <> '' THEN '3' ELSE '_' END
				) AS runners_state,
				LEAST(GREATEST(p.score_h - p.score_v, -11), 11) AS score_diff,
				go.home_won
			FROM plays p
			JOIN game_outcomes go ON
				SUBSTRING(p.gid, 4, 8) = go.date AND
				LEFT(p.gid, 3) = go.home_team AND
				RIGHT(p.gid, 1)::int = go.game_number
			WHERE p.outs_pre IS NOT NULL
			  AND p.inning IS NOT NULL
			  AND go.home_won IS NOT NULL
		)
		INSERT INTO serving_win_expectancy_state_counts (
			season,
			inning,
			is_bottom,
			outs,
			runners_state,
			score_diff,
			sample_size,
			home_win_samples,
			updated_at
		)
		SELECT
			season,
			inning,
			is_bottom,
			outs,
			runners_state,
			score_diff,
			COUNT(*)::bigint AS sample_size,
			SUM(CASE WHEN home_won THEN 1 ELSE 0 END)::bigint AS home_win_samples,
			NOW()
		FROM game_states
		GROUP BY season, inning, is_bottom, outs, runners_state, score_diff
	`, years)
	if err != nil {
		return 0, fmt.Errorf("failed to insert scoped win expectancy counts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit win expectancy counts tx: %w", err)
	}

	rows, _ := result.RowsAffected()
	return rows, nil
}

func publishWinExpectancyServingTable(ctx context.Context, database *db.DB) (int64, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin win expectancy publish tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `TRUNCATE TABLE serving_win_expectancy_historical RESTART IDENTITY`); err != nil {
		return 0, fmt.Errorf("failed to truncate serving win expectancy table: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		WITH aggregate_states AS (
			SELECT
				inning,
				is_bottom,
				outs,
				runners_state,
				score_diff,
				MIN(season) AS start_year,
				MAX(season) AS end_year,
				SUM(sample_size) AS sample_size,
				SUM(home_win_samples) AS home_win_samples
			FROM serving_win_expectancy_state_counts
			GROUP BY inning, is_bottom, outs, runners_state, score_diff
		)
		INSERT INTO serving_win_expectancy_historical (
			inning,
			is_bottom,
			outs,
			runners_state,
			score_diff,
			win_probability,
			sample_size,
			start_year,
			end_year,
			created_at,
			updated_at
		)
		SELECT
			inning,
			is_bottom,
			outs,
			runners_state,
			score_diff,
			ROUND((home_win_samples::numeric / NULLIF(sample_size::numeric, 0)), 4)::double precision AS win_probability,
			sample_size,
			start_year,
			end_year,
			NOW(),
			NOW()
		FROM aggregate_states
		WHERE sample_size >= 100
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to publish serving win expectancy table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit win expectancy publish tx: %w", err)
	}

	rows, _ := result.RowsAffected()
	return rows, nil
}

func dedupeSortedYears(years []int) []int {
	if len(years) == 0 {
		return nil
	}
	copyYears := make([]int, 0, len(years))
	for _, year := range years {
		if year > 0 {
			copyYears = append(copyYears, year)
		}
	}
	sort.Ints(copyYears)
	result := make([]int, 0, len(copyYears))
	last := 0
	for idx, year := range copyYears {
		if idx == 0 || year != last {
			result = append(result, year)
			last = year
		}
	}
	return result
}

func optionalRunIDPtr(runID int64) *int64 {
	if runID <= 0 {
		return nil
	}
	v := runID
	return &v
}
