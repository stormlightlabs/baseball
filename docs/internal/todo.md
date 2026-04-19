---
title: Parking Lot
---

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

This might be applicable to the API as well: we should normalize era labels such that
they're not abbreviated. This could be under `/meta`

Games are paginated but the UI doesn't reflect this and asserts that there are only N games (where N is the page size). We should add pagination controls to the UI and update the API to return pagination metadata.

We should remove the `jq` from cURL examples and maybe mention in `routes/docs`. These should
also be single line.

## Backend

Post-ETL we need to clean-up year specific Play-by-Play & Game log data to keep space lean.

Slog?

API version should be `ALPHA` (keep the `v1` namespace) while we iron out kinks in the system.
