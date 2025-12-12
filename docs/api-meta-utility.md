# Meta & Utility API Overview

These lightweight endpoints power monitoring and discovery functionality for the platform.

## Summary

| Endpoint                | Dataset | Highlights                                                            |
| ----------------------- | ------- | --------------------------------------------------------------------- |
| `GET /v1/health`        | -       | Readiness probe used by CLI, orchestrators, and smoke tests.          |
| `GET /v1/meta`          | L+R     | Reports API build metadata plus Lahman/Retrosheet refresh timestamps. |
| `GET /v1/meta/datasets` | L+R     | Enumerates which seasons, leagues, and tables are currently loaded.   |

## Endpoint Details

### `GET /v1/health`

- Returns HTTP 200 as long as the API process, DB connection, and configuration are valid.
- Used by `baseball server health` and infra monitors; no auth or query params.

### `GET /v1/meta`

- Response includes API semantic version, git SHA, and per-dataset checksum/hash values.
- This payload can be cached to detect when ETL refreshes have happened.

### `GET /v1/meta/datasets`

- Returns an array of datasets with name, coverage window, and freshness metadata (last ETL runtime, row counts, etc.).
- Useful for CLI tooling to warn when a requested season is missing from the warehouse.
