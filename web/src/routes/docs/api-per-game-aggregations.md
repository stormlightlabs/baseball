# Per-game Aggregations API Overview

Per-game player and team stats derived from Retrosheet plays views to power game-finder experiences, WAR calculations, and daily trend charts.

## Summary

| Endpoint                              | Dataset | Highlights                                                                                                            |
| ------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------- |
| `GET /v1/players/{id}/game-logs`      | R       | Per-game batting logs derived from plays views; supports "game finder" style lookups with coverage from 1910-2025.    |
| `GET /v1/players/{id}/stats/batting`  | R       | Same per-game aggregation surfaced from both the player and game context for symmetry in client integrations.         |
| `GET /v1/games/{id}/batting`          |         |                                                                                                                       |
| `GET /v1/players/{id}/stats/pitching` | R       | Adds per-game pitching views that unlock WAR-style calculations and advanced splits across the full 1910-2025 window. |
| `GET /v1/games/{id}/pitching`         |         |                                                                                                                       |
| `GET /v1/players/{id}/stats/fielding` | R       | Fielding aggregation keyed by position to power defensive leaderboards and situational breakdowns.                    |
| `GET /v1/games/{id}/fielding`         |         |                                                                                                                       |
| `GET /v1/teams/{id}/daily-stats`      | R       | Team-level daily stats feed for trend charts and rolling aggregates.                                                  |

## Details

- Game logs and per-game batting stats are exposed via both `/v1/players/{id}/game-logs` and `/v1/players/{id}/stats/batting`, mirroring the payload inside `/v1/games/{id}/batting`. This shared view keeps the response identical regardless of whether a consumer starts from the player or the game resource and ensures coverage from 1910-2025.
- Pitching and fielding stats reuse the same plays-derived views to provide per-game WAR inputs, situational splits, and positional defensive metrics. Surfacing the data via both player- and game-scoped endpoints simplifies caching and avoids client-side joins.
- `/v1/teams/{id}/daily-stats` aggregates the plays view at the team level so dashboards can pull multi-season rolling performance without re-implementing SQL window logic.
