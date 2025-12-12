# Pitch Sequencing

The Baseball API parses Retrosheet pitch sequence data to provide individual pitch-level access through the `/v1/pitches` endpoint. This document explains how pitch sequences are encoded, parsed, and served.

## Pitch Sequence Format

Retrosheet encodes each plate appearance's pitches as a single string where each character represents a pitch or event.
This compact format must be parsed to extract individual pitches with proper count tracking.

### Primary Pitch Types

| Code | Type                   | Description                     | Updates Count         |
| ---- | ---------------------- | ------------------------------- | --------------------- |
| B    | Ball                   | Ball outside the strike zone    | +1 ball               |
| C    | Called Strike          | Strike called by umpire         | +1 strike             |
| F    | Foul Ball              | Ball hit foul                   | +1 strike (max 2)     |
| S    | Swinging Strike        | Batter swings and misses        | +1 strike             |
| X    | Ball in Play           | Pitch results in ball in play   | Final pitch           |
| L    | Foul Bunt              | Foul ball on bunt attempt       | +1 strike (max 2)     |
| M    | Missed Bunt            | Swinging strike on bunt attempt | +1 strike             |
| O    | Foul Bunt              | Foul ball on bunt (alternative) | +1 strike (max 2)     |
| P    | Pitchout               | Intentional pitch outside zone  | +1 ball               |
| T    | Foul Tip               | Foul tip caught by catcher      | +1 strike             |
| V    | Called Strike (Appeal) | Called strike on appeal         | +1 strike             |
| H    | Hit by Pitch           | Batter hit by pitch             | Plate appearance ends |
| I    | Intentional Ball       | Intentional ball                | +1 ball               |
| N    | No Pitch               | Balk, interference, etc.        | No count change       |
| A    | Automatic Ball         | Rare automatic ball             | +1 ball               |

### Modifiers (Non-Pitch Characters)

| Code | Meaning                   | Purpose                            |
| ---- | ------------------------- | ---------------------------------- |
| .    | Play not involving batter | Pickoff attempt, stolen base, etc. |
| >    | Runner going              | Runner attempting steal            |
| \*   | Pitch blocked by catcher  | Following pitch was blocked        |
| +    | Following runner event    | Additional runner advance          |
| 1    | Pickoff to first          | Pickoff throw to first base        |
| 2    | Pickoff to second         | Pickoff throw to second base       |
| 3    | Pickoff to third          | Pickoff throw to third base        |

## Parsing Algorithm

### Count Tracking

The parser maintains a running count of balls and strikes for each pitch in the sequence:

1. Start with 0-0 count
2. For each pitch character:
    - Record current ball and strike count
    - Classify pitch type (ball, strike, in play)
    - Update count based on pitch type
    - Foul balls do not add strikes after 2 strikes
3. Skip modifier characters (not pitches)
4. Continue until sequence ends

### Example Parsing

**Sequence**: `CFBBFX`

| Seq | Pitch | Before | After | Description                 |
| --- | ----- | ------ | ----- | --------------------------- |
| 1   | C     | 0-0    | 0-1   | Called strike               |
| 2   | F     | 0-1    | 0-2   | Foul ball                   |
| 3   | B     | 0-2    | 1-2   | Ball                        |
| 4   | B     | 1-2    | 2-2   | Ball                        |
| 5   | F     | 2-2    | 2-2   | Foul ball (no strike added) |
| 6   | X     | 2-2    | -     | Ball in play                |

**Sequence**: `CSBBBX` (Full count)

| Seq | Pitch | Before | After | Description       |
| --- | ----- | ------ | ----- | ----------------- |
| 1   | C     | 0-0    | 0-1   | Called strike     |
| 2   | S     | 0-1    | 0-2   | Swinging strike   |
| 3   | B     | 0-2    | 1-2   | Ball              |
| 4   | B     | 1-2    | 2-2   | Ball              |
| 5   | B     | 2-2    | 3-2   | Ball (full count) |
| 6   | X     | 3-2    | -     | Ball in play      |

**Sequence**: `B>FB.C.X` (With modifiers)

