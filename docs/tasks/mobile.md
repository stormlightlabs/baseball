# Flutter Mobile App Tasks

- Ref spec: `docs/specs/mobile.md`
- Stack: Flutter 3.x, Dart, Flame engine, Material 3, BLoC
- Backend: Go API additions under `api/internal/`
- Scope: native mobile app + backend endpoints + design updates

## Phase 0: Project Scaffold

- [x] Initialize Flutter project in `mobile/` with Android + iOS targets.
- [x] Configure `BLoC` for state management.
- [x] Add `dio` + `retrofit` and generate typed API client from OpenAPI spec.
- [x] Set up Material 3 theme with `ColorScheme.fromSeed()` and dark mode default.
- [x] Implement team color map (30 teams → primary hex) as a static Dart map.
- [x] Add `dynamic_color` package; wrap `MaterialApp` with `DynamicColorBuilder` for fallback system palette on neutral screens.
- [x] Build bottom navigation shell (5 tabs: Home, Players, Teams, Games, More).
- [x] Configure `hive` for local cache and offline storage.

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

- [ ] Port remaining screens from mobile designs

Acceptance:

- [ ] All 5 main tabs functional with real API data.
- [x] Team color theming applies on player and team detail screens.
- [ ] Theme transitions animate smoothly (300ms ease-out via `AnimatedTheme`).

## Phase 2: Spray Chart

### Backend

- [ ] Add `GET /api/internal/spray-chart/{player_id}` endpoint in Go.
  - Query params: `season`, `vs` (L/R), `park_id`.
  - Map Retrosheet hit-location codes to standardized field coordinates (origin at home plate, y toward CF, feet).
  - Include park wall geometry as `[angle_deg, distance_ft]` control points.
  - Return batted-ball events with `x`, `y`, `result`, `exit_velo`, `launch_angle`, `pitcher`, `pitch_type`, `date`.
- [ ] Add park dimension data to the database (wall distances at standard survey angles for each park).
- [ ] Seed park geometry for all active parks (30 current + historically significant parks).

### Frontend

- [ ] Create `SprayChartPainter` extending `CustomPainter`:
  - Draw infield diamond (bases, foul lines, dirt cutout).
  - Draw outfield wall from park geometry Bézier control points.
  - Plot batted-ball events as circles (radius ∝ exit velo, color ∝ result).
  - Implement `hitTest` for tap detection on individual events.
- [ ] Wrap in `InteractiveViewer` for pinch-to-zoom and pan.
- [ ] Add filter controls: vs LHP/RHP, pitch type, count, season picker.
- [ ] Add park overlay toggle (compare home vs. away park walls).
- [ ] Tap event detail sheet: date, opponent, pitcher, exit velo, launch angle, result.
- [ ] Haptic feedback on tap (`HapticFeedback.selectionClick`).
- [ ] Add as new tab on Player Detail screen.

Acceptance:

- [ ] Spray chart renders 500+ events at 60fps.
- [ ] Park wall overlay scales correctly with zoom.
- [ ] Tapping an event shows detail bottom sheet with correct data.

## Phase 3: Pitch Tunnel Explorer

### Backend

- [ ] Add `GET /api/internal/pitch-tunnel/{pitcher_id}` endpoint.
  - Query params: `season`, `pitch_types` (comma-separated codes).
  - Aggregate pitch data by type: avg velocity, avg spin rate, spin axis, pfx_x, pfx_z, release point, usage %.
  - Return per-type trajectory parameters.
- [ ] Ensure pitch-level data includes movement vectors (pfx_x, pfx_z) where available.

### Frontend

- [ ] Create `PitchTunnelGame` Flame component:
  - Fixed camera at batter's eye position (3.5 ft height, 1 ft behind plate).
  - Render strike zone as semi-transparent rectangle.
  - Compute pitch trajectories as cubic Bézier curves using Nathan physics model (RK4 integration with drag + Magnus force).
  - Color-code by pitch type with labeled legend.
- [ ] Implement view rotation: batter's eye, catcher, overhead, side (swipe gesture).
- [ ] Add tunnel point markers (last point where two pitch types are within perceptual threshold).
- [ ] Add time scrubber: slider to animate pitch progression from release to plate.
- [ ] Tap trajectory to isolate and show pitch details.
- [ ] Haptic bump at tunnel point and plate crossing.
- [ ] Add as sub-view under Player Detail → Pitching tab, and standalone under More.

Acceptance:

- [ ] 6 simultaneous trajectories render at 60fps.
- [ ] View rotation is smooth and gesture-driven.
- [ ] Tunnel points visually mark where pitch types diverge.

