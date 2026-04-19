# Database Loading Contract (Complete Slice)

This is the source of truth for loading a **complete slice** of the database.

For this contract, a complete slice means:

- Lahman + Retrosheet data loaded for the selected year policy.
- Negro Leagues, FanGraphs constants, salary summary, Retrosheet player crosswalk, biodata, weather metadata, parks metadata, and all-star coverage loaded.
- ETL validation and readiness checks pass.

## Command Prefixes

Use one prefix consistently.

| Environment    | Prefix                             |
| -------------- | ---------------------------------- |
| Local          | `./tmp/baseball`                   |
| Docker/Coolify | `docker compose exec app baseball` |

Examples below use `<BASEBALL>` as a placeholder for your chosen prefix.

## Primary Entry Point

`<BASEBALL> etl` is the default full ETL pipeline.

`<BASEBALL> etl run` is an explicit alias with identical behavior and flags.

Shared pipeline flags:

- `--profile dev|prod`
- `--mode incremental|full`
- `--years <year-list|range|all>`
- `--era <comma-separated-era-list>`
- `--data-root <path>`

Data root resolution order:

1. `--data-root`
2. `BASEBALL_DATA_ROOT`
3. `tools/data`
4. legacy `data`

The same `--data-root` flag is available for `db` commands.

## Recommended Flows

### 1) Build and configure (local dev only)

```bash
task build
cp conf/conf.example.toml conf.toml
```

Optional full reset:

```bash
<BASEBALL> db recreate --config conf.toml
```

### 2) Apply schema migrations

```bash
<BASEBALL> db migrate
```

### 3) Run the full ETL pipeline

Representative development slice:

```bash
<BASEBALL> etl --profile dev
```

Representative development slice with explicit years:

```bash
<BASEBALL> etl --profile dev --years 2022-2025
```

Production/exhaustive load:

```bash
<BASEBALL> etl --profile prod --mode full
```

### 4) Validate completeness

```bash
<BASEBALL> etl validate --profile dev
<BASEBALL> etl status
```

Server/API checks:

```bash
curl http://localhost:8080/v1/ready
curl http://localhost:8080/v1/meta/datasets
curl http://localhost:8080/v1/health
```

### 5) Production-style external clone flow

```bash
tmpdir="$(mktemp -d)"
git clone --depth=1 <baseball-data-repo-url> "$tmpdir/baseball-data"
<BASEBALL> etl --profile prod --mode full --data-root "$tmpdir/baseball-data"
<BASEBALL> etl validate --profile prod --data-root "$tmpdir/baseball-data"
rm -rf "$tmpdir"
```

## Stage Commands (First-Class, Composable)

These commands are still supported for partial runs, debugging, and CI:

- `<BASEBALL> etl fetch retrosheet`
- `<BASEBALL> etl fetch negroleagues`
- `<BASEBALL> etl load lahman`
- `<BASEBALL> etl load retrosheet`
- `<BASEBALL> etl load retrosheet players`
- `<BASEBALL> etl load negroleagues`
- `<BASEBALL> etl load fangraphs`
- `<BASEBALL> etl load salary`
- `<BASEBALL> etl load biodata`
- `<BASEBALL> etl load weather`
- `<BASEBALL> etl load parks`
- `<BASEBALL> etl load allstar`
- `<BASEBALL> etl status`

## Command Contract Table

| Command                                 | Purpose                                                      | Write behavior                                                                                  | Prerequisites      |
| --------------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- | ------------------ |
| `<BASEBALL> etl` / `<BASEBALL> etl run` | Full ETL orchestration (extract, transform, load, validate)  | Downloads archives, loads core + supplemental datasets, refreshes materialized views, validates | Reachable Postgres |
| `<BASEBALL> etl validate --profile <P>` | Profile-aware ETL completeness validation                    | Read-only                                                                                       | Reachable Postgres |
| `<BASEBALL> etl fetch retrosheet`       | Download Retrosheet game logs, plays, and auxiliary archives | Writes files under `<data-root>/retrosheet`                                                     | Network access     |
| `<BASEBALL> etl fetch negroleagues`     | Download and extract Negro Leagues archives                  | Writes files under `<data-root>/retrosheet/negroleagues`                                        | Network access     |
| `<BASEBALL> etl load <dataset>`         | Execute a specific load stage                                | Dataset-dependent inserts/updates/truncation                                                    | Source files ready |
| `<BASEBALL> etl status`                 | Show core + supplemental dataset freshness and health        | Read-only                                                                                       | Reachable Postgres |
| `<BASEBALL> db migrate`                 | Apply SQL migrations                                         | Schema create/alter only                                                                        | Reachable Postgres |

## Retrosheet Era Contract

Supported `--era` values:

- `fed`
- `nlg`
- `boomer`
- `pitcher`
- `turf`
- `steroid`
- `moneyball`
- `statcast`
- `modern`

Invalid era values fail with the full valid-era list.
