# Awards, All-Star & Postseason API Overview

## Summary

| Endpoint                                   | Dataset | Highlights                                                                    |
| ------------------------------------------ | ------- | ----------------------------------------------------------------------------- |
| `GET /v1/awards`                           | L       | Browse award definitions and available years (MVP, Cy Young, ROY, etc.).      |
| `GET /v1/awards/{award_id}`                | L       | Deep view of a specific award including description and eligibility metadata. |
| `GET /v1/seasons/{year}/awards`            | L       | Season-specific award winners for use on recap or season pages.               |
| `GET /v1/seasons/{year}/postseason/series` | L       | Lists postseason series (LCS, WS, etc.) with participants and results.        |
| `GET /v1/seasons/{year}/postseason/games`  | L+R     | Postseason games joined with Retrosheet so box scores and plays stay linked.  |
| `GET /v1/allstar/games`                    | L+R     | All-Star Game history, including leagues, final score, and host park.         |
| `GET /v1/allstar/games/{game_id}`          | L+R     | Detail record for a particular All-Star Game with rosters and notable events. |

## Endpoint Details

- Award endpoints leverage Lahman data so they stay synchronized with yearly releases;
    - Filters include `award_id`, `league`, and `from`/`to` for multi-year dashboards.
- Postseason series/games include bracket order, series type, and coverage of tiebreaker games which often differ from regular-season scheduling.
- All-Star endpoints expose both Lahman and Retrosheet identifiers, enabling links into `/v1/games/{game_id}` or `/v1/players/{player_id}` for rosters.
