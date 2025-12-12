# Baseball API

A comprehensive REST API for baseball statistics built with Go, serving data from the Lahman Baseball Database and Retrosheet.

## Quick Start

### Build

```bash
task build
```

### Run the Server

```bash
task server:start
```

The API will be available at <http://localhost:8080>, with interactive documentation at <http://localhost:8080/docs/>.

## Features

### Natural Language Game Search

Search for games using natural language queries. The search understands team names, years, series keywords, and game numbers.

**Endpoint:** `GET /v1/search/games?q={query}&limit={limit}`

**Examples:**

```bash
# Find 2024 World Series games
curl "http://localhost:8080/v1/search/games?q=world%20series%202024"

# Find Yankees vs Red Sox games in 2024
curl "http://localhost:8080/v1/search/games?q=yankees%20red%20sox%202024"

# Find 2024 All-Star game
curl "http://localhost:8080/v1/search/games?q=all%20star%202024"

# Find Dodgers games in 2024 with limit
curl "http://localhost:8080/v1/search/games?q=dodgers%202024&limit=10"
```

The search supports:

- Team names and common aliases (e.g., "yankees", "red sox", "dodgers")
- Years (any 4-digit year)
- Postseason keywords ("world series", "playoffs", "postseason", "alcs", "nlcs", etc.)
- All-Star games ("all-star", "all star", "midsummer classic")
- Flexible query formats with automatic fuzzy matching

<!-- markdownlint-disable MD033 -->
<details>
<summary>Implementation Details</summary>

The natural language search is powered by a three-layer approach:

1. **PostgreSQL Full-Text Search**: Uses `tsvector` and `tsquery` with GIN indexes for efficient text matching
2. **Fuzzy Matching**: Trigram indexes (`pg_trgm`) provide flexible substring matching
3. **Team Alias Resolution**: 77+ team name variations mapped to official team IDs with historical date ranges

### Schema

- `games.search_text`: Precomputed searchable text including team names, dates, and keywords
- `games.search_tsv`: Full-text search vector automatically maintained via trigger
- `team_aliases`: Maps common team names to official IDs (e.g., "yankees" → "NYA")

### Searching

1. Query hits indexed `search_text` column for instant results
2. PostgreSQL ranks matches by relevance using `ts_rank`
3. Results sorted by structured filters (when detected) then text relevance then date

### Performance

- GIN indexes enable sub-millisecond search across 9000+ games
- Trigger-maintained search columns keep data synchronized on insert/update
- No runtime text processing required

</details>

## Development

## Swagger/OpenAPI Documentation

This project uses [swaggo/swag](https://github.com/swaggo/swag) for API documentation generation.

### Generating Swagger Docs

Use the task command to generate swagger documentation:

```bash
task swagger:generate
```

This will:

1. Generate swagger docs from your API annotations
2. Automatically fix known compatibility issues

### Known Issues

#### LeftDelim/RightDelim Build Errors

When generating swagger docs, swag may generate `LeftDelim` and `RightDelim` fields in `internal/docs/docs.go` that are incompatible with the current version of the swag library, causing build failures:

```log
internal/docs/docs.go:1085:2: unknown field LeftDelim in struct literal of type "github.com/swaggo/swag".Spec
internal/docs/docs.go:1086:2: unknown field RightDelim in struct literal of type "github.com/swaggo/swag".Spec
```

## Available Tasks

Run `task --list` to see all available tasks.

## Attribution

This project uses data from:

- **Lahman Baseball Database**: The information used here was obtained free of charge from and is copyrighted by Sean Lahman.
[SABR Lahman Database](https://sabr.org/lahman-database/)
- **Retrosheet**: The information used here was obtained free of charge from and is copyrighted by Retrosheet.
[Retrosheet.org](https://www.retrosheet.org/)
