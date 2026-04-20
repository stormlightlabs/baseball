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

Goal: ship a stable, scalable release with hybrid incremental materialization, an open API access model, and production-safe VM/ETL operations.

#### Phase 0 - Completed baseline (Completed 2026-04-20)

- [x] Fix Postgres tuning delivery mechanism: use `command: postgres -c ...` instead of `POSTGRES_*` env no-ops
- [x] Tune Postgres baseline for 4 GB profile (`shared_buffers=1GB`, `effective_cache_size=2GB`, `work_mem=32MB`, WAL/checkpoint controls)
- [x] Add hard app DB pool limits (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, lifetime/idle-time caps)
- [x] Add Go runtime memory/CPU bounds (`GOMEMLIMIT`, `GOMAXPROCS`)
- [x] Add service-level memory/CPU/PID limits for app/postgres/redis
- [x] Add Postgres concurrency caps (`max_connections`, parallel worker limits)
- [x] Enable checkpoint visibility (`log_checkpoints=on`)

#### Phase 1 - Platform stability and crash prevention

- [ ] Remove hardcoded dev `DATABASE_URL` from Dockerfile ENV
- [ ] Add `HOST=0.0.0.0` to prod compose environment contract
- [ ] Add host-level alerting for free disk and WAL growth thresholds
- [ ] Add runbook actions for WAL pressure (pause ETL, archive/prune strategy, checkpoint analysis)
- [ ] Add periodic `pg_stat_bgwriter` capture for `checkpoints_req`/`checkpoints_timed` trend monitoring
- [ ] Add ETL concurrency guard (single active ETL run lock)
- [ ] Add per-step timeout/cancel policy for heavy operations
- [ ] Add off-peak scheduling recommendations and safe defaults for large force/year ranges
- [ ] Add optional load-shed mode for non-critical endpoints during ETL windows
- [ ] Add documented emergency toggles (disable heavy refresh groups, pause force mode)
- [ ] Add operational checklist for resume/recovery after interruption

#### Phase 2 - Hybrid materialization design and schema

- [ ] Finalize heavy artifact list for replacement (`player_game_*`, `team_game_stats`, `season_*_leaders`, `career_*_leaders`)
- [ ] Define source-of-truth model per artifact (`MV`, `incremental table`, or mixed)
- [ ] Define incremental keys and invalidation units (season/year/team/player)
- [ ] Define cutover SLOs (max refresh time, max lock time, acceptable staleness)
- [ ] Add structural migrations for incremental target tables (no historical rewrite in migrations)
- [ ] Add `etl_watermarks` / `materialization_state` tables for resumable progress tracking
- [ ] Add indexes/constraints for incremental upsert paths
- [ ] Keep migration set structural and idempotent only (`IF NOT EXISTS`, guarded DDL)

#### Phase 3 - Incremental ETL and read-path cutover

- [ ] Replace full-refresh ETL steps with year/season-bounded recompute + upsert steps
- [ ] Make force/year runs only recompute affected years/seasons
- [ ] Add phase-level ETL events and row-count metrics per artifact update step
- [ ] Add retry-safe transactional boundaries and resumability markers
- [ ] Add dual-read verification mode (MV vs table) for sampled requests
- [ ] Add integrity checks for totals/rate metrics across old/new artifacts
- [ ] Cut over repositories to incremental tables after parity is confirmed
- [ ] Keep fallback toggle to prior read-path during rollout window
- [ ] Add runbook queries for slow phases and stale watermarks
- [ ] Add perf baselines for ETL runtime and checkpoint frequency before/after cutover
- [ ] Document the hybrid strategy in ETL and deployment docs

#### Phase 4 - Open API access model

Target policy:

- API is open (no OAuth/session/API-key requirement)
- No API-level CORS restrictions for dashboard/web/mobile use
- No app-layer rate limiting for dashboard/web/mobile clients

Execution tasks:

- [ ] Remove auth requirement middleware from API route stack
- [ ] Remove `/v1/auth/*` surface
- [ ] Remove OAuth/session token validation paths from request flow
- [ ] Remove CORS middleware from API stack (or replace with explicit open policy where browser interoperability requires headers)
- [ ] Disable API-level rate limiting middleware for first-party clients; decide whether to remove it entirely
- [ ] Keep edge-level abuse controls outside app (reverse proxy/WAF) as optional guardrail
- [ ] Remove OAuth-only env vars from prod compose/env docs (`GITHUB_*`, `CODEBERG_*`) if auth is removed
- [ ] Remove auth/rate-limit specific config/env dependencies from startup and docs
- [ ] Update middleware/API tests for open-access behavior
- [ ] Run security review for open-access posture (DoS, scrape load, origin assumptions)

#### Phase 5 - Release readiness, docs, and quality cleanup

- [ ] Document required env vars for the open-access deployment model
- [ ] Add `robots.txt` handler (allow `/`, `/docs/`)
- [ ] Add `sitemap.xml` handler (static entries for home, examples, docs)
- [ ] Add `<meta name="description">`, OpenGraph tags (`og:title`, `og:description`, `og:type`, `og:url`), and `<link rel="canonical">` to `+layout.svelte`
- [ ] Add JSON-LD structured data (`WebAPI` / `WebApplication` schema) to web app.
- [ ] Custom error types: replace `strings.Contains(err.Error(), "not found")` with typed errors
- [ ] Remove global config state (`config.Get()` panic path) via dependency injection
- [ ] Complete TODOs: `cache/repository.go:121`, `core/mlb.go:164`, `api/plays_test.go:504`, `repository/computed.go:144`
- [ ] Improve godoc coverage for exported functions
- [ ] Extract filter parsing helpers to reduce handler duplication
- [ ] Add structured logging (replace `fmt.Printf` with `log/slog` or equivalent)
- [x] Fix `writeError` infinite recursion (`api/helpers.go`)
- [x] Standardize cache integration for Stats/Awards/Manager repos
- [x] Standardize dynamic query-building pattern
