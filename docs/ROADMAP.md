# Baseball API Development Roadmap

## Database Schema & Storage

- In-Progress:
    - Views and crosswalk tables should normalize Lahman ↔ Retrosheet IDs (`player_id_map`, `team_franchise_map`).
    - Pre-aggregate leaderboards for fast reads.
- To-Do:
    - Materialized views for heavy aggregates (win probability, streaks, etc.).

## ETL & CLI Tooling

- To-Do
    - Resumable download & play-by-play ingestion with checksum validation
    - Automated cron-style Taskfile targets

### API Platform, Docs & Ops

- To-Do: Add caching, and deployment scripts after the analytics milestone.

### Exposing Joined Data

- To-Do: Dedicated SQL views (e.g., `stats_career_batting`, `player_game_logs`) and ID-mapping helpers will simplify SQLC/hand-rolled queries.
- To-Do: Future endpoints that combine season stats with play-by-play data will rely on these views/materialized views.

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
| Done   | `/v1/players/{player_id}/streaks`      | Track hitting or scoreless inning streaks.          |
| Done   | `/v1/players/{player_id}/splits`       | Home/away, handedness, month, batting order splits. |
| Done   | `/v1/teams/{team_id}/run-differential` | Season totals and rolling windows.                  |
| Done   | `/v1/games/{game_id}/win-probability`  | Win probability graphs derived from play-by-play.   |

### 10. Advanced Analytics & Enhancements

| Status      | Description                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------- |
| To-Do       | Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset. |
| In-Progress | Markdown docs                                                                                |
| In-Progress | Cache + rate limiting layer for public deployments.                                          |
| To-Do       | Performance testing and observability hooks before GA release.                               |

## Milestones & Targets

| Status | Milestone                              | Scope                                                                                                                 |
| ------ | -------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| Done   | **M1 - Lahman ETL + Basic API**        | CLI + migrations + Lahman ingestion + players/teams/stats endpoints (implemented, needs monitoring/status reporting). |
| Done   | **M2 - Retrosheet Game Logs**          | Retrosheet game-log ingestion + game/schedule endpoints (loaders and endpoints exist; substitute handling still TBD). |
| Done   | **M3 - Play-by-Play**                  | Plays ingestion + `/plays` endpoints are live; need richer event context + player game-log parity.                    |
| To-Do  | **M4 - Joined & Advanced Experiences** | Lahman + Retrosheet joins, derived analytics, caching, and deployment hardening.                                      |
