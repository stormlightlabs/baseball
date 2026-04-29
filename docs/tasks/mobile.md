# Flutter Mobile App Tasks

- Ref spec: `docs/specs/mobile.md`
- Stack: Flutter 3.x, Dart, Flame engine, Material 3, BLoC
- Backend: Go API additions under `/internal/mobile/*`, `/v1/mlb/*`, and `/v1/meta/*`
- Scope: native mobile app + backend endpoints + design updates

## Phase 0: Project Scaffold

- [x] Initialize Flutter project in `mobile/` with Android + iOS targets.
- [x] Configure `BLoC` for state management.
- [x] Add `dio` + `retrofit` and generate typed API client from OpenAPI spec.
- [x] Set up Material 3 theme with `ColorScheme.fromSeed()` and dark mode default.
- [x] Implement team color map (30 teams → primary hex) as a static Dart map.
- [x] Add `dynamic_color` package; wrap `MaterialApp` with `DynamicColorBuilder` for fallback system palette on neutral screens.
- [x] Build bottom navigation shell (5 tabs: Home, Players, Teams, Games, More).
- [x] Configure Drift/SQLite for local cache and offline storage.

Acceptance:

- [x] App launches on Android + iOS with bottom nav, dark theme, and team-aware color switching stub.
- [x] API client can fetch from `/api/v1/health` and display result.

## Phase 1: Core Screens (Port from Mobile Designs)

### Home

- [x] Search bar with entity-type pill filters (Players, Teams, Games, Franchises, Seasons).
- [x] Quick access grid (2×3) with icons linking to core sections.
- [x] API health strip sourced from `/api/v1/meta`.
- [x] Era chips for quick navigation.

### Player Detail

- [x] Bio card with pixel-art avatar placeholder, name, metadata, bio-stats grid.
- [x] Horizontal scrollable tabs: Batting, Pitching, Awards, HOF.
- [x] Career chart (line chart, selectable stats) using `fl_chart`
- [x] Season log table with sortable columns.
- [x] Awards section.
- [x] Dynamic theme: set `ColorScheme.fromSeed(team.primaryColor)` when player loads.

### Team Detail

- [x] Franchise card with badge, name, league/division, stats grid.
- [x] Segment control: Overview, Roster, Schedule, Daily.
- [x] Run differential chart.
- [x] Current roster list.
- [x] Recent games list with win/loss color coding.
- [x] Dynamic theme: seed from franchise primary color.

### Games

- [x] Filter strip: season, team, quick chips (Extra innings, Doubleheaders, Postseason).
- [x] Expandable game cards with score, metadata, and detail panel.
- [x] Win probability bar visualization.

### Seasons, Leaders, Compare, Data Sources

- [x] Port remaining screens from mobile designs

Acceptance:

- [ ] All 5 main tabs functional with real API data.
- [x] Team color theming applies on player and team detail screens.
- [ ] Theme transitions animate smoothly (300ms ease-out via `AnimatedTheme`).

## Phase 1.5: Animations (`flutter_animate`)

- Ref spec: `docs/specs/mobile.md` § Animations

### Setup

- [ ] Add `flutter_animate` package to `pubspec.yaml`.
- [ ] Create a shared `AnimationPresets` utility with reusable effect chains (list stagger, card expand, tab switch, stat count-up).
- [ ] Add a global `reduceMotion` flag derived from `MediaQuery.disableAnimations`; wire into preset chains via `.toggle(animate: !reduceMotion)`.

### Core Screen Polish

- [ ] **Lists & grids**: apply staggered `fadeIn` + `slideY` to player lists, game lists, roster lists, and leader rankings.
- [ ] **Card expand/collapse**: animate game card detail panel and collapsible standings sections with `fadeIn` + `scaleXY`.
- [ ] **Tab/segment switches**: directional `fadeIn` + `slideX` on content when switching Player Detail tabs, Standings segments, and Team Detail segments.
- [ ] **Stat values**: count-up animation on key numeric stats (batting average, ERA, HR totals) when they first appear.
- [ ] **Chip/filter selection**: subtle scale + fade on newly filtered content in game filters and spray chart filters.
- [ ] **Empty/error states**: `fadeIn` + `slideY` + `scale` entrance on empty state illustrations and error messages.

### Live Feature Animations

