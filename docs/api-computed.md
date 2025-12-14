# Computed & Advanced Analytics API Overview

These endpoints sit on top of Lahman + Retrosheet foundations with additional derived constants
for advanced metrics (wOBA, leverage index, park factors, WAR). They power sabermetric tabs in the
UI and higher-level CLI summaries.

## Summary

| Endpoint                                                    | Dataset | Highlights                                                                                           |
| ----------------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `GET /v1/players/{player_id}/stats/batting/advanced`        | Derived | Full advanced batting line (wOBA, wRC+, ISO, BABIP) with optional season/team filters.               |
| `GET /v1/players/{player_id}/stats/pitching/advanced`       | Derived | Advanced pitching package (FIP, xFIP, ERA+, K/BB, etc.) scoped by season/team.                       |
| `GET /v1/players/{player_id}/stats/baserunning`             | Derived | Baserunning value (wSB, steals, advancement runs) per player/season.                                |
| `GET /v1/players/{player_id}/stats/fielding`                | Derived | Fielding runs, positional adjustments, and rate stats per player/season.                             |
| `GET /v1/players/{player_id}/stats/war`                     | Derived | WAR + component breakdown with optional season/team filters.                                         |
| `GET /v1/players/{player_id}/leverage/summary`              | Derived | Aggregated leverage index summary (avg LI, clutch metrics) for a season.                             |
| `GET /v1/players/{player_id}/leverage/high`                 | Derived | Returns every high-leverage plate appearance above a configurable LI threshold.                      |
| `GET /v1/games/{game_id}/plate-appearances/leverage`        | Derived | Plate appearance leverage list for a single game with optional `min_li` floor.                       |
| `GET /v1/games/{game_id}/win-probability/summary`           | Derived | Biggest swings, comeback odds, and total leverage time for one game.                                 |
| `GET /v1/parks/{park_id}/factors`                           | Derived | Year-specific park factors (runs, HR, handedness splits).                                            |
| `GET /v1/parks/{park_id}/factors/series`                    | Derived | Multi-year park factor series for trend charts.                                                      |
| `GET /v1/seasons/{season}/park-factors`                     | Derived | Season-wide park factor table for every venue; filterable by factor type.                            |
| `GET /v1/seasons/{season}/leaders/batting/advanced`         | Derived | Advanced batting leaderboard with `stat`, `min_pa`, `team_id`, and `league` filters.                  |
| `GET /v1/seasons/{season}/leaders/pitching/advanced`        | Derived | Advanced pitching leaderboard with `stat`, `min_ip`, `team_id` filters.                              |
| `GET /v1/seasons/{season}/leaders/war`                      | Derived | WAR leaderboard sharing filters with batting leader endpoint (`min_pa`, `team_id`).                  |

## Endpoint Notes

### Player Advanced Stat Blocks

- Each `stats/*` endpoint accepts optional `season` and `team_id` query parameters; omitting them returns career-to-date derived values.
- Responses reflect the strongly-typed structs in `core` (e.g., `AdvancedBattingStats`, `AdvancedPitchingStats`) so downstream code can reuse the same fields rendered in the CLI.
- Baserunning/fielding/WAR endpoints share the same pattern and return HTTP 404 when the requested sample lacks enough innings/plays to compute values.

### Player & Game Leverage

- `GET /v1/players/{player_id}/leverage/summary` aggregates leverage index by role (`role` query accepts `batter` or `pitcher`) and season (default current year).
- `GET /v1/players/{player_id}/leverage/high` emits the raw plate-appearance rows (with inning, outs, LI) filtered by `min_li` (default 1.5) and `season`.
- `GET /v1/games/{game_id}/plate-appearances/leverage` mirrors the same structure but scoped to a single game; use `min_li` to highlight only pivotal moments.
- `GET /v1/games/{game_id}/win-probability/summary` summarizes swings > N%, comeback odds, and identifies the highest-leverage plays for recap content.

### Park Factors

- `GET /v1/parks/{park_id}/factors` requires a `season` query parameter.
- `GET /v1/parks/{park_id}/factors/series` expects `from_season` and `to_season`, returning a chronological array you can graph.
- `GET /v1/seasons/{season}/park-factors` optionally accepts `type` (e.g., `runs`, `hr`) to filter the response server-side.

### Advanced Leaderboards

- `GET /v1/seasons/{season}/leaders/batting/advanced` and `/pitching/advanced` both support `stat`, pagination (`page`, `per_page`), and qualification filters (`min_pa`, `min_ip`). They return a `PaginatedResponse` without additional metadata.
- `GET /v1/seasons/{season}/leaders/war` shares pagination + `min_pa`/`team_id` filters but produces `PlayerWARSummary` entries.
- These endpoints load a full season’s leaderboard in memory and page the result inside the handler, so extremely high `per_page` values can increase response size quickly.
