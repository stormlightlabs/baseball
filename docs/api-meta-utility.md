# Meta & Utility API Overview

These lightweight endpoints power monitoring and discovery functionality for the platform.

## Summary

| Endpoint                | Dataset | Highlights                                                            |
| ----------------------- | ------- | --------------------------------------------------------------------- |
| `GET /v1/health`        | -       | Liveness probe that confirms the API process is responding.           |
| `GET /v1/ready`         | Mixed   | Readiness probe for the core dataset set required by the API.         |
| `GET /v1/meta`          | Mixed   | Reports API build metadata plus dataset freshness and coverage.       |
| `GET /v1/meta/datasets` | Mixed   | Enumerates core and supplemental datasets currently loaded.           |
| `GET /v1/meta/readiness`| Mixed   | Returns the same readiness payload without a readiness status code.   |

## Endpoint Details

### `GET /v1/health`

- Returns HTTP 200 when the API process is responding.
- Use this endpoint for liveness checks.

### `GET /v1/ready`

- Returns HTTP 200 when the required datasets are loaded.
- Returns HTTP 503 when the API is live but still missing core seed data.
- Used by `baseball server health` and smoke tests that need seeded data.

### `GET /v1/meta`

- Response includes API semantic version and per-dataset checksum/hash values.
- This payload can be cached to detect when ETL refreshes have happened.

### `GET /v1/meta/datasets`

- Returns an array of datasets with name, coverage window, freshness metadata,
  and health flags.
- Useful for CLI tooling to warn when a requested season is missing from the warehouse.

### `GET /v1/meta/readiness`

- Returns the required dataset readiness summary as JSON.
- Useful when you want readiness details without relying on a 200/503 contract.
