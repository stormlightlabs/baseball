# Database Loading Contract (Complete Slice)

See also: [ETL Cutover & Narrow-Slice Runbook](./etl.md)

This is the source of truth for loading a complete database slice with an ETL-worker model:

- `baseball-etl` owns fetch/load/validate/status/cleanup operations
- Postgres is the only persistent data store for serving + analytics
- local `data/` holds source inputs and transient ETL artifacts

A complete slice means:

- Lahman + Retrosheet core coverage loaded for the requested year policy
- supplemental datasets loaded (Negro Leagues, FanGraphs constants, salaries, biodata, weather, parks, all-star)
- ETL validation + readiness checks pass

## Command Prefixes

| Environment    | Prefix                                                                                       |
| -------------- | -------------------------------------------------------------------------------------------- |
| Local          | `./tmp/baseball` (db/server), `./tmp/baseball-etl` (etl)                                     |
| Docker/Coolify | `docker compose exec app baseball` (db/server), `docker compose exec etl baseball-etl` (etl) |

Examples below use:

- `<BASEBALL>` for db/server commands
- `<BASEBALL_ETL>` for ETL commands

## ETL command signatures and queue ownership

Execution should flow through one queue consumer (`worker`):

- `<BASEBALL_ETL> worker --max-active-jobs <n> --poll-interval <dur>`
  - long-lived consumer, executes queued ETL jobs
- `<BASEBALL_ETL> run --profile <dev|prod> --years <scope> [--year-batch-size <n>] [--enqueue-only=<bool>]`
  - enqueue ingestion jobs (default enqueue-only)
- `<BASEBALL_ETL> maintenance --profile <dev|prod> --years <scope> --mv-refresh-mode <auto|non_concurrent> [--enqueue-only=<bool>]`
  - enqueue maintenance jobs; use `--enqueue-only=true` for production-like operation so maintenance also runs in the main worker loop

Current-season cron model from `docs/specs/current.md`:

- `baseball-etl cron` is a scheduling surface.
- Cron enqueues jobs and should not introduce a separate execution path.
- Scheduled jobs are still processed by the worker loop.
- Cron should run alongside the worker in the same process/container when enabled.
- Use cron settings/flags to enable or disable scheduled task registration without changing worker-loop ownership of execution.

## Primary Operational Flow

### 1) Ensure data root is present

Use checked-in CSVs under `data/` when available (for example, Lahman CSVs already restored in this repo). For missing Retrosheet files, use ETL fetch commands.

### 2) Apply DB schema migrations

```bash
<BASEBALL> db migrate
```

`db migrate` is structural/idempotent only. Heavy recompute remains an explicit ETL/load step.

### 3) Fetch source data required for your run scope

Representative dev window:

```bash
<BASEBALL_ETL> fetch retrosheet --years 2022-2025
<BASEBALL_ETL> fetch negroleagues
<BASEBALL_ETL> fetch chadwick --force
```

### 4) Start the ETL worker (long-running queue consumer)

```bash
<BASEBALL_ETL> worker
```

Run this in its own terminal/session/container.
For Docker/Coolify, run `docker compose up -d etl` once and keep it running.

### 5) Enqueue ETL ingestion jobs

Representative dev slice:

```bash
<BASEBALL_ETL> run --profile dev --years 2022-2025
```

Full historical profile:

```bash
<BASEBALL_ETL> run --profile prod --mode full
```

For VM-safe operations, prefer batched windows instead of one large full-history run:

```bash
<BASEBALL_ETL> run --profile prod --years 2022-2023
<BASEBALL_ETL> run --profile prod --years 2024-2025
```

`run` is enqueue-first by default. To enqueue and drain in one command (local-only convenience), use `--enqueue-only=false`.

### 6) Enqueue maintenance for the same scope (worker executes it)

```bash
<BASEBALL_ETL> maintenance --profile dev --years 2022-2025 --mv-refresh-mode auto --enqueue-only
```

### 7) Validate and inspect status

```bash
<BASEBALL_ETL> validate --profile dev
<BASEBALL_ETL> status
```

If jobs are queued but not draining, inspect and clear stale running rows:

```bash
<BASEBALL_ETL> jobs ls --status running,started
<BASEBALL_ETL> jobs clear --reason "recover stale running jobs"
```

### 8) Cleanup transient Retrosheet artifacts (optional but recommended)

