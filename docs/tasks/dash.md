# Dashboard Implementation Tasks

- Stack: SvelteKit 2 (SPA via `adapter-static`), Tailwind CSS v4, Chart.js 4, TypeScript
- Deploys as static SPA; backend API namespace is `/api/v1/*`
- OpenAPI spec source: `internal/docs/swagger.yaml`

## Scaffold and Tooling ✓

## Shared Layout and Components ✓

## API Client and State ✓

## Era Contracts ✓

## Home Page (`/`) ✓

## Player Explorer (`/players`) ✓

## Team and Franchise Explorer (`/teams`) ✓

## Game Finder and Detail (`/games`) ✓

## Season Hub (`/seasons`) ✓

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

## API Docs Link (`/api/v1/docs/`) ✓

## Docs (`/docs`) ✓

## Data Sources (`/data`) ✓

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
