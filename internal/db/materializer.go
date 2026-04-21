package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// MaterializedViewRefreshOptions controls observability for view refresh operations.
type MaterializedViewRefreshOptions struct {
	RunID              *int64
	Step               string
	Group              string
	SlowThreshold      time.Duration
	ForceNonConcurrent bool
	OnAttempt          func(MaterializedViewRefreshAttempt)
}

// MaterializedViewRefreshAttempt describes one refresh attempt for a single materialized view.
type MaterializedViewRefreshAttempt struct {
	ViewName   string
	Pass       int
	Attempt    int
	Mode       string
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Error      string
}

type materializedViewSessionTuning struct {
	workMem                     string
	maintenanceWorkMem          string
	maxParallelWorkersPerGather string
	jit                         string
}

type materializedViewCapabilities struct {
	populated          bool
	supportsConcurrent bool
}

// Materializer holds per-run state for materialized view refresh execution.
type Materializer struct {
	opts               MaterializedViewRefreshOptions
	capabilitiesByView map[string]materializedViewCapabilities
	attemptByView      map[string]int
	conn               *sql.Conn
}

func NewMaterializer(opts MaterializedViewRefreshOptions) *Materializer {
	return &Materializer{
		opts:               opts,
		capabilitiesByView: map[string]materializedViewCapabilities{},
		attemptByView:      map[string]int{},
	}
}

func (m *Materializer) withOptions(opts MaterializedViewRefreshOptions) *Materializer {
	return &Materializer{
		opts:               opts,
		capabilitiesByView: map[string]materializedViewCapabilities{},
		attemptByView:      map[string]int{},
	}
}

// Refresh executes materialized view refreshes for the provided views.
func (m *Materializer) Refresh(ctx context.Context, database *DB, viewNames []string) (int, error) {
	views, err := m.resolveViews(ctx, database, viewNames)
	if err != nil {
		return 0, err
	}

	for _, view := range views {
		caps, err := m.getMaterializedViewCapabilities(ctx, database, view)
		if err != nil {
			return 0, fmt.Errorf("failed to inspect materialized view %s: %w", view, err)
		}
		m.capabilitiesByView[view] = caps
	}

	if err := m.openSession(ctx, database); err != nil {
		return 0, err
	}
	defer m.closeSession()

	pending := append([]string(nil), views...)
	refreshed := 0
	maxPasses := len(views)

	for pass := 0; pass < maxPasses && len(pending) > 0; pass++ {
		nextPending := make([]string, 0)
		progressed := false

		for _, view := range pending {
			status, err := m.refreshOne(ctx, database, view, pass+1)
			if err != nil {
				return refreshed, fmt.Errorf("failed to refresh view %s: %w", view, err)
			}
			if status == "deferred_dependency" {
				nextPending = append(nextPending, view)
				continue
			}

			refreshed++
			progressed = true
		}

		if len(nextPending) == 0 {
			return refreshed, nil
		}
		if !progressed {
			return refreshed, fmt.Errorf("failed to resolve dependent materialized views: %v", nextPending)
		}

		pending = nextPending
	}

	if len(pending) > 0 {
		return refreshed, fmt.Errorf("failed to refresh all materialized views; remaining: %v", pending)
	}

	return refreshed, nil
}

