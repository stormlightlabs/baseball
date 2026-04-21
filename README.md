<!-- markdownlint-disable MD033 -->
# Big Fly monorepo

![API Banner](./docs/banner.png)

A comprehensive platform for baseball statistics built with Go, Svelte & Flutter serving data from the Lahman Baseball Database and Retrosheet.

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

## Local Development

<details>
<summary>
The CLI handles ETL, database, and server operations so you can rebuild the stack without bespoke scripts.
</summary>

### Build

```bash
task build
task build:etl
./tmp/baseball --help
./tmp/baseball-etl --help
```

### Complete Slice Loading

For the full non-optional loading contract, see [data-loading.md](./docs/internal/data-loading.md).

Dataset root resolution for `etl` and `db` commands:

1. `--data-root`
2. `BASEBALL_DATA_ROOT`
3. `data`

Data is local-first: keep source files under the resolved data root (`data` by default).
When Retrosheet files are missing for a target window, fetch them with ETL commands before running a load.

Quick local example for a complete representative slice:

```bash
cp conf/conf.example.toml conf.toml
./tmp/baseball db recreate --config conf.toml
./tmp/baseball db migrate --config conf.toml
./tmp/baseball-etl worker
./tmp/baseball-etl run --profile=dev
./tmp/baseball-etl validate --profile=dev
./tmp/baseball-etl status
```

Fetch examples (when Retrosheet files for your window are not already present):

```bash
./tmp/baseball-etl fetch retrosheet --years=2023-2025
./tmp/baseball-etl fetch negroleagues
./tmp/baseball-etl fetch chadwick
./tmp/baseball-etl cleanup retrosheet --dry-run
```

For large Retrosheet slices, keep migration and recomputation separate, and process in bounded year windows:

```bash
./tmp/baseball db migrate --config conf.toml
./tmp/baseball-etl run --profile=dev --years=2023-2025
./tmp/baseball db refresh-views player_game_batting_stats player_game_pitching_stats player_game_fielding_stats team_game_stats
./tmp/baseball db refresh-views no_hitters cycles multi_hr_games triple_plays extra_inning_games
./tmp/baseball db refresh-views season_batting_leaders season_pitching_leaders career_batting_leaders career_pitching_leaders
./tmp/baseball db refresh-views player_id_map team_franchise_map park_map
```

`db migrate` is structural/idempotent; treat materialized view refresh as an explicit incremental operation.

`./tmp/baseball-etl worker` is the long-running queue consumer.
`./tmp/baseball-etl run` is the canonical enqueue entrypoint.
Treat ETL as a batched worker flow on shared VMs: prefer scoped `--years` runs over unbounded full-history jobs unless you are operating a larger host.

`run` is enqueue-first by default; use `--enqueue-only=false` only when you explicitly want one command to enqueue + drain locally.

For exhaustive production-style ingestion:

```bash
./tmp/baseball-etl run --profile=prod --mode=full
./tmp/baseball-etl validate --profile=prod
```

Retrosheet `--era` values: `fed`, `nlg`, `boomer`, `pitcher`, `turf`, `steroid`, `moneyball`, `statcast`, `modern`.

The full ETL command also accepts `--years` and `--era` to customize the Retrosheet window.
It also accepts `--data-root` when data is mounted outside defaults.

Batched worker pattern (recommended for VM safety):

```bash
./tmp/baseball-etl fetch retrosheet --years=2022-2023
./tmp/baseball-etl run --profile=prod --years=2022-2023
./tmp/baseball-etl validate --profile=prod --years=2022-2023

./tmp/baseball-etl fetch retrosheet --years=2024-2025
./tmp/baseball-etl run --profile=prod --years=2024-2025
./tmp/baseball-etl validate --profile=prod --years=2024-2025
```

### Server

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

### Fetch

Think of `baseball server fetch` as a built-in, auth-aware `curl`. It:

- Accepts relative paths (e.g., `players?name=ruth`) and automatically targets `/v1`
- Applies syntax highlighting/pretty-printing by default, or `--raw` when you need plain JSON for `jq`
- Injects bearer tokens/API keys via `--token` or `--api-key` flags, so you can hit protected routes without manually crafting headers

</details>

## Features

### HTTP API

The REST API lives at `/v1` (or the host/port defined in `conf.toml`), covering players, teams, stats, games, events, pitches, search, metadata, and auth.

### Pitch-Level Data

Query individual pitches derived from Retrosheet sequences, including count state and game context.
See [Pitch-Level API docs](./web/src/routes/docs/api-pitches.md) and [Pitch sequencing internals](./web/src/routes/docs/pitches.md).

### Derived & Advanced Analytics

Derived endpoints provide streak detection, player splits, run differential windows, game win probability curves, and win expectancy lookups.
See [Derived & Advanced docs](./web/src/routes/docs/api-derived-advanced.md).

### Natural Language Game Search

Search supports natural language parsing for games (teams, season, postseason context, aliases).
See [Search docs](./web/src/routes/docs/api-search.md).

## Documentation

- API docs in web app: `/docs` (or `/docs/{slug}`)
- Swagger/OpenAPI explorer: `/explorer`

Recommended docs entry points:

- [Introduction](./web/src/routes/docs/introduction.md)
- [API Overview](./web/src/routes/docs/api-overview.md)
- [Auth](./web/src/routes/docs/api-auth.md)
- [Meta & Utility](./web/src/routes/docs/api-meta-utility.md)
- [Search](./web/src/routes/docs/api-search.md)
- [Players](./web/src/routes/docs/api-players.md)
- [Teams](./web/src/routes/docs/api-teams.md)
- [Games](./web/src/routes/docs/api-games.md)
- [Stats](./web/src/routes/docs/api-stats.md)
- [Pitches](./web/src/routes/docs/api-pitches.md)
- [Derived & Advanced](./web/src/routes/docs/api-derived-advanced.md)
- [Computed & Advanced](./web/src/routes/docs/api-computed.md)
- [Attribution](./web/src/routes/docs/attribution.md)

## Development Notes

- Generate Swagger docs after endpoint/comment changes:

```bash
task swagger:generate
```

- Discover available tasks:

```bash
task --list
```

## Attribution

This project uses data from:

- **Lahman Baseball Database**: The information used here was obtained free of charge from and is copyrighted by Sean Lahman.
[SABR Lahman Database](https://sabr.org/lahman-database/)
- **Retrosheet**: The information used here was obtained free of charge from and is copyrighted by Retrosheet.
[Retrosheet.org](https://www.retrosheet.org/)
- **Baseball Prospectus**: Salary data sourced from [Cot's Baseball Contracts](https://legacy.baseballprospectus.com/compensation/cots/).
- **MLB**: This project and its author are not affiliated with MLB or any MLB team. This REST API interfaces with MLB's Stats API.
Use of MLB data is subject to the notice posted at <http://gdx.mlb.com/components/copyright.txt> (is also available in every request)
- wOBA weights, league wOBA, wOBA scale, FIP constants, and park factors are taken from [FanGraphs' Guts! tool](https://www.fangraphs.com/tools/guts).
wOBA definitions follow Tom Tango's formulation as documented in the FanGraphs Library.
