# Baseball API Development Roadmap

## Database Schema & Storage

- In-Progress: Pre-aggregate leaderboards for fast reads

### API Platform, Docs & Ops

- To-Do: Add deployment scripts

## API Roadmap

### 0. Meta / Utility ✓

See [Meta & Utility Overview](./api-meta-utility.md) for summary tables and expanded details.

### 1. Players (People & Careers) - **(L)** with optional **(R)** joins ✓

See [Players Overview](./api-players.md) for the combined Lahman and Retrosheet endpoint tables and call notes.

### 2. Teams, Franchises & Seasons - **(L)** + **(R)** ✓

See [Teams, Franchises & Seasons Overview](./api-teams.md) for tables covering references, splits, and Retrosheet logs.

### 3. Games & Schedules - **(R)** ✓

See [Games & Schedules Overview](./api-games.md) for the endpoint summary and usage notes.

### 4. Play-by-Play Events & Context - **(R)** ✓

See [Play-by-Play Events Overview](./api-play-by-play.md) for tables and deeper coverage of filters.

### 5. Parks, Umpires, Managers & Other Entities - **(L+R)** ✓

See [Parks, Umpires, Managers & Entity Overview](./api-parks-umpires-managers.md) for reference tables.

### 6. Stats & Leaderboards - **(L)** (with optional **(R)** joins) ✓

See [Stats & Leaderboards Overview](./api-stats.md) for the combined stats, leader, and team-level tables.

### 7. Awards, All-Star Games, Postseason - **(L)** ✓

See [Awards, All-Star & Postseason Overview](./api-awards-postseason.md) for the completed endpoint matrix.

### 8. Search & Lookup Utilities - **(L+R)** ✓

See [Search & Lookup Overview](./api-search.md) for the fuzzy lookup endpoint details.

### 9. Derived & Advanced Endpoints ✓

See [Derived & Advanced Endpoints Overview](./api-derived-advanced.md) for the analytics-heavy APIs and implementation details.

### 10. Advanced Analytics & Enhancements

| Status      | Description                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------- |
| Done        | Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset. |
| In-Progress | Markdown docs                                                                                |
| Done        | Cache + rate limiting layer for public deployments.                                          |
| Done        | Performance testing and observability hooks before GA release.                               |

### 11. Data Coverage Expansion - **(R)** ✓

See the dedicated Data Coverage docs for the newly completed endpoints:

- [Per-game Aggregations](./api-per-game-aggregations.md)
- [Game Context & Weather](./api-game-context.md)
- [League-specific Coverage](./api-league-coverage.md)
- [Achievements & Event Feeds](./api-achievements.md)

### 12. Optimizations

- Use partitioning strategies to optimize query performance, especially for large datasets like `plays`.
    - Do it by eras
