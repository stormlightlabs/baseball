---
title: Chart and Data Visualization Standards
updated: 2026-04-18
---

## Context

The dashboard uses Chart.js 4 via a thin Svelte wrapper (`<Chart>`) for all visualization. Current charts are functional but generic: uniform grid styling, area fills, single-dataset views, no direct labeling, and no use of Tufte principles or baseball-specific viz conventions.

### Current Inventory

| Location        | Chart             | Type          | Notes                                       |
| --------------- | ----------------- | ------------- | ------------------------------------------- |
| Player batting  | Career stat trend | Line (filled) | 7 selectable stats, rate stats scaled x1000 |
| Player pitching | Career ERA trend  | Line (filled) | Single metric                               |
| Player HOF      | Voting % history  | Bar           | 0-100% scale                                |
| Home page       | Dataset coverage  | Stacked bar   | By decade, per source                       |

### Problems

- **High chartjunk**: opaque area fills, visible grid on both axes, legend boxes instead of direct labels.
- **Low data density**: one dataset per chart, no contextual reference lines (league average, era boundaries).
- **No baseball conventions**: missing sparklines, percentile strips, spray charts, rolling averages.
- **Accessibility gaps**: hardcoded hex colors, no colorblind-safe palette, no `prefers-reduced-motion` on chart animations.
- **No interactivity beyond stat selector**: no hover annotations, no cross-chart coordination, no zoom/pan for dense timelines.

## Principles

Derived from Tufte's quantitative display theory[^1][^2][^3], adapted for a dark-theme baseball dashboard on canvas (Chart.js).

### 1. Maximize data-ink ratio

Every pixel should encode data or essential structure.[^1] Remove or mute anything that doesn't: decorative fills, heavy gridlines, legend boxes, border outlines on bars.

### 2. Eliminate chartjunk

- No area fills under line charts. Use line + points only.[^1]
- No visible grid on both axes simultaneously. At most, faint horizontal rules for y-reference.
- No gradient backgrounds, drop shadows, or rounded bar corners.

### 3. Direct labeling over legends

Label datasets on the line/bar itself (annotation plugin or custom drawing).[^2] Reserve legend boxes only for charts with 4+ overlapping datasets where inline labels would collide.

### 4. Context through reference data

Every stat chart should show at least one reference layer:

- **League average** as a dashed gray line.
- **Era boundaries** as faint vertical bands with subtle labels.
- **Career/season highs** as annotated points.

### 5. Small multiples over overloaded charts

When comparing across players, seasons, or stat categories: repeat the same chart structure in a grid with shared axes.[^2] Never use dual y-axes.[^5] Never overload a single chart with unrelated scales.

### 6. Sparklines for inline context

Embed tiny, word-sized line charts in stat tables and roster lists.[^3] A 60px-wide sparkline next to a player's season AVG row communicates trend without requiring navigation.

### 7. Restrained color

- **Primary palette**: 2-3 hues max per chart. Use the existing blue (`#3b82f6`) for primary data, green (`#10b981`) for secondary/comparison, gray (`#6b7280`) for reference/context.
- **Colorblind-safe**: all categorical distinctions must work under deuteranopia. Use shape/pattern as secondary channel.[^4]
- **Emphasis**: highlight via saturation and weight, not new hues. Gray everything except the focal data.

### 8. Graphical integrity

- Bar charts must include zero baseline. No exceptions.[^1]
- Truncated y-axes on line charts must show an axis-break indicator.
- Aspect ratios should bank trend slopes to ~45 degrees for optimal perception.[^6]
- Rate stats (AVG, OBP, SLG, OPS) display as decimals (`.300`), not scaled integers (`300`).

### 9. Purposeful animation

- Transitions between states (stat change, time range change): 200-300ms ease-out.
- No entrance animations on page load.
- Respect `prefers-reduced-motion`: skip all transitions.

### 10. Tooltips are supplemental

Tooltips provide detail (exact value, game date, opponent) on hover. The chart's main message must be readable without any interaction.

## Chart Type Standards

### Line Charts (Career Trends)

```text
Config changes from current baseline:
- fill: false (remove area fill)
- pointRadius: 2 (smaller, less ink)
- tension: 0.1 (near-linear, honest shape)
- borderWidth: 2
- grid.y: { drawBorder: false, color: '#1a1e27' } (barely visible)
- grid.x: { display: false } (remove vertical grid entirely)
- Add league-average reference line as second dataset (dashed, gray, no points)
- Add era boundary annotations via chartjs-plugin-annotation
```

