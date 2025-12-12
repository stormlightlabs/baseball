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

See [Meta & Utility API Overview](./api-meta-utility.md) for summary tables and expanded details.

### 1. Players (People & Careers) - **(L)** with optional **(R)** joins ✓

See [Players API Overview](./api-players.md) for the combined Lahman and Retrosheet endpoint tables and call notes.

### 2. Teams, Franchises & Seasons - **(L)** + **(R)** ✓

See [Teams, Franchises & Seasons API Overview](./api-teams.md) for tables covering references, splits, and Retrosheet logs.

### 3. Games & Schedules - **(R)** ✓

See [Games & Schedules API Overview](./api-games.md) for the endpoint summary and usage notes.

### 4. Play-by-Play Events & Context - **(R)** ✓

See [Play-by-Play Events API Overview](./api-play-by-play.md) for tables and deeper coverage of filters.

### 5. Parks, Umpires, Managers & Other Entities - **(L+R)** ✓

See [Parks, Umpires, Managers & Entity API Overview](./api-parks-umpires-managers.md) for reference tables.

### 6. Stats & Leaderboards - **(L)** (with optional **(R)** joins) ✓

See [Stats & Leaderboards API Overview](./api-stats.md) for the combined stats, leader, and team-level tables.

### 7. Awards, All-Star Games, Postseason - **(L)** ✓

See [Awards, All-Star & Postseason API Overview](./api-awards-postseason.md) for the completed endpoint matrix.

### 8. Search & Lookup Utilities - **(L+R)** ✓

See [Search & Lookup API Overview](./api-search.md) for the fuzzy lookup endpoint details.

### 9. Derived & Advanced Endpoints

| Status | Endpoint                               | Notes                                               |
| ------ | -------------------------------------- | --------------------------------------------------- |
| To-Do  | `/v1/players/{player_id}/streaks`      | Track hitting or scoreless inning streaks.          |
| To-Do  | `/v1/players/{player_id}/splits`       | Home/away, handedness, month, batting order splits. |
| To-Do  | `/v1/teams/{team_id}/run-differential` | Season totals and rolling windows.                  |
| To-Do  | `/v1/games/{game_id}/win-probability`  | Win probability graphs derived from play-by-play.   |

### 10. Advanced Analytics & Enhancements

| Status      | Description                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------- |
| To-Do       | Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset. |
| Done        | OpenAPI docs                                                                                 |
| In-Progress | Markdown docs                                                                                |
| In-Progress | Cache + rate limiting layer for public deployments.                                          |
| To-Do       | Performance testing and observability hooks before GA release.                               |

## Milestones & Targets

| Status      | Milestone                              | Scope                                                                                                                 |
| ----------- | -------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| In-Progress | **M1 - Lahman ETL + Basic API**        | CLI + migrations + Lahman ingestion + players/teams/stats endpoints (implemented, needs monitoring/status reporting). |
| In-Progress | **M2 - Retrosheet Game Logs**          | Retrosheet game-log ingestion + game/schedule endpoints (loaders and endpoints exist; substitute handling still TBD). |
| In-Progress | **M3 - Play-by-Play**                  | Plays ingestion + `/plays` endpoints are live; need richer event context + player game-log parity.                    |
| To-Do       | **M4 - Joined & Advanced Experiences** | Lahman + Retrosheet joins, derived analytics, caching, and deployment hardening.                                      |
