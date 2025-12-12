# Players API Overview

Expands on Section 1 of `docs/ROADMAP.md`. Covers core Lahman-backed biographical data along with Retrosheet game-level joins.

## Player & Career Endpoints

| Endpoint                                   | Dataset | Highlights                                                                          |
| ------------------------------------------ | ------- | ----------------------------------------------------------------------------------- |
| `GET /v1/players`                          | L       | Filterable roster of all players (name search, era, handedness, position).          |
| `GET /v1/players/{player_id}`              | L       | Combines biographical info with career totals across batting/pitching/fielding.     |
| `GET /v1/players/{player_id}/seasons`      | L       | Detailed year-by-year splits with season metadata and league context.               |
| `GET /v1/players/{player_id}/teams`        | L       | Chronological list of teams suited up for, linked to team and franchise references. |
| `GET /v1/players/{player_id}/awards`       | L       | Paginates Lahman `Awards*` tables with award type, year, and voting results.        |
| `GET /v1/players/{player_id}/hall-of-fame` | L       | Hall of Fame ballot history including ballots, votes, and induction result.         |
| `GET /v1/players/{player_id}/salaries`     | L       | Salary records with team, league, and inflation-ready numeric fields.               |

### Details

- Collection endpoints accept `page`, `per_page`, and optional `sort` fields
    - single-player endpoints share a consistent envelope with `player_id`, Lahman IDs, and stat aggregates.
- Award/HOF/salary feeds reuse Lahman table IDs

## Retrosheet Player Game Interfaces

| Endpoint                                        | Dataset | Highlights                                                                             |
| ----------------------------------------------- | ------- | -------------------------------------------------------------------------------------- |
| `GET /v1/players/{player_id}/game-logs`         | R       | Returns Retrosheet game log rows (starter-focused today) with per-game stat lines.     |
| `GET /v1/players/{player_id}/appearances`       | R       | Enumerates every appearance with role (starter, pinch-hit, defensive swap) and inning. |
| `GET /v1/players/{player_id}/plays`             | R       | Streams every play involving the player with Lahman↔Retrosheet identity mapping.      |
| `GET /v1/players/{player_id}/plate-appearances` | R       | Normalized PA feed exposing pitcher/batter splits, leverage, and situational filters.  |

### Details

- All Retrosheet-backed endpoints share `from`, `to`, `opponent`, `home_away`, and `min_pa`/`min_ip` filters to reduce payload size.
- `game-logs` and `appearances` are keyed by Retrosheet `GAME_ID` and include helper fields for quick linking to `/v1/games/{game_id}`.
