# Baseball API Deployment Guide

This document outlines the deployment process for the Baseball API, including building Docker images, managing the database, and deploying to production.

## Architecture

The production deployment uses Docker Compose with the following services:

- **Caddy**: Reverse proxy, load balancer, and automatic HTTPS
- **App**: Baseball API Go application (scalable)
- **Postgres**: PostgreSQL 17.5 database with persistent storage
- **Redis**: Redis 7.4 cache with persistence

## Prerequisites

### Local Development

- Go 1.24+
- PostgreSQL 17.5
- Redis 7.4
- Docker and Docker Compose (for deployment)

### Production Server

- Docker and Docker Compose installed
- Minimum 4GB RAM (8GB+ recommended for full dataset); 10GB+ disk space for database
- Domain configured and VPS provisioned

## Database

The Baseball API database grows with historical data:

- **Current size**: ~3GB with partial Retrosheet data
- **Full dataset**: 5GB+ with complete play-by-play data (1914-2025)
- **Largest tables**: Partitioned plays tables (200+ MB per recent season)

## Deployment

### Authenticate with DockerHub

Before pushing images, authenticate with DockerHub:

```bash
# Interactive login (recommended for local development)
docker login -u yourusername

# Non-interactive login (for CI/CD)
echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin

# Verify authentication
docker info | grep Username
```

**Authentication Methods:**

- **Local Development**: Use `docker login` - credentials are stored securely in your OS keychain
- **CI/CD**: Use environment variables `DOCKER_USERNAME` and `DOCKER_PASSWORD` or registry tokens
- **GitHub Actions**: Use `docker/login-action@v2`
- **Access Tokens**: Create a token at <https://hub.docker.com/settings/security> and use it as the password

### Build and Push Docker Image

Use the `baseball deploy` CLI command:

```bash
# Dry run - see what would happen without executing
baseball deploy --tag v1.0.0 --registry yourusername --push --dry-run

# Build, tag, and push to DockerHub
baseball deploy --tag v1.0.0 --registry yourusername --push

# Build only (for testing)
baseball deploy --tag v1.0.0

# Push existing image without rebuilding
baseball deploy --tag v1.0.0 --registry yourusername --push --skip-build
```

The deploy command will:

- Check Docker authentication status
- Provide helpful error messages if not authenticated
- Build and tag images with the specified version
- Push both versioned and `latest` tags to your registry

### Prepare Data

The Baseball API requires two data sources:

#### Lahman Database

```bash
# Download from SABR (manual download required)
# Visit: https://sabr.org/lahman-database/
# Extract to: data/lahman/csv/

baseball etl fetch lahman  # Creates directories and shows instructions
```

#### Retrosheet Data

```bash
# Download Retrosheet event files
baseball etl fetch retrosheet --years 2020-2025

# Download Negro Leagues data
baseball etl fetch negroleagues
```

### Transfer to Production Server

Transfer the following to your production server:

```bash
# Transfer codebase and config
scp -r docker-compose.yml Caddyfile user@server:/opt/baseball/

# Transfer data files (if loading data on server)
rsync -avz data/ user@server:/opt/baseball/data/
```

### 5. Configure Environment

On the production server, create a `.env` file or set environment variables:

```bash
# .env file
DATABASE_URL=postgres://user:password@postgres:5432/baseball_prod?sslmode=disable
REDIS_URL=redis://redis:6379/0

# For managed database:
# DATABASE_URL=postgres://user:password@your-managed-db.provider.com:5432/baseball_prod?sslmode=require
```

### Deploy Services

```bash
cd /opt/baseball

# Pull latest images
docker-compose pull

# Start services
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f app
```

### Run Database Migrations

```bash
# Run migrations
docker-compose exec app baseball db migrate

# Verify migrations
docker-compose exec app baseball db migrate  # Should show "no pending migrations"
```

### Load Data (Initial Deployment Only)

The deploy command supports idempotent data loading:

```bash
# If data files are on the server, run ETL inside container:
docker-compose exec app baseball db populate --years 2020-2025

# Or use the deploy command with --skip-build:
baseball deploy --tag v1.0.0 --skip-build --skip-etl=false --years 2020-2025
```

**Idempotency**: The ETL pipeline is idempotent:

- Lahman data: Checks `dataset_refreshes` table and skips if already loaded
- Retrosheet data: Checks per-year and per-dataset, only loads missing data

Re-running data loads is safe and will not create duplicates.

### Verify

```bash
# Check API health
curl https://baseball.stormlightlabs.org/v1/health

# Check a data endpoint
curl https://baseball.stormlightlabs.org/v1/players?limit=5

# Monitor logs
docker-compose logs -f app
docker-compose logs -f caddy
```

## Scaling

### Horizontal

Scale the application to multiple instances for load balancing:

```bash
# Scale to 3 instances
docker-compose up -d --scale app=3

# Caddy will automatically load balance across all instances
```

### Vertical

Adjust resource limits in `docker-compose.yml`:

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          memory: 1G
```

## Updating Deployments

### Application Updates

```bash
# Build new version
baseball deploy --tag v1.1.0 --registry yourusername --push

# On server: pull and restart
docker-compose pull
docker-compose up -d

# Run any new migrations
docker-compose exec app baseball db migrate
```

### Data Updates

```bash
# Fetch new data
baseball etl fetch retrosheet --years 2026

# Load incrementally (idempotent)
docker-compose exec app baseball db populate retrosheet --years 2026

# Refresh materialized views
docker-compose exec app baseball db refresh-views
```

## Backup and Recovery

### Database Backups

**Docker Postgres:**

```bash
# Backup
docker-compose exec -T postgres pg_dump -U postgres baseball_prod > backup.sql

# Restore
docker-compose exec -T postgres psql -U postgres baseball_prod < backup.sql
```

**Managed Database:**
Use the provider's backup tools and follow their recovery procedures.

### Volume Backups

```bash
# Backup Postgres data volume
docker run --rm \
  -v baseball_postgres_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/postgres-backup.tar.gz -C /data .

# Restore
docker run --rm \
  -v baseball_postgres_data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/postgres-backup.tar.gz -C /data
```

## Monitoring

### Logs

```bash
# Application logs
docker-compose logs -f app

# Database logs
docker-compose logs -f postgres

# Caddy logs
docker-compose logs -f caddy

# All services
docker-compose logs -f
```

### Health Checks

```bash
# API health endpoint
curl https://baseball.stormlightlabs.org/v1/health

# Container health
docker-compose ps

# Database connection
docker-compose exec app baseball db migrate  # Should connect successfully
```

## Database Tuning

Adjust Postgres settings in docker-compose.yml based on available RAM:

```yaml
environment:
  POSTGRES_SHARED_BUFFERS: 512MB      # 25% of RAM
  POSTGRES_EFFECTIVE_CACHE_SIZE: 2GB  # 50-75% of RAM
  POSTGRES_WORK_MEM: 32MB
```

## Redis Configuration

```yaml
command: ["redis-server", "--appendonly", "yes", "--maxmemory", "1gb", "--maxmemory-policy", "allkeys-lru"]
```

## Rollback

### Application Rollback

```bash
# On server: revert to previous image
docker-compose pull yourusername/baseball-app:v1.0.0
docker tag yourusername/baseball-app:v1.0.0 baseball-app:latest
docker-compose up -d
```

### Database Rollback

If migrations need to be rolled back:

1. Restore from backup
2. Or manually revert migration SQL

**Important**: Always backup before running migrations in production.
