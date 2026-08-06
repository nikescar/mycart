# Docker Deployment Guide

This guide explains how to deploy myCart using Docker Compose with either SQLite or PostgreSQL.

## Quick Start

### SQLite (Default)

Simple single-container deployment with file-based database:

```bash
docker-compose up -d
```

Access myCart at http://localhost:8080

### PostgreSQL

Multi-container deployment with PostgreSQL database:

```bash
# Set a secure password
export POSTGRES_PASSWORD="your_secure_password_here"

# Start services
docker-compose -f docker-compose.postgres.yml up -d
```

Access myCart at http://localhost:8080

## Configuration

### SQLite Configuration

The default `docker-compose.yml` uses SQLite with the following settings:

- **Database file**: `./lc_base/mycart.db`
- **Uploads**: `./lc_uploads`
- **Digital products**: `./lc_digitals`

All data is persisted in local directories.

### PostgreSQL Configuration

The `docker-compose.postgres.yml` includes:

- **PostgreSQL 16 Alpine** (lightweight)
- **Persistent volume** for database data
- **Health checks** for both services
- **Isolated network** for security

Environment variables:
```bash
DB_TYPE=postgresql
DB_HOST=postgres
DB_PORT=5432
DB_NAME=mycart
DB_USER=mycart
DB_PASSWORD=${POSTGRES_PASSWORD}
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300
```

## Environment Variables

Create a `.env` file for custom configuration:

```bash
# PostgreSQL password
POSTGRES_PASSWORD=your_secure_password

# HTTP port
HTTP_PORT=8080

# Application settings
ADMIN_EMAIL=admin@example.com
DOMAIN=shop.example.com
```

## Data Migration

### SQLite to PostgreSQL

```bash
# 1. Stop the SQLite container
docker-compose down

# 2. Start PostgreSQL
docker-compose -f docker-compose.postgres.yml up -d postgres

# 3. Run migration
docker exec -it mycart ./mycart migrate-to-postgres \
  --database-url="postgresql://mycart:${POSTGRES_PASSWORD}@postgres:5432/mycart?sslmode=disable"

# 4. Restart with PostgreSQL
docker-compose -f docker-compose.postgres.yml up -d
```

### PostgreSQL to SQLite

```bash
# 1. Stop PostgreSQL containers
docker-compose -f docker-compose.postgres.yml down

# 2. Run migration
docker run --rm \
  -v $(pwd)/lc_base:/data \
  -e DATABASE_URL="postgresql://mycart:password@postgres:5432/mycart" \
  mycart:latest ./mycart migrate-to-sqlite --path=/data/mycart.db

# 3. Start SQLite
docker-compose up -d
```

## Backup and Restore

### SQLite Backup

```bash
# Backup
docker exec mycart cp /data/mycart.db /data/mycart.db.backup

# Or copy to host
docker cp mycart:/data/mycart.db ./backup/mycart-$(date +%Y%m%d).db
```

### PostgreSQL Backup

```bash
# Backup
docker exec mycart-postgres pg_dump -U mycart mycart > backup/mycart-$(date +%Y%m%d).sql

# Restore
docker exec -i mycart-postgres psql -U mycart mycart < backup/mycart-20260807.sql
```

## Production Deployment

For production, consider:

1. **Use secrets** for passwords (Docker secrets or environment files)
2. **Enable SSL** for PostgreSQL connections
3. **Configure reverse proxy** (nginx, Traefik, Caddy)
4. **Set up backups** (automated daily backups)
5. **Monitor health** endpoints
6. **Use volumes** for persistent data

### Production PostgreSQL Example

```yaml
services:
  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
    secrets:
      - db_password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  mycart:
    image: mycart:latest
    restart: always
    environment:
      DB_TYPE: postgresql
      DB_SSLMODE: require
    secrets:
      - db_password

secrets:
  db_password:
    file: ./secrets/db_password.txt
```

## Troubleshooting

### Container won't start

Check logs:
```bash
docker-compose logs mycart
```

### Database connection errors

Verify PostgreSQL is running:
```bash
docker-compose -f docker-compose.postgres.yml ps
docker exec mycart-postgres pg_isready -U mycart
```

### Migration issues

Check migration status:
```bash
docker exec mycart ./mycart migrate
```

### Reset database

SQLite:
```bash
docker-compose down
rm -rf lc_base/mycart.db
docker-compose up -d
```

PostgreSQL:
```bash
docker-compose -f docker-compose.postgres.yml down -v
docker-compose -f docker-compose.postgres.yml up -d
```

## Health Checks

Both configurations include health checks:

- **Application**: `http://localhost:8080/api/health`
- **PostgreSQL**: `pg_isready` command

Check health status:
```bash
docker ps
```

Healthy services show `healthy` in the STATUS column.

## Scaling

For high-traffic deployments:

1. **Increase PostgreSQL connection pool**:
   ```yaml
   environment:
     - DB_MAX_OPEN_CONNS=100
     - DB_MAX_IDLE_CONNS=25
   ```

2. **Run multiple app instances** behind a load balancer

3. **Use managed PostgreSQL** (AWS RDS, Google Cloud SQL, etc.)

4. **Enable PostgreSQL replication** for read scaling

## Support

For issues or questions:
- GitHub Issues: https://github.com/shurco/mycart/issues
- Documentation: See README.md