- [ ] **Scoreboard cards**: staggered `fadeIn` + `slideX` entrance on scoreboard `PageView` cards.
- [ ] **Live badge pulse**: looping `shimmer` effect with team accent color on "LIVE" badges.
- [ ] **Score change flash**: `tint` + `scale` burst when a score updates between refreshes.
- [ ] **Win probability line draw**: animated path-length reveal on the `LineChart` sparkline in Live Game Tracker.

### Visualization Animations

- [ ] **Spray chart dots**: staggered radial `fadeIn` + `scaleXY` entrance when spray chart data loads.
- [ ] **Pitch sequence steps**: per-pitch `fadeIn` + `slideY` in the At-Bat Sequencer.

Acceptance:

- [ ] All animated transitions complete within spec timing budgets (150–300ms typical, 600ms max for decorative).
- [ ] Reduced-motion setting disables or shortens all `flutter_animate` chains.
- [ ] No animation triggers on already-visible items during scroll.
- [ ] Animations pair with corresponding haptic feedback where specified in the spec.

## Phase 2: Spray Chart

### Backend

- [ ] Add `GET /internal/mobile/spray-chart/{player_id}` endpoint in Go.
  - Query params: `season`, `vs` (L/R), `park_id`.
  - Map Retrosheet hit-location codes (`loc`, `hittype`) to standardized field coordinates (origin at home plate, y toward CF, feet).
  - Return batted-ball events with `x`, `y`, `result`, `loc_code`, `hittype`, `pitcher`, `pitcher_hand`, `count`, `inning`, `outs`, `date`, `game_id`, `play_num`.
- [ ] Add park geometry research + ingestion workstream:
  - Research source options for park wall geometry and document provenance.
  - Define ingestion schema for per-park control points and metadata quality flags.
  - Define QA checks (angle coverage, monotonic ordering, outlier distances, season/park validity).
  - Define fallback behavior to a generic field template when park geometry is missing.
- [ ] Add explicit sample-query validation task using Shohei Ohtani IDs (`ohtansh01` / `ohtas001`) against local `plays` data.

### Frontend

- [ ] Create `SprayChartPainter` extending `CustomPainter`:
  - Draw infield diamond (bases, foul lines, dirt cutout).
  - Draw outfield wall from park geometry when available; otherwise draw a generic field wall template.
  - Plot batted-ball events as circles (color ∝ result, optional radius ∝ derived depth bucket).
  - Implement `hitTest` for tap detection on individual events.
- [ ] Wrap in `InteractiveViewer` for pinch-to-zoom and pan.
- [ ] Add filter controls: vs LHP/RHP, hit type, count, season picker.
- [ ] Add park overlay toggle (compare home vs. away park walls).
- [ ] Tap event detail sheet: date, opponent, pitcher, result, loc/hittype, count/inning/outs.
- [ ] Haptic feedback on tap (`HapticFeedback.selectionClick`).
- [ ] Add as new tab on Player Detail screen.

Acceptance:

- [ ] Spray chart renders 500+ events at 60fps.
- [ ] Park wall overlay scales correctly with zoom.
- [ ] Tapping an event shows detail bottom sheet with correct historical context fields.
- [ ] Coordinate mapping is consistent for the same `loc`/`hittype` inputs across runs.

## Phase 3: Pitch Pattern Explorer

### Backend

- [ ] Add `GET /internal/mobile/pitch-patterns/{pitcher_id}` endpoint.
  - Query params: `season`, `vs_bat` (L/R).
  - Aggregate Retrosheet pitch-call sequence data (`B`, `C`, `F`, `S`, `X`, etc.) into:
    - usage and rates by call code
    - sequence transitions by count state
    - divergence groups by shared-prefix sequences
  - Return split-ready summaries with handedness/count context.
- [ ] Add explicit sample-query validation task using Shohei Ohtani IDs (`ohtansh01` / `ohtas001`) against local `plays.pitches`.

### Frontend

- [ ] Create `PitchPatternExplorer` widget:
  - Render sequence lanes and divergence markers for pitch-call families.
  - Color-code call-code groups with labeled legend.
- [ ] Implement view/lens switching: overall, vs LHB, vs RHB, two-strike counts, hitter's counts.
- [ ] Add divergence point markers (depth where shared-prefix branches split).
- [ ] Add time/depth scrubber: slider to animate sequence depth (pitch 1 → pitch 2 → pitch 3+).
- [ ] Tap sequence lane to isolate and show pattern details.
- [ ] Haptic bump at selected divergence points.
- [ ] Add as sub-view under Player Detail → Pitching tab, and standalone under More.

