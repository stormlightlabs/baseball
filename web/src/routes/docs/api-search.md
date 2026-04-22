# Search & Lookup API Overview

This section describes fuzzy lookup helpers.

## Summary

| Endpoint             | Dataset | Highlights                                                                              |
| -------------------- | ------- | --------------------------------------------------------------------------------------- |
| `/v1/search/players` | L+R     | Fuzzy match players by name, era, league, or handedness; returns lightweight cards.     |
| `/v1/search/teams`   | L+R     | Search by city, nickname, franchise, or year ranges; helpful for UI autocomplete.       |
| `/v1/search/parks`   | L+R     | Ballpark lookup that spans historical names and locations.                              |
| `/v1/search/games`   | R       | Natural-language search like "Yankees vs Red Sox 2003 ALCS Game 7" mapped to `GAME_ID`. |

## Endpoint Details

- Player/team/park search share a consistent payload with `id`, `label`, `sub_label`, and dataset tags.
- `GET /v1/search/games` tokenizes free-form text into probable home/away teams, season, and postseason context before resolving to Retrosheet IDs; queries also accept structured params.

## Natural Language Game Search

`GET /v1/search/games` supports natural language queries and flexible phrasing. The parser understands:

- Team names and common aliases (for example: "yankees", "red sox", "dodgers")
- Years (4-digit seasons)
- Postseason keywords (`world series`, `playoffs`, `postseason`, `alcs`, `nlcs`, and similar)
- All-Star synonyms (`all-star`, `all star`, `midsummer classic`)
- Mixed free-form queries with fuzzy matching

## Examples

```bash
# Find 2024 World Series games
curl "/v1/search/games?q=world%20series%202024"

# Find Yankees vs Red Sox games in 2024
curl "/v1/search/games?q=yankees%20red%20sox%202024"

# Find 2024 All-Star game
curl "/v1/search/games?q=all%20star%202024"

# Limit result count
curl "/v1/search/games?q=dodgers%202024&limit=10"
```

## Implementation Notes

Natural-language game search is backed by three layers:

1. PostgreSQL full-text search (`tsvector`/`tsquery` + GIN indexes)
2. Trigram-based fuzzy matching (`pg_trgm`)
3. Team alias resolution to canonical IDs with historical windows

### Schema

- `games.search_text`: precomputed searchable text
- `games.search_tsv`: full-text vector maintained by trigger
- `team_aliases`: alias mapping to canonical team IDs

### Ranking Flow

1. Candidate matches come from indexed text columns.
2. PostgreSQL relevance scoring (`ts_rank`) ranks candidate rows.
3. Structured filters (when detected) are applied before final relevance/date ordering.
