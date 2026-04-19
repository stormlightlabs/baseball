---
title: Data Dashboard
updated: 2026-04-18
---

## Stack

- **Framework**: SvelteKit 2 in SPA mode (`adapter-static`, fallback `index.html`), deployed to CDN
- **Styling**: Tailwind CSS v4
- **Charts**: Chart.js 4 via a thin Svelte wrapper
- **API docs**: Existing Swagger docsite at `/api/v1/docs/` (source OpenAPI: `internal/docs/swagger.yaml`)
- **Routing**:
  - Dashboard SPA at `/`
  - API namespace is `/v1/*` (current public API surface)

## API Compatibility Baseline

The dashboard spec must track the API that is currently registered in `internal/api/*`.

### Core API families to expose

- **Meta/health**: `/v1/health`, `/v1/ready`, `/v1/meta`, `/v1/meta/datasets`, `/v1/meta/readiness`
- **Search**: `/v1/search/players`, `/v1/search/teams`, `/v1/search/parks`, `/v1/search/games`
- **Players**: `/v1/players/*` including `seasons`, `stats/*`, `game-logs/*`, `awards`, `hall-of-fame`, `teams`, `salaries`, `relatives`
- **Teams/franchises**: `/v1/teams/*`, `/v1/franchises/*`, `/v1/seasons/{year}/teams/*`
- **Games/events**: `/v1/games`, `/v1/games/{id}`, `/v1/games/{id}/boxscore`, `/v1/games/{id}/summary`, `/v1/games/{id}/events`, `/v1/games/{id}/plays`, `/v1/games/{id}/pitches`
- **Stats/leaders**: `/v1/stats/*`, `/v1/seasons/{year}/leaders/*`, `/v1/leaders/*/career`, `/v1/stats/teams/*`
- **Advanced/computed**: `/v1/players/{player_id}/stats/*/advanced`, `/v1/players/{player_id}/stats/war`, leverage + park-factor endpoints
- **Derived**: streaks, splits, run-differential, game win-probability, `/v1/win-expectancy`, `/v1/win-expectancy/eras`
- **League-specific historical slices**: `/v1/federalleague/*`, `/v1/negroleagues/*`
- **Awards/postseason/all-star/ejections/achievements/salaries/managers/umpires/coaches**
- **MLB proxy**: `/v1/mlb/*`

## Era Model (Must Be First-Class in UI)

The dashboard should expose two complementary era systems.

### A) Retrosheet load eras (static, from `internal/seed/eras.go`)

| Era                | Short code  | Years     |
| ------------------ | ----------- | --------- |
| Federal League Era | `fed`       | 1914-1915 |
| Negro Leagues Era  | `nlg`       | 1935-1949 |
| Baby Boomer Era    | `boomer`    | 1950-1962 |
| Pitcher Era        | `pitcher`   | 1963-1968 |
| Turf Time          | `turf`      | 1969-1993 |
| Steroid Era        | `steroid`   | 1994-2004 |
| Moneyball Era      | `moneyball` | 2005-2012 |
| Statcast Era       | `statcast`  | 2013-2019 |
| Modern Era         | `modern`    | 2020-2025 |

### B) Win expectancy eras (dynamic, from API)

- Source endpoint: `GET /v1/win-expectancy/eras`
- This returns era ranges based on sampled historical tables, so labels/ranges are data-driven.

### Era UX rules

- Every page that has a season or range control should show the active era badge.
- Compare mode must show both selected eras side-by-side.
- Cross-era comparisons should display source caveats (especially Federal League / Negro Leagues and pre-1918 play-event density).
- Data Source page should include an era matrix and coverage caveats, not just raw year bars.

## Search and Discovery Home

Should immediately answer: _what can this API do across eras?_

### Features

- universal search dispatcher that routes to:
  - `/v1/search/players`
  - `/v1/search/teams`
  - `/v1/search/games`
  - `/v1/search/parks`
- featured queries panel that includes at least one query from:
  - standard stats
  - derived/computed
  - league-specific historical routes
- API health + dataset coverage block sourced from meta endpoints
- era quick-jump chips (fed, nlg, boomer, pitcher, turf, steroid, moneyball, statcast, modern)

## Player Explorer

### Endpoints

- `/v1/search/players`
- `/v1/players/{id}`
- `/v1/players/{id}/seasons`
- `/v1/players/{id}/stats/batting`
- `/v1/players/{id}/stats/pitching`
- `/v1/players/{id}/game-logs/*`
- `/v1/players/{id}/awards`
- `/v1/players/{id}/hall-of-fame`
- `/v1/players/{id}/teams`
- `/v1/players/{id}/salaries`
- `/v1/players/{id}/relatives`
- Optional advanced tabs:
  - `/v1/players/{player_id}/stats/batting/advanced`
  - `/v1/players/{player_id}/stats/pitching/advanced`
  - `/v1/players/{player_id}/stats/war`
  - `/v1/players/{player_id}/splits`
  - `/v1/players/{player_id}/streaks`

