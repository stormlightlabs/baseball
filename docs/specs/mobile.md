---
title: Flutter Mobile App
updated: 2026-04-21
---

## Context

The web dashboard (SvelteKit SPA) covers the full API surface but is constrained by the browser sandbox: no GPU-accelerated custom rendering, no haptic feedback, no platform-native transitions, and limited offline capability. A Flutter app targeting Android and iOS can exploit native GPU canvas, platform gesture systems, and Material You theming to deliver experiences the web cannot — interactive spray charts with park overlays, pitch-pattern exploration, haptic-rich at-bat sequencing, and per-team dynamic color schemes.

The app is not a port of the web dashboard. It is a native companion that focuses on touch-first visualization and baseball education, backed by the same Go API and new public route families for mobile-shaped payloads.

## Stack

- **Framework**: Flutter 3.x (Dart), single codebase for Android + iOS[^1]
- **Charts**: `fl_chart` for standard charts (line, bar, sparklines, radar)[^12]; `CustomPainter` for domain-specific visualizations only[^2]
- **Interactive canvases**: `CustomPainter` + `InteractiveViewer` for spray and pattern visualizations[^2]
- **Theming**: Material 3 with `ColorScheme.fromSeed()` for per-team dynamic color[^4]
- **Haptics**: `HapticFeedback` class + `haptic_feedback` package for tactile responses[^5]
- **Animations**: `flutter_animate` for declarative staggered/sequenced animations[^13]; Hero transitions; `AnimatedSwitcher` for in-place swaps; Rive/Lottie for micro-interactions
- **Networking**: `dio` for HTTP, `retrofit` for typed API client generation
- **Caching**: `drift` (SQLite) for offline-first local storage
- **State**: `BLoC` for reactive state management

## Architecture

### API Surface

The app consumes the existing `/api/v1/` endpoints (80+ routes) and introduces a new `/internal/mobile/` namespace for mobile-optimized endpoints that aggregate multiple queries or return pre-shaped payloads:

| Internal Endpoint                                  | Purpose                                                                                       |
| -------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `GET /internal/mobile/player-card/{id}`            | Aggregated player bio + key stats + recent game log for stat card rendering                   |
| `GET /internal/mobile/spray-chart/{player_id}`     | Batted-ball coordinates from Retrosheet location/hit-type mapping, plus park overlay metadata |
| `GET /internal/mobile/pitch-patterns/{pitcher_id}` | Pitch-call sequence patterns and divergence summaries by handedness/count context             |
| `GET /internal/mobile/at-bat/{game_id}/{ab_num}`   | Full at-bat sequence: pitch-by-pitch with context, count, and result                          |
| `GET /internal/mobile/quiz/situation`              | Random game situation for learning mode quizzes                                               |
| `GET /internal/mobile/quiz/pitch-type`             | Pitch-pattern identification challenge from historical pitch-call sequences                   |

These endpoints reduce round-trips and shape data for rendering. They are not authenticated via API keys — they use a client certificate or app-token scheme separate from the public API.

### Navigation

The app preserves the 5-tab bottom navigation from the mobile designs:

```text
⌂ Home  |  ◎ Players  |  ◈ Teams  |  ◷ Games  |  ≡ More
```

New features surface within existing tabs or under More:

- **Spray Chart** → Player detail (new tab alongside Batting, Pitching, etc.)
- **Pitch Patterns** → Player detail (Pitching sub-view) or standalone under More
- **At-Bat Sequencer** → Game detail (tap any plate appearance)
- **Stat Card Generator** → Player detail (share action)
- **Learning Mode** → More tab

## Feature Specifications

### 1. Interactive Spray Chart with Park Overlay

A touch-interactive batted-ball map rendered on a `CustomPainter` canvas, overlaid on the specific ballpark's field geometry.

**Rendering approach**: `CustomPainter` draws the field outline and batted-ball events mapped from Retrosheet location codes. Event dots are colored by result and can vary by derived depth bucket (infield/shallow/deep). The canvas sits inside an `InteractiveViewer` for pinch-to-zoom and pan.[^6]

**Data requirements**:

- Batted-ball x/y coordinates derived from Retrosheet `loc` + `hittype` mapping.
- Event context fields from `plays`: date, game, inning, outs, batter/pitcher handedness, and count.
- Park overlay metadata keyed by `park_id` when available; fallback to standardized generic field geometry.

**Interactions**:

- Pinch to zoom into a region of the field
- Tap a hit dot to see detail (date, opponent, pitcher, result, loc code, hit-type, inning/outs/count context)
- Filter controls: vs LHP/RHP, by hit type, by count, by season
- Toggle park overlay between home parks and compare overlays
- Haptic pulse on tap (`HapticFeedback.selectionClick`)

**Park geometry**: A dedicated ingestion workstream populates per-park wall definitions. Until park geometry is available for a venue, the UI uses a standardized generic field template and clearly labels the overlay mode.

