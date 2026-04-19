# Game Context & Weather API Overview

Adds enhanced game-level metadata including weather, wind, DH usage, and field conditions to the canonical `GET /v1/games/{id}` payload.

## Summary

| Endpoint             | Dataset | Highlights                                                                                                         |
| -------------------- | ------- | ------------------------------------------------------------------------------------------------------------------ |
| `GET /v1/games/{id}` | R       | Enhanced metadata payload with weather, wind, start time, DH usage, and field condition context (see Notes below). |

## Weather Payload Notes

`/v1/games/{id}` endpoint now returns:

- temp_f: Temperature in Fahrenheit
- sky: Sky condition (sunny, cloudy, dome, overcast, etc.)
- wind_direction: Wind direction (rtol, ltor, fromcf, etc.)
- wind_speed_mph: Wind speed in mph
- precip: Precipitation (none, drizzle, showers, rain, snow)
- field_condition: Field condition (dry, wet, damp, soaked)
- start_time: Game start time (HH:MM:SS format)
- used_dh: Whether designated hitter was used (boolean)

### Examples

```sh
curl -s "http://localhost:8080/v1/games/ARI202404010" | jq '{id, date, home_team, temp_f, sky, wind_direction, wind_speed_mph, precip, start_time, used_dh}'

curl -s "http://localhost:8080/v1/games/BAL202404010" | jq '{id, date, home_team, temp_f, sky, wind_direction, wind_speed_mph, precip, field_condition, start_time, used_dh}'
```

```json
{
    "id": "ARI202404010",
    "date": "2024-04-01T00:00:00Z",
    "home_team": "ARI",
    "temp_f": 72,
    "sky": "dome",
    "wind_direction": null,
    "wind_speed_mph": null,
    "precip": "none",
    "start_time": "18:40:00",
    "used_dh": true
}

{
    "id": "BAL202404010",
    "date": "2024-04-01T00:00:00Z",
    "home_team": "BAL",
    "temp_f": 57,
    "sky": "overcast",
    "wind_direction": "rtol",
    "wind_speed_mph": 2,
    "precip": "none",
    "field_condition": null,
    "start_time": "18:35:00",
    "used_dh": true
}
```

### Sky Conditions

| Value    | Meaning                                           | Count        |
| -------- | ------------------------------------------------- | ------------ |
| cloudy   | Cloudy conditions                                 | 13,522 games |
| sunny    | Clear, sunny weather                              | 11,684 games |
| dome     | Indoor stadium (domed or retractable roof closed) | 5,353 games  |
| overcast | Overcast/gray sky                                 | 2,455 games  |
| night    | Night game (sky condition)                        | 342 games    |

### Wind Direction

Wind directions describe where the wind is blowing FROM, relative to home plate:

| Value  | Direction         | Description                                                       | Count       |
| ------ | ----------------- | ----------------------------------------------------------------- | ----------- |
| ltor   | Left to Right     | Wind blowing from left field toward right field                   | 5,441 games |
| rtol   | Right to Left     | Wind blowing from right field toward left field                   | 4,777 games |
| tocf   | To Center Field   | Wind blowing from home plate out toward center field              | 4,458 games |
| torf   | To Right Field    | Wind blowing from left/home area toward right field               | 4,179 games |
| tolf   | To Left Field     | Wind blowing from right/home area toward left field               | 3,081 games |
| fromrf | From Right Field  | Wind blowing from right field toward home/left field (in from RF) | 2,430 games |
| fromlf | From Left Field   | Wind blowing from left field toward home/right field (in from LF) | 2,200 games |
| fromcf | From Center Field | Wind blowing from center field toward home plate (in from CF)     | 1,978 games |

#### Impact on Gameplay

Wind directions that help hitters (carry the ball): - tocf, torf, tolf - Wind blowing out toward the outfield helps fly balls travel farther

Wind directions that help pitchers: - fromcf, fromrf, fromlf - Wind blowing in from the outfield keeps fly balls in the park

Crosswinds (affect ball trajectory sideways): - ltor, rtol - Can push fly balls toward foul territory or make curved balls break more

These values come from Retrosheet's official game logs and are particularly detailed for games for the modern era.
