# Database Loading Contract (Complete Slice)

This is the source of truth for loading a complete database slice with an ETL-worker model:

- `baseball-etl` owns fetch/load/validate/status/cleanup operations
- Postgres is the only persistent data store for serving + analytics
- local `data/` holds source inputs and transient ETL artifacts

A complete slice means:

- Lahman + Retrosheet core coverage loaded for the requested year policy
- supplemental datasets loaded (Negro Leagues, FanGraphs constants, salaries, biodata, weather, parks, all-star)
- ETL validation + readiness checks pass

## Command Prefixes

| Environment | Prefix |
| --- | --- |
| Local | `./tmp/baseball` (db/server), `./tmp/baseball-etl` (etl) |
| Docker/Coolify | `docker compose exec app baseball` (db/server), `docker compose exec app baseball-etl` (etl) |

Examples below use:

- `<BASEBALL>` for db/server commands
- `<BASEBALL_ETL>` for ETL commands

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
```

### 4) Run ETL ingestion

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

### 5) Validate and inspect status

```bash
<BASEBALL_ETL> validate --profile dev
<BASEBALL_ETL> status
```

### 6) API readiness checks

```bash
curl http://localhost:8080/v1/ready
curl http://localhost:8080/v1/meta/datasets
curl http://localhost:8080/v1/meta/datasets?strict=true
curl http://localhost:8080/v1/health
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
- Prune transient ETL artifacts periodically to keep disk usage bounded.

## Large Dataset Guidance

Run heavy steps explicitly and in order:

1. `db migrate`
2. `baseball-etl fetch retrosheet ...` (if needed)
3. `baseball-etl run ...`
4. explicit partition maintenance/analysis for heavily changed tables

Queue and batching guidance:

- keep one active ETL run at a time on shared VMs
- split full-history windows into smaller year batches
- defer additional jobs until current batch validation passes

Partition/load observability:

- `etl_step_events`
- table/partition row-count + latency checks via ETL status and SQL diagnostics

## Stage Commands

Stage commands are first-class and expected in this worker model:

- `<BASEBALL_ETL> fetch retrosheet`
- `<BASEBALL_ETL> fetch negroleagues`
- `<BASEBALL_ETL> load <dataset>`
- `<BASEBALL_ETL> validate`
- `<BASEBALL_ETL> status`

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
