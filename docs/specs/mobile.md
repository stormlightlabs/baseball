---
title: Flutter Mobile App
updated: 2026-04-18
---

## Context

The web dashboard (SvelteKit SPA) covers the full API surface but is constrained by the browser sandbox: no GPU-accelerated custom rendering, no haptic feedback, no platform-native transitions, and limited offline capability. A Flutter app targeting Android and iOS can exploit native GPU canvas, platform gesture systems, and Material You theming to deliver experiences the web cannot — interactive spray charts with park overlays, 3D pitch tunnel exploration, haptic-rich at-bat sequencing, and per-team dynamic color schemes.

The app is not a port of the web dashboard. It is a native companion that focuses on touch-first visualization and baseball education, backed by the same Go API and new public route families for mobile-shaped payloads.

## Stack

- **Framework**: Flutter 3.x (Dart), single codebase for Android + iOS[^1]
- **Charts**: `fl_chart` for standard charts (line, bar, sparklines, radar)[^12]; `CustomPainter` for domain-specific visualizations only[^2]
- **Game-like scenes**: Flame engine for complex interactive canvases (pitch tunnel)[^3]
- **Theming**: Material 3 with `ColorScheme.fromSeed()` for per-team dynamic color[^4]
- **Haptics**: `HapticFeedback` class + `haptic_feedback` package for tactile responses[^5]
- **Animations**: Hero transitions, `AnimatedSwitcher`, Rive/Lottie for micro-interactions
- **Networking**: `dio` for HTTP, `retrofit` for typed API client generation
- **Caching**: `hive` or `isar` for offline-first local storage
- **State**: `BLoC` for reactive state management

## Architecture

### API Surface

The app consumes the existing `/api/v1/` endpoints (80+ routes) and introduces a new `/v1/mobile/` namespace for mobile-optimized endpoints that aggregate multiple queries or return pre-shaped payloads:

| Internal Endpoint                             | Purpose                                                                     |
| --------------------------------------------- | --------------------------------------------------------------------------- |
| `GET /v1/mobile/player-card/{id}`          | Aggregated player bio + key stats + recent game log for stat card rendering |
| `GET /v1/mobile/spray-chart/{player_id}`   | Batted-ball coordinates with launch angle, exit velocity, and park geometry |
| `GET /v1/mobile/pitch-tunnel/{pitcher_id}` | Pitch trajectories with release point, spin vector, and movement profiles   |
| `GET /v1/mobile/at-bat/{game_id}/{ab_num}` | Full at-bat sequence: pitch-by-pitch with context, count, and result        |
| `GET /v1/mobile/quiz/situation`            | Random game situation for learning mode quizzes                             |
| `GET /v1/mobile/quiz/pitch-type`           | Pitch identification challenge with trajectory data                         |

These endpoints reduce round-trips and shape data for rendering. They are not authenticated via API keys — they use a client certificate or app-token scheme separate from the public API.

### Navigation

The app preserves the 5-tab bottom navigation from the mobile designs:

```text
⌂ Home  |  ◎ Players  |  ◈ Teams  |  ◷ Games  |  ≡ More
```

New features surface within existing tabs or under More:

- **Spray Chart** → Player detail (new tab alongside Batting, Pitching, etc.)
- **Pitch Tunnel** → Player detail (Pitching sub-view) or standalone under More
- **At-Bat Sequencer** → Game detail (tap any plate appearance)
- **Stat Card Generator** → Player detail (share action)
- **Learning Mode** → More tab

## Feature Specifications

### 1. Interactive Spray Chart with Park Overlay

A touch-interactive batted-ball map rendered on a `CustomPainter` canvas, overlaid on the specific ballpark's field geometry.

**Rendering approach**: `CustomPainter` draws the field outline (wall distances, foul lines, warning track) from park dimension data. Batted-ball events are plotted as circles sized by exit velocity and colored by result (single/double/triple/HR/out). The canvas sits inside an `InteractiveViewer` for pinch-to-zoom and pan.[^6]

**Data requirements**:

- Batted-ball x/y coordinates (available from Retrosheet hit-location codes, mappable to field coordinates)
- Park dimensions per `park_id` (wall distances at standard angles: LF line, LF, LCF, CF, RCF, RF, RF line)
- Launch angle and exit velocity where available (Statcast era, 2015+)

**Interactions**:

- Pinch to zoom into a region of the field
- Tap a hit dot to see detail (date, opponent, pitcher, exit velo, launch angle, result)
- Filter controls: vs LHP/RHP, by pitch type, by count, by season
- Toggle park overlay between home parks and compare overlays
- Haptic pulse on tap (`HapticFeedback.selectionClick`)

