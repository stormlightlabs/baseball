# ETL Binary + Container Tasks

Scope: create a separate ETL binary and dedicated ETL container, with ETL-focused performance and safety execution phases.

## Phase 0: Completed Baseline

- [x] Postgres runtime tuning moved to real `postgres -c ...` args in prod compose.
- [x] App DB pool hard caps added (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, lifetime/idle).
- [x] App runtime limits added (`GOMEMLIMIT`, `GOMAXPROCS`) with container limits.
- [x] ETL run observability tables/events added (`etl_runs`, `etl_run_steps`, `etl_step_events`, MV refresh events).
- [x] Force Retrosheet clear path made year-bounded and transaction-scoped per year.
- [x] Crosswalk refresh switched to lock-friendlier clear path (`DELETE` instead of `TRUNCATE`).

Acceptance:

- [x] Current baseline supports safer large ETL runs than previous builds.

## Phase 1: Separate ETL Binary

- [ ] Add `cli/etl/main.go` with ETL-focused root command surface.
- [ ] Build `baseball-etl` alongside `baseball` in Docker multi-stage build.
- [ ] Keep shared orchestration in `internal/seed` (no logic fork).
- [ ] Add smoke tests for `baseball-etl --help`, `run --help`, `validate --help`, `status --help`.
- [ ] Remove ETL command registration from the primary CLI entrypoint (`cmd/baseball/main.go` / `commands.NewBaseballRootCmd` wiring) once `baseball-etl` is the canonical ETL interface.

Acceptance:

- [ ] ETL can run without shipping server/cache command surfaces in its process.
- [ ] Primary `baseball` CLI no longer exposes ETL command surface after cutover.

## Phase 2: Dedicated ETL Container (Dev + Prod Compose)

- [ ] Add `etl` service to `conf/docker-compose.dev.yml`.
- [ ] Add `etl` service to `conf/docker-compose.prod.yml`.
- [ ] Keep same image as `app` unless later split is justified.
- [ ] Configure ETL-specific resource caps (`mem_limit`, `cpus`, `pids_limit`).
- [ ] Configure ETL-specific DB pool env values (separate from API defaults).
- [ ] Mount shared data root volume at `/home/app/tools/data` for `app` + `etl`.
- [ ] Keep ETL service internal-only (no exposed ports).
- [ ] Document one-shot execution path (`docker compose run --rm etl ...`) as default.

Acceptance:

- [ ] ETL can execute in its own container without `docker compose exec app`.
- [ ] API service remains independently deployable and operable during ETL runs.

## Phase 3: Safety Rails

- [ ] Add single-active ETL run guard (advisory lock or `etl_run_locks` table).
- [ ] Add per-step timeout and cancellation policy for heavy ETL phases.
- [ ] Add host-level alerting for free disk and WAL growth thresholds.
- [ ] Add runbook actions for WAL pressure (pause ETL, archive/prune strategy, checkpoint analysis).
- [ ] Add `pg_stat_bgwriter` trend capture (`checkpoints_req` / `checkpoints_timed`).
- [ ] Add off-peak scheduling recommendations and safe defaults for large force/year ranges.
- [ ] Add optional load-shed mode for non-critical endpoints during ETL windows.
- [ ] Add emergency toggles for heavy refresh groups and force-mode suppression.
- [ ] Publish resume/recovery runbook for interrupted ETL runs.

Acceptance:

- [ ] Concurrent ETL starts are rejected safely.
- [ ] Operators have deterministic controls for pause, resume, and rollback.

## Phase 4: Throughput and Refresh Scope

- [x] Keep force clears year-bounded and index-friendly (`date` predicates).
- [x] Keep per-year delete telemetry and transactional boundaries.
- [x] Keep crosswalk refresh lock-friendly (`DELETE` path, no full-table `TRUNCATE`).
- [ ] Add explicit post-load `ANALYZE` for heavily changed tables/partitions.
- [ ] Add DB backpressure throttling hooks (latency/WAL-sensitive pacing).
- [ ] Add indexes/constraints for incremental upsert paths.
- [ ] Replace global refresh default with affected-year/affected-artifact refresh plans.
- [ ] Add ETL perf baselines for range-force runs (runtime, WAL growth, checkpoint frequency).

Acceptance:

- [ ] ETL runtime variance is reduced for year-scoped loads.
- [ ] Full-system refresh is no longer the default for incremental updates.

## Phase 5: Hybrid Incremental Materialization

- [ ] Finalize heavy artifact replacement list (`player_game_*`, `team_game_stats`, `season_*_leaders`, `career_*_leaders`).
- [ ] Define source-of-truth model per artifact (`MV`, incremental table, or mixed).
- [ ] Add `etl_watermarks` / `materialization_state` for resumable progress.
- [ ] Define invalidation keys (season/year/team/player) and retry-safe transaction boundaries.
- [ ] Define cutover SLOs (max refresh time, max lock time, acceptable staleness).
- [ ] Add structural migrations for incremental target tables (no historical rewrite).
- [ ] Implement year/season-bounded recompute + upsert steps.
- [ ] Make force/year runs recompute only affected years/seasons.
- [ ] Add phase-level ETL events and row-count metrics per artifact update step.
- [ ] Add retry-safe transactional boundaries and resumability markers.
- [ ] Add runbook queries for stale watermarks and slow phases.
- [ ] Compare ETL runtime baseline before/after cutover.
- [ ] Document hybrid strategy updates in ETL and deployment docs.
- [ ] Keep migration sets structural and idempotent only (`IF NOT EXISTS`, guarded DDL).

Acceptance:

- [ ] Materialization work scales with changed data, not full history.
- [ ] Interrupted runs can resume without full rerun.

## Verification Checklist (Before Marking Complete)

- [ ] `go test ./...`
- [ ] ETL lock/concurrency behavior validated with two concurrent start attempts.
- [ ] `docker compose` ETL one-shot run succeeds in dev and prod compose layouts.
- [ ] API readiness remains healthy during ETL run window.
- [ ] ETL docs updated in `docs/internal/data-loading.md`.
- [ ] ETL docs updated in `conf/README.md`.
