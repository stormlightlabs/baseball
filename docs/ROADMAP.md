# Baseball API Development Roadmap

## API Roadmap

### 0. Meta / Utility ✓ (Completed: 2025-12-12)

See [Meta & Utility Overview](./api-meta-utility.md) for summary tables and expanded details.

### 1. Players (People & Careers) - **(L)** with optional **(R)** joins ✓ (Completed: 2025-12-12)

See [Players Overview](./api-players.md) for the combined Lahman and Retrosheet endpoint tables and call notes.

### 2. Teams, Franchises & Seasons - **(L)** + **(R)** ✓ (Completed: 2025-12-12)

See [Teams, Franchises & Seasons Overview](./api-teams.md) for tables covering references, splits, and Retrosheet logs.

### 3. Games & Schedules - **(R)** ✓ (Completed: 2025-12-12)

See [Games & Schedules Overview](./api-games.md) for the endpoint summary and usage notes.

### 4. Play-by-Play Events & Context - **(R)** ✓ (Completed: 2025-12-12)

See [Play-by-Play Events Overview](./api-play-by-play.md) for tables and deeper coverage of filters.

### 5. Parks, Umpires, Managers & Other Entities - **(L+R)** ✓ (Completed: 2025-12-12)

See [Parks, Umpires, Managers & Entity Overview](./api-parks-umpires-managers.md) for reference tables.

### 6. Stats & Leaderboards - **(L)** (with optional **(R)** joins) ✓ (Completed: 2025-12-12)

See [Stats & Leaderboards Overview](./api-stats.md) for the combined stats, leader, and team-level tables.

### 7. Awards, All-Star Games, Postseason - **(L)** ✓ (Completed: 2025-12-12)

See [Awards, All-Star & Postseason Overview](./api-awards-postseason.md) for the completed endpoint matrix.

### 8. Search & Lookup Utilities - **(L+R)** ✓ (Completed: 2025-12-12)

See [Search & Lookup Overview](./api-search.md) for the fuzzy lookup endpoint details.

### 9. Derived & Advanced Endpoints ✓ (Completed: 2025-12-12)

See [Derived & Advanced Endpoints Overview](./api-derived-advanced.md) for the analytics-heavy APIs and implementation details.

### 10. Advanced Analytics & Enhancements

| Status      | Description                                                                                  |
| ----------- | -------------------------------------------------------------------------------------------- |
| Done        | Derived stats (WAR-like measures, leverage indexes) built atop the Retrosheet plays dataset. |
| In-Progress | Markdown docs                                                                                |
| Done        | Cache + rate limiting layer for public deployments.                                          |
| Done        | Performance testing and observability hooks before GA release.                               |

### 11. Data Coverage Expansion - **(R)** ✓ (Completed: 2025-12-15)

See the dedicated Data Coverage docs for the newly completed endpoints:

- [Per-game Aggregations](./api-per-game-aggregations.md)
- [Game Context & Weather](./api-game-context.md)
- [League-specific Coverage](./api-league-coverage.md)
- [Achievements & Event Feeds](./api-achievements.md)

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

### 14. Production Deployment (Coolify/Hostinger KVM2)

Target: ~100 GB HDD VPS, 4 GB RAM, Coolify dashboard with Traefik for TLS/routing.

#### Blocking - Dockerfile & Compose

- [ ] Copy templates and static assets into Docker image (`internal/api/templates/`, `internal/api/static/` are loaded at runtime but not included in the build stage)
- [ ] Bind server to `0.0.0.0` inside the container (config default is `localhost`, unreachable through Traefik)
- [ ] Fix Postgres tuning: `POSTGRES_SHARED_BUFFERS` etc. are env-var no-ops on the official image; pass via `command: postgres -c shared_buffers=1GB ...` or mount a `postgresql.conf`
- [ ] Remove hardcoded dev `DATABASE_URL` from Dockerfile ENV (leaks creds into image layer, confusing if runtime env is missing)
- [ ] Add `GITHUB_REDIRECT_URL`, `CODEBERG_REDIRECT_URL` to prod compose and README env table (defaults are `localhost`)
- [ ] Add `HOST=0.0.0.0` to prod compose environment

#### Blocking - Auth & Security

- [ ] Fix middleware ordering: rate limiter runs before auth, so `api_key` context is always nil and all requests hit the 10 req/min unauthenticated limit
- [ ] Fix session cookie `Secure` flag: behind Traefik, `req.TLS` is always nil; check `X-Forwarded-Proto == "https"` or force `Secure: true` in non-debug mode
- [ ] Add security headers middleware: `Strict-Transport-Security`, `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Content-Security-Policy`, `Referrer-Policy`
- [ ] Add CORS middleware for cross-origin API consumers

#### Blocking - README accuracy

- [ ] Document all required env vars: `HOST`, `DEBUG_MODE`, `GITHUB_REDIRECT_URL`, `CODEBERG_REDIRECT_URL`
- [ ] Clarify which seed data is checked in (Lahman, FanGraphs, salaries) vs fetched at runtime (Retrosheet)
- [ ] Tune Postgres for 4 GB RAM: `shared_buffers=1GB`, `effective_cache_size=2GB`, `work_mem=64MB`

#### SEO

- [ ] Add `robots.txt` handler (allow `/`, `/docs/`, `/examples`; disallow `/v1/auth/`, `/debug/`)
- [ ] Add `sitemap.xml` handler (static entries for home, examples, docs, login)
- [ ] Add `<meta name="description">`, OpenGraph tags (`og:title`, `og:description`, `og:type`, `og:url`), and `<link rel="canonical">` to `base.html`
- [ ] Add JSON-LD structured data (`WebAPI` / `WebApplication` schema) to `home.html`

#### Post-deploy ETL sequence

- [ ] Draft a shell script that runs all the ETL steps

### 15. Technical Debt

- [ ] Custom error types: Replace `strings.Contains(err.Error(), "not found")` with typed errors
- [ ] Fix `writeError` infinite recursion: `api/helpers.go:47` calls itself instead of `writeInternalServerError` — any non-`NotFoundError` triggers a stack overflow crash
- [ ] Remove global config state: `config.Get()` panics, refactor to dependency injection
- [x] Standardize cache integration: Measure and add caching to Stats/Awards/Manager repos
- [x] Standardize query building: Pick one pattern for dynamic WHERE clauses
- [ ] Complete TODOs: cache/repository.go:121, core/mlb.go:164, api/plays_test.go:504, repository/computed.go:144
- [ ] Improve godoc coverage: Document all exported functions
- [ ] Extract filter parsing helpers: Reduce duplication in handler filter building
- [ ] Add structured logging: Replace fmt.Printf with proper charmbracelet/log & slog
