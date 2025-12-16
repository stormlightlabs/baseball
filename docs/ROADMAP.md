# Baseball API Development Roadmap

## Database Schema & Storage

- In-Progress: Pre-aggregate leaderboards for fast reads

### API Platform, Docs & Ops

- To-Do: Add deployment scripts

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

### 9. Derived & Advanced Endpoints ✓

See [Derived & Advanced Endpoints Overview](./api-derived-advanced.md) for the analytics-heavy APIs and implementation details.

### 10. Advanced Analytics & Enhancements

| Status      | Description                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------- |
| Done        | Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset. |
| In-Progress | Markdown docs                                                                                |
| Done        | Cache + rate limiting layer for public deployments.                                          |
| Done        | Performance testing and observability hooks before GA release.                               |

### 11. Data Coverage Expansion - **(R)**

| Status | Source       | Endpoint(s)                             | Description                                                                                                       |
| ------ | ------------ | --------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Done   | plays (view) | `/v1/players/{id}/game-logs` (enh)      | Per-game batting stats via SQL views aggregating plays table. Enables "game finder" queries. Coverage: 1910-2025. |
|        |              | `/v1/players/{id}/stats/*` (enh)        |                                                                                                                   |
|        |              | `/v1/games/{id}/batting`                |                                                                                                                   |
| Done   | plays (view) | `/v1/players/{id}/stats/*` (enh)        | Per-game pitching stats via views. Enhances advanced stats, WAR calculations, player splits. Coverage: 1910-2025. |
|        |              | `/v1/games/{id}/pitching`               |                                                                                                                   |
| Done   | plays (view) | `/v1/players/{id}/stats/fielding` (enh) | Per-game fielding stats via views. Position-specific defensive metrics. Coverage: 1910-2025.                      |
|        |              | `/v1/games/{id}/fielding`               |                                                                                                                   |
| Done   | plays (view) | `/v1/teams/{id}/daily-stats`            | Per-game team stats via views. Daily performance tracking and rolling aggregates. Coverage: 1910-2025.            |
| To-Do  | games (enh)  | `/v1/games/{id}` (enh)                  | Enhanced game metadata already in games table: park, weather, attendance, game time, umpires.                     |
| Done   | League views | `/v1/federalleague/*`                   | Federal League endpoints implemented (1914-1915). Filters games/plays/teams by league='FL'.                       |
| Done   | League views | `/v1/negroleagues/*`                    | Negro Leagues endpoints (1935-1949). Same pattern as Federal League. See implementation plan below.               |
| Done   | Achievements | `/v1/achievements/*`                    | Computed from plays/games: no-hitters (446), cycles (90), multi-HR games (236), triple plays (143), 20+ inning games (3851). |

For each dataset expansion:

1. Create SQL views (aggregating from plays table)
2. Design and implement repository interfaces
3. Create and register handlers for new routes
4. Add Swagger annotations
5. Indexes
6. Materialized views

#### Notes

**Implementation Approach:**

- **Views-based**: Per-game stats derived from existing `plays` table via SQL views (no new tables needed)
- **No CSV loading**: Retrosheet CSVs are pre-aggregated summaries of plays data - we aggregate on-demand instead
- **Coverage**: Our plays table covers 1910-2025 (missing 1898-1909 from Retrosheet CSVs, acceptable trade-off)
- **Performance**: Can use materialized views if aggregation queries become slow

**Data Integration:**

- Games/plays already contain MLB, Negro Leagues, and Federal League data (or will when loaded)
- Filter by team codes, league identifiers, or date ranges to isolate specific leagues
- **Negro Leagues coverage**: 1920-1962 (limited by plays table availability)
- **Federal League coverage**: 1914-1915

**Query Capabilities:**

- **Per-game stats**: Enable "game finder" queries (e.g., "all games where player X hit 2+ home runs")
- **Game metadata**: Park, attendance, game time, umpires already in games table
- **Achievements**: Computed dynamically from plays/games (no-hitters, cycles, etc.)

**Data Refresh:**

- Automated weekly refresh during season recommended for current plays data

### 12. Optimizations

- Use partitioning strategies to optimize query performance, especially for large datasets like `plays`.
    - Do it by eras
