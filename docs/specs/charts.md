---
title: Chart and Data Visualization Standards
updated: 2026-04-21
---

## Context

The product now has a small but real chart surface across both clients:

| Platform | Surface | Current implementation | Current gaps |
| --- | --- | --- | --- |
| Web | Player batting | Chart.js line with stat selector | Filled area, smoothed curve, rate stats scaled x1000, no reference context |
| Web | Player pitching | Chart.js line for ERA | Filled area, smoothed curve, no comparison line |
| Web | Hall of Fame | Chart.js bar for vote % | No threshold marker, no direct labels |
| Web | Game win probability | Chart.js dual-line time series | Duplicated complementary series, heavy grid, no key-event annotation |
| Web | Team run differential | Table only | API already exposes cumulative + rolling windows but no chart yet |
| Mobile | Player batting/pitching | `fl_chart` line with metric chips | Filled area, curved line, x-axis uses season index instead of year |
| Mobile | Team overview | `fl_chart` monthly run-differential bar chart | Useful summary, but not the best season-shape chart |
| Mobile | Hall of Fame | List only | No chart summary |

### Implementation Reality

- Web charts use Chart.js 4 through a thin Svelte `<Chart>` wrapper.
- Mobile charts use `fl_chart`; there is no shared cross-platform chart grammar yet.
- Historical baseball data is not uniform across eras. Some visualizations should foreground this instead of hiding it.
- Statcast-era visualizations must respect metric availability windows. MLB’s public docs and Baseball Savant indicate that Statcast is league-wide from 2015, many batted-ball quality metrics are effectively 2016+, and newer bat-tracking metrics are 2023+ with partial 2023 coverage.[^statcast][^metrics-context]

## Primary Goal

Every chart should answer one baseball question clearly, with enough context to keep the reader from making the wrong comparison.

That means:

- choose the chart type based on the baseball question, not because the component already exists
- show historical context where baseball data needs it: era, league average, rule/environment changes, partial coverage
- keep the chart readable without hover
- treat web and mobile as different reading environments, not identical canvases

## Chart Selection Rules

| Reader question | Preferred chart | Notes |
| --- | --- | --- |
| How did this player/team change over time? | Line chart | Use for season-by-season or game-by-game trends |
| How did one game swing? | Win-probability line | Annotate decisive events; leverage table is supporting detail |
| How far above/below neutral is a team? | Zero-centered bar or cumulative line | Zero is the key reference, not the max |
| Did a value clear an important threshold? | Bar/dot plot with threshold rule | Hall of Fame vote %, playoff odds cutoffs, qualification lines |
| Where did pitches or batted balls occur? | Spatial plot / heatmap / spray chart | Preserve field/zone geometry; do not force into bars |
| How does this player compare to league or peers? | Percentile strip, dot plot, or small multiples | Prefer these over radar charts |
| What is the recent shape in a dense table? | Sparkline | Decorative summary only; full detail lives elsewhere |

### Avoid

- dual y-axes
- radar charts
- donut/pie charts for baseball stat comparison
- 3D or perspective effects
- mirrored home/away lines when one series is just `100 - other`

## Baseball-Specific Display Standards

### 1. Career and Season Trends

Use line charts for season-by-season player trends and daily/cumulative team trends.

Standards:

- x-axis should use actual year/date labels, not ordinal position unless the user truly only cares about sequence
- missing seasons should appear as gaps or explicit absence, not be silently connected
- rate stats must render in baseball notation: `.300`, `.850`, `3.45`; never `300` or `850`
- use straight or near-straight interpolation; smoothing is only acceptable when the underlying process is continuous and the distortion risk is low[^line-charts]
- comparison context should be added when it changes interpretation:
  - league average or league median
  - career high / low markers
  - era bands for multi-era careers
  - strike-shortened / pandemic-shortened season annotations where relevant

### 2. Win Probability and In-Game Narrative

MLB describes win expectancy as a “story stat” and WPA as the event-to-event change in those odds.[^we][^wpa] The chart should therefore privilege narrative clarity over symmetry.

Standards:

- prefer a single focal series from one perspective, usually home-team win probability on a 0-100 scale
- away-team probability is implicit; plotting both lines is usually redundant
- include a 50% reference line
- annotate the biggest swing(s), lead changes, and walk-off / final state
- tooltips should expose inning, half-inning, event description, and delta in win probability
- leverage tables belong below the chart, not inside it

### 3. Run Differential and Season Momentum

Run differential is most useful in two complementary views:

