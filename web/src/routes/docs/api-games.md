# Games & Schedules API Overview

Focuses on global game search, summaries, and per-season schedule utilities.

## Summary

| Endpoint                                       | Dataset | Highlights                                                                     |
| ---------------------------------------------- | ------- | ------------------------------------------------------------------------------ |
| `GET /v1/games`                                | R       | Flexible search across seasons, clubs, parks, and date ranges with pagination. |
| `GET /v1/games/{game_id}`                      | R       | Canonical game metadata: teams, scores, attendance, duration, and umpire crew. |
| `GET /v1/games/{game_id}/boxscore`             | R       | Expanded boxscore including lineups, substitutions, and per-player stat lines. |
| `GET /v1/games/{game_id}/summary`              | R       | Narrative-friendly digest (winning/losing pitchers, save, highlight events).   |
| `GET /v1/seasons/{year}/schedule`              | R       | Whole-season schedule keyed by Retrosheet `GAME_ID`, league, and park data.    |
| `GET /v1/seasons/{year}/dates/{date}/games`    | R       | All games on a calendar date; convenient for scoreboard views.                 |
| `GET /v1/seasons/{year}/teams/{team_id}/games` | R       | Team-specific slice; alias of the team schedule table but kept for clarity.    |
| `GET /v1/seasons/{year}/parks/{park_id}/games` | R       | Filter games hosted at a particular park with weather/context fields.          |

## Endpoint Details

- Every endpoint emits a normalized `game` object (Retrosheet `GAME_ID`, Lahman team IDs, timestamps)
so an API client/consumers can deep-link without additional lookups.
- `GET /v1/games` supports advanced filters like `day_night`, `park_id`, `doubleheader`, and `postseason` to cut down on local filtering.
- Boxscore payloads embed both batting and pitching lines, derived win probabilities (when available), and
reference IDs needed to join into `/v1/plays`.
- Season/date/park schedule listings include `status` (scheduled, completed)
