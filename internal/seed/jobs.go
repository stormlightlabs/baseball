package seed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

const (
	defaultETLMaxActiveJobs       = 1
	defaultETLMaxQueuedJobs       = 128
	defaultETLYearBatchSize       = 2
	defaultETLPriority            = 50
	defaultETLNetworkRetryMax     = 2
	defaultETLNetworkRetryBackoff = 3 * time.Second
	defaultETLMaxJobRetries       = 2
	defaultETLWorkerPollInterval  = 5 * time.Second
)

type JobWorkerOptions struct {
	JobType             db.ETLJobType
	Priority            int
	MaxActiveJobs       int
	MaxQueuedJobs       int
	YearBatchSize       int
	BatchDelay          time.Duration
	NetworkRetryMax     int
	NetworkRetryBackoff time.Duration
	LoadChunkSize       int
	MaxJobRetries       int
	WorkerID            string
}

type QueuePipelineResult struct {
	JobIDs      []int64
	JobType     db.ETLJobType
	BatchCount  int
	BatchScopes [][]int
}

type ProcessETLQueueResult struct {
	WorkerID  string
	Processed int
	Succeeded int
	Failed    int
	Retried   int
	Cancelled int
	Rows      int64
}

func NormalizeJobWorkerOptions(opts JobWorkerOptions) JobWorkerOptions {
	if opts.Priority < 0 {
		opts.Priority = defaultETLPriority
	}
	if opts.MaxActiveJobs < 1 {
		opts.MaxActiveJobs = envIntDefault("ETL_MAX_ACTIVE_JOBS", defaultETLMaxActiveJobs)
	}
	if opts.MaxQueuedJobs < 1 {
		opts.MaxQueuedJobs = envIntDefault("ETL_MAX_QUEUED_JOBS", defaultETLMaxQueuedJobs)
	}
	if opts.YearBatchSize < 1 {
		opts.YearBatchSize = envIntDefault("ETL_YEAR_BATCH_SIZE", defaultETLYearBatchSize)
	}
	if opts.NetworkRetryMax < 0 {
		opts.NetworkRetryMax = envIntDefault("ETL_NETWORK_RETRY_MAX", defaultETLNetworkRetryMax)
	}
	if opts.NetworkRetryBackoff <= 0 {
		opts.NetworkRetryBackoff = envDurationDefault("ETL_NETWORK_RETRY_BACKOFF", defaultETLNetworkRetryBackoff)
	}
	if opts.MaxJobRetries < 0 {
		opts.MaxJobRetries = envIntDefault("ETL_JOB_MAX_RETRIES", defaultETLMaxJobRetries)
	}
	if opts.WorkerID == "" {
		host, _ := os.Hostname()
		if strings.TrimSpace(host) == "" {
			host = "local"
		}
		opts.WorkerID = fmt.Sprintf("etl-worker-%s-%d", host, time.Now().Unix())
	}
	return opts
}

func ParseETLJobType(raw string, mode PipelineMode, years []int) (db.ETLJobType, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "", "auto":
		if mode == PipelineModeFull {
			return db.ETLJobTypeFullRun, nil
		}
		if len(years) <= 1 {
			return db.ETLJobTypeYearlySync, nil
		}
		return db.ETLJobTypeYearlySync, nil
	case string(db.ETLJobTypeFullRun):
		return db.ETLJobTypeFullRun, nil
	case string(db.ETLJobTypeYearlySync):
		return db.ETLJobTypeYearlySync, nil
	case string(db.ETLJobTypeValidate):
		return db.ETLJobTypeValidate, nil
	case string(db.ETLJobTypeCleanup):
		return db.ETLJobTypeCleanup, nil
	case string(db.ETLJobTypeMaintenance):
		return db.ETLJobTypeMaintenance, nil
	case string(db.ETLJobTypeCurrentSync):
		return db.ETLJobTypeCurrentSync, nil
	default:
		return "", fmt.Errorf("invalid job type %q", raw)
	}
}

