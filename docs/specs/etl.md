---
title: ETL Worker Architecture (No External Warehouse)
updated: 2026-04-20
---

## Problem

The refactor direction must change: we do not want a separate data warehouse or external snapshot-factory workflow to operate the product.

Operationally, that means ETL should be a dedicated worker container responsible for database lifecycle work, including:

- downloading Retrosheet/Lahman source inputs when needed
- loading and validating datasets in Postgres
- running bounded post-load maintenance and recompute steps
- cleaning up temporary/intermediate Retrosheet artifacts after successful jobs

## Architectural Direction

Adopt a single-system ingestion model:

1. `baseball-etl` is the worker runtime (Sidekiq/Celery-style responsibility model).
2. Postgres is the only long-lived store.
3. `data/` is the local source-data workspace, not a warehouse product.

Core decisions:

- No external warehouse contract is required for steady-state operations.
- No snapshot-repo bootstrap/auto-clone path is required for ETL startup.
- ETL owns source acquisition and cleanup lifecycle for Retrosheet.
- ETL remains separate from API runtime (separate container/process), but targets the same operational database.
- ETL jobs must be idempotent, year-bounded where possible, and resumable after interruption.
- Materialized-view-heavy loops remain transitional; bounded table/partition maintenance is preferred.

## Responsibility Split

| Concern               | API Container (`baseball server`)     | ETL Worker Container (`baseball-etl`)             |
| --------------------- | ------------------------------------- | ------------------------------------------------- |
| HTTP traffic          | Owns API routes, auth, cache behavior | None                                              |
| Source acquisition    | None                                  | Owns `etl fetch *` workflows                      |
| DB writes for ingest  | None                                  | Owns load/upsert/reseed operations                |
| Retrosheet lifecycle  | Read-only usage through API queries   | Download, extract/parse, load, cleanup            |
| Validation/readiness  | Serves readiness endpoints            | Produces readiness inputs via ETL/validate/status |
| Operational telemetry | API metrics/logs                      | ETL run/step events + failure metadata            |

## Worker Data Contract

`baseball-etl` consumes source files from a resolved data root:

1. `--data-root`
2. `BASEBALL_DATA_ROOT`
3. `data`

Expected structure remains local-first (example):

- `lahman/csv/*`
- `retrosheet/*.zip` and extracted/generated CSVs during ETL runs
- `retrosheet/negroleagues/*`
- `retrosheet/gameinfo.csv`, `allplayers.csv`, and related side datasets as required
- `chadwick/people-*.csv`, `chadwick/people.csv`, and `chadwick/manifest.json` (worker-fetched from GitHub register shards)

Note: checked-in CSVs under `data/` are valid bootstrap inputs and reduce first-run fetch needs.

## Runtime Topology

```text
[Traefik in prod / Caddy in dev] -> [api container: baseball server start]
                     |
                     v
                [postgres] <-> [redis]
                     ^
                     |
        [etl worker container: baseball-etl run/fetch/load/validate/status]
                     |
                     v
             [local data root volume: data]
```

Deployment rules:

- API and ETL run as separate containers/processes.
- ETL exposes no public HTTP port.
- ETL has independent CPU/memory/DB-pool limits.
- ETL enforces single-active run semantics to avoid overlapping heavy jobs.

## Batched Queue Contract

ETL should execute as a batched job queue, even when manually triggered from CLI:

- queue model: enqueue scope-specific jobs (`years`, `era`, profile) and process serially by default
- concurrency default: `1` active worker per deployment unless explicitly raised
- batch unit: bounded year windows and bounded COPY/upsert chunk sizes
- pacing: optional inter-batch delay/backpressure hooks to reduce WAL/checkpoint spikes
- fairness: long full-history jobs should be split into smaller queued windows

Primary objective: avoid VM saturation while still making forward progress on ingest and maintenance.

Current runtime shape:

- `baseball-etl run` enqueues scoped jobs (enqueue-first behavior).
- `baseball-etl worker` is the long-lived process that polls and executes queued jobs.
- Queue state is persisted in `etl_jobs` with durable statuses (`queued`, `started`, `running`, `retry_wait`, `succeeded`, `failed`, `cancelled`).
- `etl status` surfaces queue state and per-job-type throughput/failure metrics.

