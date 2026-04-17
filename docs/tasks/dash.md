# Dashboard Implementation Tasks

- Stack: SvelteKit 2 (SPA via `adapter-static`), Tailwind CSS v4, Chart.js 4, TypeScript
- No Flowbite runtime -> designs reference it for wireframing only.
- Deploys to CDN as static SPA; API lives at a separate origin.
- OpenAPI spec (`internal/docs/swagger.yaml`) parsed client-side for the API Explorer.

## Scaffold & Tooling

- [x] Init SvelteKit project inside `web/` (or similar) with TypeScript, `adapter-static` (fallback `index.html` for SPA). Dashboard routes live at the root (`/`); `/api/` is reserved for the backend.
- [x] Install and configure Tailwind v4 (`@tailwindcss/vite`). Translate the design-system tokens from the wireframes into a Tailwind theme:
  - Colors: `--bg #0d0f13`, `--surface #13161c`, `--surface2 #1a1e27`, `--border #252934`, `--text #e2e8f0`, `--muted #6b7280`, `--accent #3b82f6`, `--accent2 #10b981`, `--warning #f59e0b`.
    Note: renamed to more semantic names
  - Fonts: `sans` → Inter (300–600), `heading` → Google Sans (400/500/700), `mono` → Google Sans Code.
  - Dark-only — no light mode toggle needed.
- [x] Add Chart.js 4 as a dependency.
- [ ] Create a thin `<Chart>` Svelte wrapper that handles canvas lifecycle (`onMount` / `onDestroy`) and reactive data updates.
- [x] Add an OpenAPI/Swagger parser dependency (e.g. `swagger-parser` or a lightweight alternative) to power the API Explorer page.
- [ ] Add a `task dash:dev` and `task dash:build` to `Taskfile.yml` (runs vite dev/build inside `web/`).
- [x] Configure the dev server to proxy `/v1/*` to `localhost:8080` so the SPA can call the API without CORS during development.

## Shared Layout & Components

All pages share a header, optional sidebar, and a consistent panel/card system visible in every wireframe.

### Layout

- [ ] `+layout.svelte` — sticky 56px header with logo ("Baseball API"), badge (current page context), and nav links: Home, Players, Teams, Games, Seasons, Leaders, Compare, Explorer, Data. Active link styling via `$page.url`. Account/key-management link in the right side of the header (authenticated area).
- [ ] Define two layout snippet props: **single-column** (Home, Data Sources) and **three-column** (Players, Teams, Games, Seasons, Leaders, Compare, API Explorer).
  The three-column layout is `grid-template-columns: 280px 1fr 320px` with sidebar, center scroll area, and right API/context panel.

### Reusable Components

- [ ] `Panel` — surface card with optional `label` prop (uppercase mono small text with border-bottom divider).
  Used everywhere.
- [ ] `TabRow` / `Tab` — horizontal tab strip. Active tab gets `surface2` bg + border.
  Supports keyboard nav.
- [ ] `SegmentControl` — toggle buttons (e.g. Batting/Pitching, Quick leaders/Query lab).
  Like tabs but inside sidebars.
- [ ] `SortableTable` — props: `columns` (label, key, sortable, format), `rows`, `sort` bindable.
  Mono font cells, hover row highlight, rank-bar optional column renderer.
- [ ] `Pill` / `Badge` — small rounded labels. Pills are clickable filters, badges are informational.
- [ ] `SearchInput` — input with placeholder, optional submit button.
  Mini variant for sidebars, full variant for hero search.
- [ ] `CopyButton` — copies text to clipboard, brief "Copied" feedback.
- [ ] `ApiMirrorStrip` — bottom bar showing `GET`, endpoint URL, copy URL, copy curl.
  Displayed on entity pages outside the three-column layout.
- [ ] `ApiPanel` — right-column panel used in the three-column layout. Shows: endpoint path, generated full URL, curl block, and a collapsible JSON preview. Reacts to the current page context (selected player, query params, etc.).
- [ ] `Pagination` — page/per_page controls, total count display. Wired to URL search params.
- [ ] `CoverageBar` — horizontal bar with label, range text, and filled track (percentage).
  Colored variants for each data source.

## 2. API Client & Stores

- [ ] Create a typed fetch wrapper (`$lib/api.ts`) that:
  - Reads base URL from an env var / SvelteKit `$env` (defaults to `/v1`).
  - Appends pagination params, returns typed `{ data, page, per_page, total }` envelopes.
  - Exposes per-request loading/error state.
