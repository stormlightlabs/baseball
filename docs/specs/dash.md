---
title: Data Dashboard
updated: 2026-04-17
---

## Stack

- **Framework**: SvelteKit 2 in SPA mode (`adapter-static`, fallback `index.html`), deployed to CDN
- **Styling**: Tailwind CSS v4 — dark-only theme derived from the wireframe design tokens
- **Charts**: Chart.js 4 via a thin Svelte canvas wrapper
- **API docs**: OpenAPI spec (`swagger.yaml`) parsed client-side for the API Explorer
- **Routing**: dashboard lives at the root (`/`); `/api/` is reserved for the backend API

## Search and discovery home

Should immediately answer: _what can this API do?_

### Features

- universal search for **players, teams, franchises, games**
- quick links for:
  - player lookup
  - team-season lookup
  - game finder
  - season leaders
  - today-in-history or date explorer
- a “featured queries” panel with prebuilt examples
- API health / version / dataset coverage block

### Why

Your sources are historical and broad. Lahman covers major-league batting, pitching, fielding, standings, team stats, managerial records, postseason data, and more back to 1871, while Retrosheet provides detailed game/event data and annual rosters across a very large historical range.
The home page should immediately show that this is not just a toy wrapper over one endpoint.

## Player explorer

### Endpoints

- `/players`
- `/players/{id}`
- `/players/{id}/seasons`
- `/players/{id}/teams`
- `/players/{id}/awards`
- `/players/{id}/hall-of-fame`
- `/players/{id}/salaries`
- `/players/{id}/game-logs`
- `/players/{id}/appearances`
- `/players/{id}/plays`
- `/players/{id}/plate-appearances`

### Features

- bio card
- handedness / debut / final game / birthplace
- batting career timeline
- pitching career timeline
- awards timeline
- Hall of Fame voting history
- team-by-team career path
- rate-stat vs counting-stat toggle
- era-context note when comparing across decades

### Critical UI elements

- season table with sortable columns
- line chart of season totals or rates
- team badges by season
- split batting and pitching tabs
- direct deep link to the API endpoint and raw JSON

### Why

This page proves your API is not just “lookup”; it is **narrative and historical**.
The fact that your player season response includes both batting and pitching arrays is especially strong because it supports two-way players and mixed-role careers naturally.

## Team and franchise explorer

Your spec supports:

- `/franchises`
- `/franchises/{id}`
- `/teams`
- `/teams/{id}`
- `/seasons/{year}/teams`
- `/seasons/{year}/teams/{team_id}/games`

### Franchise page

- franchise identity and active years
- franchise timeline
- team-name / era continuity
- all seasons summary
- cumulative wins/losses
- best and worst seasons
- postseason presence count if available later

### Team-season page

- wins/losses/ties
- runs scored / runs allowed
- run differential
- attendance
- division / league
- season schedule summary
- team game log
- top batting and pitching contributors for that season

### Strong additions

- franchise history view as a timeline
- compare two franchises or two team-seasons
- scatterplot: runs scored vs runs allowed for a season
- attendance vs wins

### Why

Lahman is particularly strong for team-season and historical franchise work.
A franchise explorer gives your API its own identity rather than looking like a partial clone of MLB.com.

## Game finder and game detail

### Endpoints

- `/games`
- `/games/{id}`
- `/games/{id}/boxscore`
- `/games/{id}/plays`
- `/seasons/{year}/schedule`
- `/seasons/{year}/dates/{date}/games`
- `/seasons/{year}/teams/{team_id}/games`

### Game finder should include

- filters for season, date range, home team, away team, park
- postseason toggle
- innings filter
- attendance / duration filters
- one-click “doubleheaders on this date” or “extra-inning games”

### Game detail should include

- scoreline
- date / day / venue
- attendance
- duration
- umpire crew
- season and series context
- boxscore with lineups (from `/games/{id}/boxscore`)
- play-by-play feed (from `/games/{id}/plays`)
- links to both participating teams and season pages

### What would make it excellent

- schedule heatmap by month
- team schedule calendar
- game duration trends over time
- park lookup and park-centric game browsing

### Why

Retrosheet’s event/game files are built around rich historical game descriptions and team-home-game files, which makes game-centric exploration a natural strength for your API.

## Stat leaderboard center

### Endpoints

- `/stats/batting`
- `/stats/pitching`
- `/seasons/{year}/leaders/batting`
- `/seasons/{year}/leaders/pitching`

### Build two modes

### A. Predefined leaders

For quick exploration:

- HR leaders
- AVG leaders
- RBI leaders
- SB leaders
- SO leaders
- ERA leaders
- saves leaders
- innings leaders

### B. Query lab

For power users:

- season or season range
- player filter
- team filter
- league filter
- thresholds like `min_ab`, `min_ip`, `min_gs`
- sort by stat
- sort order
- pagination controls
- generated request preview

### Best visualizations

- ranked tables
- percentile bands
- trend line across seasons
- league split toggle
- compare top N across time

### Why