func (m *Materializer) resolveViews(ctx context.Context, database *DB, viewNames []string) ([]string, error) {
	if len(viewNames) > 0 {
		return viewNames, nil
	}

	rows, err := database.QueryContext(ctx, `
		SELECT schemaname || '.' || matviewname as viewname
		FROM pg_matviews
		WHERE schemaname = 'public'
		ORDER BY matviewname
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list materialized views: %w", err)
	}
	defer rows.Close()

	views := make([]string, 0)
	for rows.Next() {
		var viewName string
		if err := rows.Scan(&viewName); err != nil {
			return nil, fmt.Errorf("failed to scan view name: %w", err)
		}
		views = append(views, viewName)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate views: %w", err)
	}
	return views, nil
}

func (m *Materializer) openSession(ctx context.Context, database *DB) error {
	conn, err := database.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire refresh connection: %w", err)
	}
	m.conn = conn

	if err := configureMVRefreshSession(ctx, conn, mvRefreshSessionTuningFromEnv()); err != nil {
		_ = conn.Close()
		m.conn = nil
		return err
	}
	return nil
}

func (m *Materializer) closeSession() {
	if m.conn == nil {
		return
	}
	_, _ = m.conn.ExecContext(context.Background(), "RESET ALL")
	_ = m.conn.Close()
	m.conn = nil
}

func (m *Materializer) recordAttempt(
	ctx context.Context,
	database *DB,
	view string,
	pass int,
	mode string,
	startedAt time.Time,
	status string,
	attemptErr error,
) {
	m.attemptByView[view]++

	errMsg := ""
	if attemptErr != nil {
		errMsg = attemptErr.Error()
	}

	event := MaterializedViewRefreshAttempt{
		ViewName:   view,
		Pass:       pass,
		Attempt:    m.attemptByView[view],
		Mode:       mode,
		Status:     status,
		StartedAt:  startedAt,
		FinishedAt: time.Now(),
		Error:      errMsg,
	}
	event.Duration = event.FinishedAt.Sub(event.StartedAt)

	if recErr := m.recordMaterializedViewRefreshEvent(ctx, database, event); recErr != nil {
		// Observability should never block the refresh operation.
	}

	if m.opts.OnAttempt != nil {
		m.opts.OnAttempt(event)
	}
}

func (m *Materializer) refreshOne(ctx context.Context, database *DB, view string, pass int) (string, error) {
	caps, ok := m.capabilitiesByView[view]
	if !ok {
		var err error
		caps, err = m.getMaterializedViewCapabilities(ctx, database, view)
		if err != nil {
			return "failed", err
		}
		m.capabilitiesByView[view] = caps
	}

	mode := "non_concurrent"
	if !m.opts.ForceNonConcurrent && caps.populated && caps.supportsConcurrent {
		mode = "concurrent"
	}

	runRefresh := func(refreshMode string) error {
		query := fmt.Sprintf("REFRESH MATERIALIZED VIEW %s", view)
		if refreshMode == "concurrent" {
			query = fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", view)
		}
		_, err := m.conn.ExecContext(ctx, query)
		return err
	}

	startedAt := time.Now()
	if err := runRefresh(mode); err == nil {
		m.recordAttempt(ctx, database, view, pass, mode, startedAt, "completed", nil)
		caps.populated = true
		m.capabilitiesByView[view] = caps
		return "completed", nil
	} else {
		if isMaterializedViewDeferredDependencyError(err) {
			m.recordAttempt(ctx, database, view, pass, mode, startedAt, "deferred_dependency", err)
			return "deferred_dependency", nil
		}

		m.recordAttempt(ctx, database, view, pass, mode, startedAt, "failed", err)

		if mode == "concurrent" && shouldRetryNonConcurrent(err) {
			fallbackMode := "non_concurrent"
			fallbackStartedAt := time.Now()
			if fallbackErr := runRefresh(fallbackMode); fallbackErr == nil {
				m.recordAttempt(ctx, database, view, pass, fallbackMode, fallbackStartedAt, "completed", nil)
				caps.populated = true
				m.capabilitiesByView[view] = caps
				return "completed", nil
			} else if isMaterializedViewDeferredDependencyError(fallbackErr) {
				m.recordAttempt(ctx, database, view, pass, fallbackMode, fallbackStartedAt, "deferred_dependency", fallbackErr)
				return "deferred_dependency", nil
			} else {
				m.recordAttempt(ctx, database, view, pass, fallbackMode, fallbackStartedAt, "failed", fallbackErr)
				return "failed", fallbackErr
			}
		}

		return "failed", err
	}
}

// RefreshMaterializedViews refreshes one or more materialized views.
// If viewNames is empty, refreshes all materialized views in the database.
func (db *DB) RefreshMaterializedViews(ctx context.Context, viewNames []string) (int, error) {
	return db.RefreshMaterializedViewsWithOptions(ctx, viewNames, MaterializedViewRefreshOptions{})
}

// RefreshMaterializedViewsWithOptions refreshes one or more materialized views with observability metadata.
// If viewNames is empty, refreshes all materialized views in the database.
func (db *DB) RefreshMaterializedViewsWithOptions(ctx context.Context, viewNames []string, opts MaterializedViewRefreshOptions) (int, error) {
	if db.materializer == nil {
		return NewMaterializer(opts).Refresh(ctx, db, viewNames)
	}
	return db.materializer.withOptions(opts).Refresh(ctx, db, viewNames)
}

func mvRefreshSessionTuningFromEnv() materializedViewSessionTuning {
	tuning := materializedViewSessionTuning{
		workMem:                     strings.TrimSpace(os.Getenv("ETL_MV_WORK_MEM")),
		maintenanceWorkMem:          strings.TrimSpace(os.Getenv("ETL_MV_MAINTENANCE_WORK_MEM")),
		maxParallelWorkersPerGather: strings.TrimSpace(os.Getenv("ETL_MV_MAX_PARALLEL_WORKERS_PER_GATHER")),
		jit:                         strings.TrimSpace(os.Getenv("ETL_MV_JIT")),
	}
	if tuning.workMem == "" {
		tuning.workMem = "16MB"
	}
	if tuning.maintenanceWorkMem == "" {
		tuning.maintenanceWorkMem = "64MB"
	}
	if tuning.maxParallelWorkersPerGather == "" {
		tuning.maxParallelWorkersPerGather = "0"
	}
	if tuning.jit == "" {
		tuning.jit = "off"
	}
	return tuning
}

func configureMVRefreshSession(ctx context.Context, conn *sql.Conn, tuning materializedViewSessionTuning) error {
	if err := setSessionConfig(ctx, conn, "work_mem", tuning.workMem); err != nil {
		return fmt.Errorf("failed to set MV refresh work_mem=%q: %w", tuning.workMem, err)
	}
	if err := setSessionConfig(ctx, conn, "maintenance_work_mem", tuning.maintenanceWorkMem); err != nil {
		return fmt.Errorf("failed to set MV refresh maintenance_work_mem=%q: %w", tuning.maintenanceWorkMem, err)
	}
	if err := setSessionConfig(ctx, conn, "max_parallel_workers_per_gather", tuning.maxParallelWorkersPerGather); err != nil {
		return fmt.Errorf("failed to set MV refresh max_parallel_workers_per_gather=%q: %w", tuning.maxParallelWorkersPerGather, err)
	}
	if err := setSessionConfig(ctx, conn, "jit", tuning.jit); err != nil {
		return fmt.Errorf("failed to set MV refresh jit=%q: %w", tuning.jit, err)
	}
	return nil
}

func setSessionConfig(ctx context.Context, conn *sql.Conn, key, value string) error {
	_, err := conn.ExecContext(ctx, `SELECT set_config($1, $2, false)`, key, value)
	return err
}

func isMaterializedViewDeferredDependencyError(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "has not been populated")
}

func shouldRetryNonConcurrent(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "concurrently cannot be used") ||
		strings.Contains(errText, "cannot refresh materialized view") ||
		strings.Contains(errText, "sqlstate 0a000") ||
		strings.Contains(errText, "sqlstate 55000")
}

func normalizeMaterializedViewName(view string) (schema, name string) {
	trimmed := strings.TrimSpace(view)
	if trimmed == "" {
		return "public", ""
	}
	parts := strings.SplitN(trimmed, ".", 2)
	if len(parts) == 1 {
		return "public", strings.Trim(parts[0], `"`)
	}
	return strings.Trim(parts[0], `"`), strings.Trim(parts[1], `"`)
}

