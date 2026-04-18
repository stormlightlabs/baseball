CREATE TABLE IF NOT EXISTS etl_runs (
    id BIGSERIAL PRIMARY KEY,
    profile TEXT NOT NULL,
    mode TEXT NOT NULL,
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'running',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    duration_ms BIGINT,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_etl_runs_status ON etl_runs(status);
CREATE INDEX IF NOT EXISTS idx_etl_runs_started_at ON etl_runs(started_at DESC);

CREATE TABLE IF NOT EXISTS etl_run_steps (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL REFERENCES etl_runs(id) ON DELETE CASCADE,
    step TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    row_count BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    duration_ms BIGINT,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_etl_run_steps_run_id ON etl_run_steps(run_id, id);
CREATE INDEX IF NOT EXISTS idx_etl_run_steps_status ON etl_run_steps(status);
