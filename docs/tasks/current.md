---
title: Current Season Pipeline Tasks
updated: 2026-04-21
---

- Ref spec: `docs/specs/current.md`
- Stack: Go 1.24, PostgreSQL, `robfig/cron/v3`, existing ETL infrastructure
- Backend: `baseball-etl cron` subcommand, `current_season` schema, API integration
- Scope: ETL cron scheduler + current-season data ingestion + API merging + UI integration

## Phase 0: Schema & Migration

- [x] Create migration `015_current_season.sql`:
  - `CREATE SCHEMA IF NOT EXISTS current_season`
  - `current_season.batting` table (see spec for columns)
  - `current_season.pitching` table
  - `current_season.standings` table
  - `current_season.games` table
  - Indexes on `game_date`, `season`, composite PKs
- [x] Add `current-season-sync` to `etl_jobs.job_type` CHECK constraint
- [x] Run migration locally and verify schema exists

Acceptance:

- [x] `\dt current_season.*` shows all four tables
- [x] `etl_jobs` accepts `job_type = 'current-season-sync'`

## Phase 1: Cron Scheduler in ETL Binary

- [ ] Add `robfig/cron/v3` dependency
- [ ] Create `internal/cli/etl_cron.go`:
  - `EtlCronCmd()` → `baseball-etl cron`
  - Flags: `--schedule`, `--profile`, `--config` (TOML schedule file)
  - Starts embedded cron scheduler + worker loop in same process
  - Scheduler enqueues `current-season-sync` jobs into `etl_jobs` via existing `EnqueueETLJob()`
  - De-duplication guard: skip enqueue if a job of same type is already queued/running
- [ ] Add `EtlCronCmd()` to `ETLCmd()` subcommand tree in `internal/cli/etl.go`
- [ ] Add TOML config parsing for multi-schedule support:

  ```toml
  [current_season]
  enabled = true
  season = 2026
  cron_stats = "0 */4 * * *"
  cron_standings = "0 * * * *"
  cron_schedule = "0 6 * * *"
  cron_rosters = "0 8 * * *"
  active_window = "03-20/11-15"
  ```

- [ ] Wire active-window check: cron ticks outside the configured date range are no-ops
- [ ] Add graceful shutdown (context cancellation stops cron + drains active job)

Acceptance:

- [ ] `baseball-etl cron --schedule "*/5 * * * *" --profile dev` enqueues a job every 5 minutes
- [ ] Duplicate enqueue is skipped when a job is already queued
- [ ] `Ctrl+C` shuts down cleanly

## Phase 2: Current-Season Sync Job

- [ ] Create `internal/seed/current_season.go`:
  - `SyncCurrentSeasonStats(ctx, db, season, httpClient)` — fetches batting + pitching stats from MLB API, upserts into `current_season.batting`/`pitching`
  - `SyncCurrentSeasonStandings(ctx, db, season, httpClient)` — fetches standings, upserts into `current_season.standings`
  - `SyncCurrentSeasonSchedule(ctx, db, season, httpClient)` — fetches schedule + scores, upserts into `current_season.games`
  - `SyncCurrentSeasonRosters(ctx, db, season, httpClient)` — fetches active rosters (used for crosswalk gap detection)
- [ ] MLB API response parsing:
  - Define Go structs for MLB Stats API response shapes (stats, standings, schedule)
  - Validate response shape before upsert (log + skip malformed entries)
- [ ] Crosswalk enrichment:
  - After fetching MLBAM data, join with `crosswalk_mlbam` to populate `player_id`, `team_id`, `franchise_id`
  - Log unmatched MLBAM IDs for crosswalk gap reporting
- [ ] Wire into ETL worker: when `job_type = 'current-season-sync'`, route to `SyncCurrentSeason*` based on job `scope.sync_type` (stats, standings, schedule, rosters, or all)
- [ ] Add `fetched_at` tracking — upserts update this column so the API can report freshness

Acceptance:

- [ ] `baseball-etl run --profile current-season --years 2026` enqueues and processes a full current-season sync
- [ ] `current_season.batting` has rows with populated `player_id` for crosswalked players
- [ ] `current_season.standings` reflects current MLB standings
- [ ] `current_season.games` has game results for completed games

## Phase 3: API Integration

### Player Stats Merge

- [ ] Modify `internal/repository/players.go` (or equivalent):
  - `GetPlayerBattingStats()` unions `Batting` (Lahman) with `current_season.batting` for seasons not present in Lahman
  - `GetPlayerPitchingStats()` same pattern for pitching
  - Add `source` field to response (`"lahman"`, `"retrosheet"`, or `"current_season"`)
- [ ] Modify `GET /v1/players/{id}/stats/batting` and `/stats/pitching` handlers to include current-season rows

### Leaderboards

- [ ] Modify `GET /v1/seasons/{year}/leaders/*`:
  - When `year` matches configured current season and `current_season.*` tables have data, query from them
  - Return results with `source: "current_season"` attribution

### Standings Endpoint

- [ ] Add `GET /v1/standings?season={year}`:
  - For current season: reads from `current_season.standings`
  - For historical seasons: reads from Lahman team records (or returns 404 if no standings data)
  - Include `last_updated` from `fetched_at`
- [ ] Add swagger annotations

### Schedule/Games Merge

- [ ] Modify `GET /v1/seasons/{year}/schedule`:
  - For current season: reads from `current_season.games`
  - Merge with any Retrosheet games if partial overlap exists
