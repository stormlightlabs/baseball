---
title: Current Season Pipeline
updated: 2026-04-21
---

## Context

The platform has two data regimes today:

1. **Historical** - Retrosheet play-by-play + Lahman career stats, loaded via the ETL pipeline (`baseball-etl`).
   Covers 1871–2025 after each off-season data release.
2. **Live proxy** - `/v1/mlb/*` proxies `statsapi.mlb.com` with Redis caching (120s TTL).
   Gives real-time scores, standings, and player lookups but produces **no local dataset** - every request hits upstream (or cache).

The gap: during the regular season (late March → early October) we have no _persisted_ current-season data in PostgreSQL.
Current-season queries either:

- fall through to the MLB proxy (ephemeral, no aggregation, no join with historical data), or
- simply don't exist until Retrosheet publishes end-of-year game logs and Lahman ships its annual update.

This spec describes a **current-season pipeline** that bridges that gap: a scheduled ETL process that periodically ingests MLB Stats API data into local tables, making current-season batting, pitching, standings, and schedule data queryable through the same API surface as historical data.

## Goals

1. **Cron-capable ETL** - extend `baseball-etl` with a built-in scheduler so it can run periodic jobs without an external cron/Kubernetes CronJob dependency.
2. **Persisted current-season dataset** - ingest MLB Stats API data into PostgreSQL tables that mirror the shape of Lahman/Retrosheet season stats, enabling unified queries across historical and current-season data.
3. **Seamless API surface** - current-season data appears through existing `/v1/players/*/stats/*`, `/v1/stats/*`, `/v1/seasons/{year}/*` endpoints without special client logic.
4. **Clean handoff** - when Retrosheet/Lahman publish end-of-year data, the current-season dataset is replaced by the canonical historical load. No permanent schema changes, no data duplication after handoff.

## Non-goals

- Replacing the live MLB proxy - real-time scores, live game feeds, and in-game data continue to flow through `/v1/mlb/*`.
- Sub-minute freshness - the pipeline targets hourly-to-daily refresh, not live updates.
- Statcast-level pitch data - current-season ingestion covers box-score-level stats, not pitch-by-pitch.

## Architecture

### ETL Cron Scheduler

The `baseball-etl` binary gains a new `cron` subcommand that runs an embedded scheduler alongside the existing worker loop.

```sh
baseball-etl cron --schedule "0 */4 * * *" --profile current-season
```

**Design decisions:**

- Uses `robfig/cron/v3` for cron expression parsing and scheduling - battle-tested, zero external dependencies beyond the library.
- The cron scheduler enqueues jobs into the existing `etl_jobs` queue. It does not bypass the worker - it _feeds_ it. This preserves all existing job lifecycle (retry, status tracking, concurrency limits).
- A single `baseball-etl cron` process combines the scheduler and worker. It runs the scheduler goroutine that enqueues on schedule, and the worker loop that dequeues and executes.
- Multiple cron expressions can be configured for different profiles/scopes via a TOML schedule file or repeated `--schedule` flags.
- Guard: the scheduler checks for existing queued/running jobs of the same type before enqueuing a duplicate. Skips if one is already pending.

**New job type:** `current-season-sync`

```sql
ALTER TABLE etl_jobs
  ALTER COLUMN job_type TYPE TEXT
  CHECK (job_type IN ('full-run', 'yearly-sync', 'validate-only', 'cleanup-only', 'maintenance', 'current-season-sync'));
```

### Data Model

#### Source: MLB Stats API

The pipeline fetches from `statsapi.mlb.com` (through our existing proxy infrastructure for caching/rate-limiting):

| Data              | MLB Endpoint                                                      | Refresh cadence |
| ----------------- | ----------------------------------------------------------------- | --------------- |
| Season batting    | `/v1/stats?stats=season&group=hitting&season={Y}&playerPool=all`  | 4h              |
| Season pitching   | `/v1/stats?stats=season&group=pitching&season={Y}&playerPool=all` | 4h              |
| Standings         | `/v1/standings?season={Y}&standingsTypes=regularSeason`           | 1h              |
| Schedule + scores | `/v1/schedule?season={Y}&sportId=1&hydrate=linescore,team`        | 1h              |
| Roster/active     | `/v1/teams/{id}/roster?rosterType=active&season={Y}`              | 24h             |
| Transactions      | `/v1/transactions?startDate=...&endDate=...`                      | 24h             |

#### Destination: PostgreSQL

New tables under a `current_season` schema to avoid collision with Lahman/Retrosheet canonical tables:

