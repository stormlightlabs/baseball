# Baseball API deployment guide

Deploy the Baseball API to a VPS using Coolify and Docker Compose.

## Architecture

- **Coolify (Traefik)**: Reverse proxy, HTTPS, load balancing
- **App**: Baseball API Go application
- **Postgres**: PostgreSQL 17.5 with persistent storage
- **Redis**: Redis 7.4 cache with AOF persistence

The repo ships two Compose files in `conf/`:

- `docker-compose.dev.yml` -- includes Caddy reverse proxy, hardcoded credentials, Postgres port exposed to host
- `docker-compose.prod.yml` -- no Caddy (Coolify's Traefik handles routing/TLS), secrets via `${VAR}` interpolation, production-sized Postgres tuning

Production Postgres tuning is applied via `command: ["postgres", "-c", ...]` entries in Compose, not `POSTGRES_*` environment variables.

## Prerequisites

- VPS with Coolify installed, 4 GB+ RAM, 10 GB+ disk
- Domain pointed to the VPS
- Go 1.24+ (local development only)

## Deploying with Coolify

1. Create a project in the Coolify dashboard.
2. Add a **Docker Compose** resource and connect the Git repo.
3. Set the Compose file path to `conf/docker-compose.prod.yml`.
4. Designate **app** as the public service on port **8080**.
5. Assign your domain -- Coolify provisions TLS automatically.
6. Set the health check path to `/v1/health`.
7. In app Git settings, enabling **Git Submodules** is optional. Production ETL can auto-clone snapshot data at runtime.
8. Configure environment variables in Coolify's UI:

| Variable                   | Example                                                         | Notes                             |
| -------------------------- | --------------------------------------------------------------- | --------------------------------- |
| `DATABASE_URL`             | `postgres://postgres:pw@postgres:5432/baseball?sslmode=disable` | Host is the Compose service name  |
| `SERVER_HOST`              | `0.0.0.0`                                                       | Bind API server to all interfaces |
| `REDIS_URL`                | `redis://redis:6379/0`                                          |                                   |
| `BASEBALL_DATA_ROOT`       | `/home/app/tools/data`                                          | Default data root checked by ETL  |
| `BASEBALL_DATA_REPO_URL`   | `https://github.com/stormlightlabs/bigflydata.git`              | Optional snapshot clone override  |
| `BASEBALL_DATA_REPO_REF`   | `main`                                                          | Optional branch/tag/SHA override  |
| `BASEBALL_DATA_AUTO_CLONE` | `true`                                                          | Set `false` to disable auto clone |
| `POSTGRES_PASSWORD`        | _(strong password)_                                             | Mark as sensitive                 |
| `POSTGRES_DB`              | `baseball`                                                      |                                   |
| `DB_MAX_OPEN_CONNS`        | `20`                                                            | App DB pool hard cap              |
| `DB_MAX_IDLE_CONNS`        | `10`                                                            | App DB pool idle cap              |
| `DB_CONN_MAX_LIFETIME`     | `30m`                                                           | App DB connection lifetime        |
| `DB_CONN_MAX_IDLE_TIME`    | `5m`                                                            | App DB idle connection timeout    |
| `GOMEMLIMIT`               | `900MiB`                                                        | Go runtime heap soft limit        |
| `GOMAXPROCS`               | `2`                                                             | Go runtime CPU concurrency cap    |

Optional: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `CODEBERG_CLIENT_ID`, `CODEBERG_CLIENT_SECRET`, `CACHE_ENABLED`, `CACHE_VERSION`.

1. Click **Deploy**.

## Data preparation and loading

Use one canonical complete-slice runbook for both local and Docker workflows:
[data-loading.md](../docs/internal/data-loading.md).

Quick Docker/Coolify example for a representative complete slice (`2022-2025`) using the streamlined ETL entrypoint:

```bash
docker compose exec app baseball db migrate
docker compose exec app baseball etl run --profile dev --years 2022-2025
docker compose exec app baseball etl validate --profile dev --years 2022-2025
docker compose exec app baseball etl status
```

If required files are missing under `/home/app/tools/data`, the ETL pipeline
automatically clones `bigflydata` into a temporary directory, uses it for the
run, then cleans it up.

If your snapshot data is mounted/cloned outside the default root:

```bash
docker compose exec app baseball etl run --profile dev --years 2022-2025 --data-root /path/to/baseball-data
docker compose exec app baseball etl validate --profile dev --years 2022-2025 --data-root /path/to/baseball-data
```

First-time full historical setup (production profile):

```bash
docker compose exec app baseball db migrate
docker compose exec app baseball etl run --profile prod --mode full
docker compose exec app baseball etl validate --profile prod
docker compose exec app baseball etl status
```

Readiness validation:

- `GET /v1/ready` for probe-style readiness status.
- `GET /v1/meta/datasets` for per-dataset health details.
- `GET /v1/health` for process liveness.
- `GET /v1/meta?strict=true` (or `/v1/meta/datasets?strict=true`) for exact row counts when auditing.
  Default mode is lightweight and each response reports `X-Count-Mode: lightweight|strict|fallback`.

## Updating

Push to the Git repo and click **Deploy** in Coolify (or enable auto-deploy). After deployment, run any new migrations:

```bash
docker compose exec app baseball db migrate
```

For new seasons:

```bash
docker compose exec app baseball etl run --profile prod --years 2026
docker compose exec app baseball etl validate --profile prod --years 2026
docker compose exec app baseball etl status
```

Manual temp-clone pattern (optional when auto-clone is enabled):

```bash
tmpdir="$(mktemp -d)"
git clone --depth=1 <baseball-data-repo-url> "$tmpdir/baseball-data"
docker compose exec app baseball etl run --profile prod --years 2026 --data-root "$tmpdir/baseball-data"
docker compose exec app baseball etl validate --profile prod --years 2026 --data-root "$tmpdir/baseball-data"
rm -rf "$tmpdir"
```

## Scaling

Set `deploy.replicas` in the Compose file or adjust in Coolify's UI.
Traefik load-balances across replicas automatically. Tune Postgres based on available RAM:

| Postgres `-c` option            | Baseline value |
| ------------------------------- | -------------- |
| `shared_buffers`                | `1GB`          |
| `effective_cache_size`          | `2GB`          |
| `work_mem`                      | `32MB`         |
| `maintenance_work_mem`          | `512MB`        |
| `max_wal_size`                  | `12GB`         |
| `min_wal_size`                  | `2GB`          |
| `checkpoint_timeout`            | `15min`        |
| `checkpoint_completion_target`  | `0.9`          |
| `wal_compression`               | `on`           |

ETL window checks:

- If logs show `checkpoints are occurring too frequently`, increase `max_wal_size`.
- Inspect `checkpoints_req` vs `checkpoints_timed` in `pg_stat_bgwriter`.
- Monitor checkpoint interval and WAL churn during force/year rewrites.

4 GB VM crash-prevention baseline:

- Keep compose resource limits enabled for `app`, `postgres`, and `redis`.
- Keep app DB pool bounded (`DB_MAX_OPEN_CONNS=20`, `DB_MAX_IDLE_CONNS=10`) unless load tests justify raising.
- Keep Postgres parallelism conservative during ETL (`max_parallel_workers_per_gather=1`).

## Backup and recovery

```bash
# Backup
docker compose exec -T postgres pg_dump -U postgres baseball > backup.sql

# Restore
docker compose exec -T postgres psql -U postgres baseball < backup.sql
```

Always back up before running migrations in production.

## Rollback

Coolify automatically rolls back if the health check fails after deployment.
For manual rollback, update the image tag in your Compose file and redeploy.

For database rollback, restore from a backup or manually revert the migration SQL.
