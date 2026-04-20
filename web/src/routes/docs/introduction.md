# Introduction

Big Fly unifies historical baseball datasets behind one unified API, currently behind the namespace: `/v1`.

Use these docs to understand coverage, query patterns, and authentication (to manage your keys).

## Quick Start

1. Check health/readiness:
   - `GET /v1/health`
   - `GET /v1/ready`
2. Open Swagger (served by the API binary): [/explorer](/explorer)
3. Create API keys from the web account page: [/account](/account)
4. Start with endpoint families:
   - `/v1/search/*`
   - `/v1/players/*`
   - `/v1/teams/*`
   - `/v1/games/*`
   - `/v1/stats/*`

## Reading Path

1. [About](/docs/about) for scope, data sources, and architecture.
2. [Apps](/docs/apps) for practical client usage patterns.
3. [API Overview](/docs/api-overview) for endpoint families and starter calls.
4. [Attribution](/docs/attribution) for dataset and constants sourcing notes.
5. API reference pages for concrete request/response contracts.

## Notes

- Swagger/OpenAPI docs remain hosted in the API service at `/v1/docs/`.
- Web and mobile clients should target `/v1` paths only, not `/api/v1`.
