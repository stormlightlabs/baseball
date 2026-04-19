# Play-by-Play Events API Overview

Captures fine-grained Retrosheet feeds used for situational analytics.

## Summary

| Endpoint                                        | Dataset | Highlights                                                                          |
| ----------------------------------------------- | ------- | ----------------------------------------------------------------------------------- |
| `GET /v1/plays`                                 | R       | Global play search with batter/pitcher/team filters and optional leverage metrics.  |
| `GET /v1/games/{game_id}/plays`                 | R       | Chronological feed for a single game; canonical play-by-play timeline.              |
| `GET /v1/games/{game_id}/events`                | R       | Alias layering raw Retrosheet event fields (pitch sequence, modifiers, base state). |
| `GET /v1/games/{game_id}/events/{event_seq}`    | R       | Lookup for a single event sequence within a game, useful for shareable links.       |
| `GET /v1/players/{player_id}/plate-appearances` | R       | Player-focused PA listing with opponent pitcher, count, and leverage data.          |
| `GET /v1/pitches`                               | R       | Derives per-pitch signals (pitch type heuristics, velocity buckets, outcomes).      |

## Endpoint Details

- `/plays` and `/plate-appearances` expose cursor pagination, `min_leverage_index`, and `count_state` filters
- Event endpoints keep raw Retrosheet columns (e.g., `EVENT_TX`, `BASERUNNER_ADVANCE`) while also surfacing parsed interpretations for easier downstream parsing.
- The `/pitches` endpoint projects each play into pitch-level rows, adding inferred pitch tags and release/velocity details
