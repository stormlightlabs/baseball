# ETL Cutover & Narrow-Slice Runbook

This runbook is the operational source of truth for ETL cutover and day-to-day narrow-slice development.

## What remains before cutover

- Deploy a build containing migration `014_etl_mv_batched_maintenance.sql` and the maintenance worker code path.
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

5. Run maintenance for the same window.

    ```bash
    baseball-etl maintenance --profile prod --years 2024-2025 --mv-refresh-mode auto
    ```

6. Validate and inspect queue/run health.

    ```bash
    baseball-etl validate --profile prod --years 2024-2025
    baseball-etl status
    ```

7. Expand scope in bounded windows.

    ```bash
    baseball-etl run --profile prod --years 2022-2023 --year-batch-size 1
    baseball-etl maintenance --profile prod --years 2022-2023 --mv-refresh-mode auto
    baseball-etl run --profile prod --years 2020-2021 --year-batch-size 1
    baseball-etl maintenance --profile prod --years 2020-2021 --mv-refresh-mode auto
    ```

8. Cleanup transient Retrosheet artifacts.

    ```bash
    baseball-etl cleanup retrosheet
    ```

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
    ./tmp/baseball-etl maintenance --profile dev --years 2024-2025 --mv-refresh-mode auto
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
- Treat `baseball-etl maintenance` as a first-class post-run phase. `run` handles extract/load/validate; maintenance handles materialized-view recompute + serving sync.