### Route and State Contract

- Canonical UI URL shape is path-first: `/players/[id]/[tab]`.
- `/players` is the search + empty-state entry route.
- `/players/[id]` redirects to `/players/[id]/batting`.
- Main tabs are nested route slugs:
  - `batting`, `pitching`, `game-logs`, `awards`, `hof`, `teams`, `salaries`, `relatives`
- Advanced tabs are always routable nested slugs:
  - `batting-adv`, `pitching-adv`, `war`, `splits`, `streaks`
- Advanced toggle only controls tab visibility in the nav; it does not gate route access.
- Query params are limited to search + tab-local controls:
  - shared/search: `q`
  - pagination: `page`, `per_page`
  - batting tab: `stat`
  - game logs tab: `log`
- Legacy query tab routing (`/players?id=...&tab=...`) is non-canonical and does not require compatibility redirects pre-release.

### Implementation Pattern (Reference for Future Sections)

- Parent layout owns persistent chrome (sidebar search/result list, selected-entity profile, API panel).
- Child nested routes own center-pane data loading and rendering.
- Data loading is route-driven (`onMount` + navigation hooks), not synchronized via cross-tab URL-state effects.
- Internal navigation uses SvelteKit routing primitives:
  - links via `href={resolve(...)}`
  - imperative query updates via `goto(resolve(...), { replaceState: true, noScroll: true, keepFocus: true })`
- Tab outlet transitions use Svelte built-ins (`crossfade` with subtle `fly` fallback) and must respect `prefers-reduced-motion`.

### Era-specific behavior

- Season charts show shaded era bands.
- Player timeline cards show "career spans X eras" with linked era chips.

## Team and Franchise Explorer

### Endpoints

- `/v1/franchises`
- `/v1/franchises/{id}`
- `/v1/search/teams`
- `/v1/teams/{id}`
- `/v1/teams/{id}/daily-stats`
- `/v1/teams/{team_id}/run-differential`
- `/v1/seasons/{year}/teams`
- `/v1/seasons/{year}/teams/{team_id}/roster`
- `/v1/seasons/{year}/teams/{team_id}/batting`
- `/v1/seasons/{year}/teams/{team_id}/pitching`
- `/v1/seasons/{year}/teams/{team_id}/fielding`
- `/v1/seasons/{year}/teams/{team_id}/games`
- `/v1/seasons/{year}/teams/{team_id}/schedule`
- `/v1/seasons/{year}/teams/{team_id}/daily-logs`

### Contract Notes (Swagger-Backed)

- `GET /v1/franchises` returns `api.FranchisesResponse` (`{ franchises: Franchise[]; total }`), not a generic `data` array.
- `GET /v1/search/teams` is paginated (`api.PaginatedResponse`) and supports `q`, `year`, `league`, `page`, `per_page`.
- `GET /v1/seasons/{year}/teams/{team_id}/roster` returns a plain `RosterPlayer[]` array (non-paginated).
- `GET /v1/seasons/{year}/teams/{team_id}/{batting|pitching|fielding}` returns aggregate objects (`core.Team*Stats`) with optional `players` splits when `players=true`.
- `GET /v1/teams/{team_id}/run-differential` requires query `season` and returns a single `core.RunDifferentialSeries` object (non-paginated).
- `GET /v1/seasons/{year}/teams/{team_id}/schedule` is a paginated alias of `/games`.

### Era-specific behavior

- Franchise view uses era segmentation for name/identity continuity.
- Team season cards show context chips (e.g., `pitcher`, `steroid`, `statcast`).

## Game Finder and Game Detail

### Endpoints

- `/v1/games`
- `/v1/games/{id}`
- `/v1/games/{id}/summary`
- `/v1/games/{id}/boxscore`
- `/v1/games/{id}/events`
- `/v1/games/{id}/events/{event_seq}`
- `/v1/games/{id}/plays`
- `/v1/games/{id}/pitches`
- `/v1/games/{game_id}/win-probability`
- `/v1/games/{game_id}/win-probability/summary`
- `/v1/games/{game_id}/plate-appearances/leverage`
- `/v1/seasons/{year}/schedule`
- `/v1/seasons/{year}/dates/{date}/games`
- `/v1/seasons/{year}/teams/{team_id}/games`
- `/v1/seasons/{year}/parks/{park_id}/games`

