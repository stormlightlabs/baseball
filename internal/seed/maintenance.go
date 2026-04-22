package seed

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"stormlightlabs.org/baseball/internal/db"
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
	_ []int,
	refreshMode MVRefreshMode,
) (int64, error) {
	count, err := RefreshPipelineMaterializedViews(ctx, database, db.MaterializedViewRefreshOptions{
		RunID:              optionalRunIDPtr(runID),
		Step:               "maintenance.refresh_pipeline_materialized_views",
		ForceNonConcurrent: refreshMode.forceNonConcurrent(),
	})
	if err != nil {
		return int64(count), err
	}
	return int64(count), nil
}

func executeMaintenanceJob(ctx context.Context, database *db.DB, job *db.ETLJob) (int64, error) {
	years, err := intSliceFromAny(job.Scope["years"])
	if err != nil {
		return 0, fmt.Errorf("invalid years in maintenance scope: %w", err)
	}
	runID := int64FromAny(job.Scope["run_id"])
	refreshMode, err := ParseMVRefreshMode(stringFromAny(job.Options["mv_refresh_mode"]))
	if err != nil {
		return 0, err
	}

	return RunMaintenanceWindow(ctx, database, runID, years, refreshMode)
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
