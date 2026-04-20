# ETL Worker Tasks (No External Warehouse)

Scope: keep ETL as an operational worker runtime for the production database. The worker owns fetch/load/validate/status/cleanup flows, including Retrosheet download lifecycle.

## Phase 0: Completed Baseline

- [x] Postgres runtime tuning moved to real `postgres -c ...` args in prod compose.
- [x] App DB pool hard caps added (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, lifetime/idle).
- [x] App runtime limits added (`GOMEMLIMIT`, `GOMAXPROCS`) with container limits.
- [x] ETL run observability tables/events added (`etl_runs`, `etl_run_steps`, `etl_step_events`, MV refresh events).
- [x] Force Retrosheet clear path made year-bounded and transaction-scoped per year.
- [x] Crosswalk refresh switched to lock-friendlier clear path (`DELETE` instead of `TRUNCATE`).

Acceptance:

- [x] Current baseline supports safer large ETL runs than previous builds.

## Phase 1: Runtime Surface Split (In Progress)

- [x] Add `cmd/baseball-etl/main.go` with ETL-focused root command surface.
- [x] Build `baseball-etl` alongside `baseball` in Docker multi-stage build.
- [x] Keep shared orchestration in `internal/seed` (no logic fork).
- [x] Consolidate command implementation into `internal/cli` and wire both binaries through that package.
- [x] Remove deprecated `tools/data` git submodule from this repo.
- [ ] Add smoke tests for `baseball-etl --help`, `run --help`, `validate --help`, `status --help`.
- [ ] Remove ETL command registration from `internal/cli.NewBaseballRootCmd` once cutover is complete.

Acceptance:

- [x] ETL can run without shipping server/cache command surfaces in its process.
- [ ] Primary `baseball` CLI no longer exposes ETL commands after cutover.

## Phase 2: Worker-Owned Data Lifecycle

- [ ] Treat local `data/` as canonical ETL input root.
- [ ] Ensure ETL can bootstrap missing Retrosheet files via `etl fetch retrosheet` as part of normal operations.
- [ ] Document and harden Retrosheet file retention policy (what stays vs what is temporary).
- [ ] Add explicit cleanup command path for transient Retrosheet artifacts after successful loads.
- [ ] Vendor a pinned Chadwick Register snapshot under `data/` and document ETL update cadence for that snapshot.
- [ ] Remove stale docs/config assumptions that require an external warehouse or snapshot repo.

Acceptance:

- [ ] ETL can execute full load windows without dependency on external warehouse publication flow.
- [ ] Operators have explicit, repeatable download + cleanup behavior for Retrosheet data.

## Phase 3: Job-Oriented ETL Worker Behavior

- [ ] Define ETL job types (`full-run`, `yearly-sync`, `validate-only`, `cleanup-only`, `maintenance`).
- [ ] Add durable run-state transitions for start/running/succeeded/failed/cancelled.
- [ ] Add job metadata for scope (`years`, `era`, profile/mode) and replayability.
- [ ] Add clear retry policy for network/download failures separate from DB write failures.
- [ ] Track throughput and failure class metrics per job type.
- [ ] Add queue controls for VM safety (max active jobs, max queued jobs, job-priority policy).
- [ ] Add batch controls (year-window chunking, load chunk sizes, optional inter-batch delay).

Acceptance:

- [ ] ETL behaves like a queue worker surface even when invoked manually.
- [ ] Failed jobs can be resumed or retried with explicit scope.
- [ ] Default queue/batch settings prevent host saturation on the production VM.

## Phase 4: Dedicated ETL Container (Dev + Prod Compose)

- [ ] Add `etl` service to `conf/docker-compose.dev.yml`.
- [ ] Add `etl` service to `conf/docker-compose.prod.yml`.
- [ ] Keep same image as `app` unless later split is justified.
- [ ] Configure ETL-specific resource caps (`mem_limit`, `cpus`, `pids_limit`).
- [ ] Configure ETL-specific DB pool env values (separate from API defaults).
- [ ] Mount shared data root volume at `/home/app/data` for `app` + `etl`.
- [ ] Keep ETL service internal-only (no exposed ports).
- [ ] Document one-shot execution path (`docker compose run --rm etl ...`) as default.

Acceptance:

- [ ] ETL can execute in its own container without `docker compose exec app`.
- [ ] API service remains independently operable during ETL runs.

## Phase 5: Safety Rails + Throughput

- [ ] Add single-active ETL run guard (advisory lock or `etl_run_locks` table).
- [ ] Add per-step timeout and cancellation policy for heavy ETL phases.
- [ ] Add explicit post-load `ANALYZE` for heavily changed tables/partitions.
- [ ] Add DB backpressure throttling hooks (latency/WAL-sensitive pacing).
- [ ] Add admission-control guardrails that reject or defer new ETL jobs when host pressure is high.
- [ ] Add host-level alerting for free disk and WAL growth thresholds.
- [ ] Add runbook actions for WAL pressure (pause ETL, archive/prune strategy, checkpoint analysis).
- [ ] Add `pg_stat_bgwriter` trend capture (`checkpoints_req` / `checkpoints_timed`).

Acceptance:

- [ ] Concurrent ETL starts are rejected safely.
- [ ] Operators have deterministic controls for pause/resume/recovery.

## Phase 6: Database Maintenance Jobs

- [ ] Mark broad materialized-view refresh loops as legacy and keep them off default ETL run path.
- [ ] Define explicit maintenance jobs for partition/table hygiene and targeted recompute.
- [ ] Add partition management policy (create/attach/retire) for high-churn datasets.
- [ ] Remove API/query dependencies on MV-only objects over time.
- [ ] Replace MV-centric observability/reporting with table/partition load metrics and checks.
- [ ] Add runbook queries for partition skew, stale partitions, and slow upsert phases.

Acceptance:

- [ ] ETL maintenance scales by changed data scope instead of full-history refreshes.
- [ ] Interrupted runs can resume without full rerun.

## Phase 7: Materialized View Decomposition (Batched)

Reference: Materialized View Decomposition and Batched Maintenance Plan section in [ETL spec](../specs/etl.md).

- [ ] Add queue-backed maintenance job model for MV replacement work (`queued`, `running`, `succeeded`, `failed`).
- [ ] Add changed-scope tracking tables (`etl_changed_games`, `etl_changed_seasons`, `etl_changed_players`).
- [ ] Migrate `player_game_*` and `team_game_stats` to serving tables with staged batched upserts.
- [ ] Migrate achievement views (`no_hitters`, `cycles`, `multi_hr_games`, `triple_plays`, `extra_inning_games`) to changed-game batch recompute.
- [ ] Migrate season leader views to changed-season recompute.
- [ ] Migrate career leader views to changed-player recompute.
- [ ] Decompose `win_expectancy_historical` into incremental state-count maintenance plus publish step.
- [ ] Add compatibility views so repository SQL can keep existing relation names during rollout.

Acceptance:

- [ ] Default ETL flow avoids single long-running MV refresh phases.
- [ ] Maintenance jobs run in bounded batches and can resume after interruption.
- [ ] API behavior remains stable during and after cutover.

## Verification Checklist

- [x] `go test ./...`
- [ ] ETL lock/concurrency behavior validated with two concurrent start attempts.
- [ ] `docker compose` ETL one-shot run succeeds in dev and prod layouts.
- [ ] API readiness remains healthy during ETL run window.
- [ ] Retrosheet fetch/cleanup behavior validated across repeated runs.
