# ETL

This runbook is the operational source of truth for ETL development.

## Command signatures and execution model

Use one execution path for all queued work:

- queue producers: `run`, `maintenance`, and `cron`
- queue consumer: `worker` loop (`RunETLWorker` / `ProcessQueuedETLJobs`)

Command contract:

- `baseball-etl worker --max-active-jobs <n> --poll-interval <dur>`
  - long-lived queue consumer; executes all ETL job types
- `baseball-etl run --profile <dev|prod|current-season> --years <scope> [--year-batch-size <n>] [--enqueue-only=<bool>]`
  - enqueue ingestion jobs; default is enqueue-only
- `baseball-etl maintenance --profile <dev|prod> --years <scope> --mv-refresh-mode <auto|non_concurrent> [--enqueue-only=<bool>]`
  - enqueue maintenance jobs; set `--enqueue-only=true` for production-like behavior so execution stays in the main worker loop
- `baseball-etl cron [--schedule <expr>|--schedule <sync_type=expr>] [--schedule-config <path>] [--disable-scheduler]`
  - runs scheduler + worker in one process; scheduler only enqueues `current-season-sync` jobs
- `baseball-etl validate --profile <dev|prod> [--years <scope>]`
  - read-only ETL coverage checks
- `baseball-etl status`
  - queue + freshness visibility

Cron model (implemented for `docs/specs/current.md`):

- `baseball-etl cron` is a scheduler surface that enqueues jobs on cadence.
- Cron does not get a separate execution path; scheduled jobs are still processed by the worker loop.
- `--schedule` is repeatable and supports either `<expr>` (defaults to `sync_type=all`) or `<sync_type=expr>`.
- Supported sync types are `all`, `stats`, `standings`, `schedule`, and `rosters`.
- Cron schedule config is read from `--schedule-config` or falls back to `--config` (default `conf.toml`).
- De-dup guard skips enqueue when a pending (`queued|started|running|retry_wait`) job already exists for the same `profile + sync_type + season`.
- `current_season.active_window` gates enqueue ticks by month/day range (for example `03-20/11-15`).

## Concurrency semantics

- `--max-active-jobs` limits concurrent started/running jobs across the shared queue.
- `--max-queued-jobs` bounds queued+active depth at enqueue time.
- `--year-batch-size` controls how `run` slices year scopes into multiple queue jobs.
- `--poll-interval` controls idle wait between worker dequeue cycles.
- Queue order is `priority ASC`, then `queued_at ASC`, then `id ASC`.
- VM-safe default for shared environments remains `--max-active-jobs=1`.

## What remains before cutover

- Deploy a build containing ETL maintenance migrations through `016_drop_etl_maintenance.sql` and the maintenance worker code path.
- Apply schema migrations in the target environment.
- Run ETL as a dedicated long-lived worker process/container (`baseball-etl worker`).
- Verify queue limits and batching defaults are set for VM safety (`max-active-jobs=1`, bounded queue depth, year batching).
- Run a narrow production slice first, then validate and expand scope.

## Production cutover checklist

1. Apply DB schema updates.

    ```bash
    baseball db migrate
    ```

2. Start API and ETL as separate processes/containers.

    ```bash
    baseball server start
    baseball-etl worker
    ```

3. Prime required source inputs for the first scope.

    ```bash
    baseball-etl fetch retrosheet --years 2024-2025
    baseball-etl fetch negroleagues
    # Optional refresh only; people.csv is committed under data/chadwick
    baseball-etl fetch chadwick --force
    ```

4. Enqueue a narrow slice first.

    ```bash
    baseball-etl run --profile prod --years 2024-2025 --year-batch-size 1
    ```

5. Enqueue maintenance for the same window (worker executes it).

    ```bash
    baseball-etl maintenance --profile prod --years 2024-2025 --mv-refresh-mode auto --enqueue-only
    ```

6. Validate and inspect queue/run health.

    ```bash
    baseball-etl validate --profile prod --years 2024-2025
    baseball-etl status
    ```

