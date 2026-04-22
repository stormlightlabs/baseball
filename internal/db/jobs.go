package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/utils"
)

type ETLJobType string

const (
	ETLJobTypeFullRun     ETLJobType = "full-run"
	ETLJobTypeYearlySync  ETLJobType = "yearly-sync"
	ETLJobTypeValidate    ETLJobType = "validate-only"
	ETLJobTypeCleanup     ETLJobType = "cleanup-only"
	ETLJobTypeMaintenance ETLJobType = "maintenance"
	ETLJobTypeCurrentSync ETLJobType = "current-season-sync"
)

type ETLJobStatus string

const (
	ETLJobStatusQueued    ETLJobStatus = "queued"
	ETLJobStatusStarted   ETLJobStatus = "started"
	ETLJobStatusRunning   ETLJobStatus = "running"
	ETLJobStatusRetryWait ETLJobStatus = "retry_wait"
	ETLJobStatusSucceeded ETLJobStatus = "succeeded"
	ETLJobStatusFailed    ETLJobStatus = "failed"
	ETLJobStatusCancelled ETLJobStatus = "cancelled"
)

var ErrETLQueueFull = errors.New("etl queue is full")

type ETLJobSpec struct {
	JobType    ETLJobType
	Priority   int
	Profile    string
	Mode       string
	Scope      map[string]any
	Options    map[string]any
	MaxRetries int
}

type ETLJob struct {
	ID           int64
	JobType      ETLJobType
	Priority     int
	Status       ETLJobStatus
	Profile      string
	Mode         string
	Scope        map[string]any
	Options      map[string]any
	MaxRetries   int
	Attempts     int
	FailureClass string
	LastError    string
	RowCount     int64
	RunID        *int64
	QueuedAt     time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
	NextRetryAt  *time.Time
	WorkerID     string
}

type ETLJobQueueSnapshot struct {
	Queued    int
	Started   int
	Running   int
	RetryWait int
	Succeeded int
	Failed    int
	Cancelled int
}

type ETLJobTypeMetrics struct {
	JobType            ETLJobType
	SucceededJobs      int
	FailedJobs         int
	CancelledJobs      int
	RetryWaitJobs      int
	RowsProcessed      int64
	NetworkFailures    int
	DBWriteFailures    int
	ValidationFailures int
}

type ETLJobListFilter struct {
	Statuses []ETLJobStatus
	JobType  ETLJobType
	Profile  string
	Limit    int
}

func (db *DB) EnqueueETLJob(ctx context.Context, spec ETLJobSpec, maxQueuedJobs int) (int64, error) {
	if spec.Priority < 0 {
		spec.Priority = 0
	}
	if spec.MaxRetries < 0 {
		spec.MaxRetries = 0
	}

	scopeJSON, err := json.Marshal(utils.NonNilMap(spec.Scope))
	if err != nil {
		return 0, fmt.Errorf("failed to marshal ETL job scope: %w", err)
	}
	optionsJSON, err := json.Marshal(utils.NonNilMap(spec.Options))
	if err != nil {
		return 0, fmt.Errorf("failed to marshal ETL job options: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin ETL enqueue transaction: %w", err)
	}
	defer tx.Rollback()

	if maxQueuedJobs > 0 {
		var pending int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM etl_jobs
			WHERE status IN ('queued', 'started', 'running', 'retry_wait')
		`).Scan(&pending); err != nil {
			return 0, fmt.Errorf("failed to inspect ETL queue depth: %w", err)
		}
		if pending >= maxQueuedJobs {
			return 0, fmt.Errorf("%w: active_or_queued=%d limit=%d", ErrETLQueueFull, pending, maxQueuedJobs)
		}
	}

	var jobID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO etl_jobs (
			job_type,
			priority,
			status,
			profile,
			mode,
			scope,
			options,
			max_retries
		) VALUES ($1, $2, 'queued', $3, $4, $5, $6, $7)
		RETURNING id
	`,
		spec.JobType,
		spec.Priority,
		spec.Profile,
		spec.Mode,
		scopeJSON,
		optionsJSON,
		spec.MaxRetries,
	).Scan(&jobID); err != nil {
		return 0, fmt.Errorf("failed to enqueue ETL job: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit ETL enqueue transaction: %w", err)
	}
	return jobID, nil
}

