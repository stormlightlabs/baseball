# Pitch-Level API Overview

Pitch routes expose individual Retrosheet pitch sequences so advanced tooling can reason about counts, sequencing, and leverage in a single request.
See [pitches](/docs/pitches) for deeper details on Retrosheet encoding and parsing rules referenced below.

## Summary

| Endpoint                                       | Dataset | Highlights                                                                                               |
| ---------------------------------------------- | ------- | -------------------------------------------------------------------------------------------------------- |
| `/v1/pitches`                                  | R       | Global pitch search with batter/pitcher/team/date filters plus count- or pitch-type constrained queries. |
| `/v1/games/{game_id}/pitches`                  | R       | Chronological slice of every pitch in a single game with pagination for long contests.                   |
| `/v1/games/{game_id}/plays/{play_num}/pitches` | R       | Returns all pitches for one plate appearance, useful when drilling into a single event from `/v1/plays`. |

## Endpoint Details

### `GET /v1/pitches`

- Accepts standard pagination plus Retrosheet-centric filters (`batter`, `pitcher`, `bat_team`, `pit_team`, `date`, `date_from`, `date_to`).
- Supports count/pitch flags like `ball_count`, `strike_count`, `pitch_type`, `is_in_play`, `is_strike`, and `is_ball`.
- Responds with `PaginatedResponse` where `data` is an array of pitch objects matching the schema documented in `docs/pitches.md`.

### `GET /v1/games/{game_id}/pitches`

- Convenience wrapper that pins `game_id` while retaining pagination controls (`page`, `per_page`, default 200).
- Response ordering mirrors play-by-play order so clients can render pitch charts alongside `/v1/games/{game_id}/plays`.

### `GET /v1/games/{game_id}/plays/{play_num}/pitches`

- Returns a compact `{ "data": [...] }` payload of every pitch parsed for a single plate appearance.
- No pagination because plate appearances typically contain \<10 pitches; callers must supply both `game_id` and `play_num`.

## Query Parameters (`GET /v1/pitches`)

- `batter`: Retrosheet batter ID
- `pitcher`: Retrosheet pitcher ID
- `bat_team`: batting team ID
- `pit_team`: pitching team ID
- `date`: game date (`YYYYMMDD`)
- `date_from` / `date_to`: date range (`YYYYMMDD`)
- `inning`: inning number
- `pitch_type`: pitch/event code (`B`, `C`, `F`, `S`, `X`, etc.)
- `ball_count`: ball count filter (`0-3`)
- `strike_count`: strike count filter (`0-2`)
- `is_in_play`: only balls in play
- `is_strike`: only strikes
- `is_ball`: only balls

## Examples

```bash
# Get all pitches from a specific pitcher
curl "/v1/pitches?pitcher=darvy001&per_page=50"

# Find all full-count pitches
curl "/v1/pitches?ball_count=3&strike_count=2"

# Filter to balls in play
curl "/v1/pitches?pitch_type=X"

# Filter by pitcher/batter matchup
curl "/v1/pitches?pitcher=darvy001&batter=ohtas001"

# Get all pitches from a game
curl "/v1/games/SDN202403200/pitches"

# Get pitches from a single plate appearance
curl "/v1/games/SDN202403200/plays/1/pitches"
```