```bash
<BASEBALL_ETL> cleanup retrosheet --dry-run
<BASEBALL_ETL> cleanup retrosheet
```

### 9) API readiness checks

```bash
curl http://localhost:8080/v1/ready
curl http://localhost:8080/v1/meta/datasets
curl http://localhost:8080/v1/meta/datasets?strict=true
curl http://localhost:8080/v1/health
```

## Narrow Slice Quickstart (Local)

Use this when you want production-like behavior with a bounded local scope.

```bash
./tmp/baseball db migrate
./tmp/baseball-etl fetch retrosheet --years 2024-2025
./tmp/baseball-etl worker --max-active-jobs 1 --poll-interval 5s
./tmp/baseball-etl run --profile dev --years 2024-2025 --year-batch-size 1
./tmp/baseball-etl maintenance --profile dev --years 2024-2025 --mv-refresh-mode auto --enqueue-only
./tmp/baseball-etl validate --profile dev --years 2024-2025
./tmp/baseball-etl status
```

## Data Root Resolution

Resolution order for ETL and DB commands:

1. `--data-root`
2. `BASEBALL_DATA_ROOT`
3. `data`

## Retrosheet Download + Cleanup Contract

Worker expectations:

- ETL fetches missing Retrosheet artifacts for requested years/eras.
- ETL load steps should only depend on files under the resolved data root.
- Temporary extraction artifacts should be cleaned after successful runs.
- Failures should preserve enough artifacts/logs for debugging before cleanup.

Operational guidance:

- Keep canonical source files (`*.zip`, core CSVs like `gameinfo.csv`, `allplayers.csv`) in the data root.
- Keep Chadwick source under `data/chadwick` as committed `people.csv` plus `manifest.json`.
- Use `baseball-etl fetch chadwick --force` when you intentionally refresh `people.csv` from upstream.
- Prune transient ETL artifacts periodically with `baseball-etl cleanup retrosheet` to keep disk usage bounded.

## Stale Data Recovery

If validation fails due to stale/incomplete local source files, recover in this order:

1. Targeted refresh:

   ```bash
   <BASEBALL_ETL> fetch retrosheet --force --years 2022-2025
   <BASEBALL_ETL> load players
   <BASEBALL_ETL> run --profile dev --years 2022-2025
   <BASEBALL_ETL> validate --profile dev --years 2022-2025
   ```

2. If source state is broadly stale in Docker/Coolify:

   ```bash
   docker compose stop app etl
   docker volume ls --format '{{.Name}}' | grep data_root
   docker volume rm <data_root_volume_name>
   docker compose up -d app etl
   ```

3. Re-run ETL + maintenance enqueue + validate for your slice.

## Large Dataset Guidance

Run heavy steps explicitly and in order:

1. `db migrate`
2. `baseball-etl fetch retrosheet ...` (if needed)
3. `baseball-etl run ...`
4. `baseball-etl maintenance ... --enqueue-only`
5. explicit partition maintenance/analysis for heavily changed tables

Queue and batching guidance:

- keep one active ETL run at a time on shared VMs
- keep one active ETL run at a time on shared VMs (`--max-active-jobs=1`)
- keep queue depth bounded (`--max-queued-jobs`, default 128)
- split full-history windows into smaller year batches
- run maintenance once per scoped window (not once per batch)
- enqueue maintenance jobs and let the main worker loop execute them
- defer additional jobs until current batch maintenance + validation passes

Partition/load observability:

- `etl_step_events`
- table/partition row-count + latency checks via ETL status and SQL diagnostics

## Stage Commands

Stage commands are first-class and expected in this worker model:

- `<BASEBALL_ETL> fetch retrosheet`
- `<BASEBALL_ETL> fetch negroleagues`
- `<BASEBALL_ETL> fetch chadwick --force` (explicit refresh)
- `<BASEBALL_ETL> cleanup retrosheet`
- `<BASEBALL_ETL> load <dataset>`
- `<BASEBALL_ETL> maintenance`
- `<BASEBALL_ETL> validate`
- `<BASEBALL_ETL> status`
- `<BASEBALL_ETL> jobs ls`
- `<BASEBALL_ETL> jobs clear`

## Retrosheet Era Contract

Supported `--era` values:

- `fed`
- `nlg`
- `boomer`
- `pitcher`
- `turf`
- `steroid`
- `moneyball`
- `statcast`
- `modern`
