---
title: ETL Binary + Data Product Architecture
updated: 2026-04-20
---

## Problem

The current deployment still runs API and ETL from the same runtime surface, and ETL remains responsible for too much upstream data handling (archive fetch/decompress/parse + DB writes).

For performance-sensitive runs, this causes avoidable contention:

- CPU/memory pressure from archive-heavy preprocessing
- long and variable ETL wall-clock time from network + decompression
- increased DB stress windows when transform + load happen in one process path

## Architectural Direction

Adopt a two-system model with strict responsibilities:

1. `bigflydata` is the upstream dataset factory.
2. `baseball-etl` is the downstream ingestion/runtime loader.

Core decisions:

- Snapshot source files should be tracked in VCS as extracted raw tabular data (not zip archives as primary artifacts).
- Zip archives are transitional only and should be removed from steady-state snapshot storage.
- Heavy transforms move to Python data tooling in `bigflydata` using Polars + NumPy (no pandas).
- Heavy derivation work moves upstream into `bigflydata`; Go ETL remains an ingestion runtime.
- Serving materialization is no longer the target strategy; partitioned table ingestion is preferred.
- `baseball-etl` focuses on pull/read/upsert/validate and DB-side safety.

## Responsibility Split

| Concern               | `bigflydata`                                           | `baseball-etl`                                            |
| --------------------- | ------------------------------------------------------ | --------------------------------------------------------- |
| Source acquisition    | Owns download/sync from upstream providers             | Does not fetch provider archives directly in steady state |
| Raw preservation      | Owns canonical raw snapshots in VCS/LFS                | Reads raw/prepared data via cloned snapshot root          |
| Transform logic       | Owns normalization/enrichment and ingest-ready outputs | Only lightweight mapping needed for upsert                |
| Snapshot metadata     | Owns manifest + schema contract metadata               | Verifies manifest/contract before load                    |
| Database writes       | None                                                   | Owns COPY/upsert/load + validation                        |
| Runtime observability | Optional build metrics                                 | Owns ETL run/step/refresh events                          |

## Snapshot Contract (V1 Target)

`bigflydata` should publish a deterministic contract under repo root:

- `snapshot.manifest.json`: file hashes, sizes, and snapshot timestamp
- `docs/spec.md`: contract version and dataset semantics
- `raw/`: extracted source-of-record tabular files
- `prepared/`: ingest-ready outputs optimized for Go loader throughput

Target prepared principles:

- stable file paths per dataset/stage
- explicit schema versioning
- deterministic row ordering where applicable
- no provider zip parsing required in `baseball-etl`

## Runtime Topology

```text
[Traefik in prod / Caddy in dev] -> [api container: baseball server start]
                     |
                     v
                [postgres] <-> [redis]
                     ^
                     |
       [etl container: baseball-etl run/validate/status]
                     ^
                     |
          [clone/pull bigflydata snapshot ref]
```

Deployment rules:

- API and ETL run in separate containers.
- Both may use the same image artifact.
- ETL exposes no HTTP port.
- ETL uses independent resource and DB pool limits.
- ETL runs are single-active via lock guard.

Ingress note:

- Production ingress uses Traefik via Coolify.
- Caddy is development-only.

## Performance Contract

- No archive decompression in steady-state ETL hot path.
- ETL reads ingest-ready files and performs bounded upsert batches.
- Keep force/year writes year-bounded and resumable.
- Avoid materialized-view rebuild loops in steady-state ETL.
- Keep partition maintenance explicit and bounded.
- Maintain ETL telemetry (`etl_runs`, `etl_run_steps`, `etl_step_events`), treating MV refresh events as legacy transition telemetry only.

## Migration Strategy

### Phase A: Upstream data-product hardening (`bigflydata`)

- Keep current sync/build/verify working.
- Add raw/prepared contract with explicit schema version.
- Add transform pipeline on Polars + NumPy.

### Phase B: Ingestion cutover (`baseball-etl`)

- Add prepared-data loader path as primary route.
- Keep legacy archive-centric path as fallback only during transition.
- Add manifest/contract preflight validation before DB writes.

### Phase C: Simplification

- Remove archive-first ETL assumptions from runbook and command surface.
- Keep `baseball-etl` focused on load/validate/status workflows.
- Remove serving materialization assumptions from ingestion runbooks and ETL hot path.

## Database Workstreams (Still In Scope)

- Single-active ETL lock and cancellation policy
- post-load analyze/backpressure hooks
- partition management and retention policy
- year/season-scoped upsert + recompute where needed
- operator runbooks for WAL/checkpoint pressure and resume/recovery

## Non-Goals (Current Scope)

- No immediate full schema rewrite to `raw/core/serving` schemas
- No API contract changes

## Acceptance Criteria

- ETL runtime no longer depends on provider zip parsing for steady-state runs.
- `bigflydata` snapshot ref + manifest deterministically reproduce ETL inputs.
- API availability is preserved during ETL windows with isolated resources.
- ETL remains observable, cancellable, and single-active.
- Heavy recompute work is moved upstream and bounded to affected data where feasible.
