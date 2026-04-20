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

## Large Dataset Guidance (Incremental First)

Retrosheet + derived leaderboards are large enough that full recomputation during schema deploy
can be disruptive. Treat materialized view population as a **separate operational step** from
`db migrate`.

Migration contract:

- `db migrate` is structural and idempotent (schema/object definitions only).
- Heavy data movement and materialized view population happen in ETL/load steps.

Recommended approach for large environments:

1. Run `db migrate` first (schema/object changes only).
2. Load source data (`etl ...` / `etl load ...`).
3. Refresh materialized views in bounded batches:

```bash
<BASEBALL> db refresh-views player_game_batting_stats player_game_pitching_stats player_game_fielding_stats team_game_stats
<BASEBALL> db refresh-views no_hitters cycles multi_hr_games triple_plays extra_inning_games
<BASEBALL> db refresh-views season_batting_leaders season_pitching_leaders career_batting_leaders career_pitching_leaders
<BASEBALL> db refresh-views player_id_map team_franchise_map park_map
```

This keeps migration time predictable and lets you pace heavy recomputes according to capacity.

Postgres runtime tuning for ETL-heavy writes should be applied via container command args (for example `postgres -c max_wal_size=12GB ...`), not `POSTGRES_*` environment variables.
This avoids frequent checkpoint churn during force clears and large refresh batches.

Materialized-view refresh observability:

- Each refresh attempt is recorded in `materialized_view_refresh_events`.
- ETL pipeline refresh attempts are linked to `etl_runs.id` via `run_id`.
- `db refresh-views` also records per-view attempts with step `db.refresh-views`.
- ETL phase-level timing and row counts are recorded in `etl_step_events` for long-running steps.

Quick diagnostics:

```sql
SELECT
  started_at,
  run_id,
  step,
  view_group,
  view_name,
  pass,
  attempt,
  mode,
  status,
  duration_ms,
  error
FROM materialized_view_refresh_events
ORDER BY started_at DESC
LIMIT 50;
```

```sql
SELECT
  started_at,
  run_id,
  step,
  phase,
  status,
  row_count,
  duration_ms,
  metadata,
  error
FROM etl_step_events
ORDER BY started_at DESC
LIMIT 100;
```

Force semantics (`etl load retrosheet --force` / `db populate retrosheet --force`):

- Force clear is year-scoped and uses date bounds (`YYYY0101` to `(YYYY+1)0101`) for both `games` and `plays`.
- Deletes execute with a per-year transaction boundary for safer resumability.
- Force mode does not delete rows outside requested year ranges.

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
curl http://localhost:8080/v1/meta/datasets?strict=true
curl -i http://localhost:8080/v1/meta | grep -i X-Count-Mode
curl http://localhost:8080/v1/health
```

### 5) Production bootstrap (automatic temp clone + cleanup)

```bash
<BASEBALL> etl --profile prod --mode full
<BASEBALL> etl validate --profile prod
```

If required snapshot files are missing under the default data root (`tools/data` locally,
`/home/app/tools/data` in Docker), the pipeline auto-clones
`https://github.com/stormlightlabs/bigflydata.git` into a temporary directory and
cleans it up after completion.

Optional overrides:

- `BASEBALL_DATA_REPO_URL` (default snapshot repo URL)
- `BASEBALL_DATA_REPO_REF` (branch/tag/SHA)
- `BASEBALL_DATA_AUTO_CLONE` (`true` by default; set `false` to disable)

### 6) Publish snapshot updates (`tools/data` submodule)

```bash
# from repo root
cd tools/data
git pull
git lfs install
uv run baseball-data sync
uv run baseball-data build
uv run baseball-data verify
git add .
git commit -m "Sync snapshot data"
git push

cd ../..
git add tools/data .gitmodules
git commit -m "Bump tools/data submodule"
git push
```

Notes:

- `tools/data` points to `https://github.com/stormlightlabs/bigflydata.git`.
- Push to `tools/data` first; the parent repo only records the submodule commit SHA.

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
- `<BASEBALL> etl status --strict` (exact-count audit mode)

## Command Contract Table

| Command                                  | Purpose                                                      | Write behavior                                                                                  | Prerequisites      |
| ---------------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- | ------------------ |
| `<BASEBALL> etl` / `<BASEBALL> etl run`  | Full ETL orchestration (extract, transform, load, validate)  | Downloads archives, loads core + supplemental datasets, refreshes materialized views, validates | Reachable Postgres |
| `<BASEBALL> etl validate --profile <P>`  | Profile-aware ETL completeness validation                    | Read-only                                                                                       | Reachable Postgres |
| `<BASEBALL> etl fetch retrosheet`        | Download Retrosheet game logs, plays, and auxiliary archives | Writes files under `<data-root>/retrosheet`                                                     | Network access     |
| `<BASEBALL> etl fetch negroleagues`      | Download and extract Negro Leagues archives                  | Writes files under `<data-root>/retrosheet/negroleagues`                                        | Network access     |
| `<BASEBALL> etl load <dataset>`          | Execute a specific load stage                                | Dataset-dependent inserts/updates/truncation                                                    | Source files ready |
| `<BASEBALL> etl status [--strict]`       | Show core + supplemental dataset freshness and health        | Read-only (lightweight by default; `--strict` runs exact counts)                                | Reachable Postgres |
| `<BASEBALL> db migrate`                  | Apply SQL migrations                                         | Structural/idempotent schema object changes only (no MV population)                             | Reachable Postgres |
| `<BASEBALL> db refresh-views <views...>` | Populate/refresh materialized views                          | Data recomputation (run in batches for large datasets)                                          | Reachable Postgres |

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