This is where your API most clearly demonstrates “programmable research.”
The presence of season range filters plus thresholds and sort controls is exactly what should be turned into a visible query builder.

## Season hub

You already have all the primitives for a strong season-centric view:

- season schedule
- season teams
- games by date
- batting leaders
- pitching leaders

### Season hub should include

- season overview
- team table
- leaderboards
- schedule calendar
- league filters
- date explorer
- notable queries for that year

### This page should answer

- Who led the league?
- Which teams played?
- What did the season calendar look like?
- What games happened on a given date?
- How can I query this season through the API?

## Compare mode

This is the feature that would make the dashboard memorable.

### Include

- player vs player
- team-season vs team-season
- season vs season
- league split comparison
- franchise-era comparison

### Example comparisons

- Ruth vs Aaron through age-30
- 1998 Yankees vs 2001 Mariners
- AL vs NL HR leaders in a season
- franchise best seasons by run differential

### Why

The underlying API is historical and normalized enough to support comparison as a first-class experience. This makes the dashboard a resource, not just an endpoint browser.

## Query builder and API mirror panel

This is essential.

Every page should have a panel that shows:

- endpoint path
- current query params
- generated URL
- example `curl`
- sample JSON response
- copy buttons

### Why this matters

The best pattern is: **visual controls on the left, results in the center, request/response on the right**

This turns the dashboard into both:

- a usable baseball research tool
- an interactive API tutorial

## Data provenance and coverage page

Because your API combines Lahman, Retrosheet, and MLB StatsAPI, you need a page that explains:

- which data comes from which source
- historical coverage
- update cadence
- where live/current-season enrichment comes from
- known gaps / caveats
- ID mapping strategy between datasets

This is important because Lahman is a structured historical database with broad statistical coverage, while Retrosheet is especially valuable for game/event history, and MLB StatsAPI exposes much richer contemporary operational data such as schedules, teams, people, leaders, standings, transactions, and metadata categories.

### Features

- source badges on records
- “last updated” timestamps
- coverage by season
- notes on missing or partial historical data
- explanation of IDs like Lahman player IDs vs Retrosheet IDs vs MLB person/team/game IDs

## Account and key management

The only authenticated area of the dashboard. All other pages are public.

### Features

- sign-in / sign-up (email + password; OAuth provider TBD)
- API key list: name, masked prefix, created date, last-used timestamp, active/revoked status
- key creation: name, optional expiry, optional scope — full key shown once on creation with copy button
- key revocation
- usage dashboard: request count by day/week/month, current rate-limit tier, quota remaining
- account settings: email, password change, account deletion

### Why

Self-service key management is table stakes for any public API. Putting it behind auth keeps keys secure while keeping the rest of the dashboard open for discovery. The usage dashboard gives developers confidence about their quota before they hit limits in production.

## Performance and reliability demo panel

- response time
- cache status
- pagination totals
- current filters
- record count
- request history

## Core UX

- global search
- advanced filters
- shareable URLs for every state
- sortable tables
- pagination controls
- raw JSON viewer
- copy as curl
- deep linking to entities

## Resource value

- saved queries
- featured research queries
- compare mode
- season hubs
- provenance notes
- glossary for baseball stats and abbreviations

## Developer value

- endpoint mirror panel
- schema snippets
- parameter docs inline
- sample requests and responses
- error examples
- rate-limit and cache visibility if applicable

## Expansion

### High value additions (not yet implemented in the API)

- standings endpoint
- team batting/pitching season stats endpoint
- park / venue endpoint
- splits endpoint
- postseason series endpoint
- franchise season history endpoint
- search autocomplete endpoint
- bulk compare endpoint

### Modern/current-season additions from MLB StatsAPI side

MLB StatsAPI exposes broader current-era concepts such as standings, team stats, rosters, transactions, venues, and richer schedule/team/person data than what is currently visible in your spec.

The dashboard would benefit a lot from eventually exposing:

- standings
- rosters
- venues
- transactions
- team leaders
- team stats
- live or near-live game context

## Page Architecture

| Route       | Page         | Layout     | Auth | Notes                                        |
| ----------- | ------------ | ---------- | ---- | -------------------------------------------- |
| `/`         | Home         | single-col | no   | search hero + dashboard grid                 |
| `/players`  | Players      | three-col  | no   | sidebar search, center stats, right API      |
| `/teams`    | Teams        | three-col  | no   | franchise + team-season combined in one page |
| `/games`    | Games        | three-col  | no   | finder filters + game detail                 |
| `/seasons`  | Seasons      | three-col  | no   | season picker, teams, leaders, schedule      |
| `/leaders`  | Leaders      | three-col  | no   | quick mode + query lab                       |
| `/compare`  | Compare      | three-col  | no   | side-by-side entity comparison               |
| `/explorer` | API Explorer | three-col  | no   | endpoint list, param builder, live response  |
| `/data`     | Data Sources | single-col | no   | provenance, coverage, ID mappings            |
| `/account`  | Account      | single-col | yes  | sign-in, API key management, usage, settings |