7. Expand scope in bounded windows.

    ```bash
    baseball-etl run --profile prod --years 2022-2023 --year-batch-size 1
    baseball-etl maintenance --profile prod --years 2022-2023 --mv-refresh-mode auto --enqueue-only
    baseball-etl run --profile prod --years 2020-2021 --year-batch-size 1
    baseball-etl maintenance --profile prod --years 2020-2021 --mv-refresh-mode auto --enqueue-only
    ```

8. Cleanup transient Retrosheet artifacts.

    ```bash
    baseball-etl cleanup retrosheet
    ```

## Current-Year Backfill (After Historical Load)

Use this flow when historical data is already loaded and you want the current season persisted in `current_season.*` without re-running full historical ETL.

1. Ensure schema changes are applied (adds `current_season.*` tables + `current-season-sync` job type).

    ```bash
    baseball db migrate
    ```

2. Run an on-demand `current-season-sync` job for the current year.

    This syncs the current-season snapshot up to the latest upstream MLB API state (including games through today) without cron.

    - If a worker is already running:

      ```bash
      baseball-etl run --profile current-season --years <current_year>
      ```

    - If no worker is running, run queue + drain in one command:

      ```bash
      baseball-etl run --profile current-season --years <current_year> --enqueue-only=false
      ```

3. Enable recurring sync after the baseline backfill.

    ```toml
    [current_season]
    enabled = true
    season = 2026 # set to the season being backfilled
    cron_stats = "0 */4 * * *"
    cron_standings = "0 * * * *"
    cron_schedule = "0 6 * * *"
    cron_rosters = "0 8 * * *"
    active_window = "03-20/11-15"
    ```

    ```bash
    baseball-etl cron --config conf.toml --profile current-season
    ```

4. Validate backfill + freshness.

    ```bash
    baseball-etl jobs ls --job-type current-season-sync --limit 10
    baseball-etl status
    curl http://localhost:8080/v1/meta/datasets
    curl http://localhost:8080/v1/meta/readiness
    ```

Notes:

- This backfill path is additive and does not require reloading Lahman/Retrosheet history.
- Re-running the same command is safe; current-season writes are upserts.
- Player stat merges exclude duplicate seasons already present in Lahman.
- `baseball-etl current-season handoff` is not implemented yet; off-season cleanup is still a manual truncate after canonical historical load validation.

## Narrow-slice local development (production-like)

Use this flow to replicate production behavior locally without a full-history load:

1. Terminal A: start server.

    ```bash
    ./tmp/baseball server start
    ```

2. Terminal B: start worker.

    ```bash
    ./tmp/baseball-etl worker --max-active-jobs 1 --poll-interval 5s
    ```

3. Fetch only the needed years.

    ```bash
    ./tmp/baseball-etl fetch retrosheet --years 2024-2025
    # Optional refresh only; people.csv is committed under data/chadwick
    ./tmp/baseball-etl fetch chadwick --force
    ```

4. Enqueue a narrow run.

    ```bash
    ./tmp/baseball-etl run --profile dev --years 2024-2025 --year-batch-size 1
    ```

5. Run maintenance for the same window.

    ```bash
    ./tmp/baseball-etl maintenance --profile dev --years 2024-2025 --mv-refresh-mode auto --enqueue-only
    ```

6. Validate + inspect.

    ```bash
    ./tmp/baseball-etl validate --profile dev --years 2024-2025
    ./tmp/baseball-etl status
    ```

## Drift guardrails (implementation vs testing)

- Keep repository query contract stable while maintenance internals evolve.
- Migrations must succeed on both:
  - fresh databases (common test path)
  - long-lived databases (real cutover path)
- When changing ETL maintenance behavior, update:
  - loader/queue tests (`internal/seed`)
  - API integration expectations (`internal/api`) where dataset shape/readiness is affected.
- Treat `baseball-etl validate` + `baseball-etl status` as required acceptance checks after each scoped run.
- Treat `baseball-etl maintenance` as a first-class post-run phase. `run` handles extract/load/validate; maintenance handles canonical materialized-view refresh.
- Keep maintenance execution in the main worker loop (`--enqueue-only`) instead of a separate one-shot drain path.
