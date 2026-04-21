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
- [x] Remove ETL command registration from `internal/cli.NewBaseballRootCmd` once cutover is complete.

Acceptance:

- [x] ETL can run without shipping server/cache command surfaces in its process.
- [x] Primary `baseball` CLI no longer exposes ETL commands after cutover.

## Phase 2: Worker-Owned Data Lifecycle

- [x] Treat local `data/` as canonical ETL input root.
- [x] Ensure ETL can bootstrap missing Retrosheet files via `etl fetch retrosheet` as part of normal operations.
- [x] Document and harden Retrosheet file retention policy (what stays vs what is temporary).
- [x] Add explicit cleanup command path for transient Retrosheet artifacts after successful loads.
- [x] Fetch required Chadwick register shards directly from GitHub and persist `data/chadwick/manifest.json`.
- [x] Remove stale docs/config assumptions that require an external warehouse or snapshot repo.

Acceptance:

- [x] ETL can execute full load windows without dependency on external warehouse publication flow.
- [x] Operators have explicit, repeatable download + cleanup behavior for Retrosheet data.

## Phase 3: Job-Oriented ETL Worker Behavior

- [x] Define ETL job types (`full-run`, `yearly-sync`, `validate-only`, `cleanup-only`, `maintenance`).
- [x] Add durable run-state transitions for start/running/succeeded/failed/cancelled.
- [x] Add job metadata for scope (`years`, `era`, profile/mode) and replayability.
- [x] Add clear retry policy for network/download failures separate from DB write failures.
- [x] Track throughput and failure class metrics per job type.
- [x] Add queue controls for VM safety (max active jobs, max queued jobs, job-priority policy).
- [x] Add batch controls (year-window chunking, load chunk sizes, optional inter-batch delay).

Acceptance:

- [x] ETL behaves like a queue worker surface even when invoked manually.
- [x] Failed jobs can be resumed or retried with explicit scope.
- [x] Default queue/batch settings prevent host saturation on the production VM.

## Phase 4: Dedicated ETL Container (Dev + Prod Compose)

- [x] Add `etl` service to `conf/docker-compose.dev.yml`.
- [x] Add `etl` service to `conf/docker-compose.prod.yml`.
- [x] Keep same image as `app` unless later split is justified.
- [x] Configure ETL-specific resource caps (`mem_limit`, `cpus`, `pids_limit`).
- [x] Configure ETL-specific DB pool env values (separate from API defaults).
- [x] Mount shared data root volume at `/home/app/data` for `app` + `etl`.
- [x] Keep ETL service internal-only (no exposed ports).
- [x] Document long-lived worker execution path (`docker compose up -d etl`) as default.

Acceptance:

- [x] ETL can execute in its own container without `docker compose exec app`.
- [x] API service remains independently operable during ETL runs.

## Phase 5: Materialized View Decomposition (Batched, `etl-mv-batching` in ETL) (Active)

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

## Phase 6: CLI Baseline and Contract Safety (Merged from Backend Refactor Tasks)

- [x] Capture baseline command contracts for `baseball-etl --help`.
- [x] Capture baseline command contracts for `baseball-etl run --help`.
- [x] Capture baseline command contracts for `baseball-etl validate --help`.
- [x] Capture baseline command contracts for `baseball-etl status --help`.
- [x] Capture baseline command contracts for `baseball db --help`.
- [x] Capture baseline command contracts for `baseball server --help`.
- [x] Add focused tests for the golden path behavior (`db migrate`, `server start`, `baseball-etl worker`, `baseball-etl run`, `baseball-etl validate`, `baseball-etl status`).
- [x] Record current ETL/status command contracts in unit tests.

Acceptance:

- [x] CLI regressions are detected early during refactor iterations.

## Phase 7: Canonical CLI Flow and Command Consolidation

- [x] Keep docs/help text consistent with `baseball-etl` as canonical data workflow entrypoint.
- [x] Remove overlapping `db populate*` and `db reset` commands from `baseball db`.
- [x] Keep `db` command scope focused on schema and explicit DB maintenance operations.
- [x] Update canonical run flow to keep API online while ETL worker runs independently.

Acceptance:

- [x] Users can complete setup with `db migrate -> server start + baseball-etl worker -> baseball-etl run -> baseball-etl validate -> baseball-etl status`.
- [x] Legacy `db` data-loading wrappers are removed to prevent redundant operational paths.
