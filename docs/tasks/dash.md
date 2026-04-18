# Dashboard Implementation Tasks

- Stack: SvelteKit 2 (SPA via `adapter-static`), Tailwind CSS v4, Chart.js 4, TypeScript
- Deploys as static SPA; backend API namespace is `/v1/*`
- OpenAPI spec source: `internal/docs/swagger.yaml`

## Scaffold and Tooling ✓

- [x] Initialize SvelteKit project in `web/` with TypeScript and `adapter-static` fallback.
- [x] Configure Tailwind v4 and map design tokens.
- [x] Add Chart.js.
- [x] Add `<Chart>` wrapper component.
- [x] Add OpenAPI parser dependency for API Explorer.
- [x] Add `task dash:dev` and `task dash:build`.
- [x] Configure dev proxy: `/v1/* -> localhost:8080`.

## Shared Layout and Components ✓

### Layout

- [x] Build shared header/nav in `+layout.svelte`.
- [x] Support single-column and three-column page layouts.

### Reusable Components

- [x] `Panel`
- [x] `TabRow` / `Tab`
- [x] `SegmentControl`
- [x] `SortableTable`
- [x] `Pill` / `Badge`
- [x] `SearchInput`
- [x] `CopyButton`
- [x] `ApiMirrorStrip`
- [x] `ApiPanel`
- [x] `Pagination`
- [x] `CoverageBar`
- [x] `EraBadge` (single era)
- [x] `EraRangeChip` (year range + era short code)
- [x] `EraLegend` (shared legend component with caveat tooltips)

## API Client and State ✓

- [x] Typed fetch wrapper that defaults base URL to `/v1`.
- [x] Shared meta store for `/v1/meta`.
- [x] Route-first state architecture established for nested explorer flows (canonical state in path segments, tab-local controls in query params).
- [x] Central endpoint-map constants derived from OpenAPI (avoid hardcoded stale params).

## Era Contracts ✓

- [x] Create a shared static era catalog from `internal/seed/eras.go`:
    - `fed`, `nlg`, `boomer`, `pitcher`, `turf`, `steroid`, `moneyball`, `statcast`, `modern`
- [x] Fetch dynamic win-expectancy eras from `GET /v1/win-expectancy/eras`.
- [x] Implement helper: `year -> static era`, with fallback when out of configured range.
- [x] Add global era disclaimer component for sparse historical/event coverage.

## Home Page (`/`) ✓

Ref: `docs/designs/home.html`

- [x] Search hero and entity pills.
- [x] Route search to `/v1/search/*` endpoints (not generic `/search`).
- [x] Quick links panel.
- [x] Featured queries panel.
- [x] API health panel from `/v1/meta`.
- [x] Dataset coverage panel from `/v1/meta/datasets`.
- [x] Add era quick-jump chips (fed -> modern) and deep-link behavior.
- [x] Add featured queries for league-specific and derived/computed endpoints.

## Player Explorer (`/players`) ✓

Ref: `docs/designs/players.html`

### Sidebar

- [x] Search players via `GET /v1/search/players?q=...`.
- [x] Selected player card from `GET /v1/players/{id}`.
- [x] Recent players list with deep-linked selection.

### Center Content

- [x] Batting/Pitching season timelines from `GET /v1/players/{id}/stats/batting` and `/stats/pitching`.
- [x] Batting stats table from `GET /v1/players/{id}/stats/batting`.
- [x] Pitching stats table from `GET /v1/players/{id}/stats/pitching`.
- [x] Game logs tabs from `GET /v1/players/{id}/game-logs/{batting|pitching|fielding}`.
- [x] Awards tab from `GET /v1/players/{id}/awards`.
- [x] Hall of Fame tab from `GET /v1/players/{id}/hall-of-fame`.
- [x] Teams tab from `GET /v1/players/{id}/teams`.
- [x] Relatives tab from `GET /v1/players/{id}/relatives`.
- [x] Salaries tab from `GET /v1/players/{id}/salaries`.
- [x] Advanced tabs (optional behind toggle):
    - `GET /v1/players/{player_id}/stats/batting/advanced`
    - `GET /v1/players/{player_id}/stats/pitching/advanced`
    - `GET /v1/players/{player_id}/stats/war`
    - `GET /v1/players/{player_id}/splits`
    - `GET /v1/players/{player_id}/streaks`

### Routing and Navigation Pattern ✓

- [x] Adopt nested routes for tab state: `/players/[id]/[tab]` (no canonical `tab=` query state).
- [x] Keep `/players` as search + empty state and redirect `/players/[id] -> /players/[id]/batting`.
- [x] Restrict query params to `q`, `page`, `per_page`, and tab-local controls (`stat`, `log`).
- [x] Keep advanced tabs always routable; toggle only affects tab visibility in navigation.
- [x] Move persistent shell concerns (sidebar/profile/API panel) to `/players/+layout.svelte`.
- [x] Move tab data ownership to child route pages under `/players/[id]/*`.
- [x] Remove players URL-sync controller/effect loop pattern; use route-driven reloads (`onMount` + navigation hooks).
- [x] Normalize internal navigation pattern:
    - template-string route paths passed to `resolve(...)`
    - `goto(resolve(...), { replaceState: true, noScroll: true, keepFocus: true })` for in-place query updates

