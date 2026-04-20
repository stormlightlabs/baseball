# Database Loading Contract (Complete Slice)

This is the source of truth for loading a complete database slice with the new upstream/downstream split:

- `bigflydata` builds versioned data snapshots (`raw` + `prepared`)
- `baseball-etl` pulls, reads, upserts, validates

A complete slice means:

- Lahman + Retrosheet core coverage loaded for the requested year policy
- supplemental datasets loaded (Negro Leagues, FanGraphs constants, salaries, biodata, weather, parks, all-star)
- ETL validation + readiness checks pass

## Command Prefixes

| Environment | Prefix |
| --- | --- |
| Local | `./tmp/baseball` (db/server), `./tmp/baseball-etl` (etl), `uv run baseball-data` (snapshot build) |
| Docker/Coolify | `docker compose exec app baseball` (db/server), `docker compose exec app baseball-etl` (etl) |

Examples below use:

- `<BASEBALL>` for db/server commands
- `<BASEBALL_ETL>` for ETL commands
- `<BIGFLYDATA>` for snapshot-build commands

## Primary Operational Flow

### 1) Build or update snapshot in `bigflydata`

```bash
cd /Users/owais/Projects/bigflydata
<BIGFLYDATA> sync
<BIGFLYDATA> build
<BIGFLYDATA> verify
```

Target contract direction:

- extracted raw tabular files are canonical artifacts in VCS/LFS
- heavy transforms happen in Python (Polars + NumPy)
- prepared ingest-ready outputs are produced in `bigflydata`

### 2) Apply DB schema migrations

```bash
<BASEBALL> db migrate
```

`db migrate` is structural/idempotent only. Heavy recompute remains an explicit ETL/load step.

### 3) Run ETL ingestion

Representative dev slice:

```bash
<BASEBALL_ETL> run --profile dev --years 2022-2025
```

Full historical profile:

```bash
<BASEBALL_ETL> run --profile prod --mode full
```

### 4) Validate and inspect status

```bash
<BASEBALL_ETL> validate --profile dev
<BASEBALL_ETL> status
```

### 5) API readiness checks

```bash
curl http://localhost:8080/v1/ready
curl http://localhost:8080/v1/meta/datasets
curl http://localhost:8080/v1/meta/datasets?strict=true
curl http://localhost:8080/v1/health
```

## Data Root Resolution

Resolution order for ETL and DB commands:

1. `--data-root`
2. `BASEBALL_DATA_ROOT`
3. `tools/data`
4. legacy `data`

## Production Bootstrap (Auto-Clone)

When required files are missing under the default root (`tools/data` locally, `/home/app/tools/data` in Docker), ETL clones the snapshot repo to a temporary directory, uses it for the run, then cleans it up.

Optional overrides:

- `BASEBALL_DATA_REPO_URL`
- `BASEBALL_DATA_REPO_REF`
- `BASEBALL_DATA_AUTO_CLONE`

Example:

```bash
<BASEBALL_ETL> run --profile prod --mode full
<BASEBALL_ETL> validate --profile prod
```

## Publishing Snapshot Updates (`bigflydata`)

```bash
tmpdir="$(mktemp -d)"
git clone --depth=1 https://github.com/stormlightlabs/bigflydata.git "$tmpdir/bigflydata"
cd "$tmpdir/bigflydata"

<BIGFLYDATA> sync
<BIGFLYDATA> build
<BIGFLYDATA> verify

# target state includes prepared outputs generated in-repo
# commit raw/prepared/manifest updates together
git add .
git commit -m "Update snapshot raw/prepared datasets"
git push

cd -
rm -rf "$tmpdir"
```

## Large Dataset Guidance

Run heavy steps explicitly and in order:

1. `db migrate`
2. `baseball-etl run ...`
3. explicit partition maintenance/analysis for heavily changed serving tables

Partition/load observability:

- `etl_step_events`
- table/partition row-count + latency checks via ETL status and SQL diagnostics

## Transitional / Legacy Commands

Legacy stage commands remain available during migration, but should not be the long-term hot path:

- `<BASEBALL_ETL> fetch retrosheet`
- `<BASEBALL_ETL> fetch negroleagues`
- `<BASEBALL_ETL> load <dataset>`

Target state is snapshot-first ingestion from `bigflydata` prepared outputs with partitioned serving-table writes (no steady-state materialized-view rebuild loop).

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