**Park geometry**: Wall outlines are Bézier curves fit to survey data. Each park stores ~12-16 control points defining the outfield wall. The `CustomPainter` scales these to canvas coordinates with the infield diamond as the anchor.

### 2. Pitch Tunnel Explorer

A 2.5D visualization of pitch trajectories from the batter's perspective, showing how different pitch types share a "tunnel" before diverging.

**Rendering approach**: Flame engine component with a fixed camera at the batter's eye position (~3.5 ft height, 1 ft behind home plate). Pitch trajectories are cubic Bézier curves computed from release point, spin-induced movement (Magnus force), and drag using Alan Nathan's trajectory model.[^7] Each pitch type gets a distinct color and trail width.

**Physics model**: The trajectory integrator uses 4th-order Runge-Kutta (RK4) with drag and Magnus force vectors derived from spin rate, spin axis, and gyro angle.[^7] Inputs per pitch:

- Release speed, release height, lateral release offset
- Spin rate (RPM), spin axis (clock position), gyro angle
- Vertical and horizontal break (pfx_x, pfx_z)

**Interactions**:

- Swipe to rotate the view (batter's eye → catcher → overhead → side)
- Tap a trajectory to isolate it and show pitch details
- Toggle "tunnel point" markers — the last point where two pitch types are indistinguishable
- Slider to scrub pitch progression in time (release → plate)
- Haptic bump at tunnel point and plate crossing

**Data requirements**: Pitch-level data with movement vectors. Available from `GET /v1/pitches` and `GET /v1/games/{id}/pitches` for Retrosheet pitch sequence data. Full Statcast movement data (pfx_x, pfx_z, spin rate) requires the internal endpoint.

### 3. At-Bat Sequencer

A pitch-by-pitch replay of any plate appearance, showing location, sequence strategy, and count progression.

**Rendering approach**: `CustomPainter` draws a strike zone (scaled to batter height) with pitch locations plotted sequentially. Each pitch animates in with a short trajectory arc. The count display updates with each pitch. Below the zone, a horizontal timeline shows the pitch sequence as labeled dots.

**Interactions**:

- Auto-play mode: pitches animate in at 1.5s intervals
- Manual mode: swipe left/right to step through pitches
- Tap any pitch dot to see type, speed, result, and count
- Long-press to compare this at-bat's sequence against the pitcher's typical patterns
- Haptic tick on each pitch arrival (`HapticFeedback.lightImpact`)

**Data source**: `GET /v1/mobile/at-bat/{game_id}/{ab_num}` returns the full sequence. Falls back to `GET /v1/games/{id}/pitches` filtered to the specific plate appearance.

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

| Module             | Format                 | Content                                                        |
| ------------------ | ---------------------- | -------------------------------------------------------------- |
| Rules & Scoring    | Animated diagrams      | Infield fly rule, balk, interference, tag-up, force play       |
| Pitch Types        | Identification trainer | Show trajectory + movement → user guesses pitch type           |
| Situation Quiz     | Game-state prompt      | "Runner on 2nd, 1 out, 2-1 count — what's the win expectancy?" |
| Stat Explainers    | Interactive calculator | What is WAR? Adjust inputs, see WAR change                     |
| Historical Moments | Narrated play-by-play  | Famous at-bats with the at-bat sequencer                       |

**Pitch identification trainer**: Uses the pitch tunnel renderer in a simplified mode. Shows a pitch trajectory, user selects from 4 options (e.g., four-seam, slider, changeup, curveball). Difficulty scales by reducing tunnel visibility and adding similar pitch types.

**Situation quiz**: Pulls from `GET /v1/mobile/quiz/situation` which returns a real historical game state. User guesses outcome or win probability. Compares against actual result and historical win expectancy from `GET /v1/win-expectancy`.

**Gamification**: Track correct answers, streaks, and category completion. Store locally (Hive). No server-side leaderboard in v1.

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

New routes are registered under `/v1/mobile/*` in the Go backend as public API surface and should be documented in Swagger with typed contracts.

### Spray Chart Endpoint

```text
GET /v1/mobile/spray-chart/{player_id}?season={year}&vs={L|R}&park_id={id}
```

Returns batted-ball events with field coordinates:

```json
{
  "player_id": "troutmi01",
  "park": { "id": "ANA01", "name": "Angel Stadium", "wall": [[...control_points]] },
  "events": [
    { "date": "2024-06-15", "x": 142.3, "y": 287.1, "result": "HR", "exit_velo": 108.2, "launch_angle": 28, "pitcher": "...", "pitch_type": "FF" }
  ]
}
```

Coordinates are in a standardized field system (origin at home plate, y-axis toward CF, units in feet). Park wall geometry is an array of `[angle_deg, distance_ft]` pairs at standard survey angles.

### Pitch Tunnel Endpoint

```text
GET /v1/mobile/pitch-tunnel/{pitcher_id}?season={year}&pitch_types={FF,SL,CH}
```

Returns aggregated pitch trajectory parameters grouped by pitch type:

```json
{
  "pitcher_id": "coMDee01",
  "pitch_types": [
    {
      "type": "FF",
      "label": "Four-seam",
      "avg_velo": 95.2,
      "avg_spin": 2340,
      "spin_axis": 210,
      "avg_pfx_x": -1.2,
      "avg_pfx_z": 14.8,
      "release": { "height": 5.8, "side": -2.1 },
      "usage_pct": 0.42,
      "count": 812
    }
  ]
}
```

### At-Bat Endpoint

```text
GET /v1/mobile/at-bat/{game_id}/{ab_num}
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
    { "seq": 1, "type": "FF", "speed": 95.1, "x": 0.3, "z": 2.8, "call": "B", "count": "1-0" },
    { "seq": 2, "type": "SL", "speed": 87.4, "x": -0.8, "z": 1.9, "call": "S", "count": "1-1" },
    {
      "seq": 3,
      "type": "FF",
      "speed": 96.0,
      "x": 0.1,
      "z": 3.1,
      "call": "X",
      "count": "1-2",
      "result": "HR",
      "exit_velo": 108.2
    }
  ]
}
```

### Quiz Endpoints

```text
GET /v1/mobile/quiz/situation?era={modern|all}
GET /v1/mobile/quiz/pitch-type?difficulty={1|2|3}
```

These pull from existing game data and pitch data, reshaping into quiz-friendly payloads with the correct answer embedded (client reveals after user input).

## Live & Current-Season Features

The MLB Stats API proxy (`/v1/mlb/*`) provides real-time access to the current season. The app surfaces this data through three primary features that complement the historical Retrosheet-backed views. We expand `/v1/mlb/*` directly (instead of introducing `/v1/mobile/*`) for UI-optimized live payloads.

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

**Offline**: Cache last-fetched scoreboard in Hive. Show stale data with "Last updated" timestamp when offline.

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

| Endpoint                                      | Source                                       | Purpose                                                                     |
| --------------------------------------------- | -------------------------------------------- | --------------------------------------------------------------------------- |
| `GET /v1/mobile/player-card/{id}`          | `/v1/players/*` (aggregated)                 | Aggregated player bio + key stats + recent game log for stat card rendering |
| `GET /v1/mobile/spray-chart/{player_id}`   | Retrosheet hit-location data + park geometry | Batted-ball coordinates with launch angle, exit velocity, and park geometry |
| `GET /v1/mobile/pitch-tunnel/{pitcher_id}` | Pitch-level data with movement vectors       | Pitch trajectories with release point, spin vector, and movement profiles   |
| `GET /v1/mobile/at-bat/{game_id}/{ab_num}` | `plays` + `pitches` tables                   | Full at-bat sequence: pitch-by-pitch with context, count, and result        |
| `GET /v1/mobile/quiz/situation`            | Historical game states                       | Random game situation for learning mode quizzes                             |
| `GET /v1/mobile/quiz/pitch-type`           | Pitch trajectory data                        | Pitch identification challenge with trajectory data                         |
| `GET /v1/mlb/schedule`                        | MLB schedule feed                            | Today's games with scores, status, linescore, and team details              |
| `GET /v1/mlb/standings`                       | MLB standings feed                           | Current standings data by league/division                                   |
| `GET /v1/meta/crosswalk/teams`                | `team_mlbam_map` + local team/franchise map  | MLB team ID → local `team_id` / `franchise_id` mapping for navigation       |
| `GET /v1/mlb/live/{gamePk}`                   | MLB game feed + win probability engine       | Real-time game state with play-by-play and win probability                  |
| `GET /v1/mlb/stats`                           | MLB stats feed                               | Current-season stat leaders and split queries                               |
| `GET /v1/mlb/people/{mlb_id}`                 | MLB people feed + local search lookups       | Current-season player stats with local player routing                       |

### ID Crosswalk

Live endpoints bridge MLB Stats API IDs (MLBAM `personId`, `teamId`) to local Retrosheet/Lahman IDs (`player_id`, `team_id`, `franchise_id`). Team/player mappings are exposed via `GET /v1/meta/crosswalk/*`, and `include=details` sidecars on `/v1/mlb/*` responses provide opt-in enrichment for labels and ID maps.

### Caching Strategy

| Endpoint group   | Cache TTL | Rationale                                        |
| ---------------- | --------- | ------------------------------------------------ |
| Scoreboard       | 30s       | Scores update frequently during games            |
| Standings        | 5min      | Changes only after games complete                |
| Live game feed   | 15s       | Near-real-time without overwhelming upstream     |
| Leaders          | 15min     | Stats update after games; no need for sub-minute |
| Player/team live | 5min      | Bio/roster data changes infrequently             |
| Spray/tunnel/ab  | 1hr       | Historical data; changes only on data loads      |
| Quiz             | No cache  | Should return varied results                     |

## Performance Targets

| Metric                              | Target                        |
| ----------------------------------- | ----------------------------- |
| Spray chart render (500 events)     | < 16ms per frame (60fps)      |
| Pitch tunnel scene (6 trajectories) | < 16ms per frame              |
| Theme switch (team color change)    | < 300ms transition            |
| Cold start to interactive           | < 2s on mid-range device      |
| Offline stat card generation        | Full capability (cached data) |
| Scoreboard refresh (15 games)       | < 500ms end-to-end            |
| Live game feed refresh              | < 300ms end-to-end            |
| Standings render (30 teams)         | < 16ms per frame              |
| APK size                            | < 25 MB                       |

## References

[^1]: Flutter. "Flutter - Build apps for any screen." [flutter.dev](https://flutter.dev). Cross-platform framework targeting Android, iOS, web, and desktop from a single Dart codebase.

[^2]: Flutter API. "CustomPainter class." [api.flutter.dev/flutter/rendering/CustomPainter-class.html](https://api.flutter.dev/flutter/rendering/CustomPainter-class.html). Low-level canvas drawing API for custom shapes, paths, and hit testing.

[^3]: Flame Engine. "A Flutter based game engine." [flame-engine.org](https://flame-engine.org). Lightweight 2D game engine with component system, collision detection, gesture handling, and sprite support. Actively maintained, 9k+ GitHub stars.

[^4]: Flutter API. "ColorScheme.fromSeed constructor." [api.flutter.dev/flutter/material/ColorScheme/ColorScheme.fromSeed.html](https://api.flutter.dev/flutter/material/ColorScheme/ColorScheme.fromSeed.html). Generates a full Material 3 color scheme from a single seed color using HCT perceptual color space.

[^5]: Flutter Gems. "haptic_feedback package." [pub.dev/packages/haptic_feedback](https://pub.dev/packages/haptic_feedback). Cross-platform haptic feedback with granular intensity control beyond the built-in `HapticFeedback` class.

[^6]: Flutter API. "InteractiveViewer class." [api.flutter.dev/flutter/widgets/InteractiveViewer-class.html](https://api.flutter.dev/flutter/widgets/InteractiveViewer-class.html). Built-in widget for pan and zoom with `Matrix4` transformation, boundary constraints, and scale limits.

[^7]: Nathan, A. M. "The Physics of Baseball." [baseball.physics.illinois.edu](http://baseball.physics.illinois.edu/). Trajectory model using drag + Magnus force with RK4 integration. Python port used by [baseball.skill-vis.com](https://baseball.skill-vis.com/).

[^8]: pub.dev. "pixel_art_generator package." [pub.dev/packages/pixel_art_generator](https://pub.dev/packages/pixel_art_generator). Template-based pixel art sprite generation in Flutter.

[^9]: Material Design. "Color system." [m3.material.io/styles/color](https://m3.material.io/styles/color). Material 3 tonal palette system ensures accessible contrast ratios across generated color roles.

[^10]: Team Color Codes. "MLB Team Color Codes." [teamcolorcodes.com/mlb-color-codes](https://teamcolorcodes.com/mlb-color-codes/). Official hex, RGB, and Pantone values for all 30 MLB teams. Cross-referenced with [teampalettes.com/mlb](https://teampalettes.com/mlb).

[^11]: pub.dev. "dynamic_color package." [pub.dev/packages/dynamic_color](https://pub.dev/packages/dynamic_color). Material.io team package (v1.8.1) for Android 12+ wallpaper-derived dynamic color schemes via `DynamicColorBuilder`.

[^12]: pub.dev. "fl_chart package." [pub.dev/packages/fl_chart](https://pub.dev/packages/fl_chart). Highly customizable Flutter chart library supporting line, bar, pie, scatter, and radar charts with touch interactions, animations, and theming. 7k+ GitHub stars.
