package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ETLStepEvent represents a fine-grained phase event within an ETL step.
type ETLStepEvent struct {
	RunID      *int64
	Step       string
	Phase      string
	Status     string
	RowCount   int64
	StartedAt  time.Time
	FinishedAt time.Time
	Metadata   map[string]any
	Error      string
}

// RecordETLStepEvent inserts one ETL phase-level event.
// This method is intentionally best-effort and returns nil if the events table
// is not present (for compatibility with older schemas).
func (db *DB) RecordETLStepEvent(ctx context.Context, event ETLStepEvent) error {
	if strings.TrimSpace(event.Step) == "" {
		return nil
	}
	if strings.TrimSpace(event.Phase) == "" {
		event.Phase = "unknown"
	}
	if strings.TrimSpace(event.Status) == "" {
		event.Status = "completed"
	}

	startedAt := event.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}

	finishedAt := event.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = time.Now()
	}
	if finishedAt.Before(startedAt) {
		finishedAt = startedAt
	}
	durationMs := finishedAt.Sub(startedAt).Milliseconds()

	metadata := event.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal ETL step event metadata: %w", err)
	}

	var runID any
	if event.RunID != nil {
		runID = *event.RunID
	}
	var errText any
	if event.Error != "" {
		errText = event.Error
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO etl_step_events (
			run_id,
			step,
			phase,
			status,
			row_count,
			started_at,
			finished_at,
			duration_ms,
			metadata,
			error
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		runID,
		event.Step,
		event.Phase,
		event.Status,
		event.RowCount,
		startedAt,
		finishedAt,
		durationMs,
		metadataJSON,
		errText,
	)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "etl_step_events") &&
			(strings.Contains(msg, "does not exist") || strings.Contains(msg, "undefined_table")) {
			return nil
		}
		return fmt.Errorf("failed to insert ETL step event: %w", err)
	}

	return nil
}