func EnqueuePipelineJobs(ctx context.Context, database *db.DB, pipelineOpts PipelineOptions, workerOpts JobWorkerOptions) (QueuePipelineResult, error) {
	workerOpts = NormalizeJobWorkerOptions(workerOpts)

	if workerOpts.JobType == db.ETLJobTypeCurrentSync {
		return enqueueCurrentSeasonSyncJobs(ctx, database, pipelineOpts, workerOpts)
	}

	pipelineOpts, err := NormalizePipelineOptions(pipelineOpts)
	if err != nil {
		return QueuePipelineResult{}, err
	}

	if workerOpts.JobType == "" {
		workerOpts.JobType, err = ParseETLJobType("auto", pipelineOpts.Mode, pipelineOpts.Years)
		if err != nil {
			return QueuePipelineResult{}, err
		}
	}

	batchScopes := splitYearsIntoBatches(pipelineOpts.Years, workerOpts.YearBatchSize)
	if workerOpts.JobType == db.ETLJobTypeValidate || workerOpts.JobType == db.ETLJobTypeCleanup || workerOpts.JobType == db.ETLJobTypeMaintenance {
		batchScopes = [][]int{slices.Clone(pipelineOpts.Years)}
	}

	result := QueuePipelineResult{
		JobType:     workerOpts.JobType,
		BatchCount:  len(batchScopes),
		BatchScopes: batchScopes,
		JobIDs:      make([]int64, 0, len(batchScopes)),
	}

	for batchIdx, batchYears := range batchScopes {
		scope := map[string]any{
			"profile":   pipelineOpts.Profile,
			"mode":      pipelineOpts.Mode,
			"years":     batchYears,
			"era_names": pipelineOpts.EraNames,
			"data_root": pipelineOpts.DataRoot,
			"batch": map[string]any{
				"index":    batchIdx + 1,
				"total":    len(batchScopes),
				"size":     len(batchYears),
				"job_type": workerOpts.JobType,
			},
		}

		options := map[string]any{
			"network_retry_max":        workerOpts.NetworkRetryMax,
			"network_retry_backoff_ms": workerOpts.NetworkRetryBackoff.Milliseconds(),
			"load_chunk_size":          workerOpts.LoadChunkSize,
			"batch_delay_ms":           workerOpts.BatchDelay.Milliseconds(),
		}

		jobID, err := database.EnqueueETLJob(ctx, db.ETLJobSpec{
			JobType:    workerOpts.JobType,
			Priority:   workerOpts.Priority,
			Profile:    string(pipelineOpts.Profile),
			Mode:       string(pipelineOpts.Mode),
			Scope:      scope,
			Options:    options,
			MaxRetries: workerOpts.MaxJobRetries,
		}, workerOpts.MaxQueuedJobs)
		if err != nil {
			return result, err
		}
		result.JobIDs = append(result.JobIDs, jobID)
	}

	return result, nil
}

func enqueueCurrentSeasonSyncJobs(
	ctx context.Context,
	database *db.DB,
	pipelineOpts PipelineOptions,
	workerOpts JobWorkerOptions,
) (QueuePipelineResult, error) {
	seasons := dedupeSortedYears(pipelineOpts.Years)
	if len(seasons) == 0 {
		seasons = []int{time.Now().Year()}
	}

	profile := strings.TrimSpace(string(pipelineOpts.Profile))
	if profile == "" {
		profile = "current-season"
	}
	mode := strings.TrimSpace(string(pipelineOpts.Mode))
	if mode == "" {
		mode = string(PipelineModeIncremental)
	}

	result := QueuePipelineResult{
		JobType:     db.ETLJobTypeCurrentSync,
		BatchCount:  len(seasons),
		BatchScopes: make([][]int, 0, len(seasons)),
		JobIDs:      make([]int64, 0, len(seasons)),
	}

	for idx, season := range seasons {
		scope := map[string]any{
			"profile":     profile,
			"mode":        mode,
			"years":       []int{season},
			"season":      season,
			"sync_type":   currentSeasonSyncAll,
			"batch_index": idx + 1,
			"batch_total": len(seasons),
		}
		options := map[string]any{
			"network_retry_max":        workerOpts.NetworkRetryMax,
			"network_retry_backoff_ms": workerOpts.NetworkRetryBackoff.Milliseconds(),
			"load_chunk_size":          workerOpts.LoadChunkSize,
			"batch_delay_ms":           workerOpts.BatchDelay.Milliseconds(),
		}

		jobID, err := database.EnqueueETLJob(ctx, db.ETLJobSpec{
			JobType:    db.ETLJobTypeCurrentSync,
			Priority:   workerOpts.Priority,
			Profile:    profile,
			Mode:       mode,
			Scope:      scope,
			Options:    options,
			MaxRetries: workerOpts.MaxJobRetries,
		}, workerOpts.MaxQueuedJobs)
		if err != nil {
			return result, err
		}

		result.JobIDs = append(result.JobIDs, jobID)
		result.BatchScopes = append(result.BatchScopes, []int{season})
	}

	return result, nil
}

