-- materialized view refresh observability.
-- Stores per-view refresh attempts for ETL and manual operations.

CREATE TABLE IF NOT EXISTS materialized_view_refresh_events (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT REFERENCES etl_runs(id) ON DELETE SET NULL,
    step TEXT,
    view_group TEXT,
    view_name TEXT NOT NULL,
    pass INTEGER NOT NULL,
    attempt INTEGER NOT NULL,
    mode TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ NOT NULL,
    duration_ms BIGINT NOT NULL,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_mv_refresh_events_run ON materialized_view_refresh_events(run_id, id);
CREATE INDEX IF NOT EXISTS idx_mv_refresh_events_view_name ON materialized_view_refresh_events(view_name, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_mv_refresh_events_status ON materialized_view_refresh_events(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_mv_refresh_events_started_at ON materialized_view_refresh_events(started_at DESC);
