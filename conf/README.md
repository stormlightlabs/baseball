# Baseball API deployment guide

Deploy the Baseball API to a VPS using Coolify and Docker Compose.

## Architecture

- **Traefik**: Reverse proxy, HTTPS, load balancing
- **App**: Baseball API Go application
- **ETL**: Long-lived `baseball-etl worker` queue consumer container
- **Postgres**: PostgreSQL 17.5 with persistent storage
- **Redis**: Redis 7.4 cache with AOF persistence

The repo ships two Compose files in `conf/`:

- `docker-compose.dev.yml` -- includes Caddy reverse proxy, hardcoded credentials, Postgres port exposed to host
- `docker-compose.prod.yml` -- no Caddy (Traefik handles routing/TLS), secrets via `${VAR}` interpolation, production-sized Postgres tuning

Production Postgres tuning is applied via `command: ["postgres", "-c", ...]` entries in Compose, not `POSTGRES_*` environment variables.

## Prerequisites

- VM with 4 GB+ RAM, 10 GB+ disk
- Domain pointed to the VM
- Go 1.24+ (local development only)

## Deploying

Configure environment variables in the docker deployment environment:

| Variable                              | Example                                                         | Notes                             |
| ------------------------------------- | --------------------------------------------------------------- | --------------------------------- |
| `DATABASE_URL`                        | `postgres://postgres:pw@postgres:5432/baseball?sslmode=disable` | Host is the Compose service name  |
| `SERVER_HOST`                         | `0.0.0.0`                                                       | Bind API server to all interfaces |
| `CORS_ALLOWED_ORIGINS`                | `https://bigfly.tech,https://www.bigfly.tech`                   | Allowlist for browser origins     |
| `RATE_LIMIT_PUBLIC_PER_MINUTE`        | `60`                                                            | Public/open API limit per IP      |
| `REDIS_URL`                           | `redis://redis:6379/0`                                          |                                   |
| `BASEBALL_DATA_ROOT`                  | `/home/app/data`                                                | Shared ETL/API data root volume   |
| `POSTGRES_PASSWORD`                   | _(strong password)_                                             | Mark as sensitive                 |
| `POSTGRES_DB`                         | `baseball`                                                      |                                   |
| `DB_MAX_OPEN_CONNS`                   | `20`                                                            | App DB pool hard cap              |
| `DB_MAX_IDLE_CONNS`                   | `10`                                                            | App DB pool idle cap              |
| `DB_CONN_MAX_LIFETIME`                | `30m`                                                           | App DB connection lifetime        |
| `DB_CONN_MAX_IDLE_TIME`               | `5m`                                                            | App DB idle connection timeout    |
| `GOMEMLIMIT`                          | `900MiB`                                                        | Go runtime heap soft limit        |
| `GOMAXPROCS`                          | `2`                                                             | Go runtime CPU concurrency cap    |

Optional: `ETL_DB_MAX_OPEN_CONNS`, `ETL_DB_MAX_IDLE_CONNS`, `ETL_DB_CONN_MAX_LIFETIME`, `ETL_DB_CONN_MAX_IDLE_TIME`, `ETL_GOMEMLIMIT`, `ETL_GOMAXPROCS`, `ETL_MAX_ACTIVE_JOBS`, `ETL_MAX_QUEUED_JOBS`, `ETL_MV_WORK_MEM`, `ETL_MV_MAINTENANCE_WORK_MEM`, `ETL_MV_MAX_PARALLEL_WORKERS_PER_GATHER`, `ETL_MV_JIT`, `CACHE_ENABLED`, `CACHE_VERSION`.

1. Click **Deploy**.

## Data preparation and loading

Use one canonical complete-slice runbook for both local and Docker workflows:
[data-loading.md](../docs/internal/data-loading.md).

Quick Docker example for a representative complete slice (`2022-2025`) using the streamlined ETL entrypoint:

```bash
docker compose exec app baseball db migrate
docker compose up -d etl
docker compose exec etl baseball-etl run --profile dev --years 2022-2025
docker compose exec etl baseball-etl validate --profile dev --years 2022-2025
docker compose exec etl baseball-etl status
```

`etl` should be kept running as the long-lived queue consumer service/process.
`baseball-etl run` enqueues ETL jobs by default.
`baseball-etl maintenance` processes queue jobs by default (`--enqueue-only=false`); pass `--enqueue-only=true` for enqueue-only behavior.

`etl` is a worker service and should not be exposed publicly (Compose sets `traefik.enable=false`).
API health probes are attached to `app` only; `etl` has healthcheck disabled.

The image now bakes `repo/data` into `/home/app/data`.
With the `data_root:/home/app/data` named volume, Docker initializes a new
empty volume with that baked content on first create.

Retrosheet windows are still worker-fetched by ETL. For explicit fetches:

