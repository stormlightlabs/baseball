-- fine-grained ETL step phase observability.

CREATE TABLE IF NOT EXISTS etl_step_events (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT REFERENCES etl_runs(id) ON DELETE SET NULL,
    step TEXT NOT NULL,
    phase TEXT NOT NULL,
    status TEXT NOT NULL,
    row_count BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    duration_ms BIGINT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_etl_step_events_run_started
ON etl_step_events(run_id, started_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_etl_step_events_step_started
ON etl_step_events(step, started_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_etl_step_events_status_started
ON etl_step_events(status, started_at DESC, id DESC);
