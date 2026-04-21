# Dashboard Implementation Tasks

- Stack: SvelteKit 2 (SPA via `adapter-static`), Tailwind CSS v4, Chart.js 4, TypeScript
- Deploys as static SPA; backend API namespace is `/api/v1/*`
- OpenAPI spec source: `internal/docs/swagger.yaml`

## Scaffold and Tooling ✓

- [x] Initialize SvelteKit project in `web/` with TypeScript and `adapter-static` fallback.
- [x] Configure Tailwind v4 and map design tokens.
- [x] Add Chart.js.
- [x] Add `<Chart>` wrapper component.
- [x] Link dashboard/app navigation to existing Swagger docsite (`/api/v1/docs/`).
- [x] Add `task dash:dev` and `task dash:build`.
- [x] Configure dev proxy: `/api/v1/* -> localhost:8080`.

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

- [x] Typed fetch wrapper that defaults base URL to `/api/v1`.
- [x] Shared meta store for `/api/v1/meta`.
- [x] Route-first state architecture established for nested dashboard flows (canonical state in path segments, tab-local controls in query params).
- [x] Central endpoint-map constants derived from OpenAPI (avoid hardcoded stale params).

## Era Contracts ✓

- [x] Create a shared static era catalog from `internal/seed/eras.go`:
  - `fed`, `nlg`, `boomer`, `pitcher`, `turf`, `steroid`, `moneyball`, `statcast`, `modern`
- [x] Fetch dynamic win-expectancy eras from `GET /api/v1/win-expectancy/eras`.
- [x] Implement helper: `year -> static era`, with fallback when out of configured range.
- [x] Add global era disclaimer component for sparse historical/event coverage.

## Home Page (`/`) ✓

Ref: `docs/designs/home.html`

- [x] Search hero and entity pills.
- [x] Route search to `/api/v1/search/*` endpoints (not generic `/search`).
- [x] Quick links panel.
- [x] Featured queries panel.
- [x] API health panel from `/api/v1/meta`.
- [x] Dataset coverage panel from `/api/v1/meta/datasets`.
- [x] Add era quick-jump chips (fed -> modern) and deep-link behavior.
- [x] Add featured queries for league-specific and derived/computed endpoints.

## Player Explorer (`/players`) ✓

Ref: `docs/designs/players.html`

### Sidebar

- [x] Search players via `GET /api/v1/search/players?q=...`.
- [x] Selected player card from `GET /api/v1/players/{id}`.
- [x] Recent players list with deep-linked selection.

### Center Content

- [x] Batting/Pitching season timelines from `GET /api/v1/players/{id}/stats/batting` and `/stats/pitching`.
- [x] Batting stats table from `GET /api/v1/players/{id}/stats/batting`.
- [x] Pitching stats table from `GET /api/v1/players/{id}/stats/pitching`.
- [x] Game logs tabs from `GET /api/v1/players/{id}/game-logs/{batting|pitching|fielding}`.
- [x] Awards tab from `GET /api/v1/players/{id}/awards`.
- [x] Hall of Fame tab from `GET /api/v1/players/{id}/hall-of-fame`.
- [x] Teams tab from `GET /api/v1/players/{id}/teams`.
- [x] Relatives tab from `GET /api/v1/players/{id}/relatives`.
- [x] Salaries tab from `GET /api/v1/players/{id}/salaries`.
- [x] Advanced tabs (optional behind toggle):
  - `GET /api/v1/players/{player_id}/stats/batting/advanced`
  - `GET /api/v1/players/{player_id}/stats/pitching/advanced`
  - `GET /api/v1/players/{player_id}/stats/war`
  - `GET /api/v1/players/{player_id}/splits`
  - `GET /api/v1/players/{player_id}/streaks`

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

## Team and Franchise Explorer (`/teams`) ✓

Ref: `docs/designs/teams.html`

### Sidebar

- [x] Franchise picker from `GET /api/v1/franchises`.
- [x] Team search from `GET /api/v1/search/teams`.
- [x] Year selector to hydrate team-season detail.

