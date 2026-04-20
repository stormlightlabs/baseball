# Database Loading Contract (Complete Slice)

This is the source of truth for loading a **complete slice** of the database.

For this contract, a complete slice means:

- Lahman + Retrosheet data loaded for the selected year policy.
- Negro Leagues, FanGraphs constants, salary summary, Retrosheet player crosswalk, biodata, weather metadata, parks metadata, and all-star coverage loaded.
- ETL validation and readiness checks pass.

## Command Prefixes

Use the matching binary prefix for each command group.

| Environment    | Prefix                             |
| -------------- | ---------------------------------- |
| Local          | `./tmp/baseball` (db/server), `./tmp/baseball-etl` (etl) |
| Docker/Coolify | `docker compose exec app baseball` (db/server), `docker compose exec app baseball-etl` (etl) |

Examples below use `<BASEBALL>` for db/server commands and `<BASEBALL_ETL>` for ETL commands.

## Primary Entry Point

`<BASEBALL_ETL> run` is the canonical full ETL pipeline.

`<BASEBALL_ETL>` and `<BASEBALL_ETL> run` are compatible aliases.

Legacy `<BASEBALL> etl ...` commands remain available during transition.

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
2. Load source data (`<BASEBALL_ETL> run ...` / `<BASEBALL_ETL> load ...`).
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
<BASEBALL_ETL> run --profile dev
```

Representative development slice with explicit years:

```bash
<BASEBALL_ETL> run --profile dev --years 2022-2025
```

Production/exhaustive load:

```bash
<BASEBALL_ETL> run --profile prod --mode full
```

### 4) Validate completeness

```bash
<BASEBALL_ETL> validate --profile dev
<BASEBALL_ETL> status
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
<BASEBALL_ETL> run --profile prod --mode full
<BASEBALL_ETL> validate --profile prod
```

If required snapshot files are missing under the default data root (`tools/data` locally,
`/home/app/tools/data` in Docker), the pipeline auto-clones
`https://github.com/stormlightlabs/bigflydata.git` into a temporary directory and
cleans it up after completion.

Optional overrides:

- `BASEBALL_DATA_REPO_URL` (default snapshot repo URL)
- `BASEBALL_DATA_REPO_REF` (branch/tag/SHA)
- `BASEBALL_DATA_AUTO_CLONE` (`true` by default; set `false` to disable)

### 6) Publish snapshot updates (external snapshot repo)

```bash
tmpdir="$(mktemp -d)"
git clone --depth=1 https://github.com/stormlightlabs/bigflydata.git "$tmpdir/bigflydata"
cd "$tmpdir/bigflydata"
git lfs install
uv run baseball-data sync
uv run baseball-data build
uv run baseball-data verify
git add .
git commit -m "Sync snapshot data"
git push

cd -
rm -rf "$tmpdir"
```

Notes:

- The application repo does not vendor snapshot data as a git submodule.
- ETL auto-clone uses `BASEBALL_DATA_REPO_URL`/`BASEBALL_DATA_REPO_REF` when default data roots are incomplete.

## Stage Commands (First-Class, Composable)

These commands are still supported for partial runs, debugging, and CI:

- `<BASEBALL_ETL> fetch retrosheet`
- `<BASEBALL_ETL> fetch negroleagues`
- `<BASEBALL_ETL> load lahman`
- `<BASEBALL_ETL> load retrosheet`
- `<BASEBALL_ETL> load retrosheet players`
- `<BASEBALL_ETL> load negroleagues`
- `<BASEBALL_ETL> load fangraphs`
- `<BASEBALL_ETL> load salary`
- `<BASEBALL_ETL> load biodata`
- `<BASEBALL_ETL> load weather`
- `<BASEBALL_ETL> load parks`
- `<BASEBALL_ETL> load allstar`
- `<BASEBALL_ETL> status`
- `<BASEBALL_ETL> status --strict` (exact-count audit mode)

## Command Contract Table

| Command                                  | Purpose                                                      | Write behavior                                                                                  | Prerequisites      |
| ---------------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- | ------------------ |
| `<BASEBALL_ETL>` / `<BASEBALL_ETL> run`  | Full ETL orchestration (extract, transform, load, validate)  | Downloads archives, loads core + supplemental datasets, refreshes materialized views, validates | Reachable Postgres |
| `<BASEBALL_ETL> validate --profile <P>`  | Profile-aware ETL completeness validation                    | Read-only                                                                                       | Reachable Postgres |
| `<BASEBALL_ETL> fetch retrosheet`        | Download Retrosheet game logs, plays, and auxiliary archives | Writes files under `<data-root>/retrosheet`                                                     | Network access     |
| `<BASEBALL_ETL> fetch negroleagues`      | Download and extract Negro Leagues archives                  | Writes files under `<data-root>/retrosheet/negroleagues`                                        | Network access     |
| `<BASEBALL_ETL> load <dataset>`          | Execute a specific load stage                                | Dataset-dependent inserts/updates/truncation                                                    | Source files ready |
| `<BASEBALL_ETL> status [--strict]`       | Show core + supplemental dataset freshness and health        | Read-only (lightweight by default; `--strict` runs exact counts)                                | Reachable Postgres |
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
