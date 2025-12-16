# Achievements & Event Feeds API Overview

Computed endpoints that highlight notable accomplishments derived directly from Retrosheet plays and game logs.

## Summary

| Endpoint                 | Dataset | Highlights                                                                                                                |
| ------------------------ | ------- | ------------------------------------------------------------------------------------------------------------------------- |
| `GET /v1/achievements/*` | R       | Computed achievements: no-hitters (446), cycles (90), multi-HR games (236), triple plays (143), 20+ inning games (3,851). |

`/v1/achievements/*` compiles notable events directly from plays/games, giving clients a story-driven feed without reprocessing the raw logs.