### Center Content

- [x] Franchise metadata from `GET /api/v1/franchises/{id}`.
- [x] Team-season snapshot from `GET /api/v1/teams/{id}?year=...`.
- [x] Team roster + stat tabs from:
  - `GET /api/v1/seasons/{year}/teams/{team_id}/roster`
  - `GET /api/v1/seasons/{year}/teams/{team_id}/batting`
  - `GET /api/v1/seasons/{year}/teams/{team_id}/pitching`
  - `GET /api/v1/seasons/{year}/teams/{team_id}/fielding`
- [x] Team schedule from `GET /api/v1/seasons/{year}/teams/{team_id}/games` and `/schedule`.
- [x] Team daily trends from `GET /api/v1/teams/{id}/daily-stats` and `/daily-logs`.
- [x] Run differential module from `GET /api/v1/teams/{team_id}/run-differential?season=...`.

### API Contract Notes (Current Implementation)

- [x] `GET /api/v1/franchises` parsed as `FranchisesResponse.franchises[]` (not `data[]`).
- [x] `GET /api/v1/seasons/{year}/teams/{team_id}/roster` handled as non-paginated array response.
- [x] `GET /api/v1/seasons/{year}/teams/{team_id}/{batting|pitching|fielding}` handled as aggregate stats object with optional `players`.
- [x] `GET /api/v1/teams/{team_id}/run-differential` includes required `season` query and handles non-paginated series response.

### Era UX

- [x] Franchise continuity timeline segmented by era bands.
- [x] Warn when comparing across coverage gaps (e.g., Negro Leagues vs modern).

## Game Finder and Detail (`/games`) ✓

Ref: `docs/designs/games.html`

### Finder

- [x] Filter form maps to `GET /api/v1/games` supported params (`season`, `home_team`, `away_team`, `park_id`, `date_from`, `date_to`, `page`, `per_page`).
- [x] Natural-language mode maps to `GET /api/v1/search/games?q=...`.
- [x] Quick toggles for Federal League and Negro Leagues using dedicated route families.

### Detail

- [x] Core metadata from `GET /api/v1/games/{id}`.
- [x] Summary box from `GET /api/v1/games/{id}/summary`.
- [x] Boxscore from `GET /api/v1/games/{id}/boxscore`.
- [x] Event stream from `GET /api/v1/games/{id}/events` (+ single-event fetch).
- [x] Plays + pitches drilldown from `GET /api/v1/games/{id}/plays` and `GET /api/v1/games/{id}/pitches`.
- [x] Win-probability modules from:
  - `GET /api/v1/games/{game_id}/win-probability`
  - `GET /api/v1/games/{game_id}/win-probability/summary`
  - `GET /api/v1/games/{game_id}/plate-appearances/leverage`

### Era UX

- [x] Show event-density confidence indicator by era.

## Season Hub (`/seasons`) ✓

Ref: `docs/designs/seasons.html`

- [x] Season list from `GET /api/v1/seasons`.
- [x] Team table from `GET /api/v1/seasons/{year}/teams`.
- [x] Leaders snapshot from:
  - `GET /api/v1/seasons/{year}/leaders/batting`
  - `GET /api/v1/seasons/{year}/leaders/pitching`
- [x] Schedule/calendar from:
  - `GET /api/v1/seasons/{year}/schedule`
  - `GET /api/v1/seasons/{year}/dates/{date}/games`
- [x] Awards/postseason modules from:
  - `GET /api/v1/seasons/{year}/awards`
  - `GET /api/v1/seasons/{year}/postseason/series`
  - `GET /api/v1/seasons/{year}/postseason/games`
- [x] Park factors snapshot from `GET /api/v1/seasons/{season}/park-factors`.
- [x] Era headline for selected season.

## Stat Leaders (`/leaders`)

Ref: `docs/designs/leaders.html`