### Era-specific behavior

- Filters include era chips in addition to year/date range.
- Detail page flags whether event richness is dense/partial for the selected era.

## Stat Leaderboard Center

### Endpoints

- `/v1/stats/batting`
- `/v1/stats/pitching`
- `/v1/stats/fielding`
- `/v1/stats/teams/batting`
- `/v1/stats/teams/pitching`
- `/v1/stats/teams/fielding`
- `/v1/seasons/{year}/leaders/batting`
- `/v1/seasons/{year}/leaders/pitching`
- `/v1/leaders/batting/career`
- `/v1/leaders/pitching/career`
- Advanced leaderboards:
  - `/v1/seasons/{season}/leaders/batting/advanced`
  - `/v1/seasons/{season}/leaders/pitching/advanced`
  - `/v1/seasons/{season}/leaders/war`

### Era-specific behavior

- Trend view should support both continuous year view and era-bucket view.

## Season Hub

### Endpoints

- `/v1/seasons`
- `/v1/seasons/{year}/teams`
- `/v1/seasons/{year}/leaders/batting`
- `/v1/seasons/{year}/leaders/pitching`
- `/v1/seasons/{year}/schedule`
- `/v1/seasons/{year}/dates/{date}/games`
- `/v1/seasons/{year}/awards`
- `/v1/seasons/{year}/postseason/series`
- `/v1/seasons/{year}/postseason/games`
- `/v1/seasons/{season}/park-factors`

## Compare Mode

### Comparisons

- player vs player
- team-season vs team-season
- season vs season
- franchise-era comparison
- win expectancy state comparison across era ranges

### Endpoints

- all relevant player/team/season endpoints above
- `/v1/win-expectancy`
- `/v1/win-expectancy/eras`

## Live & Current-Season Data

The dashboard uses the MLB Stats API proxy (`/v1/mlb/*`) as the primary source for current-season data surfaced alongside historical Retrosheet-backed views.

### Live Scoreboard (Home Page)

A scoreboard strip on the home page showing today's games with live scores.

- Source: `GET /v1/mlb/schedule?date={today}&hydrate=linescore,team`
- Renders as a horizontal scrollable row of compact game cards
- Each card: away/home abbreviations, scores, inning/status indicator, team color accents
- LIVE indicator with subtle animation on in-progress games
- Auto-refresh via `setInterval` (30s) when games are in progress; stops when all games are final
- Click a game card → `/games` with the game's Retrosheet ID if crosswalkable, otherwise show an inline detail popover with MLB data
- Fallback: "No games today" state with next game date

### Current Standings (Teams Page)

Standings panel on the Teams page, alongside the existing franchise/team-season views.

- Source: `GET /v1/mlb/standings?season={current}&standingsTypes=regularSeason`
- Renders as sortable tables grouped by division
- Columns: Rank, Team, W, L, PCT, GB, WC GB, Streak, Run Diff, L10
- Team names link to `/teams/[franchise_id]` with current-season year context
- Segment control: AL / NL / Both
- Team color dots next to names

### Today's Leaders (Home Page)

Current-season stat leaders below the scoreboard on the home page.

- Source:
  - `GET /v1/mlb/stats?stats=season&group=hitting&season={current}&playerPool=qualified&include=details`
  - `GET /v1/mlb/stats?stats=season&group=pitching&season={current}&playerPool=qualified&include=details`
- Renders as tabbed cards (one tab per stat category)
- Each card: ranked top-5 list with player name, team, stat value
- Player names link to `/players/[id]/batting` or `/players/[id]/pitching` via MLBAM crosswalk mappings
- Category tabs: HR, AVG, OPS, RBI, SB | ERA, SO, W, SV, WHIP

### Live Game Overlay (Games Page)

When viewing a game that is currently in progress (detected via crosswalk to `gamePk`), the game detail page shows a live overlay.

- Source: `GET /v1/mlb/live/{gamePk}` (expanded MLB proxy route to `/api/v1.1/game/{gamePk}/feed/live`)
- Overlay panel above the historical boxscore/events showing: current score, inning, count, runners, current play, win probability chart
- Auto-refresh 15s during live games
- Dismissible — user can toggle between live and historical views

### Player Current Season (Player Explorer)

When viewing a player who is active in the current season, an additional panel shows current-season stats from the MLB API.

- Source: `GET /v1/mlb/people/{id}?hydrate=stats(group=[hitting,pitching],type=season)` (+ local lookup via `/v1/search/players`)
- Renders as a "Current Season" card above the historical stats tabs
- Shows key current-season stats (AVG/HR/RBI for hitters, ERA/SO/W for pitchers)
- "Live data from MLB" attribution badge
- Only shows when the crosswalk maps the local player ID to an active MLBAM ID

