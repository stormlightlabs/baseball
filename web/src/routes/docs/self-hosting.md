# Self-Hosting

This guide covers how to run Big Fly on your own infrastructure using the same operational
flow used in our production environment.

## Deployment Model

Run Big Fly as four services:

- API service: serves `/v1/*` and `/explorer`
- ETL worker: fetches/loads data and runs maintenance jobs
- PostgreSQL: primary data store
- Redis: ETL queue + cache/rate-limiting support

Recommended topology:

- `app` and `etl` run as separate processes/containers
- Postgres and Redis run as managed services or dedicated containers
- A reverse proxy/ingress terminates TLS and forwards to the API service

## Minimum Runtime Configuration

Set these before starting services:

- `DATABASE_URL`: Postgres connection string
- `REDIS_URL`: Redis connection string
- `BASEBALL_DATA_ROOT`: mounted directory for source inputs/artifacts
- `SERVER_HOST`: API bind host (commonly `0.0.0.0` in containers)
- `SERVER_BASE_URL`: canonical external API base URL
- `CORS_ALLOWED_ORIGINS`: allowed web origins

Useful ETL guardrails for shared/smaller hosts:

- `ETL_MAX_ACTIVE_JOBS=1`
- `ETL_MAX_QUEUED_JOBS=16`

## Option A: Docker Compose (Recommended)

The repository includes compose definitions under `conf/`.

1. Provide required env vars (for example `POSTGRES_PASSWORD`, CORS, base URL).
2. Start the stack:

    ```bash
    docker compose -f conf/docker-compose.prod.yml up -d postgres redis app etl
    ```

3. Apply migrations:

    ```bash
    docker compose -f conf/docker-compose.prod.yml exec app baseball db migrate
    ```

4. Keep ETL worker running continuously (already configured via `etl` service command).

## Option B: Direct Process Deployment

1. Build binaries:

    ```bash
    task build
    ```

2. Start API + worker separately:

    ```bash
    ./tmp/baseball server start --config conf.toml
    ./tmp/baseball-etl worker
    ```

3. In another shell/session, run migration and ETL commands as needed.

## Initial Data Load (Safe Rollout)

Start with a narrow year window, validate, then expand.

```bash
baseball db migrate
baseball-etl fetch retrosheet --years 2024-2025
baseball-etl fetch negroleagues
baseball-etl run --profile prod --years 2024-2025 --year-batch-size 1
baseball-etl maintenance --profile prod --years 2024-2025 --mv-refresh-mode auto
baseball-etl validate --profile prod --years 2024-2025
baseball-etl status
```

After the first window succeeds, process older years in batches.

## Health and Readiness Checks

Use these probes for deployment health checks and dashboards:

```bash
curl http://localhost:8080/v1/health
curl http://localhost:8080/v1/ready
curl http://localhost:8080/v1/meta/datasets
curl http://localhost:8080/v1/meta/datasets?strict=true
```

For endpoint details, see [Meta & Utility](/docs/api-meta-utility).

## Day-2 Operations

Queue visibility and recovery commands:

```bash
baseball-etl jobs ls --status queued,running,retry_wait --limit 100
baseball-etl jobs clear --reason "recover stale running jobs"
```

Keep ETL scopes bounded on shared hosts:

- run one ETL window at a time
- use explicit `--years` windows instead of unbounded full-history jobs
- run `maintenance` after each window, then `validate`

## Upgrades

For routine upgrades:

1. Deploy updated `baseball` and `baseball-etl` binaries/images.
2. Run `baseball db migrate`.
3. Restart API and ETL services.
4. Run a scoped `maintenance` and `validate` pass.
5. Confirm readiness probes are green.

## Backup and Restore

Back up both:

- PostgreSQL data
- `BASEBALL_DATA_ROOT` (source/input artifacts)

After restore:

1. start API + ETL worker
2. run `baseball-etl validate --profile prod`
3. run `baseball-etl status`
4. re-run scoped ETL windows if dataset readiness is incomplete

## Troubleshooting

If jobs stay queued:

- confirm the ETL worker is running
- inspect queue state with `jobs ls`
- clear stale jobs only when you have confirmed they are orphaned

If readiness fails after a deployment:

- run `baseball-etl maintenance --profile prod --mv-refresh-mode auto`
- run `baseball-etl validate --profile prod`
- re-check `/v1/ready` and `/v1/meta/datasets?strict=true`
