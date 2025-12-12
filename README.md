<!-- markdownlint-disable MD033 -->
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

### CLI Toolkit

The `baseball` CLI handles ETL, database, and server operations so you can rebuild the stack without bespoke scripts.

<details>
<summary>
Usage
</summary>

#### Build

```bash
task build
./tmp/baseball --help
```

#### ETL

```bash
# Prepare folders or downloads
./tmp/baseball etl fetch lahman
./tmp/baseball etl fetch retrosheet

# Load datasets once the CSV/zip files exist
./tmp/baseball etl load lahman
./tmp/baseball etl load retrosheet
```

#### Database

```bash
# Apply migrations
./tmp/baseball db migrate

# Reseed everything (accepts Lahman/Retrosheet-specific subcommands and year ranges)
./tmp/baseball db reset --years 2023-2025
./tmp/baseball db populate --csv-dir ./data/lahman/csv --years 2023
```

#### Server

```bash
# Start the HTTP API (pass --debug to bypass auth locally)
./tmp/baseball server start --config conf.toml

# Smoke-test endpoints with formatted output
./tmp/baseball server fetch 'search/games?q=dodgers%202024'

# Check readiness & view auth instructions
./tmp/baseball server health
./tmp/baseball server auth
```

Every command accepts `--config` to point at a custom `conf.toml`, inherits rate-limits/auth from your server configuration, and prints structured output

#### Fetch

Think of `baseball server fetch` as a built-in, auth-aware `curl`. It:

- Accepts relative paths (e.g., `players?name=ruth`) and automatically targets `/v1`
- Applies syntax highlighting/pretty-printing by default, or `--raw` when you need plain JSON for `jq`
- Injects bearer tokens/API keys via `--token` or `--api-key` flags, so you can hit protected routes without manually crafting headers

</details>

### HTTP API

The REST API lives at `/v1` (or the host/port defined in `conf.toml`). Interactive Swagger UI continues to be available at `/docs` for request/response schemas.

<details>
<summary>
Endpoints
</summary>

- **Authentication**: API keys (`Authorization: Bearer sk_...`) or dashboard-issued session tokens; start the server with `--debug` while iterating locally to skip auth.
- **Health**: `GET /v1/health` exposes service/DB status; mirrors what `baseball server health` checks.
- **Primary resources**:
    - `/v1/players`
    - `/v1/teams`
    - `/v1/stats`
    - `/v1/games`
    - `/v1/plays`
    - `/v1/awards`
    - `/v1/postseason`
    - `/v1/allstar`
    - `/v1/managers`
    - `/v1/parks`
    - `/v1/umpires`
    - `/v1/ejections`
    - `/v1/pitches`
- **Other**:
    - `/v1/meta` (dataset refresh metadata)
    - `/v1/search/*` for fuzzy finding &. natural-language lookup.
- **Authentication flows**:
    - `/v1/auth/github` and `/v1/auth/codeberg` drive OAuth
    - `/dashboard` lets you mint API keys after login.

</details>

<details>
<summary>
Examples
</summary>

```bash
# Query players with fuzzy matching
curl "/v1/players?name=babe%20ruth&season=1927"

# Inspect a specific team season
curl "/v1/teams/NYY?year=2024"

# Fetch postseason metadata and recent plays
curl "/v1/postseason/series?year=2024"
curl "/v1/plays?game_id=NYN202410010"

# Retrieve dataset refresh info
curl "/v1/meta/datasets"
```

</details>

### Pitch-Level Data

Query individual pitches derived from Retrosheet play-by-play pitch sequences. Each pitch includes the count, pitch type, and full game context.

**Endpoint:** `GET /v1/pitches`

<details>
<summary>
Examples
</summary>

```bash
# Get all pitches from a specific pitcher
curl "/v1/pitches?pitcher=darvy001&per_page=50"

# Find all 3-2 count pitches
curl "/v1/pitches?ball_count=3&strike_count=2"

# Get only balls in play
curl "/v1/pitches?pitch_type=X"

# Filter by matchup
curl "/v1/pitches?pitcher=darvy001&batter=ohtas001"

# Get all pitches from a game
curl "/v1/games/SDN202403200/pitches"

# Get pitches from a specific plate appearance
curl "/v1/games/SDN202403200/plays/1/pitches"
```

</details>

<details>
<summary>
Params
</summary>

