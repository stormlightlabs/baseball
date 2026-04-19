# Teams, Franchises & Seasons API Overview

Documents Lahman team/franchise references plus Retrosheet schedule views.

## Reference & Season Catalog

| Endpoint                         | Dataset | Highlights                                                                             |
| -------------------------------- | ------- | -------------------------------------------------------------------------------------- |
| `GET /v1/teams`                  | L       | Paginated list of team seasons with filters for year, league, franchise, and min wins. |
| `GET /v1/teams/{team_id}`        | L       | Single team-season snapshot with wins/losses, scoring, and franchise metadata.         |
| `GET /v1/franchises`             | L       | Franchise directory with start/end years, current city, and active flag.               |
| `GET /v1/franchises/{franch_id}` | L       | Deep dive into a franchise with alias history, lineage, and counts of titles.          |
| `GET /v1/seasons`                | L       | Lists available seasons plus min/max coverage so the UI can gate filters.              |

### Details

- Endpoints expose Lahman IDs (`teamID`, `franchID`) alongside synthetic slugs for URL stability.
- Season results include derived fields (run differential, win pct) matching Lahman formulas for easy comparisons.

## Team Rosters & Splits

| Endpoint                                          | Dataset | Highlights                                                                    |
| ------------------------------------------------- | ------- | ----------------------------------------------------------------------------- |
| `GET /v1/seasons/{year}/teams`                    | L       | One-shot summary of every team in a season, ideal for league tables.          |
| `GET /v1/seasons/{year}/teams/{team_id}/roster`   | L       | Full roster with role (pitcher/batter), games played, and basic rate stats.   |
| `GET /v1/seasons/{year}/teams/{team_id}/batting`  | L       | Aggregated batting totals plus per-player splits keyed by player/team IDs.    |
| `GET /v1/seasons/{year}/teams/{team_id}/pitching` | L       | Team-wide pitching numbers including saves, ERA, WHIP style metrics.          |
| `GET /v1/seasons/{year}/teams/{team_id}/fielding` | L       | Fielding stats with position groupings and innings for qualification filters. |

### Details

- Each endpoint includes `league`, `division`, and `postseason` flags to support standings pages.
- Splits endpoints allow `min_pa`, `min_ip`, and `sort` query params matching stats endpoints.

## Retrosheet Schedule & Logs

| Endpoint                                            | Dataset | Highlights                                                                                     |
| --------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------- |
| `GET /v1/seasons/{year}/teams/{team_id}/schedule`   | R       | Raw game schedule with opponent, park, double-header sequence, and start times.                |
| `GET /v1/seasons/{year}/teams/{team_id}/games`      | R       | Paginated list of played games with score, record to-date, and links to `/v1/games/{game_id}`. |
| `GET /v1/seasons/{year}/teams/{team_id}/daily-logs` | R       | Rolling daily stats including win/loss streaks and aggregates useful for dashboards.           |

### Details

- Schedule and log endpoints share `from`, `to`, `home_away`, and `opponent` filters for consistent slicing.
- Daily logs include both team totals and derived per-game averages.