func (db *DB) AcquireNextETLJob(ctx context.Context, workerID string, maxActiveJobs int, jobType ETLJobType) (*ETLJob, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin ETL dequeue transaction: %w", err)
	}
	defer tx.Rollback()

	if maxActiveJobs > 0 {
		var active int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM etl_jobs
			WHERE status IN ('started', 'running')
		`).Scan(&active); err != nil {
			return nil, fmt.Errorf("failed to inspect active ETL jobs: %w", err)
		}
		if active >= maxActiveJobs {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to finalize ETL dequeue transaction: %w", err)
			}
			return nil, nil
		}
	}

	row := tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id
			FROM etl_jobs
			WHERE status IN ('queued', 'retry_wait')
			  AND (NULLIF($2, '') IS NULL OR job_type = $2)
			  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
			ORDER BY priority ASC, queued_at ASC, id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		), updated AS (
			UPDATE etl_jobs j
			SET status = 'started',
				attempts = j.attempts + 1,
				worker_id = NULLIF($1, ''),
				started_at = NOW(),
				finished_at = NULL,
				next_retry_at = NULL,
				last_error = NULL,
				failure_class = NULL,
				row_count = 0,
				run_id = NULL
			WHERE j.id = (SELECT id FROM candidate)
			RETURNING j.id,
				j.job_type,
				j.priority,
				j.status,
				j.profile,
				j.mode,
				j.scope,
				j.options,
				j.max_retries,
				j.attempts,
				j.failure_class,
				j.last_error,
				j.row_count,
				j.run_id,
				j.queued_at,
				j.started_at,
				j.finished_at,
				j.next_retry_at,
				j.worker_id
		)
			SELECT * FROM updated
		`, workerID, jobType)

	job, err := scanETLJob(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to finalize ETL dequeue transaction: %w", err)
			}
			return nil, nil
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit ETL dequeue transaction: %w", err)
	}
	return job, nil
}