```bash
docker compose exec etl baseball-etl fetch retrosheet --years 2022-2025
docker compose exec etl baseball-etl fetch negroleagues
```

If `data_root` already exists and is empty/stale, recreate that volume to
re-seed from the latest image snapshot before starting services.

### Stale Data Recovery

If ETL validation fails due to stale or incomplete local source files (for example `retrosheet_players` empty), recover in this order:

1. Targeted Retrosheet refresh + player reload:

   ```bash
   docker compose exec etl baseball-etl fetch retrosheet --force --years 2022-2025
   docker compose exec etl baseball-etl load players
   docker compose exec etl baseball-etl run --profile prod --years 2022-2025
   docker compose exec etl baseball-etl validate --profile prod --years 2022-2025
   ```

2. If broader source state is stale, recreate only the `data_root` volume (do not remove Postgres volume):

   ```bash
   docker compose stop app etl
   docker volume ls --format '{{.Name}}' | grep data_root
   docker volume rm <data_root_volume_name>
   docker compose up -d app etl
   ```

3. Re-run ETL + validation.

Incremental/year-batched ETL now reuses an already-populated `retrosheet_players`
table (instead of truncating/reloading every batch). If the table is empty, the
worker reloads it automatically.

If the queue is blocked by stale running jobs:

```bash
docker compose exec etl baseball-etl jobs ls --status running,started
docker compose exec etl baseball-etl jobs clear --reason "recover stale in-flight jobs after restart"
```

If your data root is mounted outside the default path:

```bash
docker compose exec etl baseball-etl run --profile dev --years 2022-2025 --data-root /path/to/baseball-data
docker compose exec etl baseball-etl validate --profile dev --years 2022-2025 --data-root /path/to/baseball-data
```

First-time full historical setup (production profile):

```bash
docker compose exec app baseball db migrate
docker compose up -d etl
docker compose exec etl baseball-etl run --profile prod --mode full
docker compose exec etl baseball-etl validate --profile prod
docker compose exec etl baseball-etl status
```

Readiness validation:

- `GET /v1/ready` for probe-style readiness status.
- `GET /v1/meta/datasets` for per-dataset health details.
- `GET /v1/health` for process liveness.
- `GET /v1/meta?strict=true` (or `/v1/meta/datasets?strict=true`) for exact row counts when auditing.
  Default mode is lightweight and each response reports `X-Count-Mode: lightweight|strict|fallback`.

## Updating

Push to the Git repo and redeploy. After deployment, run any new migrations:

```bash
docker compose exec app baseball db migrate
```

For new seasons:

```bash
docker compose exec etl baseball-etl run --profile prod --years 2026
docker compose exec etl baseball-etl validate --profile prod --years 2026
docker compose exec etl baseball-etl status
```

VM-safe batched pattern for new seasons or backfills:

```bash
docker compose exec etl baseball-etl fetch retrosheet --years 2022-2023
docker compose exec etl baseball-etl run --profile prod --years 2022-2023
docker compose exec etl baseball-etl validate --profile prod --years 2022-2023

docker compose exec etl baseball-etl fetch retrosheet --years 2024-2025
docker compose exec etl baseball-etl run --profile prod --years 2024-2025
docker compose exec etl baseball-etl validate --profile prod --years 2024-2025
```

## Scaling

Set `deploy.replicas` in the Compose file. Traefik load-balances across replicas
automatically. Tune Postgres based on available RAM:

| Postgres `-c` option           | Baseline value |
| ------------------------------ | -------------- |
| `shared_buffers`               | `1GB`          |
| `effective_cache_size`         | `2GB`          |
| `work_mem`                     | `32MB`         |
| `maintenance_work_mem`         | `128MB`        |
| `max_wal_size`                 | `12GB`         |
| `min_wal_size`                 | `2GB`          |
| `checkpoint_timeout`           | `15min`        |
| `checkpoint_completion_target` | `0.9`          |
| `wal_compression`              | `on`           |

ETL window checks:

- If logs show `checkpoints are occurring too frequently`, increase `max_wal_size`.
- Inspect `checkpoints_req` vs `checkpoints_timed` in `pg_stat_bgwriter`.
- Monitor checkpoint interval and WAL churn during force/year rewrites.
- Run one ETL batch at a time on shared VMs; queue additional year windows serially.

4 GB VM crash-prevention baseline:

- Keep compose resource limits enabled for `app`, `etl`, `postgres`, and `redis`.
- Keep app DB pool bounded (`DB_MAX_OPEN_CONNS=20`, `DB_MAX_IDLE_CONNS=10`) unless load tests justify raising.
- Keep ETL DB pool bounded (`ETL_DB_MAX_OPEN_CONNS=12`, `ETL_DB_MAX_IDLE_CONNS=6`) unless load tests justify raising.
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

For manual rollback, update the image tag in your Compose file and redeploy.

For database rollback, restore from a backup or manually revert the migration SQL.
