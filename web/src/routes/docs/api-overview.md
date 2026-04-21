# API Overview

Big Fly exposes a versioned REST API under `/v1`.

Swagger/OpenAPI explorer is available at [/explorer](/explorer).

## Health & Readiness

- Service probes:
  - `GET /v1/health`
  - `GET /v1/ready`

For metadata/probes, see [Meta & Utility](/docs/api-meta-utility).

## Endpoint Families

- Search and lookup: [Search](/docs/api-search)
- Player resources: [Players](/docs/api-players)
- Team/franchise resources: [Teams](/docs/api-teams)
- Games and events: [Games](/docs/api-games), [Play-by-Play](/docs/api-play-by-play)
- Stats and leaders: [Stats](/docs/api-stats)
- Awards/postseason/event feeds: [Awards & Postseason](/docs/api-awards-postseason), [Achievements](/docs/api-achievements)
- Pitch-level event access: [Pitches](/docs/api-pitches), [Pitch Sequencing internals](/docs/pitches)
- Advanced/derived analytics: [Derived & Advanced](/docs/api-derived-advanced), [Computed & Advanced](/docs/api-computed)
- MLB proxy routes: [MLB Proxy](/docs/api-mlb-proxy)

## Quick API Calls

```bash
# Query players with fuzzy matching
curl "/v1/players?name=babe%20ruth&season=1927"

# Inspect one team season
curl "/v1/teams/NYY?year=2024"

# Fetch postseason metadata and plays
curl "/v1/postseason/series?year=2024"
curl "/v1/plays?game_id=NYN202410010"

# Check dataset freshness/readiness
curl "/v1/meta/datasets"
curl "/v1/ready"
```