- [ ] Modify `GET /v1/games` filter:
  - Accept current-season `game_pk` as game IDs for games without Retrosheet IDs

### Meta/Observability

- [ ] Extend `GET /v1/meta/datasets` to include `current_season.*` table stats (row counts, `MAX(fetched_at)`)
- [ ] Extend `GET /v1/meta/readiness` to flag `current_season_stale` when data is older than 2× configured refresh interval
- [ ] Update swagger annotations for all modified/new endpoints
- [ ] Run `task swagger:generate`

Acceptance:

- [ ] `GET /v1/players/{id}/stats/batting` returns current-season row for an active player
- [ ] `GET /v1/seasons/2026/leaders/batting` returns ranked current-season leaders
- [ ] `GET /v1/standings?season=2026` returns division standings
- [ ] `GET /v1/meta/datasets` shows `current_season` table freshness
- [ ] Swagger docs are updated

## Phase 4: Season Handoff Procedure

- [ ] Add `baseball-etl current-season handoff` subcommand:
  - Validates that canonical Retrosheet + Lahman data for the given year is loaded
  - Compares row counts between `current_season.batting` and Lahman `Batting` for the handoff year
  - Reports discrepancies (players in current-season but not Lahman, and vice versa)
  - With `--execute` flag: truncates `current_season` schema tables
  - Without `--execute`: dry-run report only
- [ ] Add handoff documentation to `docs/internal/etl.md`

Acceptance:

- [ ] `baseball-etl current-season handoff --year 2025` reports comparison without modifying data
- [ ] `baseball-etl current-season handoff --year 2025 --execute` truncates current-season tables after validation

## Phase 5: Web Dashboard Updates

Ref designs: `docs/designs/current-standings.html`, `docs/designs/current-leaders.html`

### Standings

- [ ] Add `/standings` route (or segment under `/teams`):
  - Fetch `GET /api/v1/standings?season={current}`
  - Render division-grouped tables: Rank, Team, W, L, PCT, GB, WC GB, Streak, L10, Run Diff
  - Team names link to `/teams/[franchise_id]` with current season
  - Segment control: AL / NL / Both
  - "Updated {time}" footer from `last_updated`
- [ ] Add standings link to home page and teams page nav

### Season Leaders (Local)

- [ ] Update `LeaderCards` component to detect current season:
  - When season matches current year, fetch from `/api/v1/seasons/{year}/leaders/*` (local data) instead of `/v1/mlb/stats`
  - Show "Updated every 4h" attribution instead of "Live from MLB"
  - Fallback to MLB proxy if local data is empty

### Player Current Season Badge

- [ ] In player stats tables, add "current season" visual indicator for rows where `source = "current_season"`
- [ ] Show `fetched_at` as "Updated {relative_time}" tooltip

### Season 2026 Page

- [ ] Ensure `/seasons/2026` route works with current-season data:
  - Schedule calendar populated from `current_season.games`
  - Standings widget from `current_season.standings`
  - Leaders from `current_season.batting`/`pitching`
  - "Data refreshes every 4 hours" notice banner

Acceptance:

- [ ] Standings page renders with current-season data
- [ ] Season leaders load from local data with update attribution
- [ ] Player detail shows current-season stats with source badge
- [ ] `/seasons/2026` is fully functional

## Phase 6: Mobile Updates

### Home Tab

- [ ] Season leaders section reads from persisted API (not live proxy):
  - Uses same `GET /api/v1/seasons/{year}/leaders/*` endpoints as web
  - Cache in Hive for offline access
  - "Updated {time}" display
- [ ] Scoreboard continues using live proxy for real-time scores

### Players Tab

- [ ] Player detail batting/pitching stats include current-season row automatically:
  - No client-side changes needed if API integration (Phase 3) is correct
  - Add source badge rendering: "2026 (current)" chip on current-season stat rows

### Teams Tab

- [ ] Current Standings segment reads from `GET /api/v1/standings?season=2026`:
  - Faster than live proxy, offline-capable via Hive cache
  - Same visual as existing standings design

### Games Tab

- [ ] Game finder supports `season=2026` filter:
  - Returns results from `current_season.games`
  - Game cards render identically to historical games

Acceptance:

- [ ] Mobile home tab leaders load from local current-season data
- [ ] Player detail shows current-season stats with badge
- [ ] Standings work offline from cached local data
- [ ] Game finder returns current-season results

## Phase 7: Testing & Validation

- [ ] Unit tests for `internal/seed/current_season.go`:
  - Mock MLB API responses
  - Verify upsert behavior (insert, update, crosswalk enrichment)
  - Verify malformed response handling (skip bad entries, log warnings)
- [ ] Integration tests:
  - Full sync cycle against test database
  - Verify API merge behavior (historical + current-season rows)
  - Verify standings endpoint response shape
  - Verify leaderboard queries against current-season data
- [ ] Cron scheduler tests:
  - Verify enqueue de-duplication
  - Verify active-window filtering
  - Verify graceful shutdown
- [ ] End-to-end:
  - Run `baseball-etl cron` locally, verify periodic sync
  - Hit API endpoints, verify current-season data appears
  - Run `flutter analyze` after mobile changes

Acceptance:

- [ ] All unit tests pass
- [ ] Integration tests cover the sync → query → render path
- [ ] `flutter analyze` clean
- [ ] `pnpm lint` clean in `web/`
