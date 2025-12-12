# Derived & Advanced Endpoints Overview

Companion to Section 9 of `docs/ROADMAP.md`. These advanced endpoints layer analytics on top of Lahman + Retrosheet feeds and rely on precomputed streak/split aggregates.

## Summary

| Endpoint                                   | Dataset | Highlights                                                                                            |
| ------------------------------------------ | ------- | ----------------------------------------------------------------------------------------------------- |
| `GET /v1/players/{player_id}/streaks`      | L+R     | Emits hitting/on-base/pitching streak segments with linked game context and running totals.           |
| `GET /v1/players/{player_id}/splits`       | L+R     | Returns derived splits (home/away, handedness, month, lineup slot) with filters and league baselines. |
| `GET /v1/teams/{team_id}/run-differential` | L+R     | Provides season cumulative, 7-game, and 30-game scoring margins plus per-game deltas.                 |
| `GET /v1/games/{game_id}/win-probability`  | R       | Streams win-probability and leverage data per play for charting and recap narratives.                 |

## Endpoint Notes

### `GET /v1/players/{player_id}/streaks`

- Combines Lahman season aggregates with Retrosheet logs to derive streak windows (start/end dates, stat tracked, length).
- Each streak object links to `/v1/games/{game_id}` rows involved and includes per-game contribution so UI can display spark lines.

### `GET /v1/players/{player_id}/splits`

- Single endpoint powering split tabs; accepts `split` param (`home_away`, `handedness`, `lineup_slot`, `month`, etc.) and standard paging/sorting controls.
- Response includes per-split totals, rate stats, and optional league average block for quick comparisons.

### `GET /v1/teams/{team_id}/run-differential`

- Aggregates runs scored/allowed per game and emits cumulative plus rolling-window differentials (configurable via `window` query param).
- Ties back to `/v1/seasons/{year}/teams/{team_id}` for roster context.

### `GET /v1/games/{game_id}/win-probability`

- Consumes parsed plays and Win Expectancy tables to output `(event_seq, inning, outs, win_prob_home, win_prob_away, leverage_index)` tuples.
- Downstream consumers can draw charts or annotate pivotal plays; payload also includes references to `/v1/games/{game_id}/plays` for deep dives.
