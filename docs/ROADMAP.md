# Baseball API Development Roadmap

This living roadmap merges the original guidelines and endpoint specifications into a single plan that tracks scope and completion status.

## Architecture & Data Strategy

### Three-Layer Approach

1. **Raw data → Postgres**: ETL pipelines for both the Lahman Baseball Database and Retrosheet archives.
2. **Postgres → Go domain model**: Typed queries, joins, and views for consistent access patterns.
3. **Go → HTTP API**: Versioned, cached, and documented endpoints layered on top of the domain model.

### Data Sources & IDs

- **Lahman** provides season and career aggregates (batting, pitching, fielding, awards, salaries, franchises, etc.) from 1871-2024. ([SABR Lahman Database](https://sabr.org/lahman-database/))
- **Retrosheet** publishes game logs and parsed play-by-play data. ([Retrosheet.org](https://www.retrosheet.org/)) Chadwick tools help convert raw event files and maintain crosswalks for IDs.
- Opinionated split: Lahman powers reference, leaderboards, and team/season stats, while Retrosheet powers games, logs, and play-by-play.
- ID conventions: `player_id` = Lahman `playerID`, `team_id` = Lahman `teamID` or `franchID`, `game_id` = Retrosheet `GAME_ID`. Chadwick mappings keep IDs consistent across sources.

### Database Schema & Storage

- Done: Postgres is the target store with reusable DDLs checked in under `internal/db/sql` (Lahman schema, Retrosheet schema, and plays schema derived from [Baseball-PostgreSQL](https://github.com/davidbmitchell/Baseball-PostgreSQL)).
- In-Progress: Views and crosswalk tables should normalize Lahman ↔ Retrosheet IDs (`player_id_map`, `team_franchise_map`) and pre-aggregate leaderboards for fast reads.
- To-Do: Materialized views for heavy aggregates (win probability, streaks, etc.) are queued behind upcoming analytics milestones.

### ETL & CLI Tooling

- Done
    - Cobra + lipgloss CLI (`cli/`, `cmd/`) delivers:
        - `baseball etl fetch {lahman|retrosheet}` - instructions/downloaders for raw archives.
        - `baseball etl load {lahman|retrosheet}` - uses Postgres `COPY` helpers for bulk loading (see `internal/db`).
        - `baseball db migrate` - applies the bundled DDL.
        - `baseball server {start|fetch|health}` - runs the API locally and exercises endpoints.
- In-Progress
    - `baseball etl status` currently prints a placeholder; add freshness checks comparing DB tables vs. downloaded archives.
- To-Do
    - Future work: resumable downloads, checksum validation, resumable play-by-play ingestion, and automated cron-style Taskfile targets.

### API Platform, Docs & Ops

- Done: Server uses the Go standard library HTTP mux, hand-rolled repositories, and `swaggo/swag` generated docs available at `/docs`.
- Done: All routes are namespaced under `/v1/` to allow future breaking changes.
- Done: Attribution requirements are documented in `README.md`, but they are also restated here for every release:
    - Lahman: attribute SABR + Sean Lahman.
    - Retrosheet: include the required copyright statement.
- To-Do: Add caching, rate limiting, and deployment scripts after the analytics milestone.

### Exposing Joined Data

- Done: Lahman + Retrosheet datasets can already be ingested concurrently; the repositories query the normalized tables directly.
- To-Do: Dedicated SQL views (e.g., `stats_career_batting`, `player_game_logs`) and ID-mapping helpers will simplify SQLC/hand-rolled queries.
- To-Do: Future endpoints that combine season stats with play-by-play data will rely on these views/materialized views.

## API Conventions

- Base URL: `https://baseball.stormlightlabs.org/api/v1/...` (prod) and `http://localhost:8080/v1/...` (dev).
- Dataset tags reused from the original spec:
    - **(L)** Lahman, **(R)** Retrosheet, **(L+R)** joined view.
- Common query params: `page`, `per_page`, `sort`, `order`, `from`/`to` (dates as `YYYY-MM-DD`), and stat filters such as `min_pa`, `min_ip`, `min_g`.
- Paginated responses wrap payloads inside an envelope containing `data`, `page`, `per_page`, and `total`.
- Default envelope metadata keeps clients backward compatible while letting us add new fields later.

## API Roadmap

### 0. Meta / Utility ✓

| Status | Endpoint                | Dataset | Description                                                         |
| ------ | ----------------------- | ------- | ------------------------------------------------------------------- |
| Done   | `GET /v1/health`        | -       | Basic readiness check used by CLI and deploy targets.               |
| Done   | `GET /v1/meta`          | L+R     | Returns API version, dataset refresh timestamps, and schema hashes. |
| Done   | `GET /v1/meta/datasets` | L+R     | Advertises which seasons/leagues/tables are currently loaded.       |

### 1. Players (People & Careers) - **(L)** with optional **(R)** joins ✓

| Status | Endpoint                                   | Dataset | Description                                                                  |
| ------ | ------------------------------------------ | ------- | ---------------------------------------------------------------------------- |
| Done   | `GET /v1/players`                          | L       | Search/browse players with filters (name, debut year, position, handedness). |
| Done   | `GET /v1/players/{player_id}`              | L       | Biographical data plus aggregated career stats.                              |
| Done   | `GET /v1/players/{player_id}/seasons`      | L       | Year-by-year batting and pitching splits.                                    |
| Done   | `GET /v1/players/{player_id}/teams`        | L       | List every team the player suited up for by season.                          |
| Done   | `GET /v1/players/{player_id}/awards`       | L       | Awards from Lahman `Awards*` tables with pagination.                         |
| Done   | `GET /v1/players/{player_id}/hall-of-fame` | L       | Hall of Fame voting/induction history.                                       |
| Done   | `GET /v1/players/{player_id}/salaries`     | L       | Salary history via Lahman `Salaries`.                                        |

#### Player Game & Play-by-Play Views - **(R)** ✓

| Status      | Endpoint                                        | Dataset | Description                                                                                |
| ----------- | ----------------------------------------------- | ------- | ------------------------------------------------------------------------------------------ |
| Done        | `GET /v1/players/{player_id}/game-logs`         | R       | Game-by-game performance from Retrosheet logs (currently starters only; see TODO).         |
| Done        | `GET /v1/players/{player_id}/appearances`       | R       | Detailed appearance records (positions, pinch roles) sourced from Retrosheet lineups.      |
| Done        | `GET /v1/players/{player_id}/plays`             | R       | All plays involving a player as batter or pitcher with Lahman ↔ Retrosheet ID mapping.     |
| Done        | `GET /v1/players/{player_id}/plate-appearances` | R       | Normalized plate appearance feed with batter-facing filters (season, pitcher, date range).  |

### 2. Teams, Franchises & Seasons - **(L)** + **(R)** ✓

#### Team & Franchise Reference ✓

| Status | Endpoint                         | Dataset | Description                                             |
| ------ | -------------------------------- | ------- | ------------------------------------------------------- |
| Done   | `GET /v1/teams`                  | L       | List team seasons with filters for year/league.         |
| Done   | `GET /v1/teams/{team_id}`        | L       | A single team-season record (wins, losses, runs, etc.). |
| Done   | `GET /v1/franchises`             | L       | Franchise catalog with active flag.                     |
| Done   | `GET /v1/franchises/{franch_id}` | L       | Franchise details and historical names.                 |
| Done   | `GET /v1/seasons`                | L       | Summary of available seasons (min/max year, leagues).   |

#### Team Rosters & Splits ✓

| Status | Endpoint                                          | Dataset | Description                                      |
| ------ | ------------------------------------------------- | ------- | ------------------------------------------------ |
| Done   | `GET /v1/seasons/{year}/teams`                    | L       | All teams for a season with aggregate stats.     |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/roster`   | L       | Player list with positions and high-level stats. |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/batting`  | L       | Aggregated batting stats plus per-player splits. |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/pitching` | L       | Aggregated pitching stats.                       |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/fielding` | L       | Aggregated fielding stats.                       |

#### Retrosheet Team Schedule & Logs - **(R)** ✓

| Status | Endpoint                                            | Dataset | Description                                  |
| ------ | --------------------------------------------------- | ------- | -------------------------------------------- |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/schedule`   | R       | Team calendar from Retrosheet schedules.     |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/games`      | R       | All games for a team/season with pagination. |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/daily-logs` | R       | Team daily performance rollups.              |

### 3. Games & Schedules - **(R)** ✓

| Status | Endpoint                                       | Dataset | Description                                                    |
| ------ | ---------------------------------------------- | ------- | -------------------------------------------------------------- |
| Done   | `GET /v1/games`                                | R       | Search games by season, teams, park, and date range.           |
| Done   | `GET /v1/games/{game_id}`                      | R       | Game metadata, score, and key events.                          |
| Done   | `GET /v1/games/{game_id}/boxscore`             | R       | Expanded boxscore with lineups and per-player lines.           |
| Done   | `GET /v1/games/{game_id}/summary`              | R       | Narrative summary (winning pitcher, save, highlights).         |
| Done   | `GET /v1/seasons/{year}/schedule`              | R       | Full season schedule with pagination.                          |
| Done   | `GET /v1/seasons/{year}/dates/{date}/games`    | R       | All games played on a calendar date.                           |
| Done   | `GET /v1/seasons/{year}/teams/{team_id}/games` | R       | Team-specific schedule (duplicate of table above for clarity). |
| Done   | `GET /v1/seasons/{year}/parks/{park_id}/games` | R       | Games played in a specific ballpark.                           |

### 4. Play-by-Play Events & Context - **(R)**

| Status | Endpoint                                        | Dataset | Description                                                                 |
| ------ | ----------------------------------------------- | ------- | --------------------------------------------------------------------------- |
| Done   | `GET /v1/plays`                                 | R       | Query parsed plays with batter/pitcher/team filters.                        |
| Done   | `GET /v1/games/{game_id}/plays`                 | R       | Chronological plays for a game (currently the canonical play-by-play feed). |
| Done   | `GET /v1/games/{game_id}/events`                | R       | Planned raw Retrosheet events (alias/extension on top of `/plays`).         |
| Done   | `GET /v1/games/{game_id}/events/{event_seq}`    | R       | Single event lookup with structured base/out state.                         |
| Done   | `GET /v1/players/{player_id}/plate-appearances` | R       | Player-level PA list with leverage, count, and vs. pitcher filters.         |
| To-Do  | `GET /v1/pitches`                               | R       | Optional dataset if we derive per-pitch signals.                            |

### 5. Parks, Umpires, Managers & Other Entities - **(L+R)**

| Status | Endpoint                                  | Dataset | Description                                              |
| ------ | ----------------------------------------- | ------- | -------------------------------------------------------- |
| Done   | `GET /v1/parks`                           | L+R     | Ballpark directory with locations and active years.      |
| Done   | `GET /v1/parks/{park_id}`                 | L+R     | Single ballpark plus games hosted.                       |
| Done   | `GET /v1/managers`                        | L       | Manager careers, totals, and teams managed.              |
| Done   | `GET /v1/managers/{manager_id}`           | L       | Detailed manager record.                                 |
| Done   | `GET /v1/managers/{manager_id}/seasons`   | L       | Season-by-season records for a manager.                  |
| Done   | `GET /v1/umpires`                         | L+R     | Umpire list.                                             |
| Done   | `GET /v1/umpires/{umpire_id}`             | L+R     | Umpire details + officiated games.                       |
| To-Do  | `GET /v1/ejections`                       | R       | All ejection events with filters (player, umpire, year). |
| To-Do  | `GET /v1/seasons/{year}/ejections`        | R       | Season-level slice of ejections.                         |

### 6. Stats & Leaderboards - **(L)** (with optional **(R)** joins)

#### Career & Season Stats

| Status | Endpoint                                     | Dataset | Description                                                           |
| ------ | -------------------------------------------- | ------- | --------------------------------------------------------------------- |
| Done   | `GET /v1/stats/batting`                      | L       | Flexible batting query (player/team/season filters, min AB, sorting). |
| Done   | `GET /v1/stats/pitching`                     | L       | Flexible pitching query (season ranges, min IP).                      |
| Done   | `GET /v1/stats/fielding`                     | L       | Fielding stats (positions, innings).                                  |
| To-Do  | `GET /v1/players/{player_id}/stats/batting`  | L+R     | Player-specific batting summary (career + optional splits).           |
| To-Do  | `GET /v1/players/{player_id}/stats/pitching` | L+R     | Player-specific pitching summary.                                     |

#### Seasonal Leaders ✓

| Status | Endpoint                                  | Dataset | Description                 |
| ------ | ----------------------------------------- | ------- | --------------------------- |
| Done   | `GET /v1/seasons/{year}/leaders/batting`  | L       | Leaders for HR/AVG/RBI/etc. |
| Done   | `GET /v1/seasons/{year}/leaders/pitching` | L       | Leaders for ERA/SO/W/etc.   |
| Done   | `GET /v1/leaders/batting/career`          | L       | Career batting leaders.     |
| Done   | `GET /v1/leaders/pitching/career`         | L       | Career pitching leaders.    |

#### Team-Level Stats ✓

| Status | Endpoint                       | Dataset | Description                               |
| ------ | ------------------------------ | ------- | ----------------------------------------- |
| Done   | `GET /v1/stats/teams/batting`  | L       | Team batting in a season (league filter). |
| Done   | `GET /v1/stats/teams/pitching` | L       | Team pitching.                            |
| Done   | `GET /v1/stats/teams/fielding` | L       | Team fielding.                            |

### 7. Awards, All-Star Games, Postseason - **(L)**

| Status | Endpoint                                   | Dataset | Description                                   |
| ------ | ------------------------------------------ | ------- | --------------------------------------------- |
| Done   | `GET /v1/awards`                           | L       | Browse awards data (MVP, Cy Young, ROY).      |
| Done   | `GET /v1/awards/{award_id}`                | L       | Detailed view for a specific award.           |
| Done   | `GET /v1/seasons/{year}/awards`            | L       | Awards issued during a season.                |
| Done   | `GET /v1/seasons/{year}/postseason/series` | L       | Postseason series list (LCS, WS, etc.).       |
| To-Do  | `GET /v1/seasons/{year}/postseason/games`  | L+R     | Postseason games joined with Retrosheet data. |
| To-Do  | `GET /v1/allstar/games`                    | L+R     | All-Star Game history.                        |
| To-Do  | `GET /v1/allstar/games/{game_id}`          | L+R     | Specific All-Star game box/events.            |

### 8. Search & Lookup Utilities - **(L+R)**

| Status | Endpoint                 | Dataset | Description                                                         |
| ------ | ------------------------ | ------- | ------------------------------------------------------------------- |
| To-Do  | `GET /v1/search/players` | L+R     | Fuzzy player search with optional era/league filters.               |
| To-Do  | `GET /v1/search/teams`   | L+R     | Search by name, city, franchise.                                    |
| To-Do  | `GET /v1/search/games`   | R       | Natural language queries such as “Yankees vs Red Sox 2003 ALCS G7.” |
| To-Do  | `GET /v1/search/parks`   | L+R     | Ballpark lookup by name/city.                                       |

### 9. Derived & Advanced Endpoints

- To-Do `/v1/players/{player_id}/streaks` - Track hitting or scoreless inning streaks.
- To-Do `/v1/players/{player_id}/splits` - Home/away, handedness, month, batting order splits.
- To-Do `/v1/teams/{team_id}/run-differential` - Season totals and rolling windows.
- To-Do `/v1/games/{game_id}/win-probability` - Win probability graphs derived from play-by-play.

### 10. Advanced Analytics & Enhancements

- To-Do Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset.
- To-Do OpenAPI enhancements (schemas per endpoint, examples) plus Markdown docs.
- To-Do Cache + rate limiting layer for public deployments.
- To-Do Performance testing and observability hooks before GA release.

## Milestones & Targets

| Status      | Milestone                              | Scope                                                                                                                 |
| ----------- | -------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| In-Progress | **M1 - Lahman ETL + Basic API**        | CLI + migrations + Lahman ingestion + players/teams/stats endpoints (implemented, needs monitoring/status reporting). |
| In-Progress | **M2 - Retrosheet Game Logs**          | Retrosheet game-log ingestion + game/schedule endpoints (loaders and endpoints exist; substitute handling still TBD). |
| In-Progress | **M3 - Play-by-Play**                  | Plays ingestion + `/plays` endpoints are live; need richer event context + player game-log parity.                    |
| To-Do       | **M4 - Joined & Advanced Experiences** | Lahman + Retrosheet joins, derived analytics, caching, and deployment hardening.                                      |
