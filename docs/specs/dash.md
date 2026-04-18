---
title: Data Dashboard
updated: 2026-04-18
---

## Stack

- **Framework**: SvelteKit 2 in SPA mode (`adapter-static`, fallback `index.html`), deployed to CDN
- **Styling**: Tailwind CSS v4
- **Charts**: Chart.js 4 via a thin Svelte wrapper
- **API docs**: OpenAPI (`internal/docs/swagger.yaml`) for API Explorer schema + params
- **Routing**:
    - Dashboard SPA at `/`
    - API namespace is `/v1/*` (current public API surface)

## API Compatibility Baseline

The dashboard spec must track the API that is currently registered in `internal/api/*`.

### Core API families to expose

- **Meta/health**: `/v1/health`, `/v1/ready`, `/v1/meta`, `/v1/meta/datasets`, `/v1/meta/readiness`
- **Search**: `/v1/search/players`, `/v1/search/teams`, `/v1/search/parks`, `/v1/search/games`
- **Players**: `/v1/players/*` including `seasons`, `stats/*`, `game-logs/*`, `awards`, `hall-of-fame`, `teams`, `salaries`, `relatives`
- **Teams/franchises**: `/v1/teams/*`, `/v1/franchises/*`, `/v1/seasons/{year}/teams/*`
- **Games/events**: `/v1/games`, `/v1/games/{id}`, `/v1/games/{id}/boxscore`, `/v1/games/{id}/summary`, `/v1/games/{id}/events`, `/v1/games/{id}/plays`, `/v1/games/{id}/pitches`
- **Stats/leaders**: `/v1/stats/*`, `/v1/seasons/{year}/leaders/*`, `/v1/leaders/*/career`, `/v1/stats/teams/*`
- **Advanced/computed**: `/v1/players/{player_id}/stats/*/advanced`, `/v1/players/{player_id}/stats/war`, leverage + park-factor endpoints
- **Derived**: streaks, splits, run-differential, game win-probability, `/v1/win-expectancy`, `/v1/win-expectancy/eras`
- **League-specific historical slices**: `/v1/federalleague/*`, `/v1/negroleagues/*`
- **Awards/postseason/all-star/ejections/achievements/salaries/managers/umpires/coaches**
- **MLB proxy**: `/v1/mlb/*`

## Era Model (Must Be First-Class in UI)

The dashboard should expose two complementary era systems.

### A) Retrosheet load eras (static, from `internal/seed/eras.go`)

| Era                | Short code  | Years     |
| ------------------ | ----------- | --------- |
| Federal League Era | `fed`       | 1914-1915 |
| Negro Leagues Era  | `nlg`       | 1935-1949 |
| Baby Boomer Era    | `boomer`    | 1950-1962 |
| Pitcher Era        | `pitcher`   | 1963-1968 |
| Turf Time          | `turf`      | 1969-1993 |
| Steroid Era        | `steroid`   | 1994-2004 |
| Moneyball Era      | `moneyball` | 2005-2012 |
| Statcast Era       | `statcast`  | 2013-2019 |
| Modern Era         | `modern`    | 2020-2025 |

### B) Win expectancy eras (dynamic, from API)

- Source endpoint: `GET /v1/win-expectancy/eras`
- This returns era ranges based on sampled historical tables, so labels/ranges are data-driven.

### Era UX rules

- Every page that has a season or range control should show the active era badge.
- Compare mode must show both selected eras side-by-side.
- Cross-era comparisons should display source caveats (especially Federal League / Negro Leagues and pre-1918 play-event density).
- Data Source page should include an era matrix and coverage caveats, not just raw year bars.

## Search and Discovery Home

Should immediately answer: _what can this API do across eras?_

### Features

- universal search dispatcher that routes to:
    - `/v1/search/players`
    - `/v1/search/teams`
    - `/v1/search/games`
    - `/v1/search/parks`
- featured queries panel that includes at least one query from:
    - standard stats
    - derived/computed
    - league-specific historical routes
- API health + dataset coverage block sourced from meta endpoints
- era quick-jump chips (fed, nlg, boomer, pitcher, turf, steroid, moneyball, statcast, modern)

## Player Explorer

### Endpoints

- `/v1/search/players`
- `/v1/players/{id}`
- `/v1/players/{id}/seasons`
- `/v1/players/{id}/stats/batting`
- `/v1/players/{id}/stats/pitching`
- `/v1/players/{id}/game-logs/*`
- `/v1/players/{id}/awards`
- `/v1/players/{id}/hall-of-fame`
- `/v1/players/{id}/teams`
- `/v1/players/{id}/salaries`
- `/v1/players/{id}/relatives`
- Optional advanced tabs:
    - `/v1/players/{player_id}/stats/batting/advanced`
    - `/v1/players/{player_id}/stats/pitching/advanced`
    - `/v1/players/{player_id}/stats/war`
    - `/v1/players/{player_id}/splits`
    - `/v1/players/{player_id}/streaks`

### Era-specific behavior

- Season charts show shaded era bands.
- Player timeline cards show "career spans X eras" with linked era chips.

## Team and Franchise Explorer

### Endpoints