## Retrosheet Lifecycle Contract

For each ETL execution window:

1. Resolve requested year/era scope.
2. Ensure required Retrosheet files exist (download missing files).
3. Parse/load/upsert to Postgres in bounded steps.
4. Validate coverage/freshness contracts.
5. Clean temporary artifacts produced by the run (failed runs retain diagnostics where needed).

Cleanup intent:

- Keep canonical source files required for reproducibility.
- Remove transient extraction/output files that only serve in-flight load stages.
- Keep cleanup explicit and auditable in ETL logs/events.
- Use `etl cleanup retrosheet` for repeatable operator-driven pruning.

## Performance and Safety Contract

- Year/season-scoped writes are preferred over full-history rewrites.
- Batched queue execution is preferred over one unbounded monolithic run.
- Heavy steps are cancellable and resumable.
- Post-load `ANALYZE` and backpressure-aware pacing are part of worker behavior.
- ETL telemetry tables (`etl_runs`, `etl_run_steps`, `etl_step_events`) remain source of truth for operations.
- API availability must remain stable during ETL windows.

## Migration Strategy

### Phase A: Doc + Runtime Contract Realignment

- Remove warehouse/snapshot-factory assumptions from docs and runbooks.
- Make ETL worker ownership explicit for fetch/load/cleanup responsibilities.

### Phase B: Worker Job Hardening

- Complete single-active lock/cancellation guarantees.
- Add VM-safe batching controls (job chunk sizing, queue depth, optional pacing between batches).
- Harden Retrosheet download + cleanup behavior for partial/interrupted runs.
- Add explicit maintenance jobs for DB-side recompute/partition hygiene.

### Phase C: Surface Simplification

- Keep `baseball-etl` as canonical data operations surface.
- Keep `db` command group schema/maintenance focused.
- Retire stale config/doc references that imply external warehouse ownership.

### Phase D: Decompose Long MV Refreshes

- Replace monolithic materialized-view refresh maintenance with queue-driven batch jobs.
- Introduce staged intermediaries and serving tables for high-cost derived datasets.
- Keep compatibility relation names during cutover to avoid API regressions.

## Materialized View Decomposition and Batched Maintenance Plan

### Why This Exists

`refresh.materialized_views` currently executes broad materialized-view refreshes in one ETL phase after load.
Even with per-view retries and observability, this can still produce long-running maintenance windows on small VMs.

This plan decomposes heavy refresh work into batched jobs with explicit intermediaries, so each unit of work is smaller, resumable, and queueable.

### Current State (Code-Backed)

- ETL pipeline runs one `refresh.materialized_views` step after data load.
- Retrosheet-heavy refresh set includes:
  - `player_game_batting_stats`
  - `player_game_pitching_stats`
  - `player_game_fielding_stats`
  - `team_game_stats`
  - `win_expectancy_historical`
  - achievement and leader views
- Refresh mode is currently forced non-concurrent for ETL pipeline runs (`ForceNonConcurrent: true`).
- API repositories query MV names directly, so transition must preserve query compatibility.

### Target Pattern

#### 1) Queue-Driven Maintenance

Use ETL job queue semantics for maintenance work:

- one active maintenance worker by default
- explicit job types (`mv.batch.recompute`, `mv.publish`, `mv.compact`)
- bounded batches by year or game-id window
- idempotent jobs and resumable checkpoints

#### 2) Intermediary Layers

For each heavy MV family:

- `*_stage` tables: transient batch output (per season/game window)
- `*_serving` tables: durable incremental store with upsert keys
- compatibility views: keep old names stable while backing data moves to serving tables

#### 3) Scope Tracking

Add scope tables to drive only affected recomputes:

- `etl_changed_games(run_id, game_id, season)`
- `etl_changed_seasons(run_id, season)`
- `etl_changed_players(run_id, player_id, season)`

Populate these from ETL load steps and use them to fan out downstream batch jobs.

### View Family Migration Map

