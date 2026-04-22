-- 2026/04/21
--
-- current-season cron scheduler verification for ETL queue behavior.
--
-- What this checks:
-- 1) Pending current-season-sync jobs by status/profile/sync_type
-- 2) Recent enqueue/execution history for cron-driven jobs
-- 3) Constraint contract for allowed ETL job types
--
-- Notes:
-- - During smoke validation, duplicate enqueue guard should keep pending count stable
--   when a queued/running current-season-sync job already exists for the same sync_type/season.
-- - Query 2 is useful to confirm scheduler cadence and worker-loop execution order.

-- 1) Pending current-season-sync jobs grouped by status/profile/scope.
SELECT
  status,
  profile,
  COALESCE(scope->>'sync_type', '') AS sync_type,
  COALESCE(scope->>'season', '') AS season,
  COUNT(*) AS jobs
FROM etl_jobs
WHERE job_type = 'current-season-sync'
  AND status IN ('queued', 'started', 'running', 'retry_wait')
GROUP BY status, profile, COALESCE(scope->>'sync_type', ''), COALESCE(scope->>'season', '')
ORDER BY status, profile, sync_type, season;

-- 2) Most recent current-season-sync jobs for cadence/execution sanity.
SELECT
  id,
  status,
  profile,
  COALESCE(scope->>'sync_type', '') AS sync_type,
  COALESCE(scope->>'season', '') AS season,
  queued_at,
  started_at,
  finished_at,
  attempts,
  max_retries,
  last_error
FROM etl_jobs
WHERE job_type = 'current-season-sync'
ORDER BY id DESC
LIMIT 50;

-- 3) Confirm ETL job type constraint includes current-season-sync.
SELECT
  conname,
  pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE conrelid = 'etl_jobs'::regclass
  AND conname = 'chk_etl_jobs_type';