func (m *Materializer) getMaterializedViewCapabilities(ctx context.Context, database *DB, view string) (materializedViewCapabilities, error) {
	schema, name := normalizeMaterializedViewName(view)
	if name == "" {
		return materializedViewCapabilities{}, fmt.Errorf("invalid materialized view name %q", view)
	}

	var caps materializedViewCapabilities
	if err := database.QueryRowContext(
		ctx,
		`SELECT ispopulated FROM pg_matviews WHERE schemaname = $1 AND matviewname = $2`,
		schema,
		name,
	).Scan(&caps.populated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return materializedViewCapabilities{}, fmt.Errorf("materialized view %s.%s does not exist", schema, name)
		}
		return materializedViewCapabilities{}, err
	}

	if err := database.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			JOIN pg_index i ON i.indrelid = c.oid
			WHERE n.nspname = $1
			  AND c.relname = $2
			  AND i.indisunique
			  AND i.indisvalid
			  AND i.indpred IS NULL
			  AND i.indexprs IS NULL
		)`,
		schema,
		name,
	).Scan(&caps.supportsConcurrent); err != nil {
		return materializedViewCapabilities{}, err
	}

	return caps, nil
}

func (m *Materializer) recordMaterializedViewRefreshEvent(ctx context.Context, database *DB, attempt MaterializedViewRefreshAttempt) error {
	var runID any
	if m.opts.RunID != nil {
		runID = *m.opts.RunID
	}

	var step any
	if strings.TrimSpace(m.opts.Step) != "" {
		step = m.opts.Step
	}

	var group any
	if strings.TrimSpace(m.opts.Group) != "" {
		group = m.opts.Group
	}

	var errText any
	if attempt.Error != "" {
		errText = attempt.Error
	}

	_, err := database.ExecContext(ctx, `
		INSERT INTO materialized_view_refresh_events (
			run_id,
			step,
			view_group,
			view_name,
			pass,
			attempt,
			mode,
			status,
			started_at,
			finished_at,
			duration_ms,
			error
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		runID,
		step,
		group,
		attempt.ViewName,
		attempt.Pass,
		attempt.Attempt,
		attempt.Mode,
		attempt.Status,
		attempt.StartedAt,
		attempt.FinishedAt,
		attempt.Duration.Milliseconds(),
		errText,
	)
	if err != nil {
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "materialized_view_refresh_events") &&
			(strings.Contains(errText, "does not exist") || strings.Contains(errText, "undefined_table")) {
			return nil
		}
		return fmt.Errorf("failed to insert materialized view refresh event: %w", err)
	}

	return nil
}