| Current object | Intermediary strategy | Batch key | Target backing |
| --- | --- | --- | --- |
| `player_game_batting_stats` | `stage_player_game_batting_stats` -> upsert | `season`, `game_id` | `serving_player_game_batting_stats` |
| `player_game_pitching_stats` | `stage_player_game_pitching_stats` -> upsert | `season`, `game_id` | `serving_player_game_pitching_stats` |
| `player_game_fielding_stats` | `stage_player_game_fielding_stats` -> upsert | `season`, `game_id` | `serving_player_game_fielding_stats` |
| `team_game_stats` | `stage_team_game_stats` -> upsert | `season`, `game_id` | `serving_team_game_stats` |
| `no_hitters`, `cycles`, `multi_hr_games` | derive from changed games/player-game slices | `season`, `game_id` | `serving_achievement_*` tables |
| `triple_plays`, `extra_inning_games` | derive from changed games only | `season`, `game_id` | `serving_achievement_*` tables |
| `season_batting_leaders` | recompute changed seasons from serving player-game stats | `season` | `serving_season_batting_leaders` |
| `season_pitching_leaders` | recompute changed seasons from serving player-game stats | `season` | `serving_season_pitching_leaders` |
| `career_batting_leaders` | recompute changed players from season serving table | `player_id` | `serving_career_batting_leaders` |
| `career_pitching_leaders` | recompute changed players from season serving table | `player_id` | `serving_career_pitching_leaders` |
| `win_expectancy_historical` | incremental state-count table, publish from counts | `year` or `era` bucket | `serving_win_expectancy_state_counts` + view |
| `player_id_map`, `team_franchise_map`, `park_map` | keep as-is initially; migrate later if needed | dataset-specific | existing MVs first |

### Rollout Waves

#### Wave 0: Queue and Scope Foundations

- add ETL maintenance queue table(s)
- add changed-scope tables (`etl_changed_*`)
- instrument enqueue/dequeue/attempt timing

#### Wave 1: Game-Log Serving Tables (Highest Impact)

- migrate `player_game_*` + `team_game_stats` off full MV refresh
- build in batches by season/game windows
- publish compatibility views using existing object names

#### Wave 2: Achievement Incrementals

- move achievements to serving tables fed by changed game windows
- remove full-refresh requirement from ingest path

#### Wave 3: Season Leaders by Changed Season

- recompute only affected seasons
- publish season leader compatibility views

#### Wave 4: Career Leaders by Changed Player

- derive affected player set from changed seasons
- upsert career totals only for impacted players

#### Wave 5: Win Expectancy Decomposition

- replace broad full recomputation with incremental state-count updates
- keep historical rebuild as low-frequency backfill job

### Example Intermediary DDL (Shape Only)

```sql
CREATE TABLE IF NOT EXISTS etl_maintenance_jobs (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NULL,
    job_type TEXT NOT NULL,
    scope_json JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    attempts INT NOT NULL DEFAULT 0,
    queued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    error TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_etl_maintenance_jobs_status_queued_at
ON etl_maintenance_jobs(status, queued_at);

CREATE TABLE IF NOT EXISTS stage_player_game_batting_stats (
    player_id TEXT NOT NULL,
    game_id TEXT NOT NULL,
    season INT NOT NULL,
    payload JSONB NOT NULL,
    batch_key TEXT NOT NULL
);
```

### Publish/Compatibility Strategy

To avoid API breakage during migration:

1. Create serving tables with complete schemas.
2. Load serving tables in batches.
3. Replace MV usage behind compatibility views with same column contracts.
4. Keep old names (`player_game_batting_stats`, etc.) queryable throughout cutover.

### Operational Guardrails

- one active maintenance worker on shared VM
- hard timeout per batch
- retry per batch with capped attempts
- optional inter-batch delay for WAL/checkpoint pressure
- fail-fast if queue depth exceeds safety threshold

### Exit Criteria

- ETL ingest no longer depends on full-history MV refresh loops
- maintenance jobs complete in bounded batch windows
- interrupted jobs resume from remaining scopes
- API query surfaces continue to work under existing relation names

## Non-Goals (Current Scope)

- No API contract changes under `/v1/*`.
- No requirement to introduce a second persistent analytical store.

## Acceptance Criteria

- ETL worker can bootstrap required source data (including Retrosheet) without an external warehouse.
- ETL worker handles download, load, validate, and cleanup lifecycle in one operational model.
- ETL worker executes through batched queue semantics that keep shared VM usage bounded.
- API and ETL are independently deployable and resource-isolated.
- ETL runs remain observable, cancellable, and single-active.
