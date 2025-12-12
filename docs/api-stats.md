# Stats & Leaderboards API Overview

Details for Section 6 of `docs/ROADMAP.md`. These Lahman-backed endpoints power stat tables, player cards, and leaderboard pages.

## Career & Season Stats

| Endpoint                                     | Dataset | Highlights                                                                                |
| -------------------------------------------- | ------- | ----------------------------------------------------------------------------------------- |
| `GET /v1/stats/batting`                      | L       | Flexible batting query supporting filters for player/team, season range, and minimum PA.  |
| `GET /v1/stats/pitching`                     | L       | Similar query engine for pitching; supports min innings, role filters, and sorting.       |
| `GET /v1/stats/fielding`                     | L       | Fielding totals by player/position with innings data for qualification checks.            |
| `GET /v1/players/{player_id}/stats/batting`  | L+R     | Player-focused batting aggregates that combine Lahman totals with optional split toggles. |
| `GET /v1/players/{player_id}/stats/pitching` | L+R     | Player pitching summary with advanced ratios (WHIP, SO/BB) and Retrosheet overlays.       |

### Details

- Every endpoint supports `page`, `per_page`, `sort`, and `order`, mirroring CLI expectations.
- Aggregations expose both counting stats and derived rates to minimize re-computation on the client.

## Seasonal Leaders

| Endpoint                                  | Dataset | Highlights                                                                        |
| ----------------------------------------- | ------- | --------------------------------------------------------------------------------- |
| `GET /v1/seasons/{year}/leaders/batting`  | L       | Season leaders for HR/AVG/RBI/OBP/etc. with ability to select leaderboard length. |
| `GET /v1/seasons/{year}/leaders/pitching` | L       | Leaderboards for ERA/SO/W/SV and other pitching categories.                       |
| `GET /v1/leaders/batting/career`          | L       | Career-long batting leaderboards honoring Lahman definitions.                     |
| `GET /v1/leaders/pitching/career`         | L       | Same for pitching categories such as strikeouts or ERA+.                          |

### Details

- Leader endpoints cache results per season to avoid recomputing heavy aggregate queries.
- Clients can supply `stat` or `limit` parameters to tailor the leaderboard to a single metric or top-N list.

## Team-Level Stats

| Endpoint                       | Dataset | Highlights                                                                 |
| ------------------------------ | ------- | -------------------------------------------------------------------------- |
| `GET /v1/stats/teams/batting`  | L       | Aggregated team batting; accepts `league`, `season`, and `min_pa` filters. |
| `GET /v1/stats/teams/pitching` | L       | Team pitching metrics, supporting sorting on ERA, strikeouts, etc.         |
| `GET /v1/stats/teams/fielding` | L       | Team-level fielding stats with errors, assists, and defensive efficiency.  |

### Details

- Team stat endpoints align with roster endpoints so standings and leaderboard views can share the same data model.
- Useful for building comparative dashboards or powering `/seasons/{year}/teams/{team_id}` detail views.