### Endpoints

| Dashboard location | Endpoint family                                          | Refresh  |
| ------------------ | -------------------------------------------------------- | -------- |
| Home scoreboard    | `GET /v1/mlb/schedule`                                   | 30s auto |
| Home leaders       | `GET /v1/mlb/stats` (`include=details`)                  | manual   |
| Teams standings    | `GET /v1/mlb/standings` + `GET /v1/meta/crosswalk/teams` | manual   |
| Game detail live   | `GET /v1/mlb/live/{gamePk}`                              | 15s auto |
| Player current     | `GET /v1/mlb/people/{id}` + local search/crosswalk       | manual   |

## API Documentation

- The dashboard links out to the existing Swagger UI docsite at `/api/v1/docs/`.
- The dashboard does not own a separate in-app API Explorer surface.

## Docs (`/docs`)

Pre-rendered prose documentation derived from the `docs/` markdown files in the repo.

### Rendering strategy

- Use [mdsvex](https://mdsvex.pngwn.io/docs) to pre-render Markdown at build time into Svelte components.
- Each `docs/*.md` file becomes a statically generated route under `/docs/[slug]`.
- No client-side markdown parsing; all HTML is generated at build time via `adapter-static`.

### Layout

- Three-column: left sidebar (doc nav + search), center (rendered prose), right (table-of-contents / on-this-page).
- Left sidebar groups docs by category: **API Reference**, **Data & Architecture**, **Project**.
- Right column generates an on-this-page TOC from the rendered heading tree.

### Doc categories

| Category            | Source files                                                                                                                                                                                                                                                                                                                                      |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| API Reference       | `api-players`, `api-teams`, `api-games`, `api-stats`, `api-search`, `api-pitches`, `api-play-by-play`, `api-game-context`, `api-per-game-aggregations`, `api-computed`, `api-derived-advanced`, `api-achievements`, `api-awards-postseason`, `api-parks-umpires-managers`, `api-league-coverage`, `api-mlb-proxy`, `api-meta-utility`, `api-auth` |
| Data & Architecture | `id-crosswalk`, `data-loading`, `statistical-methodology`                                                                                                                                                                                                                                                                                         |
| Project             | `README`, `ROADMAP`, `testing`                                                                                                                                                                                                                                                                                                                    |

### Prose styling

- Headings, tables, inline `code`, and fenced code blocks styled to match the dashboard dark theme.
- Coverage/status pills rendered for percentage values in tables.

## Data Sources and Coverage Page

### Must include

- source cards: Lahman, Retrosheet, MLB StatsAPI
- historical coverage + known caveats
- explicit era matrix (the 9 seed eras + dynamic WE eras)
- crosswalk strategy for Lahman IDs, Retrosheet IDs, MLB IDs
- league-specific caveats for Federal League and Negro Leagues endpoints

## Account and Key Management

Authenticated area behind `/account`.

### API hooks

- `/v1/auth/me`
- `/v1/auth/keys`
- `/v1/auth/keys/{id}` (delete/revoke)
- OAuth entrypoints exist for GitHub/Codeberg callback flow

## Page Architecture

| Route           | Page               | Layout     | Auth | Notes                                                     |
| --------------- | ------------------ | ---------- | ---- | --------------------------------------------------------- |
| `/`             | Home               | single-col | no   | search + meta + era jump + live scoreboard + leaders      |
| `/players`      | Players            | three-col  | no   | search/empty route                                        |
|                 |                    |            |      | canonical deep links use `/players/[id]/[tab]`            |
| `/teams`        | Teams              | three-col  | no   | franchise + team-season + era context + current standings |
| `/games`        | Games              | three-col  | no   | finder + game detail + event richness + live overlay      |
| `/seasons`      | Seasons            | three-col  | no   | season hub + awards/postseason                            |
| `/leaders`      | Leaders            | three-col  | no   | quick leaders + query lab + advanced                      |
| `/compare`      | Compare            | three-col  | no   | side-by-side + era normalization                          |
| `/api/v1/docs/` | API Docs (Swagger) | external   | no   | Existing Swagger UI docsite linked from app navigation    |
| `/data`         | Data Sources       | single-col | no   | provenance + era matrix + caveats                         |
| `/docs`         | Docs               | three-col  | no   | mdsvex pre-rendered prose from `docs/*.md`                |
| `/docs/[slug]`  | Doc page           | three-col  | no   | individual doc with sidebar nav + TOC                     |
| `/account`      | Account            | single-col | yes  | API keys + usage                                          |
