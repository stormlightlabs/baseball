# Baseball API Development Roadmap

## API Roadmap

### 0. Meta / Utility ✓ (Completed: 2025-12-12)

See [Meta & Utility Overview](../../web/src/routes/docs/api-meta-utility.md) for summary tables and expanded details.

### 1. Players (People & Careers) - **(L)** with optional **(R)** joins ✓ (Completed: 2025-12-12)

See [Players Overview](./api-players.md) for the combined Lahman and Retrosheet endpoint tables and call notes.

### 2. Teams, Franchises & Seasons - **(L)** + **(R)** ✓ (Completed: 2025-12-12)

See [Teams, Franchises & Seasons Overview](../../web/src/routes/docs/api-teams.md) for tables covering references, splits, and Retrosheet logs.

### 3. Games & Schedules - **(R)** ✓ (Completed: 2025-12-12)

See [Games & Schedules Overview](../../web/src/routes/docs/api-games.md) for the endpoint summary and usage notes.

### 4. Play-by-Play Events & Context - **(R)** ✓ (Completed: 2025-12-12)

See [Play-by-Play Events Overview](../../web/src/routes/docs/api-play-by-play.md) for tables and deeper coverage of filters.

### 5. Parks, Umpires, Managers & Other Entities - **(L+R)** ✓ (Completed: 2025-12-12)

See [Parks, Umpires, Managers & Entity Overview](../../web/src/routes/docs/api-parks-umpires-managers.md) for reference tables.

### 6. Stats & Leaderboards - **(L)** (with optional **(R)** joins) ✓ (Completed: 2025-12-12)

See [Stats & Leaderboards Overview](../../web/src/routes/docs/api-stats.md) for the combined stats, leader, and team-level tables.

### 7. Awards, All-Star Games, Postseason - **(L)** ✓ (Completed: 2025-12-12)

See [Awards, All-Star & Postseason Overview](../../web/src/routes/docs/api-awards-postseason.md) for the completed endpoint matrix.

### 8. Search & Lookup Utilities - **(L+R)** ✓ (Completed: 2025-12-12)

See [Search & Lookup Overview](../../web/src/routes/docs/api-search.md) for the fuzzy lookup endpoint details.

### 9. Derived & Advanced Endpoints ✓ (Completed: 2025-12-12)

See [Derived & Advanced Endpoints Overview](../../web/src/routes/docs/api-derived-advanced.md) for the analytics-heavy APIs and implementation details.

### 10. Advanced Analytics & Enhancements

| Status      | Description                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------- |
| Done        | Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset. |
| In-Progress | Markdown docs                                                                                |
| Done        | Cache + rate limiting layer for public deployments.                                          |
| Done        | Performance testing and observability hooks before GA release.                               |

### 11. Data Coverage Expansion - **(R)** ✓ (Completed: 2025-12-15)

See the dedicated Data Coverage docs for the newly completed endpoints:

- [Per-game Aggregations](../../web/src/routes/docs/api-per-game-aggregations.md)
- [Game Context & Weather](../../web/src/routes/docs/api-game-context.md)
- [League-specific Coverage](../../web/src/routes/docs/api-league-coverage.md)
- [Achievements & Event Feeds](../../web/src/routes/docs/api-achievements.md)

### 12. Optimizations ✓ (Completed: 2025-12-16)

- Era-based partitioning (61 partitions)
  - Partitioned by year:
    - pre-1914, 1914-1915 (Federal)
    - 1916-1934 (sparse)
    - 1935-1949 (Negro Leagues yearly)
    - 1950-1962 (Baby Boomer)
    - 1963-1968 (Pitcher),
    - 1969-1993 (Turf Time in 5-year chunks)
    - 1994-2025+ (yearly by era - steroid, moneyball, statcast)
- League column denormalization - Added league columns to plays table
  - Columns: `home_team_league`, `visiting_team_league`
  - Eliminates expensive joins with games table for league filtering and creates bitmap indexes for fast league lookups
- Implicit date filters for league queries - Enable partition pruning
- League mappings
  - 19th Century: UA (1884), PL (1890), AA (1882-1891)
  - Early 20th: FL (1914-1915)
  - Negro Leagues: NAL, NNL, NN2, ECL, ANL, EWL, NSL, IND (1935-1949)
  - Modern: AL/NL (no restriction - full historical range)

#### Maybes?

- **Filtered indexes** - Create partial indexes for specific league + date combinations if query patterns show benefit
- **Composite partition key** - Repartition by (league, year) only if league-specific queries dominate and current performance is insufficient

### 13. Release

Release planning for ETL safety/performance now follows a single operational track:

1. Baseball API ETL worker runtime:
   - [ETL Architecture Spec](../specs/etl.md)
   - [ETL Task List](../tasks/etl.md)
   - [Database Loading Contract](./data-loading.md)

Execution rule:

- `baseball-etl` owns source acquisition, load/upsert, validation, cleanup, and DB-side observability/safety rails.
- ETL execution should use batched queue semantics to keep shared VM resource usage bounded.
- Serving materialization remains a transition concern; bounded table/partition maintenance is preferred in worker jobs.

#### Deferred after ETL release scope

- Open API access model changes (auth/CORS/rate-limiter policy changes)
- General web SEO/discovery tasks (`robots.txt`, `sitemap.xml`, metadata/JSON-LD)
- Broad backend tech-debt cleanup and non-ETL refactors