func (db *DB) ListETLJobs(ctx context.Context, filter ETLJobListFilter) ([]ETLJob, error) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 6)
	argIdx := 1

	if len(filter.Statuses) > 0 {
		statusArgs := make([]string, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			statusArgs = append(statusArgs, fmt.Sprintf("$%d", argIdx))
			args = append(args, status)
			argIdx++
		}
		clauses = append(clauses, fmt.Sprintf("j.status IN (%s)", strings.Join(statusArgs, ", ")))
	}
	if filter.JobType != "" {
		clauses = append(clauses, fmt.Sprintf("j.job_type = $%d", argIdx))
		args = append(args, filter.JobType)
		argIdx++
	}
	if profile := strings.TrimSpace(filter.Profile); profile != "" {
		clauses = append(clauses, fmt.Sprintf("j.profile = $%d", argIdx))
		args = append(args, profile)
		argIdx++
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	query := `
		SELECT
			j.id,
			j.job_type,
			j.priority,
			j.status,
			j.profile,
			j.mode,
			j.scope,
			j.options,
			j.max_retries,
			j.attempts,
			j.failure_class,
			j.last_error,
			j.row_count,
			j.run_id,
			j.queued_at,
			j.started_at,
			j.finished_at,
			j.next_retry_at,
			j.worker_id
		FROM etl_jobs j
	`
	if len(clauses) > 0 {
		query += "WHERE " + strings.Join(clauses, " AND ") + "\n"
	}
	query += fmt.Sprintf("ORDER BY j.id DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ETL jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]ETLJob, 0, limit)
	for rows.Next() {
		job, scanErr := scanETLJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		jobs = append(jobs, *job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating ETL jobs: %w", err)
	}
	return jobs, nil
}

func (db *DB) HasPendingETLJob(ctx context.Context, jobType ETLJobType, profile, syncType, season string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM etl_jobs
			WHERE job_type = $1
			  AND status IN ('queued', 'started', 'running', 'retry_wait')
			  AND (NULLIF($2, '') IS NULL OR profile = $2)
			  AND (NULLIF($3, '') IS NULL OR COALESCE(scope->>'sync_type', '') = $3)
			  AND (NULLIF($4, '') IS NULL OR COALESCE(scope->>'season', '') = $4)
		)
	`, jobType, strings.TrimSpace(profile), strings.TrimSpace(syncType), strings.TrimSpace(season)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to inspect pending ETL jobs: %w", err)
	}
	return exists, nil
}

func (db *DB) MarkETLJobRunning(ctx context.Context, jobID int64) error {
	_, err := db.ExecContext(ctx, `
		UPDATE etl_jobs
		SET status = 'running'
		WHERE id = $1
	`, jobID)
	if err != nil {
		return fmt.Errorf("failed to mark ETL job running: %w", err)
	}
	return nil
}

func (db *DB) MarkETLJobSucceeded(ctx context.Context, jobID int64, runID *int64, rowCount int64) error {
	_, err := db.ExecContext(ctx, `
		UPDATE etl_jobs
		SET status = 'succeeded',
			row_count = $2,
			run_id = $3,
			finished_at = NOW(),
			last_error = NULL,
			failure_class = NULL,
			next_retry_at = NULL
		WHERE id = $1
	`, jobID, rowCount, runID)
	if err != nil {
		return fmt.Errorf("failed to mark ETL job succeeded: %w", err)
	}
	return nil
}

func (db *DB) MarkETLJobRetryWait(ctx context.Context, jobID int64, nextRetryAt time.Time, failureClass, errMsg string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE etl_jobs
		SET status = 'retry_wait',
			finished_at = NOW(),
			next_retry_at = $2,
			last_error = NULLIF($3, ''),
			failure_class = NULLIF($4, '')
		WHERE id = $1
	`, jobID, nextRetryAt, errMsg, failureClass)
	if err != nil {
		return fmt.Errorf("failed to mark ETL job retry_wait: %w", err)
	}
	return nil
}

func (db *DB) MarkETLJobFailed(ctx context.Context, jobID int64, failureClass, errMsg string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE etl_jobs
		SET status = 'failed',
			finished_at = NOW(),
			last_error = NULLIF($2, ''),
			failure_class = NULLIF($3, ''),
			next_retry_at = NULL
		WHERE id = $1
	`, jobID, errMsg, failureClass)
	if err != nil {
		return fmt.Errorf("failed to mark ETL job failed: %w", err)
	}
	return nil
}

func (db *DB) MarkETLJobCancelled(ctx context.Context, jobID int64, reason string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE etl_jobs
		SET status = 'cancelled',
			finished_at = NOW(),
			last_error = NULLIF($2, ''),
			next_retry_at = NULL
		WHERE id = $1
	`, jobID, reason)
	if err != nil {
		return fmt.Errorf("failed to mark ETL job cancelled: %w", err)
	}
	return nil
}

func (db *DB) ClearRunningETLJobs(ctx context.Context, reason string) (int64, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "cleared by operator"
	}

	result, err := db.ExecContext(ctx, `
		UPDATE etl_jobs
		SET status = 'retry_wait',
			finished_at = NOW(),
			next_retry_at = NOW(),
			last_error = $1,
			failure_class = 'operator_clear',
			worker_id = NULL
		WHERE status = 'running'
	`, reason)
	if err != nil {
		return 0, fmt.Errorf("failed to clear running ETL jobs: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed reading cleared ETL job count: %w", err)
	}
	return rows, nil
}