- [ ] Build a class encapsulating Svelte runes for API health/meta (fetched once on app init from `/v1/meta`). Provides version, dataset date ranges, source coverage to the Home page and header badge.
- [ ] URL-driven state: every filter, search term, page number, and sort column should be reflected in `$page.url.searchParams` so that every dashboard state is shareable and deep-linkable.

## 3. Home Page (`/`)

Ref: `docs/designs/home.html`

- [ ] Search hero — centered heading, subtitle, full-width `SearchInput`, and entity-type pill row (Players, Teams, Franchises, Games, Seasons). Typing + enter or pill click navigates to the appropriate entity page with a `?q=` param.
- [ ] Dashboard grid (3-col, gap-1px, border-radius-8) containing:
  - **Quick Links panel** — list of nav items with endpoint hint and arrow. Links to Players, Teams, Games, Leaders, Seasons pages.
  - **Featured Queries panel** — hardcoded list of interesting API queries with title + mono endpoint string. Clicking navigates to the API Explorer pre-filled.
  - **API Health panel** — 2x3 grid of stat boxes sourced from `/v1/meta` (status, version, data-from, data-to, avg response, source count).
  - **Dataset Coverage panel** (col-span-2) — three `CoverageBar`s (Lahman, Retrosheet, MLB StatsAPI) plus a stacked bar Chart.js chart showing coverage density by decade.
  - **Endpoints panel** — simple list of top-level endpoint paths, colored accent.
- [ ] `ApiMirrorStrip` at the bottom reflecting the last/default query.

## 4. Player Explorer (`/players`)

Ref: `docs/designs/players.html`

Three-column layout.

### Left Sidebar

- [ ] Player search input calling `GET /v1/players?search=...`. Debounced, results populate the player list below.
- [ ] Bio card for the selected player — avatar placeholder, name, birth info, handedness, debut/final year, and a row of career summary stats (HR, AVG, RBI, Seasons). Data from `GET /v1/players/{id}`.
- [ ] Player list — scrollable list of search results or recent selections. Active item highlighted.

### Center Content

- [ ] Tab row: Batting, Pitching, Awards, Hall of Fame.
- [ ] **Batting tab** — line chart of a selectable stat (HR, AVG, RBI, SB) across seasons via `<Chart>` wrapper. Data from `GET /v1/players/{id}/seasons`. Include era-context label. Below the chart, a `SortableTable` of all batting seasons (Year, Team badge, G, AB, H, HR, RBI, AVG, SB, etc.).
- [ ] **Pitching tab** — same pattern for pitching seasons (W, L, ERA, SO, IP, SV, WHIP, etc.). Show "no pitching data" gracefully for pure position players.
- [ ] **Awards tab** — timeline list from `GET /v1/players/{id}/awards`. Year + award name rows.
- [ ] **Hall of Fame tab** — bar chart of vote percentages by year from `GET /v1/players/{id}/hall-of-fame`. 75% threshold line.

### Right Panel

- [ ] `ApiPanel` reflecting the current endpoint (`/players/{id}/seasons`, etc.) with generated URL, curl, and sample JSON preview updated as the user switches tabs or players.

## 5. Team & Franchise Explorer (`/teams`)

Ref: `docs/designs/teams.html`

Three-column layout.

### Left Sidebar

- [ ] Franchise list from `GET /v1/franchises` with search filter. Clicking a franchise loads its detail.
- [ ] Year selector (dropdown or input) to drill into a specific team-season.

### Center Content

- [ ] **Franchise view** — name, active years, active flag. Timeline of historical team names/eras from `GET /v1/franchises/{id}`. Summary table of all seasons (year, team name, W, L, attendance) from `GET /v1/teams?franchise={id}`.
- [ ] **Team-season view** — W/L/T, runs scored/allowed, run differential, attendance, division, league. Data from `GET /v1/teams/{id}`. Below: team game log via `GET /v1/seasons/{year}/teams/{team_id}/games` in a `SortableTable`.

### Right Panel

- [ ] `ApiPanel` for the current franchise or team-season endpoint.

## 6. Game Finder & Detail (`/games`)

Ref: `docs/designs/games.html`

Three-column layout.

### Left Sidebar (Filters)

- [ ] Filter form: season (number input), date range (from/to), home team, away team, park, postseason toggle, min innings, attendance range. All params map to `GET /v1/games?...`.
- [ ] Quick filter pills: "Extra innings", "Doubleheaders", "Postseason".