func ProcessQueuedETLJobs(ctx context.Context, database *db.DB, workerOpts JobWorkerOptions) (ProcessETLQueueResult, error) {
	workerOpts = NormalizeJobWorkerOptions(workerOpts)

	result := ProcessETLQueueResult{WorkerID: workerOpts.WorkerID}

	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		job, err := database.AcquireNextETLJob(ctx, workerOpts.WorkerID, workerOpts.MaxActiveJobs, workerOpts.JobType)
		if err != nil {
			return result, err
		}
		if job == nil {
			return result, nil
		}

		result.Processed++
		echo.Infof("Processing ETL job id=%d type=%s attempt=%d/%d priority=%d", job.ID, job.JobType, job.Attempts, job.MaxRetries+1, job.Priority)

		if err := database.MarkETLJobRunning(ctx, job.ID); err != nil {
			return result, err
		}

		rows, runID, failureClass, retryable, execErr := executeETLJob(ctx, database, job, workerOpts)
		if execErr == nil {
			if err := database.MarkETLJobSucceeded(ctx, job.ID, runID, rows); err != nil {
				return result, err
			}
			result.Succeeded++
			result.Rows += rows
			echo.Successf("✓ ETL job %d succeeded rows=%d", job.ID, rows)
		} else {
			if errors.Is(execErr, context.Canceled) || errors.Is(execErr, context.DeadlineExceeded) {
				if markErr := database.MarkETLJobCancelled(ctx, job.ID, execErr.Error()); markErr != nil {
					return result, markErr
				}
				result.Cancelled++
				return result, execErr
			}

			if retryable && job.Attempts <= job.MaxRetries {
				delay := time.Duration(job.Attempts) * workerOpts.NetworkRetryBackoff
				if delay <= 0 {
					delay = workerOpts.NetworkRetryBackoff
				}
				nextRetryAt := time.Now().Add(delay)
				if err := database.MarkETLJobRetryWait(ctx, job.ID, nextRetryAt, failureClass, execErr.Error()); err != nil {
					return result, err
				}
				result.Retried++
				echo.Infof("  job=%d scheduled retry_at=%s class=%s err=%v", job.ID, nextRetryAt.Format(time.RFC3339), failureClass, execErr)
			} else {
				if err := database.MarkETLJobFailed(ctx, job.ID, failureClass, execErr.Error()); err != nil {
					return result, err
				}
				result.Failed++
				echo.Errorf("✗ ETL job %d failed class=%s err=%v", job.ID, failureClass, execErr)
			}
		}

		if workerOpts.BatchDelay > 0 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(workerOpts.BatchDelay):
			}
		}
	}
}

func RunETLWorker(ctx context.Context, database *db.DB, workerOpts JobWorkerOptions, pollInterval time.Duration) error {
	workerOpts = NormalizeJobWorkerOptions(workerOpts)
	if pollInterval <= 0 {
		pollInterval = defaultETLWorkerPollInterval
	}

	echo.Infof("ETL worker online id=%s max_active=%d poll_interval=%s", workerOpts.WorkerID, workerOpts.MaxActiveJobs, pollInterval)

	for {
		if ctx.Err() != nil {
			return nil
		}

		result, err := ProcessQueuedETLJobs(ctx, database, workerOpts)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}
		if result.Processed > 0 {
			echo.Infof("Worker cycle complete id=%s processed=%d succeeded=%d failed=%d retried=%d cancelled=%d rows=%d",
				result.WorkerID, result.Processed, result.Succeeded, result.Failed, result.Retried, result.Cancelled, result.Rows)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(pollInterval):
		}
	}
}

