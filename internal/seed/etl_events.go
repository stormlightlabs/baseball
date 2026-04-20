package seed

import (
	"context"
	"time"

	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

type etlStepContextKey struct{}

type etlStepContextValue struct {
	runID int64
	step  string
}

func withETLStepContext(ctx context.Context, runID int64, step string) context.Context {
	if runID <= 0 || step == "" {
		return ctx
	}
	return context.WithValue(ctx, etlStepContextKey{}, etlStepContextValue{
		runID: runID,
		step:  step,
	})
}

func etlStepContextValues(ctx context.Context) (*int64, string) {
	v, ok := ctx.Value(etlStepContextKey{}).(etlStepContextValue)
	if !ok || v.runID <= 0 || v.step == "" {
		return nil, ""
	}
	runID := v.runID
	return &runID, v.step
}

func recordETLPhaseEvent(
	ctx context.Context,
	database *db.DB,
	defaultStep string,
	phase string,
	status string,
	rowCount int64,
	startedAt time.Time,
	metadata map[string]any,
	eventErr error,
) {
	if database == nil || phase == "" {
		return
	}

	runID, step := etlStepContextValues(ctx)
	if step == "" {
		step = defaultStep
	}
	if step == "" {
		return
	}
	if status == "" {
		status = "completed"
	}

	finishedAt := time.Now()
	etlEvent := db.ETLStepEvent{
		RunID:      runID,
		Step:       step,
		Phase:      phase,
		Status:     status,
		RowCount:   rowCount,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Metadata:   metadata,
	}
	if eventErr != nil {
		etlEvent.Error = eventErr.Error()
	}

	if err := database.RecordETLStepEvent(ctx, etlEvent); err != nil {
		echo.Infof("  ⚠ Failed to record ETL phase event step=%s phase=%s: %v", step, phase, err)
	}
}
