# ETL Binary + Data Product Tasks

Scope: keep ETL runtime performance-safe by moving upstream data engineering to `bigflydata`, while keeping `baseball-etl` focused on ingest/read/upsert/validate.

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
- [x] Remove `tools/data` git submodule from this repo.
- [ ] Add smoke tests for `baseball-etl --help`, `run --help`, `validate --help`, `status --help`.
- [ ] Remove ETL command registration from `internal/cli.NewBaseballRootCmd` once cutover is complete.

Acceptance:

- [x] ETL can run without shipping server/cache command surfaces in its process.
- [ ] Primary `baseball` CLI no longer exposes ETL commands after cutover.

## Phase 2: Upstream Snapshot Contract (`bigflydata`)

- [ ] Define Snapshot Contract V1 (`raw/`, `prepared/`, `snapshot.manifest.json`, schema version).
- [ ] Store extracted raw source files in VCS/LFS as canonical artifacts (zip files are transitional inputs only).
- [ ] Implement transform stages in Python using Polars + NumPy (no pandas).
- [ ] Produce ingest-ready prepared outputs with deterministic schema/path contract.
- [ ] Add data quality checks and row-level invariants in `bigflydata` CI.
- [ ] Publish documentation and runbooks in `/Users/owais/Projects/bigflydata/docs/spec.md` + `todo.md`.

Acceptance:

- [ ] A pinned `bigflydata` ref fully defines ETL input files without upstream fetch variance.
- [ ] Prepared outputs are deterministic and directly ingestible by Go ETL.

## Phase 3: Ingestion Contract in `baseball-etl`

- [ ] Add manifest/contract preflight validation before load.
- [ ] Add prepared-data loader path as default ingest route.
- [ ] Keep legacy archive-centric loader path only as fallback during migration.
- [ ] Add per-dataset upsert strategy docs (keys, conflict policy, idempotency guarantees).
- [ ] Track load throughput metrics (rows/s, duration, failure class) per phase.

Acceptance:

- [ ] Steady-state ETL does not require zip decompression/network fetch from providers.
- [ ] Load behavior is idempotent and resumable with deterministic inputs.

## Phase 4: Dedicated ETL Container (Dev + Prod Compose)

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
- [ ] API service remains independently operable during ETL runs.

## Phase 5: Safety Rails + Throughput

- [ ] Add single-active ETL run guard (advisory lock or `etl_run_locks` table).
- [ ] Add per-step timeout and cancellation policy for heavy ETL phases.
- [ ] Add explicit post-load `ANALYZE` for heavily changed tables/partitions.
- [ ] Add DB backpressure throttling hooks (latency/WAL-sensitive pacing).
- [ ] Add host-level alerting for free disk and WAL growth thresholds.
- [ ] Add runbook actions for WAL pressure (pause ETL, archive/prune strategy, checkpoint analysis).
- [ ] Add `pg_stat_bgwriter` trend capture (`checkpoints_req` / `checkpoints_timed`).

Acceptance:

- [ ] Concurrent ETL starts are rejected safely.
- [ ] Operators have deterministic controls for pause/resume/recovery.

## Phase 6: De-Materialization + Partitioned Serving

- [ ] Mark materialized-view refresh loops as legacy and remove from default ETL run path.
- [ ] Define partitioned serving-table targets for current MV-backed heavy artifacts.
- [ ] Add partition management policy (create/attach/retire) for serving tables.
- [ ] Implement upsert/load steps from `bigflydata` prepared artifacts into partitioned serving tables.
- [ ] Remove API/query dependencies on materialized-view-only objects.
- [ ] Replace MV-centric observability/reporting with partitioned-table load metrics and checks.
- [ ] Add runbook queries for partition skew, stale partitions, and slow upsert phases.

Acceptance:

- [ ] Serving ingestion scales with changed partitions, not full-history MV rebuilds.
- [ ] Interrupted runs can resume without full rerun.

## Verification Checklist

- [x] `go test ./...`
- [ ] ETL lock/concurrency behavior validated with two concurrent start attempts.
- [ ] `docker compose` ETL one-shot run succeeds in dev and prod layouts.
- [ ] API readiness remains healthy during ETL run window.
- [x] ETL docs updated in `docs/internal/data-loading.md`.
- [x] ETL docs updated in `conf/README.md`.
