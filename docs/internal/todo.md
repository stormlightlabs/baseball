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

## Web

This might be applicable to the API as well: we should normalize era labels such that
they're not abbreviated. This could be under `/meta`

The go project's static content (html templates, css, js) should be moved to the `web`
project.

For `api.bigfly.tech` to work, we have to namespace the API routes behind `/v1` not `/api/v1`