```sql
CREATE SCHEMA IF NOT EXISTS current_season;

-- Player season batting stats (mirrors Lahman "Batting" shape)
CREATE TABLE current_season.batting (
    mlb_id          INTEGER NOT NULL,
    player_id       VARCHAR,          -- lahman ID via crosswalk, nullable until matched
    season          INTEGER NOT NULL,
    team_mlb_id     INTEGER NOT NULL,
    team_id         VARCHAR,          -- local team ID via crosswalk
    g   INT, pa  INT, ab  INT, r   INT, h   INT,
    "2b" INT, "3b" INT, hr  INT, rbi INT, sb  INT, cs  INT,
    bb  INT, so  INT, hbp INT, sf  INT, sh  INT,
    avg NUMERIC(4,3), obp NUMERIC(4,3), slg NUMERIC(4,3), ops NUMERIC(4,3),
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (mlb_id, season, team_mlb_id)
);

-- Player season pitching stats (mirrors Lahman "Pitching" shape)
CREATE TABLE current_season.pitching (
    mlb_id          INTEGER NOT NULL,
    player_id       VARCHAR,
    season          INTEGER NOT NULL,
    team_mlb_id     INTEGER NOT NULL,
    team_id         VARCHAR,
    w INT, l INT, g INT, gs INT, sv INT, ip NUMERIC(5,1),
    h INT, r INT, er INT, hr INT, bb INT, so INT, hbp INT,
    era NUMERIC(4,2), whip NUMERIC(4,2),
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (mlb_id, season, team_mlb_id)
);

-- Standings snapshot
CREATE TABLE current_season.standings (
    season          INTEGER NOT NULL,
    division_id     INTEGER NOT NULL,
    division_name   VARCHAR NOT NULL,
    team_mlb_id     INTEGER NOT NULL,
    team_id         VARCHAR,
    franchise_id    VARCHAR,
    w INT, l INT, pct NUMERIC(4,3), gb VARCHAR,
    wc_gb VARCHAR, streak VARCHAR, l10 VARCHAR,
    run_diff INT, rs INT, ra INT,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (season, team_mlb_id)
);

-- Game results (box-score level, not play-by-play)
CREATE TABLE current_season.games (
    game_pk         INTEGER PRIMARY KEY,
    season          INTEGER NOT NULL,
    game_date       DATE NOT NULL,
    status          VARCHAR NOT NULL,     -- Final, In Progress, Scheduled, etc.
    away_mlb_id     INTEGER NOT NULL,
    away_team_id    VARCHAR,
    away_score      INT,
    home_mlb_id     INTEGER NOT NULL,
    home_team_id    VARCHAR,
    home_score      INT,
    venue           VARCHAR,
    innings         INT,
    day_night       VARCHAR,
    doubleheader    VARCHAR,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cs_games_date ON current_season.games(game_date);
CREATE INDEX idx_cs_games_season ON current_season.games(season);
```

#### Crosswalk Integration

The pipeline uses `crosswalk_mlbam` (already loaded by the existing ETL) to map MLBAM IDs to local Lahman/Retrosheet IDs. Where crosswalk entries exist, `player_id`, `team_id`, and `franchise_id` columns are populated. Where they don't (rookies, new call-ups), those columns remain NULL and the MLBAM ID is the primary key.

The API layer joins `current_season.*` tables with the crosswalk to serve unified responses. A player with both historical and current-season data gets both in a single response.

### API Integration

Current-season data merges into existing endpoints transparently:

| Endpoint family                          | Behavior during active season                                                           |
| ---------------------------------------- | --------------------------------------------------------------------------------------- |
| `GET /v1/players/{id}/stats/batting`     | Appends current-season row (source: `current_season.batting`) to historical Lahman rows |
| `GET /v1/players/{id}/stats/pitching`    | Same pattern for pitching stats                                                         |
| `GET /v1/stats/batting?season={current}` | Falls through to `current_season.batting` when requested season has no Lahman data      |
| `GET /v1/seasons/{year}/leaders/*`       | Queries `current_season.batting`/`pitching` for the current year's leaderboards         |
| `GET /v1/standings`                      | New endpoint backed by `current_season.standings`                                       |
| `GET /v1/seasons/{year}/schedule`        | For current season, reads exclusively from `current_season.games` (no Retrosheet merge) |

**Source attribution:** API responses include a `source` field (`"retrosheet"`, `"lahman"`, or `"current_season"`) so clients can distinguish data provenance.
**Schedule payload guardrail:** current-season schedule endpoints must not return mixed-source payloads. A response for the active season is sourced entirely from `current_season.games`, while historical season responses continue to use historical sources.

### Lifecycle: Season Handoff

```text
Mar         Apr ────────────── Sep    Oct         Dec/Jan
 │           │                  │      │            │
 │  current-season-sync cron    │      │  Retrosheet/Lahman
 │  starts (Opening Day)        │      │  annual release
 │           ├──────────────────┤      │            │
 │           │  current_season  │      │  TRUNCATE current_season.*
 │           │  tables active   │      │  Load canonical historical
 │           │                  │      │  data via normal ETL
 └───────────┴──────────────────┴──────┴────────────┘
```