Acceptance:

- [ ] 6 simultaneous sequence lanes render at 60fps.
- [ ] Lens switching is smooth and gesture-driven.
- [ ] Divergence markers reflect real sequence split depth.

### Internal API Scope Alignment

- [ ] Do a comprehensive internal API namespace consistency pass in docs/tasks:
  - Use `/internal/mobile/*` for player-card, spray-chart, pitch-patterns, at-bat, and quiz routes.
  - Remove stale legacy internal-route references.

## Phase 4: At-Bat Sequencer

### Backend

- [ ] Add `GET /internal/mobile/at-bat/{game_id}/{ab_num}` endpoint.
  - Return batter/pitcher info, pitch sequence (call, description, count, seq), and at-bat result.
  - Derive from existing `plays` and `pitches` tables.

### Frontend

- [ ] Create `AtBatSequencerWidget`:
  - `CustomPainter` draws strike zone scaffold and sequence markers by pitch order.
  - Pitches animate in as ordered sequence steps (200ms per pitch).
  - Count display updates with each pitch.
  - Horizontal timeline below zone shows sequence as labeled dots.
- [ ] Auto-play mode: pitches appear at 1.5s intervals.
- [ ] Manual mode: swipe left/right to step through.
- [ ] Tap pitch dot for detail (call, description, result, count).
- [ ] Long-press for pitcher's typical pattern comparison overlay.
- [ ] Haptic tick on each pitch arrival (`HapticFeedback.lightImpact`).
- [ ] Integrate into Game Detail: tap any plate appearance row to open sequencer.

Acceptance:

- [ ] Pitch sequence order and count progression render accurately.
- [ ] Auto-play and manual modes both work smoothly.
- [ ] Sequencer accessible from any game's play-by-play.

## Phase 5: Stat Card Generator

- [ ] Implement deterministic pixel-art avatar generator:
  - Seed from `player_id` hash.
  - Generate 32×32 sprite with team-colored cap and jersey.
  - Render at 128×128 with nearest-neighbor scaling.
  - Use `pixel_art_generator` package.
- [ ] Create `StatCardWidget` template:
  - Player avatar, name, position, team.
  - 4-6 headline stats with inline sparklines.
  - Team accent gradient background.
  - Season/career scope label.
  - Big Fly watermark.
- [ ] Capture card as PNG via `RenderRepaintBoundary`.
- [ ] Share via `share_plus` platform share sheet.
- [ ] Add share action to Player Detail screen (FAB).

Acceptance:

- [ ] Pixel-art avatars are deterministic (same player → same avatar across sessions).
- [ ] Card renders as shareable PNG at ≥1080px width.
- [ ] Share sheet opens with card image on both platforms.

## Phase 6: Baseball Learning Mode

### Backend

- [ ] Add `GET /internal/mobile/quiz/situation` endpoint:
  - Pull random historical game state (inning, outs, runners, score, count).
  - Include actual outcome and win expectancy from materialized view.
- [ ] Add `GET /internal/mobile/quiz/pitch-type` endpoint:
  - Pull pitch-call sequence context for pattern identification challenge.
  - Include 4 likely next-call/outcome options with correct answer.

### Frontend

- [ ] Create Learning Mode hub screen under More tab.
- [ ] **Rules & Scoring module**: animated diagrams explaining infield fly, balk, tag-up, force play.
  - Use Rive or Lottie animations for field diagrams.
- [ ] **Pitch identification trainer**:
  - Simplified pitch-pattern view showing partial sequence context.
  - 4-option multiple choice for likely next call/outcome.
  - Difficulty scaling: reduce visible context, add lookalike pattern families.
- [ ] **Situation quiz**:
  - Display game state visually (diamond with runners, scoreboard).
  - User guesses outcome or win probability range.
  - Reveal actual result + historical win expectancy.
- [ ] **Stat explainers**:
  - Interactive WAR calculator: adjust component inputs, see WAR change.
  - wOBA breakdown with weighted contribution bars.
- [ ] **Historical moments**:
  - Curated list of famous at-bats.
  - Opens at-bat sequencer with narration overlays.
- [ ] Local progress tracking via Drift (correct answers, streaks, completion %).

Acceptance:

- [ ] All 5 learning modules functional.
- [ ] Pitch identification trainer uses actual historical pitch-call sequence rendering.
- [ ] Progress persists across app sessions.

