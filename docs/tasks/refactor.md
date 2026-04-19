# Backend Refactor Tasks

Scope: simplify backend structure and CLI operations while retaining current behavior.

## Phase 0: Baseline and Safety Rails

- [ ] Capture baseline command contracts for:
    - `baseball etl --help`
    - `baseball etl run --help`
    - `baseball etl validate --help`
    - `baseball etl status --help`
    - `baseball db --help`
    - `baseball server --help`
- [ ] Add focused tests for the current golden path behavior (`db migrate`, `etl`, `etl validate`, `etl status`).
- [ ] Record current ETL/status outputs in test fixtures where practical.

Acceptance:

- [ ] We can detect behavior regressions before/after refactor.

## Phase 1: Establish Canonical CLI Flow

- [ ] Make docs and help text consistently present `etl` as the canonical data workflow entrypoint.
- [ ] Mark overlapping `db populate*` and `db reset` commands as deprecated in help/long descriptions.
- [ ] Keep wrappers functional while emitting migration guidance to canonical commands.

Acceptance:

- [ ] Users can complete full setup with: `db migrate -> etl -> etl validate -> etl status -> server start`.
- [ ] Existing scripts using old commands still run during migration window.

## Phase 2: Extract Shared Command Runtime

- [ ] Create shared runtime/bootstrap package for config + DB + Redis/cache setup.
- [ ] Move repeated command setup code into reusable helpers/services.
- [ ] Update command handlers to consume shared runtime constructors.

Acceptance:

- [ ] DB and cache initialization logic is not duplicated across command handlers.
- [ ] Command files become thinner and easier to reason about.

## Phase 3: Move Orchestration Out of `cmd`

- [ ] Move ETL orchestration-style logic from `cmd/etl.go` and `cmd/db.go` into `internal/seed` or a dedicated service package.
- [ ] Keep `cmd` package focused on Cobra definitions, parsing, and output.
- [ ] Ensure `db` commands that remain are DB-lifecycle only.

Acceptance:

- [ ] ETL execution paths originate from one orchestration layer.
- [ ] Command package does not duplicate seed pipeline logic.

## Phase 4: Unify Status and Validation Contracts

- [ ] Introduce a shared dataset check registry/contract used by both `etl status` and `etl validate`.
- [ ] Refactor both commands to consume this shared contract with different output modes.
- [ ] Keep existing dataset coverage checks and thresholds unless explicitly changed.

Acceptance:

- [ ] No drift between what `status` reports and what `validate` enforces.
- [ ] Adding a new dataset check requires one contract change, not two implementations.

## Phase 5: Route Introspection Cleanup

- [ ] Replace AST-based route discovery for `server routes` with registration-time metadata.
- [ ] Ensure route listing includes all runtime-registered endpoints and utility endpoints.

Acceptance:

- [ ] `server routes` reflects actual registered routes without AST parsing fragility.

## Phase 6: Documentation and Migration Finish

- [ ] Update root `README.md` and `docs/internal/data-loading.md` with the simplified canonical flow.
- [ ] Add a deprecation timeline and removal criteria for overlapping commands.
- [ ] Publish a short migration guide for automation/scripts.

Acceptance:

- [ ] One clear workflow is documented for new users.
- [ ] Existing users have a predictable migration path.

## Verification Checklist (Run Before Merge)

- [ ] `go test ./...`
- [ ] CLI golden path smoke test passes locally.
- [ ] Command help text is internally consistent and matches docs.
- [ ] No API behavior regressions on `/v1` endpoints.
