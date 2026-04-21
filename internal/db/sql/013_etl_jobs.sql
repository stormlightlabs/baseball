-- queue-backed ETL job orchestration and durable worker states.

CREATE TABLE IF NOT EXISTS etl_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_type TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 50,
    status TEXT NOT NULL DEFAULT 'queued',
    profile TEXT NOT NULL,
    mode TEXT NOT NULL,
    scope JSONB NOT NULL DEFAULT '{}'::jsonb,
    options JSONB NOT NULL DEFAULT '{}'::jsonb,
    max_retries INTEGER NOT NULL DEFAULT 2,
    attempts INTEGER NOT NULL DEFAULT 0,
    failure_class TEXT,
    last_error TEXT,
    row_count BIGINT NOT NULL DEFAULT 0,
    run_id BIGINT REFERENCES etl_runs(id) ON DELETE SET NULL,
    queued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    worker_id TEXT,
    CONSTRAINT chk_etl_jobs_type CHECK (job_type IN ('full-run', 'yearly-sync', 'validate-only', 'cleanup-only', 'maintenance')),
    CONSTRAINT chk_etl_jobs_status CHECK (status IN ('queued', 'started', 'running', 'retry_wait', 'succeeded', 'failed', 'cancelled')),
    CONSTRAINT chk_etl_jobs_retries_nonnegative CHECK (max_retries >= 0),
    CONSTRAINT chk_etl_jobs_attempts_nonnegative CHECK (attempts >= 0)
);

CREATE INDEX IF NOT EXISTS idx_etl_jobs_status_priority_queued
ON etl_jobs(status, priority ASC, queued_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_etl_jobs_status_next_retry
ON etl_jobs(status, next_retry_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_etl_jobs_worker_status
ON etl_jobs(worker_id, status, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_etl_jobs_type_status
ON etl_jobs(job_type, status, queued_at DESC);
