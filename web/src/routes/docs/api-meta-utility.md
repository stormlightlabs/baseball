# Meta & Utility API Overview

These lightweight endpoints power monitoring and discovery functionality for the platform.

## Summary

| Endpoint                         | Dataset | Highlights                                                              |
| -------------------------------- | ------- | ----------------------------------------------------------------------- |
| `GET /v1/health`                 | -       | Liveness probe that confirms the API process is responding.             |
| `GET /v1/ready`                  | Mixed   | Readiness probe for the core dataset set required by the API.           |
| `GET /v1/meta`                   | Mixed   | Reports API build metadata plus dataset freshness and coverage.         |
| `GET /v1/meta/datasets`          | Mixed   | Enumerates core and supplemental datasets currently loaded.             |
| `GET /v1/meta/readiness`         | Mixed   | Returns the same readiness payload without a readiness status code.     |
| `GET /v1/meta/crosswalk/teams`   | Mixed   | Season-scoped MLBAM team ID to local `team_id`/`franchise_id` mappings. |
| `GET /v1/meta/crosswalk/players` | Mixed   | MLBAM person ID to local Lahman/Retrosheet ID mappings.                 |

## Endpoint Details

### `GET /v1/health`

- Returns HTTP 200 when the API process is responding.
- Use this endpoint for liveness checks.

### `GET /v1/ready`

- Returns HTTP 200 when the required datasets are loaded.
- Returns HTTP 503 when the API is live but still missing core seed data.
- Used by `baseball server health` and smoke tests that need seeded data.
- Always runs in lightweight count mode and reports `X-Count-Mode: lightweight`.

### `GET /v1/meta`

- Response includes API semantic version and per-dataset checksum/hash values.
- This payload can be cached to detect when ETL refreshes have happened.
- Supports `?strict=true` for exact row counts.
- Default mode is lightweight, with response header `X-Count-Mode: lightweight|strict|fallback`.

### `GET /v1/meta/datasets`

- Returns an array of datasets with name, coverage window, freshness metadata,
  and health flags.
- Useful for CLI tooling to warn when a requested season is missing from the warehouse.
- Supports `?strict=true` for exact row counts.
- Default mode is lightweight, with response header `X-Count-Mode: lightweight|strict|fallback`.

### `GET /v1/meta/readiness`

- Returns the required dataset readiness summary as JSON.
- Useful when you want readiness details without relying on a 200/503 contract.
- Always runs in lightweight count mode and reports `X-Count-Mode: lightweight`.

### `GET /v1/meta/crosswalk/teams`

- Canonical team crosswalk endpoint (replaces legacy `/v1/mlb/crosswalk/teams`).
- Supports season-scoped lookup (`season`), all-season mode (`all_seasons=true`), and filters by local or MLBAM IDs.
- Lookup helpers are available at:
  - `/v1/meta/crosswalk/teams/by-team/{team_id}`
  - `/v1/meta/crosswalk/teams/by-franchise/{franchise_id}`
  - `/v1/meta/crosswalk/teams/by-mlbam/{mlbam_team_id}`

### `GET /v1/meta/crosswalk/players`

- Canonical player crosswalk endpoint for Lahman/Retrosheet/MLBAM IDs.
- Supports filters by `player_id`, `retro_id`, and `mlbam_id`.
- Lookup helpers are available at:
  - `/v1/meta/crosswalk/players/by-player/{player_id}`
  - `/v1/meta/crosswalk/players/by-retro/{retro_id}`
  - `/v1/meta/crosswalk/players/by-mlbam/{mlbam_id}`
