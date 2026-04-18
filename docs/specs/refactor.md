---
title: Backend Refactor (CLI and Runtime Simplification)
updated: 2026-04-18
---

## Context

The backend currently works and has broad capability, but the operational surface is harder than it needs to be.

Primary signals from the current code/docs:

- CLI surface area is large: 35 `RunE` handlers and 40 `AddCommand` registrations.
- Command implementation is concentrated in very large files (`cmd/etl.go` 924 lines, `cmd/server.go` 514 lines, `cmd/db.go` 468 lines, `cmd/status.go` 397 lines).
- DB/bootstrap wiring is repeated across command handlers (`db.Connect` appears in ~20 command paths).
- ETL behavior exists in two layers with overlapping responsibility:
    - Orchestration in `internal/seed/pipeline.go`
    - Additional orchestration-style logic in `cmd/etl.go` and `cmd/db.go`
- User-facing command model has overlap between `etl` and `db` (`etl run`, `etl load`, `db populate`, `db reset`, `db refresh-views`).

## Goals

- Keep all current backend functionality and data coverage.
- Make the default operational path obvious and minimal.
- Reduce command-layer duplication and hidden coupling.
- Keep backward compatibility for existing command users during migration.

## Non-Goals

- No API contract changes under `/v1/*`.
- No schema redesign of core baseball data tables.
- No removal of advanced ETL controls (`--years`, `--era`, profile/mode semantics).

## Simplified Backend Model

### 1) Canonical operational path

Define one golden path for day-to-day usage:

1. `baseball db recreate` (optional)
2. `baseball db migrate`
3. `baseball etl` (or `baseball etl run`)
4. `baseball etl validate`
5. `baseball etl status`
6. `baseball server start`

This path becomes the primary docs path and is the default support target.

### 2) CLI information architecture

Clarify command ownership:

- `etl`: data pipeline orchestration and dataset stage operations
- `db`: schema lifecycle and DB-level maintenance only
- `server`: API runtime and diagnostics
- `cache`: cache diagnostics and invalidation

Planned consolidation:

- Keep `etl` as the canonical data entrypoint.
- Keep stage commands under `etl fetch/*` and `etl load/*` for debugging/partial runs.
- Deprecate overlapping `db populate*` and `db reset` flows in favor of `etl` plus explicit DB commands.
- During deprecation window, retain wrappers that forward to canonical implementations and print migration hints.

### 3) Runtime/bootstrap extraction

Introduce a shared command runtime package (for example `internal/app` or `internal/bootstrap`) that centralizes:

- Config loading and validation
- DB connection lifecycle
- Redis/cache connection lifecycle
- Common repository wiring for command use-cases

Command handlers should consume this runtime instead of performing repeated local setup.

### 4) Command handlers become thin

Move operational logic out of `cmd/*.go` into focused use-case services.

- `cmd` package responsibilities:
    - Cobra command/flag definitions
    - Input parsing and output formatting
    - Calling a use-case/service API
- `internal/seed` and new service packages responsibilities:
    - Pipeline orchestration
    - Dataset load/fetch logic
    - Validation/status logic

Target outcome: `cmd/*.go` mostly declarative and testable via command contracts.

### 5) Single dataset registry for status and validation

Create a shared dataset contract used by both:

- `etl status` (human-readable report)
- `etl validate` (machine-actionable pass/fail)

This avoids duplicated dataset checks and drift between status output and validation criteria.

### 6) Route introspection simplification

Current `server routes` AST parsing is brittle and disconnected from runtime registration.

Refactor to route registration metadata generated at route registration time, then use that for:

- `server routes`
- future docs verification checks

## Compatibility Strategy

- Keep existing command names during migration.
- Add explicit deprecation notices for overlapping commands before removal.
- Preserve current flag behavior (`--config`, `--profile`, `--mode`, `--years`, `--era`) unless a replacement is fully equivalent.

## Success Criteria

- New users can execute the full data-to-server workflow from one short docs path.
- Command layer line count and duplication materially decrease.
- `etl status` and `etl validate` derive from one shared dataset contract.
- No regression in ETL outcomes, API behavior, or existing automation scripts during migration window.