### Center Content

- [ ] Results table: date, matchup (away @ home), score, innings, attendance, duration. Paginated via `Pagination`. Clicking a row opens game detail.
- [ ] Game detail panel (inline or modal): scoreline, date, day of week, venue, attendance, duration, umpire crew, season/series context. Links to both team pages. Data from `GET /v1/games/{id}`.

### Right Panel

- [ ] `ApiPanel` reflecting current search or game detail endpoint.

## 7. Season Hub (`/seasons`)

Ref: `docs/designs/seasons.html`

Three-column layout.

### Left Sidebar

- [ ] Season picker (input or dropdown, 1871–present). League filter (AL/NL/Both).

### Center Content

- [ ] Season overview: team table from `GET /v1/seasons/{year}/teams` — team, W, L, runs, attendance.
- [ ] Leaderboard snapshots: top 5 batting and pitching leaders for the season from `GET /v1/seasons/{year}/leaders/{type}?limit=5`. Link to full Leaders page.
- [ ] Date explorer: calendar-style or date-input view. Selecting a date shows games from `GET /v1/seasons/{year}/dates/{date}/games`.
- [ ] Schedule summary: full paginated schedule from `GET /v1/seasons/{year}/schedule`.

### Right Panel

- [ ] `ApiPanel` for the current season endpoint.

## 8. Stat Leaders (`/leaders`)

Ref: `docs/designs/leaders.html`

Three-column layout.

### Left Sidebar

- [ ] Mode toggle: Quick Leaders vs Query Lab (`SegmentControl`).
- [ ] **Quick Leaders mode**: batting/pitching segment, quick-stat pill row (HR, AVG, RBI, SB, SO / ERA, W, SV, SO, IP), season dropdown, league filter. Calls `GET /v1/seasons/{year}/leaders/{type}?stat={stat}`.
- [ ] **Query Lab mode**: season_from, season_to, min_ab/min_ip, sort_by (dropdown), sort_order, page_size. "Run query" button calls `GET /v1/stats/{type}?...`. Show the generated request URL live as params change.

### Center Content

- [ ] Tab row: Table, Bar Chart, Trend.
- [ ] **Table view** — `SortableTable` with rank, player, team, year, stat value (with rank-bar inline visualization), AB/IP, percentage.
- [ ] **Bar chart view** — horizontal bar chart via Chart.js ranking the leaders.
- [ ] **Trend view** — line chart of the league leader value for the selected stat across decades/seasons.

### Right Panel

- [ ] `ApiPanel` reflecting the current leader query with endpoint, URL, curl, and sample JSON.

## 9. Compare Mode (`/compare`)

Ref: `docs/designs/compare.html`

- [ ] Comparison type selector: Player vs Player, Team-Season vs Team-Season.
- [ ] Two entity pickers (search inputs that resolve to player IDs or team-season IDs).
- [ ] Side-by-side stat table: fetch both entities' data and render columns A vs B with highlighted differences.
- [ ] Overlay chart: line chart of both players' career stat or both teams' season stats on the same axes.
- [ ] `ApiPanel` showing both endpoints being called.

## 10. API Explorer (`/explorer`)

Ref: `docs/designs/api-explorer.html`

Three-column layout. This page is the interactive API reference.

### Left Sidebar

- [ ] Parse the OpenAPI spec (`swagger.yaml` fetched from the API (& cached)) and render a grouped endpoint list: Players, Teams & Franchises, Games, Seasons, Stats, Plays, Meta. Each item shows `GET` badge + path. Clicking selects it.

### Center Content

- [ ] URL bar: method badge, editable full URL, "Send" button.
- [ ] Parameters builder: dynamically generated from the OpenAPI spec for the selected endpoint. Grid of labeled inputs with type hints and doc strings. Path params are inlined, query params are separate fields. Changes update the URL bar live.
- [ ] Response viewer: on "Send", actually call the API. Show response status bar (status code, time, size, cache, total records) and syntax-highlighted JSON below. Use CSS classes for keys/strings/numbers/booleans (`.jk`, `.js`, `.jn`, `.jb` from the wireframe).

### Right Panel

- [ ] Curl block reflecting the current URL.
- [ ] Schema viewer: for the selected endpoint, render the response schema fields (name, type, description) from the OpenAPI spec.
- [ ] Request history: last N requests shown as clickable items that restore the URL bar.