- `batter` - Retrosheet batter ID
- `pitcher` - Retrosheet pitcher ID
- `bat_team` - Batting team ID
- `pit_team` - Pitching team ID
- `date` - Game date (YYYYMMDD)
- `date_from` / `date_to` - Date range (YYYYMMDD)
- `inning` - Inning number
- `pitch_type` - Pitch type code (B, C, F, S, X, etc.)
- `ball_count` - Filter by ball count (0-3)
- `strike_count` - Filter by strike count (0-2)
- `is_in_play` - Only pitches in play (X)
- `is_strike` - Only strikes
- `is_ball` - Only balls

</details>

See [pitching](./docs/pitches.md) for implementation details

### Natural Language Game Search

Search for games using natural language queries. The search understands team names, years, series keywords, and game numbers.

**Endpoint:** `GET /v1/search/games?q={query}&limit={limit}`

<details>
<summary>
Examples
</summary>

```bash
# Find 2024 World Series games
curl "/v1/search/games?q=world%20series%202024"

# Find Yankees vs Red Sox games in 2024
curl "/v1/search/games?q=yankees%20red%20sox%202024"

# Find 2024 All-Star game
curl "/v1/search/games?q=all%20star%202024"

# Find Dodgers games in 2024 with limit
curl "/v1/search/games?q=dodgers%202024&limit=10"
```

</details>

The search supports:

- Team names and common aliases (e.g., "yankees", "red sox", "dodgers")
- Years (any 4-digit year)
- Postseason keywords ("world series", "playoffs", "postseason", "alcs", "nlcs", etc.)
- All-Star games ("all-star", "all star", "midsummer classic")
- Flexible query formats with automatic fuzzy matching

<details>
<summary>Implementation Details</summary>

The natural language search is powered by a three-layer approach:

1. **PostgreSQL Full-Text Search**: Uses `tsvector` and `tsquery` with GIN indexes for efficient text matching
2. **Fuzzy Matching**: Trigram indexes (`pg_trgm`) provide flexible substring matching
3. **Team Alias Resolution**: 77+ team name variations mapped to official team IDs with historical date ranges

#### Schema

- `games.search_text`: Precomputed searchable text including team names, dates, and keywords
- `games.search_tsv`: Full-text search vector automatically maintained via trigger
- `team_aliases`: Maps common team names to official IDs (e.g., "yankees" → "NYA")

#### Searching

1. Query hits indexed `search_text` column for instant results
2. PostgreSQL ranks matches by relevance using `ts_rank`
3. Results sorted by structured filters (when detected) then text relevance then date

#### Performance

- GIN indexes enable sub-millisecond search across 9000+ games
- Trigger-maintained search columns keep data synchronized on insert/update
- No runtime text processing required

</details>

## Development

### Swagger/OpenAPI Documentation

This project uses [swaggo/swag](https://github.com/swaggo/swag) for API documentation generation.

#### Generating Swagger Docs

Use the task command to generate swagger documentation:

```bash
task swagger:generate
```

This will:

1. Generate swagger docs from your API annotations
2. Automatically fix known compatibility issues

<details>
<summary>
Notes
</summary>

#### Known Issues

##### LeftDelim/RightDelim Build Errors

When generating swagger docs, swag may generate `LeftDelim` and `RightDelim` fields in `internal/docs/docs.go` that are incompatible with the current version of the swag library, causing build failures:

```log
internal/docs/docs.go:1085:2: unknown field LeftDelim in struct literal of type "github.com/swaggo/swag".Spec
internal/docs/docs.go:1086:2: unknown field RightDelim in struct literal of type "github.com/swaggo/swag".Spec
```

</details>

## Available Tasks

Run `task --list` to see all available tasks.

## Attribution

This project uses data from:

- **Lahman Baseball Database**: The information used here was obtained free of charge from and is copyrighted by Sean Lahman.
[SABR Lahman Database](https://sabr.org/lahman-database/)
- **Retrosheet**: The information used here was obtained free of charge from and is copyrighted by Retrosheet.
[Retrosheet.org](https://www.retrosheet.org/)
- **MLB**: This project and its author are not affiliated with MLB or any MLB team. This REST API interfaces with MLB's Stats API.
Use of MLB data is subject to the notice posted at <http://gdx.mlb.com/components/copyright.txt> (is also available in every request)