## Phase 7: Expanded MLB Proxy Namespace

### Scaffold

- [ ] Expand `internal/api/mlb.go` with UI-oriented live routes under `/v1/mlb/*`.
- [ ] Register expanded MLB proxy routes in `internal/api/server.go`.
- [ ] Add MLBAM-to-local ID crosswalk helpers for players and teams.
- [ ] Add MLB team ID → team color map (30 teams) as a static Go map.

### Scoreboard Endpoint

- [ ] Implement scoreboard view via `GET /v1/mlb/schedule?date={YYYY-MM-DD}&hydrate=linescore,team,probablePitcher`.
  - Map MLB team IDs to local franchise records for color theming.
  - Extract game status, scores, linescore, venue, probable pitchers.
  - Cache at 30s TTL.
- [ ] Add `core.InternalScoreboardResponse` type with game cards, team colors, and status.

### Standings Endpoint

- [ ] Implement standings view via `GET /v1/mlb/standings?season={year}&standingsTypes=regularSeason`.
  - Use `GET /v1/meta/crosswalk/teams?season={year}` for local routing IDs.
  - Group by division, enrich with team colors and franchise IDs.
  - Include wins, losses, PCT, GB, wild card GB, streak, run differential, last 10.
  - Cache at 5min TTL.
- [ ] Add `core.InternalStandingsResponse` type.

### Live Game Feed Endpoint

- [ ] Implement `GET /v1/mlb/live/{gamePk}`.
  - Proxy to MLB game feed (`/api/v1.1/game/{gamePk}/feed/live`).
  - Extract linescore, current play, recent plays, runners, count.
  - Merge with local win-probability engine when play-by-play state permits.
  - Cache at 15s TTL (or no cache for active games).
- [ ] Add `core.InternalLiveGameResponse` type.

### Leaders Endpoint

- [ ] Implement leaders view via `GET /v1/mlb/stats` (`stats=season&group=hitting|pitching&sortStat={stat}&limit=5`).
  - Call `/v1/mlb/stats` with `stats=season&group=hitting|pitching&sortStat={stat}&limit=5` per category.
  - Crosswalk MLBAM person IDs to local `player_id` for deep linking.
  - Merge team colors.
  - Cache at 15min TTL.
- [ ] Add `core.InternalLeadersResponse` type.

### Player Live Endpoint

- [ ] Implement current player card via `GET /v1/mlb/people/{mlb_id}?hydrate=stats(group=[hitting,pitching],type=season)`.
  - Fetch current-season stats from `/v1/mlb/people/{id}?hydrate=stats(group=[hitting,pitching],type=season)`.
  - Crosswalk to local player record for historical context.
  - Return merged bio + current stats + historical summary.
  - Cache at 5min TTL.
- [ ] Add `core.InternalPlayerLiveResponse` type.

### Team Live Endpoint

- [ ] Implement team live card via `GET /v1/mlb/teams/{mlb_id}`.
  - Fetch current team info from `/v1/mlb/teams/{id}`.
  - Crosswalk to local franchise record.
  - Return merged team info + franchise history + team colors.
  - Cache at 5min TTL.
- [ ] Add `core.InternalTeamLiveResponse` type.

Acceptance:

- [ ] Expanded `/v1/mlb/*` routes return shaped payloads matching spec response schemas.
- [ ] MLBAM → local ID crosswalk works for players and teams.
- [ ] Cache TTLs are respected per endpoint group.
- [ ] Expanded MLB proxy routes follow standard `/v1/*` auth and rate-limit behavior.

## Phase 8: Live Scoreboard (Mobile)

### Frontend

- [ ] Create `ScoreboardWidget` for Home tab:
  - Horizontal `PageView` of game cards.
  - Each card shows team abbreviations, scores, inning/status, and linescore row.
  - Team primary colors as gradient accents on each card.
  - "LIVE" badge with pulse animation on in-progress games.
  - "Final" / "Scheduled" badges for completed/upcoming games.
- [ ] Implement `ScoreboardBloc`:
  - Fetch from `GET /v1/mlb/schedule?date={today}&hydrate=linescore,team,probablePitcher`.
  - Auto-refresh every 30s when `games_in_progress > 0`.
  - Cache last response in Drift for offline display.
- [ ] Tap game card → navigate to Live Game Tracker (in progress) or Game Detail (final).
- [ ] Pull-to-refresh gesture.
- [ ] Haptic tick on score changes between refreshes.
- [ ] Date picker to view previous/future days.

