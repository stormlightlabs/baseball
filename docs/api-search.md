# Search & Lookup API Overview

Supplement for Section 8 of `docs/ROADMAP.md`, describing fuzzy lookup helpers.

## Summary

| Endpoint                 | Dataset | Highlights                                                                              |
| ------------------------ | ------- | --------------------------------------------------------------------------------------- |
| `GET /v1/search/players` | L+R     | Fuzzy match players by name, era, league, or handedness; returns lightweight cards.     |
| `GET /v1/search/teams`   | L+R     | Search by city, nickname, franchise, or year ranges; helpful for UI autocomplete.       |
| `GET /v1/search/parks`   | L+R     | Ballpark lookup that spans historical names and locations.                              |
| `GET /v1/search/games`   | R       | Natural-language search like "Yankees vs Red Sox 2003 ALCS Game 7" mapped to `GAME_ID`. |

## Endpoint Details

- Player/team/park search share a consistent payload with `id`, `label`, `sub_label`, and dataset tags.
- `GET /v1/search/games` tokenizes free-form text into probable home/away teams, season, and postseason context before resolving to Retrosheet IDs; queries also accept structured params.