- `/v1/franchises`
- `/v1/franchises/{id}`
- `/v1/search/teams`
- `/v1/teams/{id}`
- `/v1/teams/{id}/daily-stats`
- `/v1/teams/{team_id}/run-differential`
- `/v1/seasons/{year}/teams`
- `/v1/seasons/{year}/teams/{team_id}/roster`
- `/v1/seasons/{year}/teams/{team_id}/batting`
- `/v1/seasons/{year}/teams/{team_id}/pitching`
- `/v1/seasons/{year}/teams/{team_id}/fielding`
- `/v1/seasons/{year}/teams/{team_id}/games`
- `/v1/seasons/{year}/teams/{team_id}/schedule`
- `/v1/seasons/{year}/teams/{team_id}/daily-logs`

### Era-specific behavior

- Franchise view uses era segmentation for name/identity continuity.
- Team season cards show context chips (e.g., `pitcher`, `steroid`, `statcast`).

## Game Finder and Game Detail

### Endpoints

- `/v1/games`
- `/v1/games/{id}`
- `/v1/games/{id}/summary`
- `/v1/games/{id}/boxscore`
- `/v1/games/{id}/events`
- `/v1/games/{id}/events/{event_seq}`
- `/v1/games/{id}/plays`
- `/v1/games/{id}/pitches`
- `/v1/games/{game_id}/win-probability`
- `/v1/games/{game_id}/win-probability/summary`
- `/v1/games/{game_id}/plate-appearances/leverage`
- `/v1/seasons/{year}/schedule`
- `/v1/seasons/{year}/dates/{date}/games`
- `/v1/seasons/{year}/teams/{team_id}/games`
- `/v1/seasons/{year}/parks/{park_id}/games`

### Era-specific behavior

- Filters include era chips in addition to year/date range.
- Detail page flags whether event richness is dense/partial for the selected era.

## Stat Leaderboard Center

### Endpoints

- `/v1/stats/batting`
- `/v1/stats/pitching`
- `/v1/stats/fielding`
- `/v1/stats/teams/batting`
- `/v1/stats/teams/pitching`
- `/v1/stats/teams/fielding`
- `/v1/seasons/{year}/leaders/batting`
- `/v1/seasons/{year}/leaders/pitching`
- `/v1/leaders/batting/career`
- `/v1/leaders/pitching/career`
- Advanced leaderboards:
    - `/v1/seasons/{season}/leaders/batting/advanced`
    - `/v1/seasons/{season}/leaders/pitching/advanced`
    - `/v1/seasons/{season}/leaders/war`

### Era-specific behavior

- Trend view should support both continuous year view and era-bucket view.

## Season Hub

### Endpoints

- `/v1/seasons`
- `/v1/seasons/{year}/teams`
- `/v1/seasons/{year}/leaders/batting`
- `/v1/seasons/{year}/leaders/pitching`
- `/v1/seasons/{year}/schedule`
- `/v1/seasons/{year}/dates/{date}/games`
- `/v1/seasons/{year}/awards`
- `/v1/seasons/{year}/postseason/series`
- `/v1/seasons/{year}/postseason/games`
- `/v1/seasons/{season}/park-factors`

## Compare Mode

### Comparisons

- player vs player
- team-season vs team-season
- season vs season
- franchise-era comparison
- win expectancy state comparison across era ranges

### Endpoints

- all relevant player/team/season endpoints above
- `/v1/win-expectancy`
- `/v1/win-expectancy/eras`

## API Explorer

The explorer is the interactive API reference and must reflect the current OpenAPI document.

### Required groups

- meta/health
- search
- players
- teams/franchises
- games/plays/pitches
- stats/leaders
- computed/derived
- achievements/awards/postseason/allstar/ejections/salaries
- managers/umpires/coaches
- league-specific (`federalleague`, `negroleagues`)
- mlb proxy

## Data Sources and Coverage Page

### Must include

- source cards: Lahman, Retrosheet, MLB StatsAPI
- historical coverage + known caveats
- explicit era matrix (the 9 seed eras + dynamic WE eras)
- crosswalk strategy for Lahman IDs, Retrosheet IDs, MLB IDs
- league-specific caveats for Federal League and Negro Leagues endpoints

## Account and Key Management

Authenticated area behind `/account`.

### API hooks

- `/v1/auth/me`
- `/v1/auth/keys`
- `/v1/auth/keys/{id}` (delete/revoke)
- OAuth entrypoints exist for GitHub/Codeberg callback flow

## Page Architecture

| Route       | Page         | Layout     | Auth | Notes                                  |
| ----------- | ------------ | ---------- | ---- | -------------------------------------- |
| `/`         | Home         | single-col | no   | search + meta + era jump               |
| `/players`  | Players      | three-col  | no   | player profile + advanced/derived tabs |
| `/teams`    | Teams        | three-col  | no   | franchise + team-season + era context  |
| `/games`    | Games        | three-col  | no   | finder + game detail + event richness  |
| `/seasons`  | Seasons      | three-col  | no   | season hub + awards/postseason         |
| `/leaders`  | Leaders      | three-col  | no   | quick leaders + query lab + advanced   |
| `/compare`  | Compare      | three-col  | no   | side-by-side + era normalization       |
| `/explorer` | API Explorer | three-col  | no   | OpenAPI-driven endpoint explorer       |
| `/data`     | Data Sources | single-col | no   | provenance + era matrix + caveats      |
| `/account`  | Account      | single-col | yes  | API keys + usage                       |
