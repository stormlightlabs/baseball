# Derived & Advanced Endpoints Overview

These endpoints layer derived analytics on top of Lahman + Retrosheet data.

## Summary

| Endpoint                                   | Dataset | Highlights                                                                                       |
| ------------------------------------------ | ------- | ------------------------------------------------------------------------------------------------ |
| `GET /v1/players/{player_id}/streaks`      | L+R     | Finds streak windows for hitting or scoreless innings with start/end dates and game context.     |
| `GET /v1/players/{player_id}/splits`       | L+R     | Returns batting splits by `home_away`, `pitcher_handed`, or `month`.                             |
| `GET /v1/teams/{team_id}/run-differential` | L+R     | Provides season run differential with cumulative and rolling windows.                            |
| `GET /v1/games/{game_id}/win-probability`  | R       | Streams win-probability and leverage data per play for charting and recap narratives.            |
| `GET /v1/win-expectancy`                   | R       | Returns historical win probability for an arbitrary game state (inning, base state, score diff). |
| `GET /v1/win-expectancy/eras`              | R       | Lists available historical eras so callers can scope probability lookups to specific seasons.    |

## Endpoint Notes

### `GET /v1/players/{player_id}/streaks`

- Required query params:
  - `kind`: `hitting` or `scoreless_innings`
  - `season`: season year
- Optional query params:
  - `min_length`: minimum streak length (default `5`)
- Response includes streak metadata (start/end dates, game IDs, length).

Examples:

```bash
# Hitting streaks of 10+ games in 2024
curl "/v1/players/reynb001/streaks?kind=hitting&season=2024&min_length=10"

# Scoreless innings streaks of 15+ innings
curl "/v1/players/flord002/streaks?kind=scoreless_innings&season=2024&min_length=15"
```

### `GET /v1/players/{player_id}/splits`

- Required query params:
  - `dimension`: `home_away`, `pitcher_handed`, or `month`
  - `season`: season year
- Response includes grouped split rows with counting stats and slash line fields.

Examples:

```bash
# Home/away split
curl "/v1/players/judga001/splits?dimension=home_away&season=2024"

# Splits vs left/right handed pitchers
curl "/v1/players/sotoj001/splits?dimension=pitcher_handed&season=2024"

# Monthly split
curl "/v1/players/ohtas001/splits?dimension=month&season=2024"
```

### `GET /v1/teams/{team_id}/run-differential`

- Required query params:
  - `season`: season year
- Optional query params:
  - `windows`: comma-separated rolling windows (default `10,20,30`)
- Returns season totals, per-game cumulative values, and rolling window aggregates.

Examples:

```bash
# Default windows (10,20,30)
curl "/v1/teams/NYA/run-differential?season=2024"

# Custom windows
curl "/v1/teams/LAN/run-differential?season=2024&windows=5,10,15"
```

### `GET /v1/games/{game_id}/win-probability`

- Returns per-event win probability progression for a game.
- Payload includes game state context and home/away win probability values.

Example:

```bash
curl "/v1/games/BAL202404010/win-probability"
```

### `GET /v1/win-expectancy`

- Required query params:
  - `inning` (`1-9`)
  - `is_bottom` (`true`/`false`)
  - `outs` (`0-2`)
  - `runners` (for example `___`, `1__`, `12_`, `123`)
  - `score_diff` (home perspective, typically `-11` to `+11`)
- Optional query params:
  - `start_year`
  - `end_year`

Examples:

```bash
# Bottom 9th, 2 outs, bases empty, tied
curl "/v1/win-expectancy?inning=9&is_bottom=true&outs=2&runners=___&score_diff=0"

# Bottom 9th, bases loaded, no outs, tied
curl "/v1/win-expectancy?inning=9&is_bottom=true&outs=0&runners=123&score_diff=0"
```

### `GET /v1/win-expectancy/eras`

- Lists available historical windows loaded into the win expectancy table.
- Useful for UI filters and pre-validating `start_year`/`end_year`.

Example:

```bash
curl "/v1/win-expectancy/eras"
```

## Implementation Notes

- Streaks are derived from player game logs and segmented into contiguous windows.
- Run differential uses rolling window aggregation over ordered season game points.
- Win probability queries use historical game state data and can fall back when a state is sparse.
