# Testing

This project uses Go's standard test runner for automated checks. The current
suite focuses on integration-style API tests that run against a real Postgres
instance, seeded with a slice of sampled data.

## Coverage Scope

The main automated coverage lives under `internal/api`. The test harness does the
following before the package `tests` run:

- Starts a temporary Postgres 16 container with `testcontainers-go`.
- Runs the full migration set from `internal/db/sql`.
- Loads CSV fixtures from `internal/testutils/testdata`.
- Seeds supplemental SQL fixtures for wOBA constants, league constants, park factors,
  and salary summaries.
- Refreshes materialized views so repository and API queries see the same shape the
  application expects in normal use.

The current suite does not include full end-to-end worker-process tests for ETL
loaders. It also does not require Redis for test execution.

The migration/test path primarily exercises fresh-database behavior. Keep this
aligned with long-lived cutover behavior by validating migration + narrow-slice
runs from the [ETL Cutover & Narrow-Slice Runbook](./etl-cutover.md) when ETL
maintenance semantics change.

## Prerequisites

You need the following before running tests:

- Go 1.24 or newer.
- The Docker daemon running locally, because the tests start Postgres in a
  container.
- Network access on the first run if Go modules or the `postgres:16-alpine`
  image are not already cached locally.

You do not need a manually provisioned local Postgres instance for the API tests.

## Run the tests

Run the full suite as you normally would with `go test`:

```bash
task test         # or optionally filter: test:filter -- -run TestPlayerEndpoints
go test ./... -v  # optionally filter: -run TestPlayerEndpoints
```

## Next steps

If you change ETL behavior, add or update integration coverage for the loader
path as well as the API surface that depends on the loaded data.