### Transitions and Motion Pattern ✓

- [x] Add nested-outlet transition baseline using Svelte built-ins (`crossfade` + subtle `fly` fallback).
- [x] Respect reduced-motion preferences (`prefers-reduced-motion`) when applying transitions.
- [x] Keep sidebar and API panel stable while animating only center-pane route swaps.

### Era UX

- [x] Show era chip on each season row.
- [x] Show career-spans-era summary near player header.

## Team and Franchise Explorer (`/teams`)

Ref: `docs/designs/teams.html`

### Sidebar

- [x] Franchise picker from `GET /v1/franchises`.
- [x] Team search from `GET /v1/search/teams`.
- [x] Year selector to hydrate team-season detail.

### Center Content

- [x] Franchise metadata from `GET /v1/franchises/{id}`.
- [x] Team-season snapshot from `GET /v1/teams/{id}?year=...`.
- [x] Team roster + stat tabs from:
    - `GET /v1/seasons/{year}/teams/{team_id}/roster`
    - `GET /v1/seasons/{year}/teams/{team_id}/batting`
    - `GET /v1/seasons/{year}/teams/{team_id}/pitching`
    - `GET /v1/seasons/{year}/teams/{team_id}/fielding`
- [x] Team schedule from `GET /v1/seasons/{year}/teams/{team_id}/games` and `/schedule`.
- [x] Team daily trends from `GET /v1/teams/{id}/daily-stats` and `/daily-logs`.
- [x] Run differential module from `GET /v1/teams/{team_id}/run-differential?season=...`.

### API Contract Notes (Current Implementation)

- [x] `GET /v1/franchises` parsed as `FranchisesResponse.franchises[]` (not `data[]`).
- [x] `GET /v1/seasons/{year}/teams/{team_id}/roster` handled as non-paginated array response.
- [x] `GET /v1/seasons/{year}/teams/{team_id}/{batting|pitching|fielding}` handled as aggregate stats object with optional `players`.
- [x] `GET /v1/teams/{team_id}/run-differential` includes required `season` query and handles non-paginated series response.

### Era UX

- [x] Franchise continuity timeline segmented by era bands.
- [x] Warn when comparing across coverage gaps (e.g., Negro Leagues vs modern).

## Game Finder and Detail (`/games`)

Ref: `docs/designs/games.html`

### Finder

- [x] Filter form maps to `GET /v1/games` supported params (`season`, `home_team`, `away_team`, `park_id`, `date_from`, `date_to`, `page`, `per_page`).
- [x] Natural-language mode maps to `GET /v1/search/games?q=...`.
- [x] Quick toggles for Federal League and Negro Leagues using dedicated route families.

### Detail

- [x] Core metadata from `GET /v1/games/{id}`.
- [x] Summary box from `GET /v1/games/{id}/summary`.
- [x] Boxscore from `GET /v1/games/{id}/boxscore`.
- [x] Event stream from `GET /v1/games/{id}/events` (+ single-event fetch).
- [x] Plays + pitches drilldown from `GET /v1/games/{id}/plays` and `GET /v1/games/{id}/pitches`.
- [x] Win-probability modules from:
    - `GET /v1/games/{game_id}/win-probability`
    - `GET /v1/games/{game_id}/win-probability/summary`
    - `GET /v1/games/{game_id}/plate-appearances/leverage`

### Era UX

- [x] Show event-density confidence indicator by era.

## Season Hub (`/seasons`)

Ref: `docs/designs/seasons.html`

- [ ] Season list from `GET /v1/seasons`.
- [ ] Team table from `GET /v1/seasons/{year}/teams`.
- [ ] Leaders snapshot from:
    - `GET /v1/seasons/{year}/leaders/batting`
    - `GET /v1/seasons/{year}/leaders/pitching`
- [ ] Schedule/calendar from:
    - `GET /v1/seasons/{year}/schedule`
    - `GET /v1/seasons/{year}/dates/{date}/games`
- [ ] Awards/postseason modules from:
    - `GET /v1/seasons/{year}/awards`
    - `GET /v1/seasons/{year}/postseason/series`
    - `GET /v1/seasons/{year}/postseason/games`
- [ ] Park factors snapshot from `GET /v1/seasons/{season}/park-factors`.
- [ ] Era headline for selected season.

## Stat Leaders (`/leaders`)

Ref: `docs/designs/leaders.html`