- **Cumulative run differential line** for season shape and inflection points
- **Rolling-window line or zero-centered bars** for short-term form

Standards:

- for per-game season views, cumulative line is the default primary chart
- monthly aggregation is acceptable as a compact summary card, but should not be the only view when per-game data exists
- zero should be visually explicit in any differential chart
- positive/negative color can be used, but the zero line does most of the communication

### 4. Threshold History

Hall of Fame vote history is not just a sequence of percentages; it is a threshold story.

Standards:

- preserve zero baseline for bars[^bars-zero]
- draw the 75% induction threshold as a labeled reference line
- directly label bars when the number of bars is small enough
- distinguish inducted vs non-inducted visually, but do not rely on color alone
- if a long HOF record becomes cramped, consider a dot plot instead of narrow bars

### 5. Spatial Baseball Charts

Baseball Savant’s public surfaces make clear that baseball analysis commonly uses pitch movement charts, field visualizers, percentile rankings, rolling windows, spray/location views, and attack/zone-oriented visuals.[^savant-home][^savant-search]

Use spatial charts when geometry is the story:

- spray charts for batted-ball direction and depth
- strike-zone heatmaps for pitch location or outcomes by zone
- pitch movement scatter plots for arsenal shape
- field-position maps for defensive opportunity or spray distribution

Standards:

- preserve baseball geometry; avoid distorting the field or zone into arbitrary rectangles when exact location matters
- density views need clear legends and minimum sample-size messaging
- handedness and pitcher/batter perspective must be explicit
- historical filters must respect the fact that these charts are mostly Statcast-era products

### 6. Percentile and Distribution Context

Percentile strips are useful for advanced player profiles because they communicate relative standing without asking the user to decode raw ranges first.

Standards:

- use a single perceptually ordered sequential palette, not rainbow color scales[^color-coding]
- label the percentile directly at the marker
- include the metric name and sample/season scope nearby
- use percentile strips for quick scanning; use dot plots or distributions when the exact peer spread matters

### 7. Small Multiples and Sparklines

Tufte’s small multiples principle is a good fit for baseball comparisons across players, teams, eras, and pitch types.[^tufte-envisioning] Use them when overlap would make a single chart unreadable. Modern chart guidance reaches the same conclusion for overlapping line series.[^small-multiples]

Standards:

- keep shared axes where cross-panel comparison matters
- use panel titles or direct labels instead of a detached legend
- on desktop, small multiples are preferred once a comparison chart would exceed roughly 4-5 overlapping series
- on mobile, stack one chart per row
- sparklines should be tiny summaries only: no axes, no legend, no tooltip, no interaction

## Visual Language

### Tufte-Style Defaults

- prioritize data-ink over decoration[^tufte-visual-display]
- no gradient backgrounds, shadows, or ornamental fills
- no area fills under ordinary comparison lines unless the chart is explicitly about accumulation/share
- at most one subtle grid direction; usually faint horizontal rules only
- direct labels beat legends when there are few series[^line-labels]
- annotations should be sparse, specific, and low-opacity[^annotations]

### Color and Emphasis

- use CSS/design tokens, not hardcoded hex values
- default chart palette:
  - primary series
  - muted comparison/context series
  - alert/threshold color only when semantics justify it
- ensure colorblind-safe distinctions and add a second channel when categories matter: dash, symbol, label, or pattern[^wong][^colorblind-access]
- do not use hue to encode ordered quantities when position or length already does the job[^color-coding]

### Axes and Labels

- bars must start at zero[^bars-zero]
- line charts do not have to start at zero, but the chosen range must be honest and legible[^line-charts]
- axis labels should be minimized when the title, subtitle, tick formatting, and direct labels already explain the measure[^axis-labels]
- numeric formatting must match baseball idiom:
  - AVG / OBP / SLG / OPS as leading-dot decimals
  - ERA / WHIP / K-BB% with sport-appropriate precision
  - win probability and vote share as percentages

## Historical Context and Coverage Rules

Do not let the chart imply comparability the data cannot support.

Required handling:

- show era context when a time series spans materially different run environments or rules
- surface known partial-coverage eras in UI copy or annotations, especially Federal League and Negro Leagues coverage already called out elsewhere in the app
- gate Statcast visualizations by metric availability window
- avoid side-by-side comparisons that quietly mix full modern tracking data with sparse historical event data

## Interaction Rules

- tooltips are supplemental; the chart’s core message must survive without them
- hover-only interactions need a touch equivalent on mobile
- no entrance animation on page load
- short state-change animation only; disable motion when `prefers-reduced-motion` is active[^reduced-motion]
- zoom/pan is opt-in for dense game-level views, not a default crutch for poor chart design

