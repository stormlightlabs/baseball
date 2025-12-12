# Parks, Umpires, Managers & Other Entities API Overview

Supporting people/places exposed through Lahman and Retrosheet joins.

## Summary

| Endpoint                                | Dataset | Highlights                                                                   |
| --------------------------------------- | ------- | ---------------------------------------------------------------------------- |
| `GET /v1/parks`                         | L+R     | Ballpark directory with Lahman metadata enriched by Retrosheet park IDs.     |
| `GET /v1/parks/{park_id}`               | L+R     | Detailed park record including capacity, location, and hosted games.         |
| `GET /v1/managers`                      | L       | Manager careers with win/loss totals and franchise associations.             |
| `GET /v1/managers/{manager_id}`         | L       | Full profile for a manager including playing background when available.      |
| `GET /v1/managers/{manager_id}/seasons` | L       | Season-by-season records aligned with Lahman managerial tables.              |
| `GET /v1/umpires`                       | L+R     | Combined Lahman/Retrosheet umpire roster, mapped so IDs stay consistent.     |
| `GET /v1/umpires/{umpire_id}`           | L+R     | Umpire biography plus officiated games split by home/away plate assignments. |
| `GET /v1/ejections`                     | R       | List of ejection events with participants, inning, and reason text.          |
| `GET /v1/seasons/{year}/ejections`      | R       | Season-scoped ejection slice to support league trend charts.                 |

## Endpoint Details

- Park and umpire objects expose both Lahman and Retrosheet identifiers to simplify cross-linking when data only exists in one source.
- Manager endpoints share the same pagination envelope as players and include derived stats (win pct, playoff appearances) so clients do not need to recompute.
- Ejection endpoints are driven entirely by Retrosheet event logs; filters like `player_id`, `umpire_id`, `team_id`, or `from`/`to` allow targeted lookups when reviewing discipline trends.