Acceptance:

- [ ] Scoreboard renders all daily games in a swipeable card view.
- [ ] Auto-refresh updates scores without user interaction during live games.
- [ ] Offline mode shows cached scoreboard with "Last updated" indicator.

## Phase 9: Current Standings (Mobile)

### Frontend

- [ ] Create `StandingsScreen` accessible from Teams tab (segment control: Standings / Franchises).
- [ ] Division-grouped list with collapsible sections:
  - Row per team: rank, team name (with color dot), W, L, PCT, GB, WC GB, streak, L10.
  - Division leader indicator.
  - Wild card separator line.
- [ ] Segment control: AL / NL / Both.
- [ ] Sort by any column (tap header).
- [ ] Implement `StandingsBloc`:
  - Fetch from `GET /v1/mlb/standings?season={current}&standingsTypes=regularSeason`.
  - Cache in Drift for offline.
- [ ] Tap team row → Team Detail with current-season year pre-selected.
- [ ] Haptic on section collapse/expand.

Acceptance:

- [ ] All 6 divisions render with correct team ordering.
- [ ] Sorting works on all columns.
- [ ] Team tap navigates to correct team detail with current season context.

## Phase 10: Live Game Tracker (Mobile)

### Frontend

- [ ] Create `LiveGameScreen`:
  - Scoreboard header with full linescore grid (innings × team).
  - Diamond `CustomPainter`: infield diamond with filled/empty base indicators.
  - Count `CustomPainter`: balls (green dots), strikes (red dots), outs (white dots).
  - Current play description with animated text transition (`AnimatedSwitcher`).
  - Win probability sparkline (`fl_chart` `LineChart`) updating in real-time.
  - Scrollable recent plays list.
- [ ] Implement `LiveGameBloc`:
  - Fetch from `GET /v1/mlb/live/{gamePk}`.
  - Auto-refresh every 15s during active games.
  - Stop auto-refresh when game status is Final or Scheduled.
- [ ] Tap win probability chart → expand to full-screen view.
- [ ] Tap play in recent list → play detail bottom sheet.
- [ ] Haptic bump on scoring plays and third outs.
- [ ] Swipe down or back to return to scoreboard.
- [ ] Pre-game state: show probable pitchers, venue, weather (if available), first pitch time.
- [ ] Post-game state: show final score, winning/losing pitcher, save, notable stats.

Acceptance:

- [ ] Diamond and count indicators update correctly with each refresh.
- [ ] Win probability chart renders and updates smoothly.
- [ ] Transitions between pre-game, live, and post-game states are handled.

## Phase 11: Today's Leaders (Mobile)

### Frontend

- [ ] Create `LeadersWidget` for Home tab (below scoreboard):
  - Horizontal `PageView` of stat category cards.
  - Chip row above cards for category selection.
  - Each card: ranked list of 5 players with team-colored accent bars.
  - Hitting categories: HR, AVG, OPS, RBI, SB.
  - Pitching categories: ERA, SO, W, SV, WHIP.
- [ ] Implement `LeadersBloc`:
  - Fetch from `GET /v1/mlb/stats` category queries for season leaders.
  - Cache in Drift; refresh on pull-to-refresh.
- [ ] Tap player row → Player Detail (via crosswalked `player_id`).
- [ ] Swipe or tap chip to change category.

Acceptance:

- [ ] Leader cards show correct top-5 rankings per category.
- [ ] Player tap navigates to local player detail when crosswalk exists.
- [ ] Graceful fallback when crosswalk ID is unavailable (show stats only, no deep link).

## Phase 13: Scorekeeper

- Ref spec: `docs/specs/mobile.md` § 10 — Scorekeeper

### Storage & Data Model

- [x] Define Drift relational schema for scorecard:
  - `ScorecardGame`: uuid, away/home team names, venue, date, status (in_progress | final), innings list.
  - `ScorecardInning`: half (top/bottom), plays list.
  - `ScorecardPlay`: batter name/pos, outcome code, putout sequence, pitch log (`List<String>` Retrosheet codes), base_state_before, base_state_after, rbi, scored flag.
- [x] Register Drift database tables and repository for `scorecards`.

### Navigation

