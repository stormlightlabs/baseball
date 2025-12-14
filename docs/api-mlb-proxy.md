# MLB Stats Proxy API Overview

`internal/api/mlb.go` exposes a curated subset of `https://statsapi.mlb.com/api/v1` through `/v1/mlb/*`.
Every proxy call adds caching, shared auth, and a consistent schema so clients do not need to hit MLB's public API directly.

## Summary

| Endpoint                        | Dataset        | Highlights                                                                                   |
| ------------------------------- | -------------- | -------------------------------------------------------------------------------------------- |
| `GET /v1/mlb`                   | MLB Stats API  | Catalog of proxied routes (`routes` array) plus base URL metadata.                           |
| `GET /v1/mlb/people`            | MLB Stats API  | Pass-through to `/v1/people`; supports `personIds`, `sportId`, and `hydrate`.                |
| `GET /v1/mlb/people/{id}`       | MLB Stats API  | Single-player lookup mirroring `/v1/people/{personId}`.                                      |
| `GET /v1/mlb/teams`             | MLB Stats API  | Team directory; accepts `sportId` (defaults to 1) and `season`.                              |
| `GET /v1/mlb/teams/{id}`        | MLB Stats API  | Individual team record, optionally filtered by `season`.                                     |
| `GET /v1/mlb/schedule`          | MLB Stats API  | Schedule endpoint relayed directly from MLB Stats; any query params are forwarded.           |
| `GET /v1/mlb/seasons`           | MLB Stats API  | Season metadata (start/end dates, postseason flags).                                         |
| `GET /v1/mlb/stats`             | MLB Stats API  | The generic stats endpoint for situational queries (leaders, splits, etc.).                  |
| `GET /v1/mlb/standings`         | MLB Stats API  | League/division standings; accepts the same query flags as Stats API.                        |
| `GET /v1/mlb/awards`            | MLB Stats API  | Awards directory.                                                                            |
| `GET /v1/mlb/awards/{id}`       | MLB Stats API  | Individual MLB award definition or recipient list.                                          |
| `GET /v1/mlb/venues`            | MLB Stats API  | Venue directory mirrored from `/v1/venues`.                                                  |

## Endpoint Notes

- All query parameters are passed directly to MLB unless explicitly overridden (default `sportId=1` for a few routes).
- Responses include an `X-Proxy-Target: statsapi.mlb.com` header so consumers can log/trace upstream hits.
- The proxy uses a short-lived HTTP cache keyed on method + URL; repeated requests for the same MLB endpoint avoid re-fetching.
- Errors from MLB (non-2xx status or invalid JSON) are surfaced as `500` responses with context in the error message.
