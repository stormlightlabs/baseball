# League-specific Coverage API Overview

Dedicated endpoint aliases for historical leagues that filter the full plays/games dataset to a specific competition window.

## Summary

| Endpoint                  | Dataset | Highlights                                                                                |
| ------------------------- | ------- | ----------------------------------------------------------------------------------------- |
| `GET /v1/federalleague/*` | R       | Filters all relevant tables to Federal League coverage (1914-1915) via dedicated aliases. |
| `GET /v1/negroleagues/*`  | R       | Same schema as core endpoints but scoped to Negro Leagues seasons (1935-1949).            |

League views reuse the same schema as the base endpoints but pre-filter the dataset by league, acting as convenience aliases plus documentation anchors for historically under-documented competitions.