### Bar Charts (HOF Voting, Coverage)

```text
- Zero baseline enforced (already correct for HOF)
- Remove grid lines on x-axis
- Direct value labels above bars (chartjs-plugin-datalabels)
- Muted bar color with highlight on the decisive ballot (75%+ threshold line)
```

### Sparklines

```text
- Canvas-only, no axes, no labels, no tooltips
- Fixed height: 20-24px, width: fills container
- Single-pixel line, no points
- Color: inherit from text color (muted by default, primary on hover)
- Use for: season stat tables, roster lists, leaderboard rows
```

### Percentile Strips (for Advanced Stats)

```text
- Horizontal bar from 0-100
- Color gradient: single hue, sequential (blue)
- Player's percentile marked with a dot + direct label
- Used for: Baseball Savant-style stat profiles[^7] (exit velo, hard hit%, K%, BB%)
```

### Era Band Overlays

```text
- Vertical semi-transparent rectangles behind data area
- Alternating subtle opacity (0.03 / 0.06) to distinguish eras
- Tiny era label at top of band, rotated or abbreviated
- Applied to any time-series chart spanning multiple eras
```

## Dependencies

- `chartjs-plugin-annotation` — era bands, reference lines, threshold markers.
- `chartjs-plugin-datalabels` — direct value labels on bars and key points.
- No new charting library. Stay on Chart.js 4.

## Color Tokens

Chart colors should reference CSS custom properties, not hardcoded hex values:

| Token               | Current Hex | Usage                                 |
| ------------------- | ----------- | ------------------------------------- |
| `--chart-primary`   | `#3b82f6`   | Main data series                      |
| `--chart-secondary` | `#10b981`   | Comparison / secondary series         |
| `--chart-muted`     | `#6b7280`   | Reference lines, axis text            |
| `--chart-grid`      | `#1a1e27`   | Subtle grid lines (y-axis only)       |
| `--chart-surface`   | `#252934`   | Era band fills                        |
| `--chart-danger`    | `#ef4444`   | Below-threshold / negative indicators |
| `--chart-warn`      | `#f59e0b`   | Caution indicators                    |

## Accessibility

- All charts must have a visually hidden `<p>` or `aria-label` summarizing the data trend in text.
- Color alone must never be the sole differentiator — pair with dash patterns, point shapes, or labels.
- Tooltip content must be keyboard-accessible (not hover-only on critical charts).
- Respect `prefers-reduced-motion` — disable all Chart.js animations when active.
- Minimum contrast ratio: 4.5:1 for all text labels against chart background.[^8]

## References

[^1]: Tufte, E. R. (2001). *The Visual Display of Quantitative Information* (2nd ed.). Graphics Press. Defines data-ink ratio, chartjunk, graphical integrity, and the lie factor. The foundational argument for removing non-data ink.
[^2]: Tufte, E. R. (1990). *Envisioning Information*. Graphics Press. Covers small multiples, direct labeling, layering/separation, and micro/macro readings — techniques for encoding dense multivariate data without clutter.
[^3]: Tufte, E. R. (2006). *Beautiful Evidence*. Graphics Press. Introduces sparklines ("intense, simple, word-sized graphics") and the principle of data/content adjacency.
[^4]: Wong, B. (2011). "Points of view: Color blindness." *Nature Methods*, 8(6), 441. Establishes the 8-color colorblind-safe palette widely adopted in scientific visualization.
[^5]: Few, S. (2006). "Dual-Scaled Axes in Graphs: Are They Ever the Best Solution?" *Visual Business Intelligence Newsletter*. Argues dual y-axes create false implied correlations; recommends small multiples with shared x-axis instead.
[^6]: Cleveland, W. S. (1993). *Visualizing Data*. Hobart Press. Formalizes banking to 45 degrees for optimal slope judgment in trend lines; introduces Cleveland dot plots as bar chart alternatives.
[^8]: W3C. (2023). *Web Content Accessibility Guidelines (WCAG) 2.2*, Success Criterion 1.4.3. Minimum contrast ratio of 4.5:1 for normal text; 3:1 for large text.