- [ ] Quick leaders mode using season leader endpoints.
- [ ] Query lab using:
    - `GET /v1/stats/batting`
    - `GET /v1/stats/pitching`
    - `GET /v1/stats/fielding`
    - `GET /v1/stats/teams/batting`
    - `GET /v1/stats/teams/pitching`
    - `GET /v1/stats/teams/fielding`
- [ ] Career mode using:
    - `GET /v1/leaders/batting/career`
    - `GET /v1/leaders/pitching/career`
- [ ] Advanced mode using:
    - `GET /v1/seasons/{season}/leaders/batting/advanced`
    - `GET /v1/seasons/{season}/leaders/pitching/advanced`
    - `GET /v1/seasons/{season}/leaders/war`
- [ ] Era-bucket trend charts.

## Compare Mode (`/compare`)

Ref: `docs/designs/compare.html`

- [ ] Player vs player and team-season vs team-season comparison.
- [ ] Era-normalized comparison mode (show both entities' eras).
- [ ] Win expectancy scenario compare using:
    - `GET /v1/win-expectancy`
    - `GET /v1/win-expectancy/eras`

## API Explorer (`/explorer`)

Ref: `docs/designs/api-explorer.html`

- [ ] Parse OpenAPI at runtime and build endpoint tree.
- [ ] Group endpoints by current API families:
    - meta/health, search, players, teams/franchises, games/plays/pitches, stats/leaders,
    computed/derived, awards/postseason/allstar/ejections/achievements/salaries,
    managers/umpires/coaches, federalleague/negroleagues, mlb proxy.
- [ ] Parameter builder must use real parameter names from schema (e.g., `page`/`per_page`, `name`, `q`, `sort_by`).
- [ ] Era-aware query helpers for endpoints that support era-style filters/ranges.

## Data Sources (`/data`)

Ref: `docs/designs/data-sources.html`

- [ ] Source cards for Lahman, Retrosheet, MLB StatsAPI.
- [ ] Coverage timeline and caveats from `/v1/meta/datasets`.
- [ ] Era matrix panel (static 9 eras + dynamic WE eras).
- [ ] ID crosswalk table (Lahman, Retrosheet, MLB).
- [ ] League-specific caveat block for Federal League and Negro Leagues routes.

## League-Specific Views

- [ ] Add explicit UI entry points for:
    - `/v1/federalleague/games`, `/teams`, `/plays`, `/seasons/{year}/schedule`, `/seasons/{year}/teams/{team_id}/games`
    - `/v1/negroleagues/games`, `/teams`, `/plays`, `/seasons/{year}/schedule`, `/seasons/{year}/teams/{team_id}/games`
- [ ] Show these as first-class historical contexts, not hidden filters.

## Account and API Keys (`/account`)

- [ ] Auth state wiring to `/v1/auth/me`.
- [ ] API key list/create/revoke against `/v1/auth/keys`.
- [ ] OAuth launch links for GitHub/Codeberg.
- [ ] Usage dashboard remains TBD pending dedicated usage endpoint.

## Cross-Cutting

- [ ] Loading and error states.
- [ ] Deep-link restoration for all views.
- [ ] Responsive behavior for three-column pages.
- [ ] Keyboard navigation/accessibility.
- [ ] API panel always shows canonical `/v1/...` request and runnable curl.

## Build and Deploy

- [ ] Static build with SPA fallback.
- [ ] Bundle OpenAPI spec into static output.
- [ ] Prod API base should target `/v1` namespace (example: `https://bigfly.tech/api/v1`).
- [ ] Validate CDN-hosted routing for all SPA routes.

## Endpoint Map by Page

| Page         | Endpoint families                                                                                                           |
| ------------ | --------------------------------------------------------------------------------------------------------------------------- |
| Home         | `/v1/meta`, `/v1/meta/datasets`, `/v1/search/*`                                                                             |
| Players      | `/v1/search/players`, `/v1/players/*`, computed/derived player endpoints                                                    |
| Teams        | `/v1/franchises/*`, `/v1/search/teams`, `/v1/teams/*`, `/v1/seasons/{year}/teams/*`, `/v1/teams/{team_id}/run-differential` |
| Games        | `/v1/games*`, `/v1/search/games`, `/v1/seasons/{year}/*games*`, game win-probability/leverage endpoints                     |
| Seasons      | `/v1/seasons`, `/v1/seasons/{year}/teams`, leaders, schedule, awards, postseason                                            |
| Leaders      | `/v1/stats/*`, `/v1/leaders/*/career`, `/v1/seasons/{season}/leaders/*advanced*`                                            |
| Compare      | Player/team/season endpoints + `/v1/win-expectancy*`                                                                        |
| Explorer     | Full OpenAPI surface                                                                                                        |
| Data Sources | `/v1/meta/datasets`, `/v1/win-expectancy/eras`, league-specific families                                                    |
| Account      | `/v1/auth/*`                                                                                                                |
