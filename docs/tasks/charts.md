# Chart and Data Visualization Tasks

- Ref spec: `docs/specs/charts.md`
- Web stack: Chart.js 4 in Svelte 5
- Mobile stack: `fl_chart` in Flutter
- Goal: bring existing chart surfaces up to the spec before adding novel chart types

## Phase 0: Inventory, Semantics, and Shared Rules

- [ ] Update internal chart inventory references so docs and planning match the actual app surface:
  - Web: batting, pitching, Hall of Fame, win probability.
  - Web gap: no run-differential chart yet.
  - Mobile: player metric line chart, team monthly run-differential bar chart.
- [ ] Document metric availability windows in code-facing notes where needed:
  - Statcast-era visualizations are 2015+.
  - Several quality-of-contact metrics are 2016+.
  - Bat-tracking metrics are 2023+ with partial 2023 coverage.
- [ ] Define shared chart-copy rules for both clients:
  - Rate stats display in baseball notation (`.300`, `.812`, `3.45`).
  - Time-series axes use real years/dates, not ordinal sequence, unless sequence is the story.
  - Missing periods are handled intentionally, not silently interpolated.

Acceptance:

- [ ] Spec, tasks, and implementation inventory no longer contradict each other.
- [ ] Historical/statcast coverage constraints are explicit in planning notes.

## Phase 1: Web Chart Foundation

- [ ] Create shared Chart.js defaults module for web:
  - muted horizontal grid only
  - no chart border
  - legend off by default
  - direct-label/annotation-friendly plugin defaults
  - reduced-motion-aware animation defaults
- [ ] Register annotation/datalabel plugins centrally in the Svelte chart wrapper.
- [ ] Replace hardcoded chart hex colors with theme/CSS token reads.
- [ ] Add a small formatting helper module for baseball chart values:
  - leading-dot rate stats
  - percentages
  - signed point changes for win probability

Acceptance:

- [ ] Web charts share one visual baseline instead of per-chart config drift.
- [ ] No web chart needs to hand-roll basic grid/color/motion defaults.

## Phase 2: Mobile Chart Foundation

- [ ] Create shared `fl_chart` helpers for mobile:
  - grid style
  - axis text style
  - zero-line / threshold rule styling
  - reduced-motion behavior where supported
- [ ] Add common baseball number formatters for chart axes and summary labels.
- [ ] Ensure mobile chart widgets accept real x-axis labels instead of deriving labels from list index.

Acceptance:

- [ ] Player and team mobile charts no longer encode seasons as `1, 2, 3...`.
- [ ] Mobile chart styling is consistent across screens.

## Phase 3: Fix Existing Player Trend Charts

### Web batting + pitching

- [ ] Remove area fills.
- [ ] Reduce curve smoothing to near-linear.
- [ ] Display batting rate stats as raw decimals, not x1000 integers.
- [ ] Use only faint horizontal rules; remove vertical grid.
- [ ] Add comparison/reference context where available:
  - league average line for selected stat
  - career high marker
  - era band overlays for multi-era careers
- [ ] Add accessible chart summary text below or on the chart container.

### Mobile batting + pitching

- [ ] Remove below-line fill for standard trend views.
- [ ] Reduce curvature or use straight segments unless smoothing is justified.
- [ ] Change bottom-axis labels from season index to actual year.
- [ ] Keep detailed values outside the plot area; do not rely on precise touch hitboxes alone.

Acceptance:

- [ ] Player trend charts in both apps read as baseball season timelines, not generic app charts.
- [ ] Rate-stat formatting is correct everywhere.

## Phase 4: Hall of Fame Threshold View

### Web

- [ ] Add 75% threshold annotation.
- [ ] Add direct value labels for small ballot histories.
- [ ] Highlight inducted season without relying on color alone.
- [ ] Re-evaluate whether long ballot histories should switch from bars to dots.

### Mobile

- [ ] Add a compact Hall of Fame summary chart when vote percentages exist.
- [ ] Keep the existing list as the detailed fallback.
- [ ] Prefer a simple threshold visualization over a dense multi-interaction chart.

Acceptance:

- [ ] Hall of Fame charts communicate threshold crossing first, raw vote totals second.

## Phase 5: Win Probability Cleanup (Web First)

- [ ] Refactor web win-probability chart from dual complementary series to one focal home-team curve.
- [ ] Add 50% reference rule.
- [ ] Annotate biggest swing(s), walk-off/final event, and optionally lead changes.
- [ ] Keep leverage rows/table beneath the chart as supporting context.
- [ ] Add an explicit note about game-state density / sparse historical context when applicable.

Acceptance:

- [ ] The chart tells the game story at a glance without requiring tooltip hunting.
- [ ] Home vs away is still obvious even though only one probability line is plotted.

## Phase 6: Run Differential Views

### Web

- [ ] Add a primary run-differential chart to `teams/[id]/run-diff` using the existing endpoint.
- [ ] Make cumulative run differential the default first view.
- [ ] Add a rolling-window toggle if the API payload already supports it cleanly.
- [ ] Keep the sortable game table as the audit/detail view.

### Mobile

- [ ] Keep monthly run-differential bars as a compact overview card.
- [ ] Add a season-shape chart when the mobile data model exposes cumulative/rolling values.
- [ ] Ensure zero is visually explicit and symmetric in any differential chart.

Acceptance:

- [ ] Web no longer exposes run differential only as a table.
- [ ] Mobile monthly bars become summary, not the only run-diff story.

## Phase 7: Baseball-Specific Future Chart Families

Only start these after existing surfaces are coherent.

- [ ] Sparklines in tables/lists for recent trend context.
- [ ] Percentile strips for advanced player profiles.
- [ ] Spray charts for batted-ball direction and depth.
- [ ] Strike-zone heatmaps for pitch location / outcome views.
- [ ] Pitch-movement scatter plots for arsenal shape.
- [ ] Small multiples for player/team comparison screens.

For each new family, require before implementation:

- [ ] clear baseball question
- [ ] confirmed data availability window
- [ ] web/mobile interaction plan
- [ ] accessibility fallback

## Phase 8: Accessibility and QA

- [ ] Add chart summaries / accessible names for all production charts.
- [ ] Verify colorblind-safe distinctions for all multi-series states.
- [ ] Verify reduced-motion behavior on web and sensible non-distracting transitions on mobile.
- [ ] Test chart readability at mobile widths before accepting direct labels or annotations.
- [ ] Add screenshot/regression coverage for key chart states where practical.
- [ ] Add formatter tests for baseball decimal/percentage display.

Acceptance:

- [ ] Charts remain understandable without hover precision.
- [ ] Accessibility and formatting regressions are covered by tests or explicit QA steps.

## Data / API Follow-Ups

- [ ] Expose league-average series for player trend metrics if not already available.
- [ ] Confirm run-differential payload shape in mobile domain models for cumulative and rolling windows.
- [ ] Define sample-size rules for spatial charts and percentile strips.
- [ ] Define era-band source of truth so UI overlays are not coupled to ad hoc labels.
- [ ] Decide whether historical win-probability views need era-scoped lookup controls in the UI.
