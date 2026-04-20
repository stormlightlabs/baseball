---
title: ETL Binary + Container Architecture
updated: 2026-04-20
---

## Problem

The current system runs API and ETL from the same `baseball` CLI/runtime. In practice, ETL is executed inside the `app` container (`docker compose exec app baseball etl ...`), so API traffic and ETL contend for:

- CPU and memory
- DB pool capacity (`DB_MAX_OPEN_CONNS` is shared)
- Postgres write/read headroom during large loads and view refreshes

This is the main operational risk for performance-sensitive runs.

## Current System Reality

- ETL orchestration already exists in `internal/seed/pipeline.go` with tracked runs in `etl_runs`, `etl_run_steps`, and `etl_step_events`.
- Bulk ingest is already `COPY`-based (`games_temp`, `plays_temp`, etc.), not row-by-row inserts.
- `plays` is already date-partitioned by migration; `games` is not partitioned.
- Force reload path is year-bounded and transaction-scoped per year (already hardened).
- Materialized views are still globally refreshed in ETL pipeline steps (non-concurrent by default in ETL).
- Compose has no dedicated ETL service today (`app`, `postgres`, `redis`, optional `caddy` only).

## Design Decisions

| Recommendation                                   | Current system                               | Decision                                                          |
| ------------------------------------------------ | -------------------------------------------- | ----------------------------------------------------------------- |
| Separate ETL worker from API runtime             | Not isolated yet                             | Adopt now: dedicated ETL binary + container                       |
| Use `COPY` + staging                             | Already true via temp staging per load       | Keep and standardize across ETL stages                            |
| Idempotent run metadata/checkpoints              | `etl_runs` + step events exist               | Extend with lock + watermarks for resumability                    |
| Partition heavy fact tables                      | `plays` partitioned, `games` non-partitioned | Keep `plays`; evaluate `games` partitioning behind migration gate |
| Incremental serving refresh (not full every run) | Full refresh groups currently                | Move to affected-year/affected-artifact refresh plan              |

## Target Architecture

```text
[Traefik in prod / Caddy in dev] -> [api container: baseball server start]
                     |
                     v
                [postgres] <-> [redis]
                     ^
                     |
       [etl container: baseball-etl run/validate/status]
```

Key rules:

- API and ETL run in separate containers.
- Both may use the same Docker image artifact.
- ETL has its own resource limits and DB pool limits.
- ETL container exposes no HTTP port.
- ETL runs are single-active via DB lock guard.

## Binary Split Plan

Introduce a dedicated ETL executable while preserving shared packages:

- New entrypoint: `cli/etl/main.go`
- Binary name: `baseball-etl`
- Command surface: `run`, `validate`, `status`, `fetch`, `load`
- Keep existing `baseball` binary unchanged for API/server workflows

This is a runtime split, not a logic rewrite. `internal/seed` remains the orchestration core.

## Container Plan (Same Image, Separate Service)

Add `etl` service in `conf/docker-compose.dev.yml` and `conf/docker-compose.prod.yml`:

- image: same as `app` (`baseball-app:latest`)
- command: ETL binary entrypoint (`baseball-etl ...`)
- depends_on: `postgres` (not `redis`, unless future ETL stage requires it)
- no ports exposed
- dedicated limits (`mem_limit`, `cpus`, `pids_limit`)
- dedicated env for ETL DB pool (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, lifetime/idle)
- shared data root mount (`/home/app/tools/data`) between `app` and `etl`

Operational mode:

- Default: one-shot ETL jobs (`docker compose run --rm etl ...`)
- Optional later: scheduled ETL worker/cron trigger

Ingress note:

- Production ingress is Traefik via Coolify.
- Caddy is development-only (`docker-compose.dev.yml`).

## Performance and Safety Requirements

- Single active ETL run lock (DB advisory lock or `etl_run_locks` table)
- Bounded stage timeouts with cancellation propagation
- Backpressure controls when DB latency/WAL pressure rises
- Year-scoped recompute for heavy artifacts; avoid full global refresh on every run
- Post-load `ANALYZE` on heavily changed tables/partitions
- Maintain phase-level telemetry (`etl_step_events`, `materialized_view_refresh_events`)

## Database Workstreams

### Operational safety and crash prevention

- Host-level alerting for low disk and WAL growth thresholds
- Runbook actions for WAL pressure (pause ETL, archive/prune strategy, checkpoint analysis)
- Periodic `pg_stat_bgwriter` capture for `checkpoints_req` and `checkpoints_timed`
- ETL concurrency guard (single active ETL run lock)
- Per-step timeout and cancellation policy for heavy operations
- Off-peak scheduling recommendations and safe defaults for large force/year ranges
- Optional load-shed mode for non-critical endpoints during ETL windows
- Documented emergency toggles (disable heavy refresh groups, pause force mode)
- Operational resume/recovery checklist after interruption

### Force/year write-path performance

- Keep force-clear Retrosheet deletes year-bounded and date-index-friendly for `plays` and `games`
- Emit per-year delete telemetry (rows and duration by table)
- Keep per-year transactional boundaries for resumability
- Keep crosswalk refresh lock-friendly (`DELETE` path, avoid full-table `TRUNCATE`)
- Add indexes/constraints for incremental upsert paths
- Maintain ETL performance baselines for range-force runs (runtime, WAL growth, checkpoint frequency)

### Hybrid incremental materialization and cutover

- Finalize heavy artifact list for replacement (`player_game_*`, `team_game_stats`, `season_*_leaders`, `career_*_leaders`)
- Define source-of-truth model per artifact (`MV`, incremental table, or mixed)
- Define incremental keys/invalidation units (season/year/team/player)
- Define cutover SLOs (max refresh time, max lock time, acceptable staleness)
- Add structural migrations for incremental target tables (no historical rewrite)
- Add `etl_watermarks` / `materialization_state` for resumable progress tracking
- Replace full-refresh ETL steps with year/season-bounded recompute + upsert steps
- Make force/year runs recompute only affected years/seasons
- Add phase-level ETL events and row-count metrics per artifact update step
- Add retry-safe transactional boundaries and resumability markers
- Add runbook queries for slow phases and stale watermarks
- Compare ETL runtime baseline before/after cutover
- Keep migration sets structural and idempotent only (`IF NOT EXISTS`, guarded DDL)

## Non-Goals (For This Plan)

- No immediate full schema rewrite to `raw/core/serving` schemas
- No immediate replacement of all materialized views with denormalized serving tables
- No API contract changes

## Acceptance Criteria

- API remains available while ETL runs in separate container.
- ETL cannot starve API DB connections by configuration.
- ETL runs are observable, cancellable, and single-active.
- Heavy refresh work is scoped to affected data where feasible.