- [x] Add Scorekeeper entry to More tab hub list.
- [x] Route `more/scorekeeper` → `ScorecardHubScreen`.
- [x] Route `more/scorekeeper/new` → `GameSetupScreen`.
- [x] Route `more/scorekeeper/{uuid}` → `ActiveScoringScreen`.
- [x] Route `more/scorekeeper/{uuid}/grid` → `ScorecardGridScreen` (swipe from active scoring).
- [x] Route `more/scorekeeper/{uuid}/export` → `ExportSheet` (bottom sheet).

### Hub Screen

- [ ] `ScorecardHubScreen`:
  - Load all saved `ScorecardGame` records from Drift, sorted by last-modified descending.
  - Filter chips: All / In Progress / Final.
  - Each row: team names, status badge, score, venue/date, pitch count, action buttons (Resume · Box Score · Export).
  - Empty state with illustration and "Start your first scorecard" prompt.
  - New Game FAB (primary blue button).
  - Import button in header (future: paste JSON or scan QR).

### Game Setup Screen

- [ ] `GameSetupScreen`:
  - Away/home team name text fields + 3-letter abbreviation auto-generated.
  - Venue and date fields (date picker).
  - Optional lineup table: 9 rows, each with batting order number, name input, position input.
  - "Import from API" chip: fetch current roster from `/v1/mlb/teams/{mlb_id}` and populate names.
  - Start Scoring button → create local record, navigate to `ActiveScoringScreen`.

### Active Scoring Screen

- [ ] `ActiveScoringScreen`:
  - Persistent score header: away/home scores, inning/half badge with live dot.
  - Compact linescore (horizontal scroll; innings 1–9 + R/H/E totals).
  - Current batter strip: batting order #, name, position, season AVG (fetched from crosswalk if available, else blank).
  - `BaseDiamondPainter` (`CustomPainter`): 96×96 dp diamond; occupied bases highlighted amber; SVG path lines trace runner paths.
  - Count panel: ball dots (blue), strike dots (red), out dots (white); pitch counter.
  - Pitch log toggle: optional per-pitch Retrosheet code entry (B/C/S/F/X) with tap-to-add dots row.
  - Outcome button grid (see spec § 10 for groupings); all buttons min 48dp height.
  - Positional out selector: 9-button grid (1–9 with position abbreviations); tap to build sequence; confirm button records the play.
  - After recording: advance batter, update linescore, update base state, trigger haptics.
  - Undo FAB: revert last recorded play and restore previous state.
  - Swipe-right gesture → `ScorecardGridScreen`.
- [ ] `ScoringBloc`: manages game state (current inning, half, batter index, base state, count, outs); pure in-memory during play; persists to Drift on each play recorded.

### Scorecard Grid Screen

- [ ] `ScorecardGridScreen`:
  - Team toggle (away / home).
  - Horizontally scrollable table: rows = batting order, columns = innings.
  - Each cell: `AbCellPainter` — 42×48 dp; mini diamond (22×22 SVG) with traced path lines (green = scored, red = out, blue = reach); outcome code below; run dot top-right if scored.
  - Totals columns: AB, R, H, RBI per batter.
  - Innings totals row at bottom.
  - Long-press cell → bottom sheet to review or edit the play.
- [ ] `AbCellWidget` using `CustomPainter` for the mini diamond; paths derived from `base_state_before → base_state_after`.

### Export

- [ ] `ExportSheet` (modal bottom sheet):
  - Three rows: PDF, JSON, Markdown.
  - **PDF**: use `pdf` package to render a two-page (one per team) traditional scorecard. Layout: game header, lineup × innings grid with notation, linescore, box score totals. Share via `share_plus`.
  - **JSON**: serialize `ScorecardGame` to the canonical schema (spec § 10). Share as `.json` file.
  - **Markdown**: render linescore as GFM table; per-batter stats as GFM table; prepend game metadata header. Share as `.md` file or copy to clipboard.
- [ ] Haptic on run scored (`HapticFeedback.mediumImpact`); haptic on each recorded PA (`HapticFeedback.lightImpact`).

Acceptance:

- [ ] Full game (9 innings, 2 teams, 27+ plate appearances per team) can be scored without crashes.
- [ ] Scorecard grid cells match the notation entered in active scoring.
- [ ] PDF export renders correctly on Letter and A4 paper sizes.
- [ ] JSON export validates against the canonical schema structure.
- [ ] Markdown export tables render correctly in GitHub Flavored Markdown.
- [ ] Undo correctly reverts the last play and restores base/count state.
- [ ] All data survives app kill and relaunch (Drift persistence).