## Phase 4: At-Bat Sequencer

### Backend

- [ ] Add `GET /api/internal/at-bat/{game_id}/{ab_num}` endpoint.
  - Return batter/pitcher info, pitch sequence (type, speed, location x/z, call, count), and at-bat result.
  - Derive from existing `plays` and `pitches` tables.

### Frontend

- [ ] Create `AtBatSequencerWidget`:
  - `CustomPainter` draws strike zone scaled to batter height.
  - Pitches animate in with trajectory arc (200ms per pitch).
  - Count display updates with each pitch.
  - Horizontal timeline below zone shows sequence as labeled dots.
- [ ] Auto-play mode: pitches appear at 1.5s intervals.
- [ ] Manual mode: swipe left/right to step through.
- [ ] Tap pitch dot for detail (type, speed, result, count).
- [ ] Long-press for pitcher's typical pattern comparison overlay.
- [ ] Haptic tick on each pitch arrival (`HapticFeedback.lightImpact`).
- [ ] Integrate into Game Detail: tap any plate appearance row to open sequencer.

Acceptance:

- [ ] Pitch locations render accurately within the strike zone.
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

- [ ] Add `GET /api/internal/quiz/situation` endpoint:
  - Pull random historical game state (inning, outs, runners, score, count).
  - Include actual outcome and win expectancy from materialized view.
- [ ] Add `GET /api/internal/quiz/pitch-type` endpoint:
  - Pull pitch trajectory data for identification challenge.
  - Include 4 pitch type options with correct answer.

### Frontend

- [ ] Create Learning Mode hub screen under More tab.
- [ ] **Rules & Scoring module**: animated diagrams explaining infield fly, balk, tag-up, force play.
  - Use Rive or Lottie animations for field diagrams.
- [ ] **Pitch identification trainer**:
  - Simplified pitch tunnel view showing single trajectory.
  - 4-option multiple choice (e.g., four-seam, slider, changeup, curve).
  - Difficulty scaling: reduce tunnel time, add similar pitch types.
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
- [ ] Local progress tracking via Hive (correct answers, streaks, completion %).

Acceptance:

- [ ] All 5 learning modules functional.
- [ ] Pitch identification trainer uses actual pitch trajectory rendering.
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
  - Use `GET /v1/mlb/crosswalk/teams?season={year}` for local routing IDs.
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
  - Cache last response in Hive for offline display.
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
  - Cache in Hive for offline.
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
  - Cache in Hive; refresh on pull-to-refresh.
- [ ] Tap player row → Player Detail (via crosswalked `player_id`).
- [ ] Swipe or tap chip to change category.

Acceptance:

- [ ] Leader cards show correct top-5 rankings per category.
- [ ] Player tap navigates to local player detail when crosswalk exists.
- [ ] Graceful fallback when crosswalk ID is unavailable (show stats only, no deep link).

## Phase 12: Design Updates

- [ ] Add spray chart screen to `docs/designs/mobile/`:
  - Full-field view with park overlay and hit dots.
  - Filter controls and detail bottom sheet.
- [ ] Add pitch tunnel screen:
  - Batter's-eye perspective with multiple colored trajectories.
  - View rotation controls and time scrubber.
- [ ] Add at-bat sequencer screen:
  - Strike zone with plotted pitches and timeline.
  - Auto-play and manual controls.
- [ ] Add stat card screen:
  - Card template with pixel-art avatar.
  - Share action.
- [ ] Add learning mode screens:
  - Hub with module tiles.
  - Pitch identification trainer.
  - Situation quiz with diamond visualization.
- [ ] Add live scoreboard screen:
  - Horizontal game card carousel with team colors and linescore.
  - LIVE badge, score display, inning indicator.
- [ ] Add standings screen:
  - Division-grouped table with sortable columns.
  - AL/NL segment control.
- [ ] Add live game tracker screen:
  - Linescore grid, diamond with runners, count dots.
  - Win probability sparkline.
  - Recent plays list.
- [ ] Add today's leaders widget:
  - Stat category cards with ranked player lists.
- [ ] Update `docs/designs/mobile/index.html` to include new screens in the gallery.
- [ ] Update `docs/designs/mobile/players.html` to show spray chart and pitch tunnel tabs.
- [ ] Update `docs/designs/mobile/games.html` to show at-bat sequencer and live game tracker entry points.
- [ ] Update `docs/designs/mobile/home.html` to show scoreboard and leaders widgets.

Acceptance:

- [ ] All new feature screens match existing design system (colors, typography, spacing, dark mode).
- [ ] Index gallery shows all screens including new additions.
