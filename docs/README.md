# Baseball Stats API

## Architecture & Data Strategy

Postgres is the target store with reusable DDLs checked in under `internal/db/sql` (Lahman schema, Retrosheet schema, and plays schema derived from [Baseball-PostgreSQL](https://github.com/davidbmitchell/Baseball-PostgreSQL)).

### Three-Layer Approach

1. **Raw data → Postgres**: ETL pipelines for both the Lahman Baseball Database and Retrosheet archives.
2. **Postgres → Go domain model**: Typed queries, joins, and views for consistent access patterns.
3. **Go → HTTP API**: Versioned, cached, and documented endpoints layered on top of the domain model.

### Data Sources & IDs

- **Lahman** provides season and career aggregates (batting, pitching, fielding, awards, salaries, franchises, etc.) from 1871-2024. ([SABR Lahman Database](https://sabr.org/lahman-database/))
- **Retrosheet** publishes game logs and parsed play-by-play data. ([Retrosheet.org](https://www.retrosheet.org/)) Chadwick tools help convert raw event files and maintain crosswalks for IDs.
- Opinionated split: Lahman powers reference, leaderboards, and team/season stats, while Retrosheet powers games, logs, and play-by-play.
- ID conventions: `player_id` = Lahman `playerID`, `team_id` = Lahman `teamID` or `franchID`, `game_id` = Retrosheet `GAME_ID`. Chadwick mappings keep IDs consistent across sources.

### API Conventions

- Base URL: `https://baseball.stormlightlabs.org/api/v1/...` (prod) and `http://localhost:8080/v1/...` (dev).
- Common query params: `page`, `per_page`, `sort`, `order`, `from`/`to` (dates as `YYYY-MM-DD`), and stat filters such as `min_pa`, `min_ip`, `min_g`.
- Paginated responses wrap payloads inside an envelope containing `data`, `page`, `per_page`, and `total`.
- Default envelope metadata keeps clients backward compatible while letting us add new fields later.