- **Start:** Cron begins running on or before Opening Day. First sync populates full season schedule + rosters.
- **During season:** Periodic syncs update stats, standings, and game results. Frequency configurable (default: every 4 hours for stats, every hour for standings/scores).
- **Postseason:** Pipeline continues through World Series.
- **Off-season handoff:** When Retrosheet/Lahman release annual data, the normal ETL pipeline loads canonical historical data. After successful load, `current_season` schema is truncated. The cron schedule is paused (or the `baseball-etl cron` process is stopped).

### Configuration

```toml
[current_season]
enabled = true
season = 2026
cron_stats = "0 */4 * * *"      # every 4 hours
cron_standings = "0 * * * *"     # every hour during game hours
cron_schedule = "0 6 * * *"      # daily at 6 AM ET
cron_rosters = "0 8 * * *"       # daily at 8 AM ET
active_window = "03-20/11-15"    # only run between these dates
```

## Web Dashboard Integration

### Standings Page (`/standings` or `/teams` tab)

The existing Teams page gets a "Current Standings" segment (already spec'd in `docs/specs/dash.md`). With the current-season pipeline, this data comes from local PostgreSQL rather than hitting the MLB proxy on every page load.

- **Before this pipeline:** standings came from `GET /v1/mlb/standings` (proxied, ephemeral)
- **After:** standings come from `GET /v1/standings?season=2026` backed by `current_season.standings` (persisted, joined with franchise/team crosswalks)

### Current Season Leaders

The home page's "Today's Leaders" section can switch from proxied MLB stats to local queries:

- `GET /v1/seasons/2026/leaders/batting` → queries `current_season.batting`
- `GET /v1/seasons/2026/leaders/pitching` → queries `current_season.pitching`

Renders identically to historical leaderboards, with a "Current Season - updated {fetched_at}" attribution.

### Player Current Season Card

The Player Explorer's "Current Season" card (spec'd in `docs/specs/dash.md`) now reads from local data:

- `GET /v1/players/{id}/stats/batting` returns a current-season row alongside historical rows
- The dashboard renders this row with a "2026 (current)" badge and last-updated timestamp

### Season Page (`/seasons/2026`)

The Season page for the current year works like any historical season page:

- Schedule calendar from `current_season.games`
- Team standings from `current_season.standings`
- Leader boards from `current_season.batting`/`pitching`
- "Data refreshes every 4 hours" notice

## Mobile Integration

### Home Tab

- Scoreboard continues using the live proxy for real-time scores during games
- "Season Leaders" section below scoreboard reads from persisted current-season data (faster, works offline)
- Cached in Hive for offline access with "Last updated" timestamp

### Players Tab

- Player detail batting/pitching stats include the current season row
- Source badge: "2026 season - updated 4h ago"
- Same data shape as historical seasons - no special client handling needed

### Teams Tab

- "Current Standings" segment reads from persisted standings
- Faster load, offline-capable, consistent with historical team views
- Tap team → team detail with current season pre-selected

### Games Tab

- Current season schedule/results available in game finder
- Filter by "2026" season works like any other year
- Final games show box-score data from `current_season.games`

## Observability

- ETL job status visible via existing `GET /v1/meta/datasets` - add `current_season.*` table row counts and last-refresh timestamps.
- `GET /v1/meta/readiness` reports `current_season_stale: true` if the most recent sync is older than 2× the configured interval.
- Worker logs tagged with `job_type=current-season-sync` for filtering.
- Prometheus-compatible metrics: `etl_current_season_sync_duration_seconds`, `etl_current_season_rows_upserted`, `etl_current_season_last_success_timestamp`.

## Risks and Mitigations

| Risk                                      | Mitigation                                                                                                       |
| ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| MLB Stats API rate limiting               | Use existing Redis cache layer; pipeline fetches are batched and scheduled, not per-request                      |
| Crosswalk gaps (rookies, callups)         | Serve data by MLBAM ID when no crosswalk exists; API returns `mlb_id` alongside nullable `player_id`             |
| Stale data after game                     | Standings/schedule cron runs hourly; stats cron runs every 4h; configurable per deployment                       |
| Schema drift from MLB API                 | Pipeline validates response shape before upsert; unknown fields are dropped, missing fields are nulled           |
| Data duplication after Retrosheet release | Handoff procedure truncates `current_season` schema after canonical load is validated                            |
| First-party rate limit bypass abuse       | `current-season-sync` jobs are tagged as first-party; existing rate limiter already exempts internal ETL traffic |