## 11. Data Sources (`/data`)

Ref: `docs/designs/data-sources.html`

Single-column layout.

- [ ] Source cards for Lahman, Retrosheet, and MLB StatsAPI — each with description, coverage years, and a `CoverageBar`.
- [ ] Timeline chart: stacked bar or area chart showing which sources cover which decades.
- [ ] "Last updated" timestamps from `GET /v1/meta/datasets`. Per-table freshness if available.
- [ ] ID mapping reference table: Lahman player ID, Retrosheet player ID, MLB person ID — formats, examples, and which endpoints use which.
- [ ] Known gaps and caveats section — prose explaining partial historical data, starter-only game logs limitation (from `docs/TODO.md`), and any dataset-specific notes.

## 12. Account & Key Management (`/account`)

Authenticated area. Requires login; unauthenticated users see a sign-in prompt.

- [ ] Auth flow: sign-in / sign-up (email + password or OAuth TBD). Store session token; gate `/account` routes behind an auth guard in `+layout.ts`.
- [ ] **API Keys page** — list the user's API keys (name, prefix, created date, last used, status). Actions: create, revoke, copy key (shown once on creation).
- [ ] **Key creation form** — name, optional expiry, optional scope/permissions. On submit, display the full key once with a copy button and a warning that it won't be shown again.
- [ ] **Usage dashboard** — request count over time (daily/weekly/monthly chart), current rate-limit tier, quota remaining. Data from a backend usage endpoint (TBD).
- [ ] **Account settings** — email, password change, delete account.

## 13. Cross-Cutting Concerns

- [ ] Global search: the home search bar and a compact header search should route to the appropriate entity page with `?q=` populated. Implement a simple dispatcher that guesses entity type from the query or lets the user pick.
- [ ] Deep linking: every table sort, filter, tab, page number, and selected entity should be encoded in the URL search params. Navigating to a URL should restore the full state.
- [ ] Loading and error states: skeleton panels for loading, error panels with retry for failed fetches.
- [ ] Responsive: at minimum, collapse three-column layouts to stacked on narrow viewports. Sidebar becomes a slide-out drawer or collapsible section.
- [ ] Keyboard navigation: tab through nav links, enter to activate, escape to close panels.

## 14. Build & Deploy

- [ ] Static build via `adapter-static` with `fallback: 'index.html'` for SPA client-side routing.
- [ ] Bundle the OpenAPI spec into the static output (copy `swagger.yaml` to `static/`).
- [ ] Environment-based API base URL: dev proxies to localhost, prod points to `https://baseball.stormlightlabs.org/api`.
- [ ] Verify the built output works when served from a CDN with no server-side logic — all routes resolve to `index.html` and SvelteKit's client router takes over.

## Dependency Order

```text
0 (scaffold) → 1 (layout + components) → 2 (API client)
  → then pages in any order, but suggested:
    3 (Home) → 4 (Players) → 5 (Teams) → 6 (Games)
    → 7 (Seasons) → 8 (Leaders) → 9 (Compare)
    → 10 (Explorer) → 11 (Data Sources) → 12 (Account)
  → 13 (cross-cutting) throughout
  → 14 (build/deploy) last
```

## API Endpoints Used Per Page

| Page         | Endpoints                                                                                                              |
| ------------ | ---------------------------------------------------------------------------------------------------------------------- |
| Home         | `/meta`                                                                                                                |
| Players      | `/players`, `/players/{id}`, `/players/{id}/seasons`, `/players/{id}/awards`, `/players/{id}/hall-of-fame`             |
| Teams        | `/franchises`, `/franchises/{id}`, `/teams`, `/teams/{id}`, `/seasons/{year}/teams/{team_id}/games`                    |
| Games        | `/games`, `/games/{id}`                                                                                                |
| Seasons      | `/seasons/{year}/teams`, `/seasons/{year}/leaders/*`, `/seasons/{year}/schedule`, `/seasons/{year}/dates/{date}/games` |
| Leaders      | `/seasons/{year}/leaders/*`, `/stats/batting`, `/stats/pitching`                                                       |
| Compare      | `/players/{id}/seasons`, `/teams/{id}` (any two entities)                                                              |
| Explorer     | All (dynamically from OpenAPI spec)                                                                                    |
| Data Sources | `/meta/datasets`                                                                                                       |
| Account      | TBD — key management, usage metrics (backend endpoints not yet implemented)                                            |
