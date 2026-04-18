# Chart and Data Visualization Tasks

- Ref spec: `docs/specs/charts.md`
- Stack: Chart.js 4, chartjs-plugin-annotation, chartjs-plugin-datalabels, Svelte 5
- Scope: improve existing charts, add new chart types, establish shared conventions

## Phase 0: Foundation

- [ ] Add `chartjs-plugin-annotation` and `chartjs-plugin-datalabels` to `web/package.json`.
- [ ] Register both plugins in `Chart.svelte` alongside Chart.js registerables.
- [ ] Define CSS custom properties for chart color tokens (`--chart-primary`, `--chart-secondary`, `--chart-muted`, `--chart-grid`, `--chart-surface`, `--chart-danger`, `--chart-warn`) in `layout.css`.
- [ ] Create shared chart defaults module (`$lib/charts/defaults.ts`):
    - Base scale config (muted y-grid only, no x-grid, no border).
    - Base plugin config (legend hidden by default, datalabels off by default).
    - Reduced-motion detection helper that disables Chart.js animations.
    - Tooltip style defaults (dark background, small font, no caret).
- [ ] Refactor `CHART_SCALES` in `$lib/players/constants.ts` to use the shared defaults.

Acceptance:

- [ ] Plugins register without error; shared defaults importable from `$lib/charts/defaults`.
- [ ] `prefers-reduced-motion` disables chart animations globally.

## Phase 1: Fix Existing Charts

### Career Line Charts (Batting + Pitching)

- [ ] Remove area fill (`fill: false`).
- [ ] Reduce tension to `0.1` and point radius to `2`.
- [ ] Remove x-axis grid lines entirely.
- [ ] Reduce y-axis grid to barely visible (`#1a1e27`, no border).
- [ ] Display rate stats as decimals (`.300`) instead of scaled integers (`300`).
    - Update y-axis tick callback for rate stats.
    - Store raw decimal values in dataset, not multiplied.
- [ ] Add league-average reference line as a second dataset:
    - Dashed gray line, no points, no legend entry.
    - Source: hardcoded era-average lookup or future API endpoint.
- [ ] Hide legend box (single-dataset charts don't need it).

### HOF Voting Bar Chart

- [ ] Remove x-axis grid lines.
- [ ] Add 75% threshold line via annotation plugin (dashed, labeled "75%").
- [ ] Add direct value labels above bars via datalabels plugin.
- [ ] Highlight the decisive ballot bar (the one where inducted) in a distinct color.

### Dataset Coverage Chart (Home Page)

- [ ] Remove tooltip and legend disabling hacks; use shared defaults.
- [ ] Add direct era labels above each decade group.
- [ ] Mute grid to match new shared scale config.

Acceptance:

- [ ] All existing charts render with reduced ink, no area fills, minimal grid.
- [ ] Rate stats display as readable decimals.
- [ ] HOF chart shows threshold line and direct labels.

## Phase 2: Era Band Overlays

- [ ] Create era band annotation generator (`$lib/charts/era-bands.ts`):
    - Input: year range of the chart's x-axis.
    - Output: array of `chartjs-plugin-annotation` box annotations.
    - Alternating opacity (`0.03` / `0.06`) using `--chart-surface`.
    - Abbreviated era label at top of each band.
- [ ] Apply era bands to career batting and pitching line charts.
- [ ] Apply era bands to any future time-series chart via shared helper.

Acceptance:

- [ ] Era bands visible as subtle background stripes on career charts.
- [ ] Bands auto-clip to the chart's actual year range.

## Phase 3: Sparklines

- [ ] Create `Sparkline.svelte` component:
    - Minimal Chart.js line config: no axes, no grid, no tooltips, no legend.
    - Fixed height (20-24px), fluid width.
    - Single-pixel line, zero point radius.
    - Color inherits from text (muted default, primary on row hover).
    - Props: `data: number[]`, optional `color`, optional `height`.
- [ ] Add sparklines to player batting stats table (trailing trend for selected stat).
- [ ] Add sparklines to player pitching stats table (ERA trend).
- [ ] Add sparklines to salary table (salary trend).

Acceptance:

- [ ] Sparklines render inline in stat tables without layout shift.
- [ ] Sparklines are purely decorative — no tooltips, no interaction.

## Phase 4: Percentile Strips

- [ ] Create `PercentileStrip.svelte` component:
    - Horizontal bar, 0-100 scale.
    - Sequential single-hue fill (blue gradient via canvas).
    - Player dot + direct label at their percentile.
    - Props: `value: number`, `label: string`, optional `thresholds`.
- [ ] Use on advanced batting tab for key metrics (if available from API): hard hit%, K%, BB%, barrel%.
- [ ] Use on WAR tab for positional percentile context.

Acceptance:

- [ ] Strips render at correct percentiles with readable labels.
- [ ] Color scale is perceptually uniform (sequential blue, not rainbow).

## Phase 5: Accessibility

- [ ] Add `aria-label` or visually hidden description to every `<Chart>` instance summarizing the trend in text.
- [ ] Ensure all multi-series charts use dash patterns or point shapes alongside color.
- [ ] Verify all chart text labels meet 4.5:1 contrast against `--color-crust`.
- [ ] Verify `prefers-reduced-motion` fully disables transitions (test with system setting).
- [ ] Replace hardcoded hex values in chart configs with CSS custom property reads (via `getComputedStyle`).

Acceptance:

- [ ] Screen reader announces chart summary on focus.
- [ ] Charts are distinguishable without color perception.
- [ ] No hardcoded color hex values remain in chart config code.

## Phase 6: Future Chart Types (Blocked on API/Data)

These require endpoints or data not yet available. Track here for planning.

- [ ] **Win probability line chart** — needs `GET /v1/games/{id}/win-probability`.
- [ ] **Pitch movement scatter plot** — needs pitch-level data with pfx_x/pfx_z.
- [ ] **Spray chart** — needs batted-ball coordinate data.
- [ ] **Rolling average overlay** — needs game-log granularity; compute client-side from `game-logs/*`.
- [ ] **Small multiples for comparison** — needs `/compare` page buildout; reuse career chart in grid layout.
- [ ] **Leaderboard sparklines** — needs `/leaders` page buildout.
- [ ] **Run differential trend** — needs `GET /v1/teams/{id}/daily-stats` integration on teams page.
