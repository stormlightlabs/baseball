---
title: Parking Lot
---

Should the color hex/hue in `team-branding.ts` be in sync with `colors.dart` between
codebases?

## Mobile

Default Player should be `ohtansh01` (Shohei Ohtani).

Default Team should be `LAN` (Los Angeles Dodgers).

We should add a settings screen:
    - disable animations
    - change default player
    - change default team
    - remove color hue (team color) or set default color

Remove api endpoints from Quick Access cards

Mobile needs a way to handle CORS from the backend.

## Web

System health should actually display Database & API health.

This might be applicable to the API as well: we should normalize era labels such that
they're not abbreviated. This could be under `/meta`

Games are paginated but the UI doesn't reflect this and asserts that there are only N games (where N is the page size). We should add pagination controls to the UI and update the API to return pagination metadata.

The doc file badge (ex. `docs/introduction.md`) should link to the Github source file
(ex. `https://github.com/stormlightlabs/baseball/blob/main/web/src/routes/docs/introduction.md`)

Featured Queries should link to dashboard pages, not the swagger docs/explorer-redirect

### Live/Home Cards

These have endpoints as a subtitle. We should turn these into a special component with a footer
that displays the endpoint with Copy URL & Copy cURL buttons.

### Players

Season Log Team column should link to team page, with a tooltip showing the full team name.

Table should be sortable

## Backend

Post-ETL we need to clean-up year specific Play-by-Play & Game log data to keep space lean.

API version should be `ALPHA` (keep the `v1` namespace) while we iron out kinks in the system.

Do we ingest the Chadwick Register more exhaustively to enrich persons data?

Remove `core.MLBTeamCrosswalk* structs` if no longer needed anywhere.

Repository-level unit tests for crosswalk query ambiguity behavior beyond API integration tests.

Current Season games & stats should be their on Repositories, not added to GameRepository & StatsRepository.

## Open Questions

Should standings should include richer historical GB/WCGB fields?

Should GET /v1/games/{id}/boxscore eventually support current-season game_pk IDs?