### 2. Pitch Pattern Explorer

A sequence-pattern visualization of pitcher behavior using Retrosheet pitch-call codes (`B`, `C`, `F`, `S`, `X`, etc.), highlighting where patterns diverge by count and batter handedness.

**Rendering approach**: `CustomPainter` + chart widgets render pattern lanes, transition bands, and sequence divergence markers. The main view emphasizes call-code sequences and frequency, not physical pitch trajectory simulation.

**Pattern model**: Aggregate `plays.pitches` into:

- Usage and rate by call code.
- Sequence transitions (e.g., `B->C`, `C->X`) by count state.
- Split views by batter handedness (`vs_bat=L|R`) and season.
- Divergence markers where two pattern families separate after shared prefixes.

**Interactions**:

- Swipe or tap to switch pattern lenses (overall, vs LHB, vs RHB, two-strike, hitter's counts).
- Tap a sequence lane to isolate and show supporting stats.
- Toggle divergence markers for shared-prefix pattern families.
- Use a scrubber to step through sequence depth (pitch 1, pitch 2, pitch 3+).
- Haptic bump on selected divergence points.

**Data requirements**: Retrosheet pitch-call sequence data from `plays.pitches`, with count/handedness context from `plays`. No Statcast movement vectors are required.

### 3. At-Bat Sequencer

A pitch-by-pitch replay of any plate appearance, showing sequence strategy and count progression from historical pitch-call data.

**Rendering approach**: `CustomPainter` draws a strike-zone frame and sequence markers by pitch order. The count display updates with each pitch. Below the zone, a horizontal timeline shows the pitch sequence as labeled dots.

**Interactions**:

- Auto-play mode: pitches animate in at 1.5s intervals
- Manual mode: swipe left/right to step through pitches
- Tap any pitch dot to see call code, result, and count
- Long-press to compare this at-bat's sequence against the pitcher's typical patterns
- Haptic tick on each pitch arrival (`HapticFeedback.lightImpact`)

**Data source**: `GET /internal/mobile/at-bat/{game_id}/{ab_num}` returns the full sequence.
Falls back to `GET /v1/games/{id}/pitches` filtered to the specific plate appearance.

### 4. Stat Card Generator

Shareable, visually rich player cards with pixel-art avatars, key stats, and team-colored backgrounds.

**Design**: A card template (~1080×1920 or ~1080×1080 for social sharing) with:

- Pixel-art player avatar (see below)
- Player name, position, team in team-branded typography
- 4-6 headline stats with sparkline trends
- Team accent color gradient background
- Season or career scope indicator
- Big Fly branding watermark

**Pixel-art avatars**: Since real player photos cannot be used, the app generates deterministic pixel-art portraits from player attributes. The `pixel_art_generator` package[^8] provides template-based sprite generation. Inputs: a seed derived from `player_id`, mapped to hair style, skin tone, cap color (team primary), jersey color (team secondary), and facial features. The avatar is rendered at 32×32 native resolution, displayed at 128×128 with nearest-neighbor scaling for crisp pixel edges.

**Sharing**: `RenderRepaintBoundary` captures the card widget as a PNG. Share via platform share sheet (`share_plus` package).

### 5. Baseball Learning Mode

An interactive educational module that teaches baseball rules, strategy, and analytics through guided lessons and quizzes.

**Modules**:

| Module             | Format                 | Content                                                                   |
| ------------------ | ---------------------- | ------------------------------------------------------------------------- |
| Rules & Scoring    | Animated diagrams      | Infield fly rule, balk, interference, tag-up, force play                  |
| Pitch Patterns     | Identification trainer | Show pitch-call sequence snippets → user guesses likely next call/outcome |
| Situation Quiz     | Game-state prompt      | "Runner on 2nd, 1 out, 2-1 count — what's the win expectancy?"            |
| Stat Explainers    | Interactive calculator | What is WAR? Adjust inputs, see WAR change                                |
| Historical Moments | Narrated play-by-play  | Famous at-bats with the at-bat sequencer                                  |

**Pitch pattern trainer**: Uses the pitch-pattern renderer in a simplified mode. Shows a partial sequence context, user selects from 4 likely next-call/outcome options. Difficulty scales by reducing context and introducing lookalike pattern families.

**Situation quiz**: Pulls from `GET /internal/mobile/quiz/situation` which returns a real historical game state. User guesses outcome or win probability. Compares against actual result and historical win expectancy from `GET /v1/win-expectancy`.

**Gamification**: Track correct answers, streaks, and category completion. Store locally (Drift). No server-side leaderboard in v1.

## Team Accent Colors and Dynamic Theming

### Material You Integration

Flutter's Material 3 implementation generates a full `ColorScheme` from a single seed color via `ColorScheme.fromSeed(seedColor: teamPrimary)`.[^4] This produces harmonious primary, secondary, tertiary, surface, and on-surface variants automatically. The `dynamicSchemeVariant` parameter controls the palette algorithm — `tonalSpot` (default) works well for sports branding.

At runtime, when a user navigates to a player or team screen, the app rebuilds the theme with that team's primary color as the seed:

```dart
ColorScheme.fromSeed(
  seedColor: team.primaryColor,
  brightness: Brightness.dark,
  dynamicSchemeVariant: DynamicSchemeVariant.tonalSpot,
)
```

This shifts surface tones, app bar colors, FAB colors, and ripple effects to harmonize with the team's brand without breaking contrast or accessibility guarantees — Material 3's tonal system enforces WCAG contrast ratios by construction.[^9]

### Color Palette

All 30 MLB teams' official primary and secondary colors[^10]:

| Team                  | Primary   | Secondary |
| --------------------- | --------- | --------- |
| Arizona Diamondbacks  | `#A71930` | `#E3D4AD` |
| Atlanta Braves        | `#CE1141` | `#13274F` |
| Baltimore Orioles     | `#DF4601` | `#000000` |
| Boston Red Sox        | `#BD3039` | `#0C2340` |
| Chicago Cubs          | `#0E3386` | `#CC3433` |
| Chicago White Sox     | `#27251F` | `#C4CED4` |
| Cincinnati Reds       | `#C6011F` | `#000000` |
| Cleveland Guardians   | `#E31937` | `#0C2340` |
| Colorado Rockies      | `#33006F` | `#000000` |
| Detroit Tigers        | `#0C2340` | `#FA4616` |
| Houston Astros        | `#002D62` | `#EB6E1F` |
| Kansas City Royals    | `#004687` | `#BD9B60` |
| Los Angeles Angels    | `#BA0021` | `#003263` |
| Los Angeles Dodgers   | `#005A9C` | `#FFFFFF` |
| Miami Marlins         | `#000000` | `#0077C8` |
| Milwaukee Brewers     | `#12284B` | `#FFC52F` |
| Minnesota Twins       | `#002B5C` | `#D31145` |
| New York Mets         | `#002D72` | `#FF5910` |
| New York Yankees      | `#132448` | `#C4CED4` |
| Oakland Athletics     | `#003831` | `#EFB21E` |
| Philadelphia Phillies | `#E81828` | `#002D72` |
| Pittsburgh Pirates    | `#27251F` | `#FDB827` |
| San Diego Padres      | `#2F241D` | `#A2AAAD` |
| San Francisco Giants  | `#FD5A1E` | `#27251F` |
| Seattle Mariners      | `#0C2C56` | `#005C5C` |
| St. Louis Cardinals   | `#C41E3A` | `#0C2340` |
| Tampa Bay Rays        | `#092C5C` | `#8FBCE6` |
| Texas Rangers         | `#003278` | `#C0111F` |
| Toronto Blue Jays     | `#134A8E` | `#1D2D5C` |
| Washington Nationals  | `#AB0003` | `#14225A` |

The primary color is used as the `fromSeed` input. The secondary is available for accent elements (stat highlights, chart series) that need to contrast with the generated scheme.

### Theme Transition

When navigating from a team-colored screen (e.g., player profile with NYM blue) to another (e.g., team page for ATL red), the theme animates via `AnimatedTheme` with a 300ms ease-out curve. Surface colors cross-fade; the effect is subtle — tonal shifts, not jarring hue jumps.

### Fallback

On Android 12+, the `dynamic_color` package[^11] can also pull the user's wallpaper-derived system palette. The app defaults to team-specific theming on detail screens but falls back to the system dynamic color on neutral screens (home, search, settings).

## Design System Alignment

The Flutter app inherits the mobile design system's foundations:

| Token      | Value     | Usage                                                                  |
| ---------- | --------- | ---------------------------------------------------------------------- |
| Background | `#0d0f13` | Scaffold background (overridden by team tonal surface in detail views) |
| Surface    | `#13161c` | Cards, panels                                                          |
| Surface2   | `#1a1e27` | Secondary surfaces, hover/pressed states                               |
| Border     | `#252934` | Dividers, outlines                                                     |
| Text       | `#e2e8f0` | Primary text                                                           |
| Muted      | `#6b7280` | Secondary text, labels                                                 |
| Accent     | `#3b82f6` | Default accent (overridden by team primary on detail screens)          |
| Accent2    | `#10b981` | Success, positive states                                               |

Typography: Google Sans (titles), Google Sans Code (stats/monospace), Inter (body). Bottom nav: 5 tabs, 64px height, active tab uses team accent when on a team-scoped screen.

## API Additions (Backend)

New routes are registered under `/internal/mobile/*` in the Go backend as public API surface and should be documented in Swagger with typed contracts.

### Spray Chart Endpoint

```text
GET /internal/mobile/spray-chart/{player_id}?season={year}&vs={L|R}&park_id={id}
```

Returns batted-ball events with field coordinates:

```json
{
  "player_id": "troutmi01",
  "season": 2024,
  "park": { "id": "ANA01", "name": "Angel Stadium", "overlay_mode": "generic" },
  "events": [
    {
      "game_id": "ANA202406150",
      "play_num": 213,
      "date": "2024-06-15",
      "x": 142.3,
      "y": 287.1,
      "result": "HR",
      "loc_code": "89XD+",
      "hittype": "F",
      "pitcher": "gilbl001",
      "pitcher_hand": "R",
      "count": "1-2",
      "inning": 4,
      "outs": 2
    }
  ]
}
```

Coordinates are in a standardized field system (origin at home plate, y-axis toward CF, units in feet) derived from Retrosheet location mapping.

### Pitch Pattern Endpoint

```text
GET /internal/mobile/pitch-patterns/{pitcher_id}?season={year}&vs_bat={L|R}
```

Returns aggregated pitch-call sequence metrics and divergence groups:

```json
{
  "pitcher_id": "ohtansh01",
  "season": 2024,
  "vs_bat": "L",
  "call_usage": [
    { "call_code": "C", "label": "Called strike", "count": 412, "usage_pct": 0.278 },
    { "call_code": "B", "label": "Ball", "count": 365, "usage_pct": 0.246 }
  ],
  "transitions": [{ "from": "B", "to": "C", "count": 91, "rate": 0.249 }],
  "divergence_groups": [{ "shared_prefix": "BC", "branches": ["BCF", "BCX", "BCB"], "split_depth": 3 }]
}
```

### At-Bat Endpoint

```text
GET /internal/mobile/at-bat/{game_id}/{ab_num}
```

Returns the full pitch sequence for a specific plate appearance:

```json
{
  "game_id": "ANA202406150",
  "ab_num": 23,
  "batter": { "id": "troutmi01", "name": "Mike Trout", "stance": "R" },
  "pitcher": { "id": "coMDee01", "name": "...", "throws": "R" },
  "result": "Home Run",
  "pitches": [
    { "seq": 1, "call": "B", "description": "Ball", "count": "1-0" },
    { "seq": 2, "call": "S", "description": "Swinging strike", "count": "1-1" },
    { "seq": 3, "call": "X", "description": "Ball in play", "count": "1-2", "result": "HR" }
  ]
}
```

### Quiz Endpoints

```text
GET /internal/mobile/quiz/situation?era={modern|all}
GET /internal/mobile/quiz/pitch-type?difficulty={1|2|3}
```

These pull from existing game data and pitch data, reshaping into quiz-friendly payloads with the correct answer embedded (client reveals after user input).

## Live & Current-Season Features

The MLB Stats API proxy (`/v1/mlb/*`) provides real-time access to the current season. The app surfaces this data through three primary features that complement the historical Retrosheet-backed views. We expand `/v1/mlb/*` directly (instead of introducing additional `/internal/mobile/*` live routes) for UI-optimized payloads.

### 6. Live Scoreboard

A real-time scoreboard of today's MLB games, prominently featured on the Home tab.

**Data source**: `GET /v1/mlb/schedule?date={YYYY-MM-DD}&hydrate=linescore,team,probablePitcher&include=details` plus `GET /v1/meta/crosswalk/teams?season={year}` for canonical team/franchise mapping.

**Response shape**:

```json
{
  "date": "2026-04-19",
  "games_in_progress": 3,
  "games": [
    {
      "game_pk": 822750,
      "status": "In Progress",
      "inning": 5,
      "inning_half": "Top",
      "away": {
        "mlb_id": 137,
        "abbreviation": "SF",
        "name": "Giants",
        "score": 3,
        "record": { "wins": 9, "losses": 12 },
        "probable_pitcher": "Robbie Ray",
        "primary_color": "#FD5A1E"
      },
      "home": {
        "mlb_id": 120,
        "abbreviation": "WSH",
        "name": "Nationals",
        "score": 1,
        "record": { "wins": 7, "losses": 14 },
        "probable_pitcher": "MacKenzie Gore",
        "primary_color": "#AB0003"
      },
      "venue": "Nationals Park",
      "start_time": "2026-04-19T17:35:00Z",
      "linescore": { "innings": [...], "balls": 2, "strikes": 1, "outs": 1 }
    }
  ]
}
```

**Rendering**: Horizontal `PageView` of game cards, each showing team abbreviations, scores, inning indicator, and linescore. Cards use team primary colors as subtle gradient accents. A "LIVE" badge pulses on in-progress games.

**Interactions**:

- Swipe between game cards
- Tap a game card to navigate to Live Game Tracker (if in progress) or Game Detail (if final)
- Pull-to-refresh; auto-refresh every 30s when games are in progress
- Haptic tick on score changes during auto-refresh

**Offline**: Cache last-fetched scoreboard in Drift. Show stale data with "Last updated" timestamp when offline.

### 7. Current Standings

Division standings with current records, streaks, and wild card positioning.

**Data source**: `GET /v1/mlb/standings?season={year}&standingsTypes=regularSeason&include=details` plus `GET /v1/meta/crosswalk/teams?season={year}` for local franchise IDs and navigation.

**Response shape**:

```json
{
  "season": 2026,
  "last_updated": "2026-04-19T18:00:00Z",
  "divisions": [
    {
      "id": 201,
      "name": "AL East",
      "league": "American League",
      "teams": [
        {
          "mlb_id": 139,
          "abbreviation": "TB",
          "name": "Tampa Bay Rays",
          "wins": 14,
          "losses": 7,
          "pct": ".667",
          "games_back": "-",
          "wild_card_gb": "+3.0",
          "streak": "W4",
          "run_differential": 28,
          "last_10": { "wins": 7, "losses": 3 },
          "primary_color": "#092C5C",
          "franchise_id": "TBD"
        }
      ]
    }
  ]
}
```

**Rendering**: Vertical list grouped by division. Each division is a collapsible section with a `SortableTable`-style row per team. Division leader gets a subtle crown indicator. Wild card contenders show a separator line.

**Interactions**:

- Toggle between AL / NL / both (segment control)
- Tap a team row to navigate to Team Detail with current-season context
- Sort by W, L, PCT, GB, streak, or run differential
- Haptic on division header collapse/expand

**Navigation**: Surfaces under the Teams tab as a "Current Standings" segment alongside the existing franchise/historical team views.

### 8. Live Game Tracker

A real-time game view combining MLB live feed data with the app's win probability model.

**Data source**: `GET /v1/mlb/live/{gamePk}` proxies the MLB game feed (`/api/v1.1/game/{gamePk}/feed/live`) and merges with the local win-probability engine when play-by-play state permits. This endpoint is not cached (or cached at 15s TTL max).

**Response shape**:

```json
{
  "game_pk": 822750,
  "status": "In Progress",
  "inning": 5,
  "inning_half": "Top",
  "away": { "abbreviation": "SF", "score": 3, "hits": 7, "errors": 0 },
  "home": { "abbreviation": "WSH", "score": 1, "hits": 4, "errors": 1 },
  "linescore": {
    "innings": [
      { "num": 1, "away": { "runs": 0 }, "home": { "runs": 1 } }
    ]
  },
  "current_play": {
    "description": "Trout doubles to left field",
    "event": "Double",
    "count": { "balls": 1, "strikes": 2, "outs": 1 },
    "runners": ["1B", "3B"]
  },
  "recent_plays": [...],
  "win_probability": 0.62,
  "win_probability_series": [...]
}
```

**Rendering**: Full-screen game view with:

- Scoreboard header with linescore grid
- Diamond visualization showing runners (filled bases)
- Count indicator (balls/strikes/outs as filled dots)
- Current play description with animated transition
- Win probability sparkline (using `fl_chart`) updating in real time
- Scrollable recent plays list below

**Interactions**:

- Pull-to-refresh; auto-refresh every 15s during live games
- Tap the win probability chart to expand to full-screen historical view
- Tap a play in the recent list to see detail
- Haptic bump on scoring plays and outs
- Swipe down to return to scoreboard

**Rendering approach**: The diamond, count dots, and base indicators use a single `CustomPainter` for performance. The linescore is a standard `Row` of styled cells.

### 9. Today's Leaders

Current-season stat leaders surfaced on the Home tab below the scoreboard.

**Data source**: `/v1/mlb/stats` calls (`stats=season&group=hitting|pitching&sortStat={stat}&limit=5&include=details`) for requested categories, plus `/v1/meta/crosswalk/players` lookup routes as fallback for deep linking.

**Response shape**:

```json
{
  "season": 2026,
  "categories": [
    {
      "stat": "HR",
      "label": "Home Runs",
      "group": "hitting",
      "leaders": [
        {
          "mlb_id": 660271,
          "name": "Shohei Ohtani",
          "team_abbr": "LAD",
          "value": "8",
          "player_id": "ohtansh01",
          "primary_color": "#005A9C"
        }
      ]
    }
  ]
}
```

**Rendering**: Horizontal `PageView` of leader category cards. Each card shows a ranked list of 5 players with team-colored accent bars. The active stat category is shown as a chip row above the card.

**Interactions**:

- Swipe between stat categories
- Tap a chip to jump to a specific category
- Tap a player row to navigate to Player Detail
- Category chips: HR, AVG, OPS, RBI, SB (hitting) | ERA, SO, W, SV, WHIP (pitching)

## Expanded `/v1/mlb/*` Namespace

The MLB namespace provides UI-optimized live endpoints as part of the public API contract. These routes aggregate multiple data sources, map IDs across systems, and return pre-shaped payloads for rendering.

### Authentication

Live MLB endpoints use the same auth and API key policies as the rest of `/v1/*`.

### Live Endpoint Catalog

| Endpoint                                           | Source                                       | Purpose                                                                                 |
| -------------------------------------------------- | -------------------------------------------- | --------------------------------------------------------------------------------------- |
| `GET /internal/mobile/player-card/{id}`            | `/v1/players/*` (aggregated)                 | Aggregated player bio + key stats + recent game log for stat card rendering             |
| `GET /internal/mobile/spray-chart/{player_id}`     | Retrosheet hit-location data + park geometry | Batted-ball coordinate mapping with contextual play metadata and optional park overlays |
| `GET /internal/mobile/pitch-patterns/{pitcher_id}` | `plays.pitches` sequence data                | Pitch-call usage, transitions, and divergence patterns by handedness/count context      |
| `GET /internal/mobile/at-bat/{game_id}/{ab_num}`   | `plays` + `pitches` tables                   | Full at-bat sequence: pitch-by-pitch with context, count, and result                    |
| `GET /internal/mobile/quiz/situation`              | Historical game states                       | Random game situation for learning mode quizzes                                         |
| `GET /internal/mobile/quiz/pitch-type`             | Pitch-call sequence data                     | Pitch-pattern identification challenge                                                  |
| `GET /v1/mlb/schedule`                             | MLB schedule feed                            | Today's games with scores, status, linescore, and team details                          |
| `GET /v1/mlb/standings`                            | MLB standings feed                           | Current standings data by league/division                                               |
| `GET /v1/meta/crosswalk/teams`                     | `team_mlbam_map` + local team/franchise map  | MLB team ID → local `team_id` / `franchise_id` mapping for navigation                   |
| `GET /v1/mlb/live/{gamePk}`                        | MLB game feed + win probability engine       | Real-time game state with play-by-play and win probability                              |
| `GET /v1/mlb/stats`                                | MLB stats feed                               | Current-season stat leaders and split queries                                           |
| `GET /v1/mlb/people/{mlb_id}`                      | MLB people feed + local search lookups       | Current-season player stats with local player routing                                   |

### ID Crosswalk

Live endpoints bridge MLB Stats API IDs (MLBAM `personId`, `teamId`) to local Retrosheet/Lahman IDs (`player_id`, `team_id`, `franchise_id`). Team/player mappings are exposed via `GET /v1/meta/crosswalk/*`, and `include=details` sidecars on `/v1/mlb/*` responses provide opt-in enrichment for labels and ID maps.

### Caching Strategy

| Endpoint group    | Cache TTL | Rationale                                        |
| ----------------- | --------- | ------------------------------------------------ |
| Scoreboard        | 30s       | Scores update frequently during games            |
| Standings         | 5min      | Changes only after games complete                |
| Live game feed    | 15s       | Near-real-time without overwhelming upstream     |
| Leaders           | 15min     | Stats update after games; no need for sub-minute |
| Player/team live  | 5min      | Bio/roster data changes infrequently             |
| Spray/patterns/ab | 1hr       | Historical data; changes only on data loads      |
| Quiz              | No cache  | Should return varied results                     |

## 10. Scorekeeper

An offline-first manual scorecard tool for fans scoring games in real time. Uses traditional baseball notation (positional numbers 1–9, Retrosheet pitch codes) and exports in three formats.

### Navigation

Accessible from the More tab as a standalone entry. A persistent scorecard list (hub) is the landing screen; in-game scoring is a full-screen modal that replaces the bottom nav during active play.

### Screens

| Screen         | Purpose                                                                                                            |
| -------------- | ------------------------------------------------------------------------------------------------------------------ |
| Hub            | Saved scorecard list with resume, box score, and export shortcuts; New Game button                                 |
| Game Setup     | Team names, venue/date, optional batting lineup (9 rows); Import from API populates names from current roster data |
| Active Scoring | Score header + linescore, base-state diamond, count/outs, optional pitch-by-pitch log, outcome buttons             |
| Scorecard Grid | Traditional batter × inning grid with mini-diamond cells showing notation and base advancement                     |
| Export Sheet   | Format selector: PDF, JSON, Markdown                                                                               |

### Notation Model

The app uses standard baseball scorecard notation:

- **Fielding positions**: 1 (P), 2 (C), 3 (1B), 4 (2B), 5 (3B), 6 (SS), 7 (LF), 8 (CF), 9 (RF)
- **Plate appearance outcomes**: `BB`, `HBP`, `CI` (reach base); `1B`, `2B`, `3B`, `HR` (hits); `K` (swinging strikeout), `Ꝁ` (called strikeout); `FC`, `E#`, `SAC`, `SF`, `DP`, `TP`
- **Putout sequences**: positional strings separated by `–` (e.g., `6–3`, `4–6–3`)
- **Pitch log (optional)**: Retrosheet call codes per pitch — `B` (ball), `C` (called strike), `S` (swinging strike), `F` (foul), `X` (ball in play)
- **Scoring**: Diamond in each cell is traced by path lines as the runner advances; filled/circled when a run scores

### Outcome Buttons

Touch targets (min 48px height) organized in rows:

| Group           | Buttons                                                   |
| --------------- | --------------------------------------------------------- |
| Hits            | 1B · 2B · 3B · HR                                         |
| Reach / Special | BB · HBP · FC · E                                         |
| Strikeouts      | K (swing) · Ꝁ (called) · SAC                              |
| Other           | SF · CI · DP                                              |
| Ground/fly out  | 9-position selector → sequence picker → Record Out / + DP |

After selecting fielder sequence and confirming, the app records the putout code, advances runners, and moves to the next batter.

### Base State & Count Panel

- **Diamond `CustomPainter`**: infield diamond with four base squares; bases highlighted amber when occupied; SVG path lines trace runner advancement within the cell
- **Count dots**: balls (blue), strikes (red), outs (white), matching the Live Game Tracker visual language
- **Pitch counter**: running total for pitcher pitch count

### Scorecard Grid

Horizontal scroll; rows = batters, columns = innings. Each cell (42×48 dp) contains:

1. Miniature diamond SVG (22×22) with traced path lines colored by result (green = scored, red = out, blue = reach)
2. Outcome code label below the diamond (`1B`, `K`, `6–3`, etc.)
3. Run dot (top-right corner) when the batter scored

Total columns: AB, R, H, RBI per batter. Inning row at bottom shows runs per inning.

### Export Formats

| Format   | Content                                                                                                                                                            |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| PDF      | Print-ready traditional scorecard; 2 pages (one per team); Letter or A4; includes game header, lineup, inning cells, linescore, and box score totals               |
| JSON     | Full play-by-play structured data: `{ game, teams, innings: [{ half, plays: [{ batter, outcome, pitches, base_state_before, base_state_after, rbi, scored }] }] }` |
| Markdown | Linescore table + per-batter stats table; paste-ready for Obsidian, Notion, or any notes app                                                                       |

Export uses `pdf` package (Flutter) for PDF rendering and `share_plus` for the platform share sheet.

### Interactions

- New game → setup form → Start Scoring; returns to hub on game completion or back-button
- Swipe right from active scoring → scorecard grid (both teams via toggle)
- Long-press a scorecard cell → edit / annotate play
- Haptic tick (`HapticFeedback.lightImpact`) on each recorded plate appearance outcome
- Haptic bump (`HapticFeedback.mediumImpact`) on run scored
- Undo last action (FAB secondary action)

### Storage

All scorecard data stored in Drift/SQLite, keyed by game UUID generated at setup. No server sync in v1 — fully offline. Exported files are ephemeral; only the local record is persisted.

## Animations

### Approach

The app uses `flutter_animate` as the primary animation toolkit for declarative, composable motion. It provides a chainable API for fades, slides, scales, shimmers, and custom effects without manual `AnimationController` boilerplate. Heavier one-off animations (logo reveals, achievement badges) use Rive or Lottie files.

### Animation Catalog

| Context                        | Effect                                                                | Timing                   |
| ------------------------------ | --------------------------------------------------------------------- | ------------------------ |
| List item entrance             | `fadeIn` + `slideY(begin: 0.05)` with stagger interval                | 200ms per item, 50ms gap |
| Card expand/collapse           | `fadeIn` + `scaleXY(begin: 0.97)` ease-out                            | 250ms                    |
| Tab/segment switch content     | `fadeIn` + `slideX` directional (left/right based on tab index delta) | 200ms                    |
| Stat value count-up            | `custom` effect interpolating numeric text from 0 → value             | 400ms ease-out           |
| Chip/filter selection          | `scale(begin: 0.95)` + `fadeIn` on newly visible filtered content     | 150ms                    |
| Scoreboard card entrance       | `fadeIn` + `slideX(begin: 0.1)` staggered per card                    | 300ms, 80ms stagger      |
| Live badge pulse               | `shimmer` looping with team accent color                              | 1.5s loop                |
| Score change flash             | `tint(color: accent)` + `scale(begin: 1.05)` then settle              | 300ms                    |
| Spray chart dot entrance       | `fadeIn` + `scaleXY(begin: 0)` staggered radially from home plate     | 15ms stagger, 200ms each |
| Pitch sequence step            | `fadeIn` + `slideY(begin: -0.1)` per pitch marker                     | 200ms per pitch          |
| Win probability line draw      | `custom` effect animating `LineChart` path length from 0% → 100%      | 600ms ease-in-out        |
| Theme transition               | `AnimatedTheme` (built-in), 300ms ease-out — not `flutter_animate`    | 300ms                    |
| Page route transitions         | Hero + `fadeIn`/`slideY` on non-hero content                          | 300ms                    |
| Empty state / error appearance | `fadeIn` + `slideY(begin: 0.08)` + `scale(begin: 0.98)`               | 350ms                    |

### Guidelines

- **Stagger, don't flood**: when multiple items animate in (lists, grids, chart dots), use `interval` or explicit delays so items cascade rather than all appearing at once.
- **Keep durations short**: most transitions 150–300ms; nothing over 600ms except looping decorative effects.
- **Respect reduced motion**: check `MediaQuery.disableAnimations` and skip or shorten all `flutter_animate` chains when true (the library supports `.toggle(animate: !reduceMotion)`).
- **No animation on scroll**: items already visible during scroll should not re-animate. Use `Animate.onPlay` or visibility-gated triggers, not rebuild-based animations.
- **Pair with haptics**: significant state changes (score update, play recorded, quiz answer) combine a visual animation with the corresponding haptic from the spec.

## Performance Targets

| Metric                           | Target                        |
| -------------------------------- | ----------------------------- |
| Spray chart render (500 events)  | < 16ms per frame (60fps)      |
| Pitch pattern explorer (6 lanes) | < 16ms per frame              |
| Theme switch (team color change) | < 300ms transition            |
| Cold start to interactive        | < 2s on mid-range device      |
| Offline stat card generation     | Full capability (cached data) |
| Scoreboard refresh (15 games)    | < 500ms end-to-end            |
| Live game feed refresh           | < 300ms end-to-end            |
| Standings render (30 teams)      | < 16ms per frame              |
| APK size                         | < 25 MB                       |

## References

[^1]: Flutter. "Flutter - Build apps for any screen." [flutter.dev](https://flutter.dev). Cross-platform framework targeting Android, iOS, web, and desktop from a single Dart codebase.

[^2]: Flutter API. "CustomPainter class." [api.flutter.dev/flutter/rendering/CustomPainter-class.html](https://api.flutter.dev/flutter/rendering/CustomPainter-class.html). Low-level canvas drawing API for custom shapes, paths, and hit testing.

[^4]: Flutter API. "ColorScheme.fromSeed constructor." [api.flutter.dev/flutter/material/ColorScheme/ColorScheme.fromSeed.html](https://api.flutter.dev/flutter/material/ColorScheme/ColorScheme.fromSeed.html). Generates a full Material 3 color scheme from a single seed color using HCT perceptual color space.

[^5]: Flutter Gems. "haptic_feedback package." [pub.dev/packages/haptic_feedback](https://pub.dev/packages/haptic_feedback). Cross-platform haptic feedback with granular intensity control beyond the built-in `HapticFeedback` class.

[^6]: Flutter API. "InteractiveViewer class." [api.flutter.dev/flutter/widgets/InteractiveViewer-class.html](https://api.flutter.dev/flutter/widgets/InteractiveViewer-class.html). Built-in widget for pan and zoom with `Matrix4` transformation, boundary constraints, and scale limits.

[^8]: pub.dev. "pixel_art_generator package." [pub.dev/packages/pixel_art_generator](https://pub.dev/packages/pixel_art_generator). Template-based pixel art sprite generation in Flutter.

[^9]: Material Design. "Color system." [m3.material.io/styles/color](https://m3.material.io/styles/color). Material 3 tonal palette system ensures accessible contrast ratios across generated color roles.

[^10]: Team Color Codes. "MLB Team Color Codes." [teamcolorcodes.com/mlb-color-codes](https://teamcolorcodes.com/mlb-color-codes/). Official hex, RGB, and Pantone values for all 30 MLB teams. Cross-referenced with [teampalettes.com/mlb](https://teampalettes.com/mlb).

[^11]: pub.dev. "dynamic_color package." [pub.dev/packages/dynamic_color](https://pub.dev/packages/dynamic_color). Material.io team package (v1.8.1) for Android 12+ wallpaper-derived dynamic color schemes via `DynamicColorBuilder`.

[^12]: pub.dev. "fl_chart package." [pub.dev/packages/fl_chart](https://pub.dev/packages/fl_chart). Highly customizable Flutter chart library supporting line, bar, pie, scatter, and radar charts with touch interactions, animations, and theming. 7k+ GitHub stars.

[^13]: pub.dev. "flutter_animate package." [pub.dev/packages/flutter_animate](https://pub.dev/packages/flutter_animate). Declarative animation library by gskinner providing chainable effects (fade, slide, scale, shimmer, blur, tint, custom) with stagger/interval support, reduced-motion awareness, and adapter architecture.