func executeETLJob(
	ctx context.Context,
	database *db.DB,
	job *db.ETLJob,
	workerOpts JobWorkerOptions,
) (int64, *int64, string, bool, error) {
	switch job.JobType {
	case db.ETLJobTypeMaintenance:
		rows, err := executeMaintenanceJob(ctx, database, job)
		if err != nil {
			return rows, nil, "db_write", false, err
		}
		return rows, nil, "", false, nil
	case db.ETLJobTypeValidate:
		opts, err := pipelineOptionsFromJob(job)
		if err != nil {
			return 0, nil, "db_write", false, err
		}
		validation, err := ValidatePipeline(ctx, database, opts.Profile, opts.Years)
		if err != nil {
			return 0, nil, "db_write", false, err
		}
		if !validation.OK() {
			parts := make([]string, 0, len(validation.Errors()))
			for _, issue := range validation.Errors() {
				parts = append(parts, fmt.Sprintf("[%s] %s", issue.Dataset, issue.Message))
			}
			return int64(len(validation.Issues)), nil, "validation", false, fmt.Errorf("validation failed: %s", strings.Join(parts, "; "))
		}
		return int64(len(validation.Issues)), nil, "", false, nil
	case db.ETLJobTypeCleanup:
		opts, err := pipelineOptionsFromJob(job)
		if err != nil {
			return 0, nil, "db_write", false, err
		}
		cleanupResult, err := CleanupRetrosheetArtifacts(opts.RetrosheetDataDir, false)
		if err != nil {
			return 0, nil, "db_write", false, err
		}
		return int64(len(cleanupResult.Removed)), nil, "", false, nil
	case db.ETLJobTypeCurrentSync:
		rows, err := executeCurrentSeasonSync(ctx, database, job, nil)
		if err != nil {
			failureClass, retryable := classifyCurrentSeasonSyncFailure(err)
			return rows, nil, failureClass, retryable, err
		}
		return rows, nil, "", false, nil
	case db.ETLJobTypeFullRun, db.ETLJobTypeYearlySync:
		opts, err := pipelineOptionsFromJob(job)
		if err != nil {
			return 0, nil, "db_write", false, err
		}
		opts.NetworkRetryMax = workerOpts.NetworkRetryMax
		opts.NetworkRetryBackoff = workerOpts.NetworkRetryBackoff
		if v := intFromAny(job.Options["network_retry_max"]); v >= 0 {
			opts.NetworkRetryMax = v
		}
		if ms := int64FromAny(job.Options["network_retry_backoff_ms"]); ms > 0 {
			opts.NetworkRetryBackoff = time.Duration(ms) * time.Millisecond
		}
		if chunk := intFromAny(job.Options["load_chunk_size"]); chunk > 0 {
			opts.LoadChunkSize = chunk
		}
		result, err := RunPipeline(ctx, database, opts)
		if err != nil {
			failureClass, retryable := classifyPipelineFailure(err)
			return result.TotalRows, nil, failureClass, retryable, err
		}
		runID := result.RunID
		return result.TotalRows, &runID, "", false, nil
	default:
		return 0, nil, "db_write", false, fmt.Errorf("unsupported ETL job type %q", job.JobType)
	}
}

func classifyPipelineFailure(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "cancelled", false
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "validation failed") || strings.Contains(msg, "[coverage]") {
		return "validation", false
	}

	if strings.Contains(msg, "extract.") ||
		strings.Contains(msg, "download") ||
		strings.Contains(msg, "http") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "tls") {
		return "network", true
	}

	return "db_write", false
}

func splitYearsIntoBatches(years []int, batchSize int) [][]int {
	if len(years) == 0 {
		return [][]int{{}}
	}
	if batchSize < 1 || len(years) <= batchSize {
		return [][]int{slices.Clone(years)}
	}
	batches := make([][]int, 0, (len(years)+batchSize-1)/batchSize)
	for start := 0; start < len(years); start += batchSize {
		end := start + batchSize
		if end > len(years) {
			end = len(years)
		}
		batches = append(batches, slices.Clone(years[start:end]))
	}
	return batches
}

func pipelineOptionsFromJob(job *db.ETLJob) (PipelineOptions, error) {
	profile := PipelineProfile(strings.TrimSpace(job.Profile))
	mode := PipelineMode(strings.TrimSpace(job.Mode))
	years, err := intSliceFromAny(job.Scope["years"])
	if err != nil {
		return PipelineOptions{}, fmt.Errorf("invalid years in ETL job %d scope: %w", job.ID, err)
	}
	eras, err := stringSliceFromAny(job.Scope["era_names"])
	if err != nil {
		return PipelineOptions{}, fmt.Errorf("invalid era_names in ETL job %d scope: %w", job.ID, err)
	}
	dataRoot := stringFromAny(job.Scope["data_root"])

	opts, err := NormalizePipelineOptions(PipelineOptions{
		Profile:  profile,
		Mode:     mode,
		Years:    years,
		EraNames: eras,
		DataRoot: dataRoot,
	})
	if err != nil {
		return PipelineOptions{}, err
	}
	return opts, nil
}

func envIntDefault(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envDurationDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func intSliceFromAny(value any) ([]int, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []int:
		return slices.Clone(v), nil
	case []any:
		out := make([]int, 0, len(v))
		for _, item := range v {
			out = append(out, intFromAny(item))
		}
		slices.Sort(out)
		return slices.Compact(out), nil
	default:
		return nil, fmt.Errorf("unsupported type %T", value)
	}
}

func stringSliceFromAny(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []string:
		return slices.Clone(v), nil
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, stringFromAny(item))
		}
		slices.Sort(out)
		return slices.Compact(out), nil
	default:
		return nil, fmt.Errorf("unsupported type %T", value)
	}
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	default:
		return 0
	}
}

func int64FromAny(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}
