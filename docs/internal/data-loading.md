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
- `<BASEBALL_ETL> run --profile <dev|prod|current-season> --years <scope> [--year-batch-size <n>] [--enqueue-only=<bool>]`
  - enqueue ingestion jobs (default enqueue-only)
- `<BASEBALL_ETL> maintenance --profile <dev|prod> --years <scope> --mv-refresh-mode <auto|non_concurrent> [--enqueue-only=<bool>]`
  - enqueue maintenance jobs; use `--enqueue-only=true` for production-like operation so maintenance also runs in the main worker loop
- `<BASEBALL_ETL> cron [--schedule <expr>|--schedule <sync_type=expr>] [--schedule-config <path>] [--disable-scheduler]`
  - scheduler+worker process for `current-season-sync` queue jobs

Current-season cron model from `docs/specs/current.md`:

- `baseball-etl cron` is a scheduling surface.
- Cron enqueues jobs and should not introduce a separate execution path.
- Scheduled jobs are still processed by the worker loop.
- Cron should run alongside the worker in the same process/container when enabled.
- `--schedule` supports `<expr>` (defaults `sync_type=all`) or `<sync_type=expr>`.
- Supported sync types: `all`, `stats`, `standings`, `schedule`, `rosters`.
- Queue de-dup skips new ticks while a matching `profile + sync_type + season` job is pending.

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

### 10) Backfill current year after historical load (no full reload)

Use this when Lahman/Retrosheet history is already loaded and you only need the current-season bridge tables populated.

1. Ensure migration `015_current_season.sql` is applied:

   ```bash
   <BASEBALL> db migrate
   ```

2. Run an on-demand `current-season-sync` job:

   This refreshes current-season tables up to the latest upstream MLB API state (including games through today) without requiring cron.

   - If worker is already running:

     ```bash
     <BASEBALL_ETL> run --profile current-season --years <current_year>
     ```

   - If worker is not running:

     ```bash
     <BASEBALL_ETL> run --profile current-season --years <current_year> --enqueue-only=false
     ```

3. Enable recurring refresh with cron (recommended after baseline seed):

   Add to your config file (for example `conf.toml`):

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

   Start scheduler+worker:

   ```bash
   <BASEBALL_ETL> cron --config conf.toml --profile current-season
   ```

4. Verify current-season backfill:

   ```bash
   <BASEBALL_ETL> jobs ls --job-type current-season-sync --limit 10
   curl http://localhost:8080/v1/meta/datasets
   curl http://localhost:8080/v1/meta/readiness
   ```

Notes:

- This path does not require re-running full historical ETL.
- Re-running the same command is safe; current-season tables are updated via upserts.
- API merges avoid duplicate season rows when Lahman already has that season.
- Off-season handoff is still manual; no `baseball-etl current-season handoff` subcommand yet.

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
   <BASEBALL_ETL> load retrosheet players
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
- keep `--max-active-jobs=1` on shared VMs
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
