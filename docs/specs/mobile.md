---
title: Flutter Mobile App
updated: 2026-04-18
---

## Context

The web dashboard (SvelteKit SPA) covers the full API surface but is constrained by the browser sandbox: no GPU-accelerated custom rendering, no haptic feedback, no platform-native transitions, and limited offline capability. A Flutter app targeting Android and iOS can exploit native GPU canvas, platform gesture systems, and Material You theming to deliver experiences the web cannot — interactive spray charts with park overlays, 3D pitch tunnel exploration, haptic-rich at-bat sequencing, and per-team dynamic color schemes.

The app is not a port of the web dashboard. It is a native companion that focuses on touch-first visualization and baseball education, backed by the same Go API with an additional `api/internal/` namespace for client-specific endpoints.

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

The app consumes the existing `/api/v1/` endpoints (80+ routes) and introduces a new `/api/internal/` namespace for mobile-optimized endpoints that aggregate multiple queries or return pre-shaped payloads:

| Internal Endpoint                             | Purpose                                                                     |
| --------------------------------------------- | --------------------------------------------------------------------------- |
| `GET /api/internal/player-card/{id}`          | Aggregated player bio + key stats + recent game log for stat card rendering |
| `GET /api/internal/spray-chart/{player_id}`   | Batted-ball coordinates with launch angle, exit velocity, and park geometry |
| `GET /api/internal/pitch-tunnel/{pitcher_id}` | Pitch trajectories with release point, spin vector, and movement profiles   |
| `GET /api/internal/at-bat/{game_id}/{ab_num}` | Full at-bat sequence: pitch-by-pitch with context, count, and result        |
| `GET /api/internal/quiz/situation`            | Random game situation for learning mode quizzes                             |
| `GET /api/internal/quiz/pitch-type`           | Pitch identification challenge with trajectory data                         |

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

**Data source**: `GET /api/internal/at-bat/{game_id}/{ab_num}` returns the full sequence. Falls back to `GET /v1/games/{id}/pitches` filtered to the specific plate appearance.

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

**Situation quiz**: Pulls from `GET /api/internal/quiz/situation` which returns a real historical game state. User guesses outcome or win probability. Compares against actual result and historical win expectancy from `GET /v1/win-expectancy`.

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

New routes registered under `api/internal/` in the Go backend. These are not part of the public API contract and do not require Swagger documentation.

### Spray Chart Endpoint

```text
GET /api/internal/spray-chart/{player_id}?season={year}&vs={L|R}&park_id={id}
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
GET /api/internal/pitch-tunnel/{pitcher_id}?season={year}&pitch_types={FF,SL,CH}
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
GET /api/internal/at-bat/{game_id}/{ab_num}
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
GET /api/internal/quiz/situation?era={modern|all}
GET /api/internal/quiz/pitch-type?difficulty={1|2|3}
```

These pull from existing game data and pitch data, reshaping into quiz-friendly payloads with the correct answer embedded (client reveals after user input).

## Performance Targets

| Metric                              | Target                        |
| ----------------------------------- | ----------------------------- |
| Spray chart render (500 events)     | < 16ms per frame (60fps)      |
| Pitch tunnel scene (6 trajectories) | < 16ms per frame              |
| Theme switch (team color change)    | < 300ms transition            |
| Cold start to interactive           | < 2s on mid-range device      |
| Offline stat card generation        | Full capability (cached data) |
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