- [ ] Quick leaders mode using season leader endpoints.
- [ ] Query lab using:
  - `GET /api/v1/stats/batting`
  - `GET /api/v1/stats/pitching`
  - `GET /api/v1/stats/fielding`
  - `GET /api/v1/stats/teams/batting`
  - `GET /api/v1/stats/teams/pitching`
  - `GET /api/v1/stats/teams/fielding`
- [ ] Career mode using:
  - `GET /api/v1/leaders/batting/career`
  - `GET /api/v1/leaders/pitching/career`
- [ ] Advanced mode using:
  - `GET /api/v1/seasons/{season}/leaders/batting/advanced`
  - `GET /api/v1/seasons/{season}/leaders/pitching/advanced`
  - `GET /api/v1/seasons/{season}/leaders/war`
- [ ] Era-bucket trend charts.

## Compare Mode (`/compare`)

Ref: `docs/designs/compare.html`

- [ ] Player vs player and team-season vs team-season comparison.
- [ ] Era-normalized comparison mode (show both entities' eras).
- [ ] Win expectancy scenario compare using:
  - `GET /api/v1/win-expectancy`
  - `GET /api/v1/win-expectancy/eras`

## API Docs Link (`/api/v1/docs/`)

- [x] Remove in-app API Explorer as a dashboard feature.
- [x] Add direct Swagger docsite link in top-level app navigation.
- [x] Route former `/explorer` dashboard path to Swagger docs for compatibility.

## Docs (`/docs`) ✓

Ref: `docs/designs/docs.html`

- [x] Install and configure mdsvex in the SvelteKit project.
- [x] Add `docs/` as a source directory for mdsvex pre-rendering (copy or symlink `docs/*.md` into the SvelteKit source tree, or configure mdsvex glob to reach them).
- [x] Generate a static route per doc file: `/docs/[slug]` from each `*.md` filename.
- [x] Build three-column docs layout (`/docs/+layout.svelte`): sidebar nav, center prose outlet, right TOC.
- [x] Left sidebar: grouped doc list (API Reference / Data & Architecture / Project) with search filter.
- [x] Right column: on-this-page TOC extracted from rendered heading tree at build time.
- [x] Apply prose stylesheet matching dark dashboard theme (headings, tables, code blocks, inline code).
- [x] Redirect `/docs` to first doc (e.g. `/docs/README` or `/docs/api-players`).

## Data Sources (`/data`) ✓

Ref: `docs/designs/data-sources.html`

- [x] Source cards for Lahman, Retrosheet, MLB StatsAPI.
- [x] Coverage timeline and caveats from `/api/v1/meta/datasets`.
- [x] Era matrix panel (static 9 eras + dynamic WE eras).
- [x] ID crosswalk table (Lahman, Retrosheet, MLB).
- [x] League-specific caveat block for Federal League and Negro Leagues routes.

## League-Specific Views

Ref: `docs/designs/federal-league.html`, `docs/designs/negro-leagues.html`

Both leagues get their own SvelteKit route families (`/federal-league` and `/negro-leagues`) that share a common page structure and a set of dedicated components. Neither is a filtered view of the main `/games` or `/teams` routes — they are isolated historical contexts backed by their own API families.

### Shared: `LeagueContextBanner`

- [ ] Create `LeagueContextBanner` component:
  - Full-width sub-header strip rendered below the site nav on league-specific pages.
  - Slots: era chip (e.g. `fed · 1914–1915`), short descriptor text, one or more caveat pills.
  - Caveat pill uses orange accent and warns of partial play-by-play coverage.
  - Accept props: `era` (short code), `years` (string), `description`, `caveats: string[]`.
  - Used on both `/federal-league` and `/negro-leagues` pages.

### Shared: `LeagueEraPanel`

- [ ] Create `LeagueEraPanel` component:
  - Top-of-center-pane summary card.
  - Shows: era name + era chip badge, prose description, and a row of quick-stats (teams count, seasons count, game count, PBP coverage quality indicator).
  - Coverage quality indicator renders as a string (`partial` / `variable`) in orange when data is incomplete.
  - Accept props: `name`, `eraCode`, `years`, `description`, `stats: {teams, seasons, games, coverageLabel}`.

### Shared: `LeagueCoverageCaveat`

- [ ] Create `LeagueCoverageCaveat` component:
  - Orange-bordered inline callout block (left border accent, muted background).
  - Used above the Plays view to warn that play-by-play is fragmentary for these eras.
  - Accept props: `league` (`'Federal League' | 'Negro Leagues'`), optional `message` override.
  - Reuses the same visual pattern as the existing global era disclaimer component.

### Shared: `LeagueTeamChipGrid`

- [ ] Create `LeagueTeamChipGrid` component:
  - Compact wrapping grid of team ID chips rendered in the sidebar.
  - Each chip is a small pill styled with the page's era accent color.
  - Clicking a chip sets the `home_team` filter and triggers a results reload.
  - Accept props: `teams: { id: string, city?: string }[]`, `onSelect: (id: string) => void`.
  - Used in both league sidebars.

### Negro Leagues (`/negro-leagues`)

- [ ] Add SvelteKit route `/negro-leagues` with three-column layout.
- [ ] Shared: `SubLeaguePillGroup` component:
  - Segmented control rendering NAL and NNL as distinct colored pills (NAL blue, NNL green).
  - Selecting a pill sets `?league=NAL` or `?league=NNL` query param on all fetch calls.
  - An "All" option clears the param.
  - Used in sidebar and within the Teams view filter row.
- [ ] Sidebar:
  - `SubLeaguePillGroup` at top.
  - `LeagueTeamChipGrid` sourced from `GET /api/v1/negroleagues/teams`.
  - Season select: full range 1935–1949.
  - Home team and away team selects.
  - View switcher (Games / Teams / Plays / Schedule).
  - Quick-filter chips: Extra innings, NAL only, NNL only, KC Monarchs.
  - Apply filters button.
  - `LeagueCoverageCaveat` pinned to sidebar bottom.
- [ ] Center — Games view (default):
  - `LeagueEraPanel` at top.
  - Results table from `GET /api/v1/negroleagues/games` with `league` param applied (columns: Date, Home, Away, Score, Inn, League pill, Season).
  - League column renders `SubLeaguePill` (NAL blue / NNL green) per row.
  - Inline game detail on row click.
  - Pagination.
- [ ] Center — Teams view:
  - `SubLeaguePillGroup` filter row above table.
  - Table from `GET /api/v1/negroleagues/teams?league=...` (columns: Team ID chip, Name, City, League pill, Active years).
- [ ] Center — Plays view:
  - `LeagueCoverageCaveat` at top.
  - Batter ID, pitcher ID, team, `SubLeaguePillGroup`, date-range form mapping to `GET /api/v1/negroleagues/plays`.
  - Results table with League pill column.
  - Inline gap indicator row for partial sequences.
- [ ] Center — Schedule view:
  - Season select (1935–1949) and `SubLeaguePillGroup` side-by-side.
  - Fetches `GET /api/v1/negroleagues/seasons/{year}/schedule?league=...`.
  - Results table with League pill column.
  - Note linking to team-scoped endpoint: `GET /api/v1/negroleagues/seasons/{year}/teams/{team_id}/games`.
  - Pagination.
- [ ] Center — Games-per-season chart:
  - Stacked bar chart (Chart.js) of recorded games per season, NAL and NNL stacked.
  - Always visible below the active view to communicate coverage variability at a glance.
- [ ] `ApiPanel` (right column): reflects active endpoint + league param as URL, curl, and sample JSON. Includes sub-league param reference block (NAL / NNL / omit).

### Federal League (`/federal-league`)

- [ ] Add SvelteKit route `/federal-league` with three-column layout (sidebar, center, API panel).
- [ ] Sidebar:
  - `LeagueTeamChipGrid` with the 8 Federal League franchises sourced from `GET /api/v1/federalleague/teams`.
  - Season select: 1914 / 1915 / All.
  - Home team and away team selects populated from teams list.
  - View switcher (Games / Teams / Plays / Schedule) as sidebar tab group.
  - Quick-filter chips: Extra innings, Shutouts, CHF home.
  - Apply filters button.
  - `LeagueCoverageCaveat` pinned to sidebar bottom.
- [ ] Center — Games view (default):
  - `LeagueEraPanel` at top.
  - Sortable results table from `GET /api/v1/federalleague/games` (columns: Date, Home, Away, Score, Inn, Park, Season).
  - Clicking a row expands an inline game detail strip (score, park, innings, season).
  - Pagination via `Pagination` component.
- [ ] Center — Teams view:
  - Table from `GET /api/v1/federalleague/teams` (columns: Team ID chip, City, Nickname, Seasons, W, L, PCT).
- [ ] Center — Plays view:
  - `LeagueCoverageCaveat` at top.
  - Batter ID, pitcher ID, team, date-range filter form mapping to `GET /api/v1/federalleague/plays` params.
  - Results table (columns: Game ID, Inning, Batter, Pitcher, Event, Runs).
  - Inline gap indicator row when event sequence is known-partial.
- [ ] Center — Schedule view:
  - Season selector buttons (1914 / 1915) that fetch `GET /api/v1/federalleague/seasons/{year}/schedule`.
  - Results table (columns: Date, Home, Away, Score, Inn, Park).
  - Note below table linking to team-scoped schedule endpoint: `GET /api/v1/federalleague/seasons/{year}/teams/{team_id}/games`.
  - Pagination.
- [ ] Center — Standings chart:
  - Bar chart (Chart.js) showing wins per team grouped by season (1914 vs 1915), always visible below the active view.
- [ ] `ApiPanel` (right column): reflects active endpoint + params as URL, curl, and sample JSON response shape.

### Navigation

- [ ] Add "Federal League" and "Negro Leagues" links to the top-level app nav (`+layout.svelte`).
- [ ] Both routes listed under a "Historical Leagues" group or divider in the nav.

## Live & Current-Season Data

### Home: Live Scoreboard

- [x] Create `ScoreboardStrip` component:
  - Horizontal scrollable row of compact game cards.
  - Each card: away/home team abbreviations, scores, inning/status, team color accents.
  - LIVE badge with CSS animation on in-progress games.
  - "No games today" empty state with next game date.
- [x] Fetch from `GET /v1/mlb/schedule?date={today}&hydrate=linescore,team` (proxy-first implementation for live scoreboard).
- [x] Auto-refresh via `setInterval` (30s) when `games_in_progress > 0` & tab is focused; clear interval when all final.
  - Refresh button
  - Pause auto-refresh toggle
- [x] Click game card → navigate to `/games` with Retrosheet game ID (if crosswalkable) or show inline MLB detail popover.
- [x] Add scoreboard strip to Home page above existing featured queries panel.

### Home: Today's Leaders

- [x] Create `LeaderCards` component:
  - Tabbed panel with one tab per stat category.
  - Each tab: ranked top-5 list with player name, team abbreviation, stat value.
  - Player names link to `/players/[id]/batting` or `/pitching` via crosswalk.
  - Hitting tabs: HR, AVG, OPS, RBI, SB.
  - Pitching tabs: ERA, SO, W, SV, WHIP.
- [x] Fetch via MLB proxy:
  - `GET /v1/mlb/stats?stats=season&group=hitting&season={current}&playerPool=qualified`
  - `GET /v1/mlb/stats?stats=season&group=pitching&season={current}&playerPool=qualified`
- [x] Add leaders panel to Home page below scoreboard.

### Teams: Current Standings

- [x] Create `StandingsPanel` component:
  - Sortable tables grouped by division (collapsible sections).
  - Columns: Rank, Team (with color dot), W, L, PCT, GB, WC GB, Streak, Run Diff, L10.
  - Division leader highlight.
  - Segment control: AL / NL / Both.
- [x] Fetch from `GET /v1/mlb/standings?season={current}&standingsTypes=regularSeason`.
- [x] Team name links to `/teams/[franchise_id]` with current year when crosswalkable (fallback: `/teams?q=`).
- [x] Add standings panel on Teams landing view alongside existing franchise and team-season flows.

### Games: Live Game Overlay

- [ ] Create `LiveGameOverlay` component:
  - Panel showing current score, inning, count, runners, current play description.
  - Win probability line chart (Chart.js) updating with each refresh.
  - "Live data from MLB" attribution badge.
    - Render the copyright attribution
  - Dismissible toggle to switch between live overlay and historical view.
- [ ] Detect live game: when viewing a game detail, check if the game crosswalks to an active `gamePk`.
- [ ] Fetch from `GET /v1/mlb/live/{gamePk}` (expanded MLB proxy route).
- [ ] Auto-refresh 15s during in-progress games.

### Players: Current Season Stats

- [ ] Create `CurrentSeasonCard` component:
  - Compact card showing key current-season stats from MLB API.
  - Hitters: AVG, HR, RBI, OPS, SB.
  - Pitchers: ERA, SO, W-L, WHIP, IP.
  - "Live data from MLB" badge.
  - Render the copyright attribution
- [ ] On Player Detail load, check if player has an active `mlb_id` via crosswalk.
- [ ] If active, fetch from `GET /v1/mlb/people/{mlb_id}?hydrate=stats(group=[hitting,pitching],type=season)` and enrich with local lookup.
- [ ] Render above historical stats tabs.
- [ ] Hide card when player is not active in current season.

### Endpoint Map Update

| Page    | New endpoint families                                            |
| ------- | ---------------------------------------------------------------- |
| Home    | `/v1/mlb/schedule`, `/v1/mlb/stats`, `/v1/mlb/teams`             |
| Teams   | `/v1/mlb/standings`, `/v1/mlb/teams`, `/v1/meta/crosswalk/teams` |
| Games   | `/v1/mlb/live/{gamePk}`                                          |
| Players | `/v1/mlb/people/{mlb_id}`                                        |

## Cross-Cutting

- [ ] Loading and error states.
- [ ] Deep-link restoration for all views.
- [ ] Responsive behavior for three-column pages.
- [ ] Keyboard navigation/accessibility.
- [ ] API panel always shows canonical `/api/v1/...` request and runnable curl.

## Build and Deploy

- [ ] Static build with SPA fallback.
- [ ] Prod API base should target `/api/v1` namespace (example: `https://bigfly.tech/api/v1`).
- [ ] Validate CDN-hosted routing for all SPA routes.

## Endpoint Map by Page

| Page         | Endpoint families                                                                                                                                    |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Home         | `/api/v1/meta`, `/api/v1/meta/datasets`, `/api/v1/search/*`, `/v1/mlb/schedule`, `/v1/mlb/stats`, `/v1/mlb/teams`                                    |
| Players      | `/api/v1/search/players`, `/api/v1/players/*`, computed/derived player endpoints, `/v1/mlb/people/*`                                                 |
| Teams        | `/api/v1/franchises/*`, `/api/v1/search/teams`, `/api/v1/teams/*`, `/api/v1/seasons/{year}/teams/*`, `/v1/mlb/standings`, `/v1/meta/crosswalk/teams` |
| Games        | `/api/v1/games*`, `/api/v1/search/games`, `/api/v1/seasons/{year}/*games*`, win-probability/leverage, `/v1/mlb/live/*`                               |
| Seasons      | `/api/v1/seasons`, `/api/v1/seasons/{year}/teams`, leaders, schedule, awards, postseason                                                             |
| Leaders      | `/api/v1/stats/*`, `/api/v1/leaders/*/career`, `/api/v1/seasons/{season}/leaders/*advanced*`                                                         |
| Compare      | Player/team/season endpoints + `/api/v1/win-expectancy*`                                                                                             |
| API Docs     | Swagger UI docsite at `/api/v1/docs/`                                                                                                                |
| Data Sources | `/api/v1/meta/datasets`, `/api/v1/win-expectancy/eras`, league-specific families                                                                     |
