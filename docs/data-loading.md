# Database Loading Contract (Complete Slice)

This is the source of truth for loading a **complete slice** of the database.

For this contract, a complete slice means:

- Lahman + Retrosheet data loaded for your chosen year window.
- Negro Leagues, FanGraphs constants, salary summary, Retrosheet player crosswalk, biodata, weather metadata, parks metadata, and all-star coverage loaded.
- Dataset freshness/health checks pass in CLI and API readiness endpoints.

If you skip any command in the sequence below, the slice is incomplete.

## Command Prefixes

Use one prefix consistently.

| Environment    | Prefix                             |
| -------------- | ---------------------------------- |
| Local          | `./tmp/baseball`                   |
| Docker/Coolify | `docker compose exec app baseball` |

Examples below use `<BASEBALL>` as a placeholder for your chosen prefix.

## Complete-Slice Sequence

Choose a year window once and use it consistently.

- Example focused slice: `YEARS=2022-2025`
- Full history slice: `YEARS=all`

### 1) Build and configure (local dev only)

```bash
task build
cp conf/conf.example.toml conf.toml
```

If needed for a full reset:

```bash
<BASEBALL> db recreate --config conf.toml
```

### 2) Apply schema migrations

```bash
<BASEBALL> db migrate
```

### 3) Fetch external archives for the target slice

```bash
<BASEBALL> etl fetch retrosheet --years "${YEARS}"
<BASEBALL> etl fetch negroleagues
```

### 4) Populate core warehouse (Lahman + Retrosheet)

```bash
<BASEBALL> db populate all --years "${YEARS}"
```

`db populate` also accepts `db repopulate` as a compatibility alias.

### 5) Load required supplemental datasets

Run in this order:

```bash
<BASEBALL> etl load negroleagues
<BASEBALL> etl load fangraphs
<BASEBALL> etl load salary
<BASEBALL> etl load retrosheet players
<BASEBALL> etl load biodata
<BASEBALL> etl load weather
<BASEBALL> etl load parks
<BASEBALL> etl load allstar
```

### 6) Verify completeness

```bash
<BASEBALL> etl status
```

Server/API checks:

```bash
curl http://localhost:8080/v1/ready
curl http://localhost:8080/v1/meta/datasets
curl http://localhost:8080/v1/health
```

## Command Contract Table

This table documents what each command does, including defaults and write behavior.

| Command                                           | Purpose                                                                                           | Write behavior                                                          | Prerequisites                                             | Default behavior                                         |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- | --------------------------------------------------------- | -------------------------------------------------------- |
| `<BASEBALL> db migrate`                           | Applies all SQL migrations                                                                        | Schema create/alter only                                                | Reachable Postgres                                        | None                                                     |
| `<BASEBALL> etl fetch retrosheet --years <YEARS>` | Downloads Retrosheet game logs, plays, ejections, allplayers, gameinfo, allstar, biodata archives | Writes files under `data/retrosheet`                                    | Network access                                            | If `--years` omitted, uses `2023-2025`                   |
| `<BASEBALL> etl fetch negroleagues`               | Downloads and extracts Negro Leagues archives                                                     | Writes files under `data/retrosheet/negroleagues`                       | Network access                                            | None                                                     |
| `<BASEBALL> db populate all --years <YEARS>`      | Loads Lahman + Retrosheet core datasets and refreshes materialized views                          | Truncates/reloads Lahman tables; inserts/upserts Retrosheet games/plays | Migrations complete; source files present or downloadable | If `--years` omitted, Retrosheet defaults to `2023-2025` |
| `<BASEBALL> etl load negroleagues`                | Loads Negro Leagues `gameinfo.csv` and `plays.csv`                                                | Inserts into `games`/`plays`                                            | `etl fetch negroleagues` completed                        | None                                                     |
| `<BASEBALL> etl load fangraphs`                   | Loads wOBA, league constants, park factors                                                        | Truncates/reloads constants tables                                      | `data/fangraphs/*` files available                        | None                                                     |
| `<BASEBALL> etl load salary`                      | Loads salary summary and per-year salary enrichments                                              | Truncates/reloads `salary_summary`; updates/inserts Lahman salaries     | `data/salaries/*` available                               | None                                                     |
| `<BASEBALL> etl load retrosheet players`          | Loads Retrosheet player crosswalk                                                                 | Truncates/reloads `retrosheet_players`                                  | `allplayers.zip` or extracted CSV available               | None                                                     |
| `<BASEBALL> etl load biodata`                     | Loads player biographical data, relatives, coaches, umpires                                       | Upserts/loads biodata-related tables                                    | `biodata.zip` available                                   | None                                                     |
| `<BASEBALL> etl load weather`                     | Applies weather/game metadata to existing games                                                   | Updates `games` metadata columns                                        | `gameinfo.csv` available                                  | None                                                     |
| `<BASEBALL> etl load parks`                       | Fills missing park metadata and refreshes park map                                                | Updates park metadata + refreshes `park_map`                            | Migrations complete                                       | None                                                     |
| `<BASEBALL> etl load allstar`                     | Loads all-star gameinfo + plays into `games`/`plays`                                              | Inserts/upserts all-star rows                                           | `allstar.zip` available                                   | None                                                     |
| `<BASEBALL> etl status`                           | Shows core + supplemental dataset health/freshness                                                | Read-only                                                               | Database reachable                                        | None                                                     |

## Retrosheet Era Contract

Supported `--era` values for Retrosheet load/populate commands:

- `fed`
- `nlg`
- `boomer`
- `pitcher`
- `turf`
- `steroid`
- `moneyball`
- `statcast`
- `modern`

If an invalid era is passed, CLI errors include the valid era list.
