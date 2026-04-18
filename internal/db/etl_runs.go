package db

import (
	"context"
	"encoding/json"
	"fmt"
)

// StartETLRun creates an ETL run record and returns its ID.
func (db *DB) StartETLRun(ctx context.Context, profile, mode string, params map[string]any) (int64, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal ETL params: %w", err)
	}

	var runID int64
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO etl_runs (profile, mode, params, status) VALUES ($1, $2, $3, 'running') RETURNING id`,
		profile,
		mode,
		paramsJSON,
	).Scan(&runID); err != nil {
		return 0, fmt.Errorf("failed to insert ETL run: %w", err)
	}

	return runID, nil
}

// FinishETLRun marks an ETL run as completed or failed.
func (db *DB) FinishETLRun(ctx context.Context, runID int64, status, errMsg string) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE etl_runs
		SET status = $2,
			finished_at = NOW(),
			duration_ms = (EXTRACT(EPOCH FROM (NOW() - started_at)) * 1000)::BIGINT,
			error = NULLIF($3, '')
		WHERE id = $1
	`, runID, status, errMsg); err != nil {
		return fmt.Errorf("failed to finalize ETL run %d: %w", runID, err)
	}

	return nil
}

// StartETLStep creates a step record for a run and returns its ID.
func (db *DB) StartETLStep(ctx context.Context, runID int64, step string) (int64, error) {
	var stepID int64
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO etl_run_steps (run_id, step, status) VALUES ($1, $2, 'running') RETURNING id`,
		runID,
		step,
	).Scan(&stepID); err != nil {
		return 0, fmt.Errorf("failed to insert ETL run step: %w", err)
	}

	return stepID, nil
}

// FinishETLStep marks a run step as completed or failed.
func (db *DB) FinishETLStep(ctx context.Context, stepID int64, status string, rowCount int64, errMsg string) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE etl_run_steps
		SET status = $2,
			row_count = $3,
			finished_at = NOW(),
			duration_ms = (EXTRACT(EPOCH FROM (NOW() - started_at)) * 1000)::BIGINT,
			error = NULLIF($4, '')
		WHERE id = $1
	`, stepID, status, rowCount, errMsg); err != nil {
		return fmt.Errorf("failed to finalize ETL run step %d: %w", stepID, err)
	}

	return nil
}