## Accessibility Rules

- every chart needs a text summary or accessible name stating the trend / takeaway
- focusable data points are required only where interaction changes understanding; otherwise prefer a summary plus table fallback
- contrast for labels and rules must meet WCAG expectations[^wcag]
- never encode critical meaning with color alone

## Platform Guidance

### Web

- can support richer annotation, direct labels, threshold rules, and synchronized table/chart context
- should use the existing table surfaces as detail views beneath summary charts
- should centralize defaults so every chart does not hand-roll colors, grid, tooltip styles, and motion behavior

### Mobile

- should favor one strong story per card; cramped multi-series charts degrade quickly on touch screens
- x-axis labels must use real seasons/dates where possible, even if abbreviated
- supporting numbers should sit outside the plot area because tooltip precision is harder to deliver well on touch
- monthly summary bars are acceptable as compact cards, but season-shape charts should prefer cumulative or rolling lines when the data exists

## Implications for Current Implementations

- Web player charts should stop scaling rate stats by 1000 and should stop filling the area under the line.
- Web win probability should shift from complementary dual lines to a single focal curve plus event annotations.
- Web team run differential should add a chart, because the endpoint already returns the data shape needed for one.
- Mobile player metric charts should switch the x-axis from season index (`1, 2, 3`) to actual year labels.
- Mobile run differential should keep the monthly bar summary only as a secondary view once cumulative/rolling views are available.
- Mobile Hall of Fame can remain list-first, but should gain a compact threshold summary if HOF vote % data exists.

## References

[^tufte-visual-display]: Edward R. Tufte, *The Visual Display of Quantitative Information* (2nd ed.).
[^tufte-envisioning]: Edward R. Tufte, *Envisioning Information*.
[^statcast]: [Statcast | Glossary | MLB](https://www.mlb.com/glossary/statcast)
[^metrics-context]: [Statcast Metrics Context | Baseball Savant](https://baseballsavant.mlb.com/statcast-metrics-context)
[^we]: [Win Expectancy (WE) | Glossary | MLB](https://www.mlb.com/glossary/advanced-stats/win-expectancy)
[^wpa]: [Win Probability Added (WPA) | Glossary | MLB](https://www.mlb.com/glossary/advanced-stats/win-probability-added)
[^savant-home]: [Baseball Savant Home / Leaderboards / Visuals](https://baseballsavant.mlb.com/)
[^savant-search]: [Statcast Search | Baseball Savant](https://baseballsavant.mlb.com/statcast_search)
[^line-charts]: [What to Consider When Creating Line Charts | Datawrapper Academy](https://academy.datawrapper.de/article/129-what-to-consider-when-creating-line-charts)
[^small-multiples]: [What to Consider When Creating Small Multiple Line Charts | Datawrapper Blog](https://www.datawrapper.de/blog/what-to-consider-when-creating-small-multiple-line-charts)
[^line-labels]: [Customizing Your Line Chart | Datawrapper Academy](https://www.datawrapper.de/academy/customizing-your-line-chart)
[^annotations]: [Customizing Your Line Chart | Datawrapper Academy](https://www.datawrapper.de/academy/customizing-your-line-chart)
[^axis-labels]: [Why Many Datawrapper Charts Don't Include Axis Labels | Datawrapper Academy](https://academy.datawrapper.de/article/239-why-datawrapper-does-not-include-axis-labels-for-many-charts)
[^bars-zero]: [Why Our Column and Bar Charts Start at Zero | Datawrapper Academy](https://www.datawrapper.de/academy/why-our-column-and-bar-charts-start-at-zero)
[^wong]: [Points of View: Color Blindness | Nature Methods](https://www.nature.com/articles/nmeth.1618)
[^color-coding]: [Color Coding | Nature Methods](https://www.nature.com/articles/nmeth0810-573)
[^colorblind-access]: [How We Make Sure Our Charts, Maps and Tables Are Accessible | Datawrapper Academy](https://www.datawrapper.de/academy/how-we-make-sure-our-charts-maps-and-tables-are-accessible)
[^reduced-motion]: [SCR40: Using the CSS prefers-reduced-motion query in JavaScript to prevent motion | W3C WAI](https://www.w3.org/WAI/WCAG22/Techniques/client-side-script/SCR40)
[^wcag]: [WCAG 2.2](https://www.w3.org/TR/WCAG22/)