func (db *DB) ETLJobQueueSnapshot(ctx context.Context) (ETLJobQueueSnapshot, error) {
	var snapshot ETLJobQueueSnapshot
	err := db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'queued') AS queued,
			COUNT(*) FILTER (WHERE status = 'started') AS started,
			COUNT(*) FILTER (WHERE status = 'running') AS running,
			COUNT(*) FILTER (WHERE status = 'retry_wait') AS retry_wait,
			COUNT(*) FILTER (WHERE status = 'succeeded') AS succeeded,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled
		FROM etl_jobs
	`).Scan(
		&snapshot.Queued,
		&snapshot.Started,
		&snapshot.Running,
		&snapshot.RetryWait,
		&snapshot.Succeeded,
		&snapshot.Failed,
		&snapshot.Cancelled,
	)
	if err != nil {
		return ETLJobQueueSnapshot{}, fmt.Errorf("failed to read ETL job queue snapshot: %w", err)
	}
	return snapshot, nil
}

func (db *DB) ETLJobMetricsByType(ctx context.Context) ([]ETLJobTypeMetrics, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			job_type,
			COUNT(*) FILTER (WHERE status = 'succeeded') AS succeeded_jobs,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed_jobs,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_jobs,
			COUNT(*) FILTER (WHERE status = 'retry_wait') AS retry_wait_jobs,
			COALESCE(SUM(row_count) FILTER (WHERE status = 'succeeded'), 0) AS rows_processed,
			COUNT(*) FILTER (WHERE failure_class = 'network') AS network_failures,
			COUNT(*) FILTER (WHERE failure_class = 'db_write') AS db_write_failures,
			COUNT(*) FILTER (WHERE failure_class = 'validation') AS validation_failures
		FROM etl_jobs
		GROUP BY job_type
		ORDER BY job_type ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query ETL job metrics by type: %w", err)
	}
	defer rows.Close()

	metrics := make([]ETLJobTypeMetrics, 0)
	for rows.Next() {
		var m ETLJobTypeMetrics
		if err := rows.Scan(
			&m.JobType,
			&m.SucceededJobs,
			&m.FailedJobs,
			&m.CancelledJobs,
			&m.RetryWaitJobs,
			&m.RowsProcessed,
			&m.NetworkFailures,
			&m.DBWriteFailures,
			&m.ValidationFailures,
		); err != nil {
			return nil, fmt.Errorf("failed to scan ETL job metrics: %w", err)
		}
		metrics = append(metrics, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating ETL job metrics rows: %w", err)
	}

	return metrics, nil
}

type etlJobScanner interface {
	Scan(dest ...any) error
}

func scanETLJob(scanner etlJobScanner) (*ETLJob, error) {
	var (
		job       ETLJob
		scopeJSON []byte
		optsJSON  []byte
		runID     sql.NullInt64
		failure   sql.NullString
		lastError sql.NullString
		startedAt sql.NullTime
		finished  sql.NullTime
		nextRetry sql.NullTime
		workerID  sql.NullString
	)
	if err := scanner.Scan(
		&job.ID,
		&job.JobType,
		&job.Priority,
		&job.Status,
		&job.Profile,
		&job.Mode,
		&scopeJSON,
		&optsJSON,
		&job.MaxRetries,
		&job.Attempts,
		&failure,
		&lastError,
		&job.RowCount,
		&runID,
		&job.QueuedAt,
		&startedAt,
		&finished,
		&nextRetry,
		&workerID,
	); err != nil {
		return nil, err
	}

	job.Scope = map[string]any{}
	if len(scopeJSON) > 0 {
		if err := json.Unmarshal(scopeJSON, &job.Scope); err != nil {
			return nil, fmt.Errorf("failed to decode ETL job scope: %w", err)
		}
	}

	job.Options = map[string]any{}
	if len(optsJSON) > 0 {
		if err := json.Unmarshal(optsJSON, &job.Options); err != nil {
			return nil, fmt.Errorf("failed to decode ETL job options: %w", err)
		}
	}

	if runID.Valid {
		v := runID.Int64
		job.RunID = &v
	}
	if startedAt.Valid {
		v := startedAt.Time
		job.StartedAt = &v
	}
	if finished.Valid {
		v := finished.Time
		job.FinishedAt = &v
	}
	if nextRetry.Valid {
		v := nextRetry.Time
		job.NextRetryAt = &v
	}
	if workerID.Valid {
		job.WorkerID = workerID.String
	}
	if failure.Valid {
		job.FailureClass = failure.String
	}
	if lastError.Valid {
		job.LastError = lastError.String
	}

	return &job, nil
}