| Seq | Pitch | Before | After | Description               |
| --- | ----- | ------ | ----- | ------------------------- |
| 1   | B     | 0-0    | 1-0   | Ball                      |
| -   | >     | -      | -     | Runner going (ignored)    |
| 2   | F     | 1-0    | 1-1   | Foul ball                 |
| 3   | B     | 1-1    | 2-1   | Ball                      |
| -   | .     | -      | -     | Play not involving batter |
| 4   | C     | 2-1    | 2-2   | Called strike             |
| -   | .     | -      | -     | Play not involving batter |
| 5   | X     | 2-2    | -     | Ball in play              |

## API Implementation

### Database Schema

Pitch sequences are stored in the `plays.pitches` column as text:

```sql
SELECT pitches FROM plays WHERE gid = 'SDN202403200' AND pn = 1;
-- Result: "BFBBV"
```

### Repository Layer

The `PitchRepository` parses sequences on-demand:

1. Query `plays` table with filters (pitcher, batter, date, etc.)
2. For each matching play, parse the `pitches` field
3. Generate individual `Pitch` records with context
4. Apply pitch-level filters (pitch type, count, etc.)
5. Return paginated results

### Performance Considerations

Pitch parsing happens at query time. For large datasets:

- Database filters reduce plays before parsing
- Pitch-level filters applied in memory after parsing
- Pagination limits result set size
- Typical query parses 50-200 plate appearances

## API Response Format

Each pitch includes full context from the parent play:

```json
{
    "game_id": "SDN202403200",
    "play_num": 1,
    "inning": 1,
    "top_bot": 0,
    "bat_team": "LAN",
    "pit_team": "SDN",
    "date": "20240320",
    "batter": "bettm001",
    "pitcher": "darvy001",
    "bat_hand": "R",
    "pit_hand": "R",
    "outs_pre": 0,
    "seq_num": 3,
    "pitch_type": "B",
    "ball_count": 1,
    "strike_count": 1,
    "is_in_play": false,
    "is_strike": false,
    "is_ball": true,
    "description": "Ball",
    "event": null
}
```

### Key Fields

- **seq_num**: Pitch number within the plate appearance (1-indexed)
- **pitch_type**: Single character Retrosheet code
- **ball_count**: Balls in count before this pitch
- **strike_count**: Strikes in count before this pitch
- **is_in_play**: True for X pitches
- **is_strike**: True for C, S, F, L, M, O, T, V
- **is_ball**: True for B, P, I, H
- **description**: Human-readable pitch type
- **event**: Play result (only on final pitch)

## Query Examples

### Filter by Pitcher

Get all pitches thrown by Yu Darvish:

```bash
GET /v1/pitches?pitcher=darvy001&per_page=50
```

### Filter by Count

Find all 3-2 count pitches:

```bash
GET /v1/pitches?ball_count=3&strike_count=2
```

### Filter by Pitch Type

Get all balls in play:

```bash
GET /v1/pitches?pitch_type=X
```

### Game-Specific Pitches

Get all pitches from a specific game:

```bash
GET /v1/games/SDN202403200/pitches
```

### Plate Appearance Pitches

Get pitches from a single plate appearance:

```bash
GET /v1/games/SDN202403200/plays/1/pitches
```

## Limitations

### What's Available

- Pitch type (ball, strike, foul, in play)
- Count progression
- Batter/pitcher matchup
- Game situation (inning, outs, runners)
- Play outcome (for final pitch)

### What's NOT Available

Retrosheet pitch sequences do not include:

- Pitch velocity
- Pitch movement/break
- Pitch location (inside/outside, high/low)
- Specific pitch classification (fastball, curveball, etc.)
- Catcher framing data
- Umpire-specific tendencies

For detailed pitch-by-pitch data with velocity and location, use MLB's Statcast data source instead.

## References

- [Retrosheet Play-by-Play Specification](https://www.retrosheet.org/eventfile.htm)
- [Retrosheet Pitch Sequence Codes](https://www.retrosheet.org/eventfile.htm#6)
