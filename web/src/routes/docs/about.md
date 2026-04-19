# About

## What Big Fly Provides

Big Fly is a baseball data platform that combines:

- Lahman career + seasonal records
- Retrosheet game/event/play data
- Derived and computed sabermetric outputs
- Targeted MLB Stats API proxy endpoints (`/v1/mlb/*`)

Everything is exposed through the same REST namespace: `/v1`.

## Data Coverage

- Long-run historical coverage starts in the 19th century where source data supports it.
- Retrosheet-backed features are strongest in eras with complete event logs.
- Readiness/coverage details are published via:
  - `GET /v1/meta`
  - `GET /v1/meta/datasets`
  - `GET /v1/meta/readiness`

## App Split

- API service: `api.bigfly.tech`
- Static web app: `bigfly.tech` (Cloudflare Pages)
- Mobile app: Flutter client using the same `/v1` surface

Swagger remains served by the API binary at `/v1/docs/` and is linked from the web app via `/explorer`.
