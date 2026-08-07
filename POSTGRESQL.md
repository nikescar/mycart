# PostgreSQL Support for myCart

myCart now supports both SQLite and PostgreSQL databases, giving you flexibility in choosing the right database for your deployment.

## Overview

- **SQLite** (default): Simple, file-based, perfect for small to medium deployments
- **PostgreSQL**: Powerful, scalable, ideal for production and high-traffic sites

## Quick Start

### Using PostgreSQL

#### Option 1: Connection String (Recommended)

```bash
DATABASE_URL="postgresql://user:password@localhost:5432/mycart?sslmode=require" ./mycart serve
```

#### Option 2: Individual Parameters

```bash
DB_TYPE=postgresql \
DB_HOST=localhost \
DB_PORT=5432 \
DB_NAME=mycart \
DB_USER=mycart_user \
DB_PASSWORD=secure_password \
DB_SSLMODE=require \
./mycart serve
```

#### Option 3: Environment File

Create `.env`:
```bash
DB_TYPE=postgresql
DATABASE_URL=postgresql://user:password@localhost:5432/mycart?sslmode=require
```

Then run:
```bash
./mycart serve
```

### Installation with PostgreSQL

```bash
./mycart install \
  --email=admin@example.com \
  --password=secure_password \
  --domain=shop.example.com \
  --db-type=postgresql \
  --database-url="postgresql://user:pass@localhost:5432/mycart"
```

## Configuration

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DB_TYPE` | Database type | `sqlite` | `postgresql` |
| `DATABASE_URL` | Full connection string | - | `postgresql://user:pass@host:5432/db` |
| `DB_HOST` | Database host | `localhost` | `db.example.com` |
| `DB_PORT` | Database port | `5432` | `5432` |
| `DB_NAME` | Database name | `mycart` | `mycart_prod` |
| `DB_USER` | Database user | `postgres` | `mycart_user` |
| `DB_PASSWORD` | Database password | - | `secure_password` |
| `DB_SSLMODE` | SSL mode | `require` | `disable`, `require`, `verify-ca`, `verify-full` |
| `DB_SCHEMA` | PostgreSQL schema | `public` | `public` |
| `DB_TIMEZONE` | Database timezone | `UTC` | `America/New_York` |
| `DB_CONNECT_TIMEOUT` | Connection timeout (seconds) | `10` | `30` |
| `DB_MAX_OPEN_CONNS` | Max open connections | `25` | `100` |
| `DB_MAX_IDLE_CONNS` | Max idle connections | `5` | `10` |
| `DB_CONN_MAX_LIFETIME` | Connection max lifetime (seconds) | `300` | `600` |

### SSL Modes

- **`disable`**: No SSL (development only)
- **`require`**: Use SSL, don't verify certificate
- **`verify-ca`**: Use SSL, verify certificate
- **`verify-full`**: Use SSL, verify certificate and hostname

For production, use `verify-full` with proper certificates.

## Database Migration

### SQLite to PostgreSQL

Migrate your existing SQLite data to PostgreSQL:

```bash
# 1. Ensure PostgreSQL is running and accessible
createdb mycart

# 2. Run migration command
./mycart migrate-to-postgres \
  --database-url="postgresql://user:password@localhost:5432/mycart?sslmode=disable"

# 3. Update environment to use PostgreSQL
export DB_TYPE=postgresql
export DATABASE_URL="postgresql://user:password@localhost:5432/mycart"

# 4. Restart application
./mycart serve
```

### PostgreSQL to SQLite

Migrate from PostgreSQL back to SQLite:

```bash
# 1. Set up PostgreSQL connection in environment
export DATABASE_URL="postgresql://user:password@localhost:5432/mycart"

# 2. Run migration
./mycart migrate-to-sqlite --path=./lc_base/data.db

# 3. Update environment to use SQLite
export DB_TYPE=sqlite
export SQLITE_PATH=./lc_base/data.db

# 4. Restart application
./mycart serve
```

## Managed PostgreSQL Services

### Supabase

```bash
DATABASE_URL="postgresql://postgres:your_password@db.xxx.supabase.co:5432/postgres" ./mycart serve
```

### Amazon RDS

```bash
DATABASE_URL="postgresql://username:password@mydb.xxx.rds.amazonaws.com:5432/mycart?sslmode=require" ./mycart serve
```

### Google Cloud SQL

```bash
DATABASE_URL="postgresql://username:password@xxx.xxx.xxx.xxx:5432/mycart?sslmode=require" ./mycart serve
```

### Azure Database for PostgreSQL

```bash
DATABASE_URL="postgresql://username@servername:password@servername.postgres.database.azure.com:5432/mycart?sslmode=require" ./mycart serve
```

### DigitalOcean Managed Databases

```bash
DATABASE_URL="postgresql://doadmin:password@dbname.db.ondigitalocean.com:25060/mycart?sslmode=require" ./mycart serve
```

## Performance Tuning

### Connection Pooling

For high-traffic deployments:

```bash
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=600
```

Guidelines:
- **Max open connections**: `(number of CPU cores) × 2 + effective_spindle_count`
- **Max idle connections**: 10-20% of max open connections
- **Connection lifetime**: 5-10 minutes to allow connection recycling

### PostgreSQL Configuration

Recommended `postgresql.conf` settings for production:

```ini
# Connection Settings
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 2621kB
min_wal_size = 1GB
max_wal_size = 4GB
```

## Backup and Restore

### PostgreSQL Backup

```bash
# Full backup
pg_dump -U mycart_user -h localhost mycart > backup_$(date +%Y%m%d).sql

# Compressed backup
pg_dump -U mycart_user -h localhost mycart | gzip > backup_$(date +%Y%m%d).sql.gz

# Custom format (recommended)
pg_dump -U mycart_user -h localhost -F c mycart > backup_$(date +%Y%m%d).dump
```

### PostgreSQL Restore

```bash
# From SQL dump
psql -U mycart_user -h localhost mycart < backup_20260807.sql

# From compressed dump
gunzip -c backup_20260807.sql.gz | psql -U mycart_user -h localhost mycart

# From custom format
pg_restore -U mycart_user -h localhost -d mycart backup_20260807.dump
```

### Automated Backups

Using cron:

```bash
# Daily backups at 2 AM
0 2 * * * pg_dump -U mycart_user mycart | gzip > /backups/mycart_$(date +\%Y\%m\%d).sql.gz

# Keep only last 7 days
0 3 * * * find /backups -name "mycart_*.sql.gz" -mtime +7 -delete
```

## Monitoring

### Health Checks

The application provides health check endpoints:

```bash
curl http://localhost:8080/api/health
```

### Database Connection Status

Check if the database is accessible:

```bash
# PostgreSQL
psql -U mycart_user -h localhost -c "SELECT version();" mycart

# Using mycart CLI
./mycart migrate
```

### Query Performance

Enable slow query logging in `postgresql.conf`:

```ini
log_min_duration_statement = 1000  # Log queries taking >1s
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d '
log_statement = 'ddl'
```

## Security Best Practices

1. **Use strong passwords**: Generate secure random passwords
2. **Enable SSL**: Always use SSL in production (`sslmode=require` or higher)
3. **Restrict network access**: Use firewall rules to limit database access
4. **Regular updates**: Keep PostgreSQL updated for security patches
5. **Principle of least privilege**: Create dedicated database user with minimal permissions
6. **Connection encryption**: Use SSL certificates for connections
7. **Audit logging**: Enable PostgreSQL audit logging for production

### Creating a Dedicated Database User

```sql
-- Create database
CREATE DATABASE mycart;

-- Create user with secure password
CREATE USER mycart_user WITH ENCRYPTED PASSWORD 'secure_random_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE mycart TO mycart_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO mycart_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO mycart_user;

-- Connect to the database and set default privileges
\c mycart
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO mycart_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO mycart_user;
```

## Troubleshooting

### Connection Refused

```bash
# Check if PostgreSQL is running
systemctl status postgresql

# Check if port is open
netstat -an | grep 5432

# Test connection
psql -U mycart_user -h localhost mycart
```

### SSL Connection Errors

```bash
# Verify SSL is enabled in postgresql.conf
ssl = on

# Check certificate files exist
ls -l /var/lib/postgresql/data/server.crt
ls -l /var/lib/postgresql/data/server.key

# Test with SSL disabled temporarily
DATABASE_URL="postgresql://user:pass@localhost:5432/mycart?sslmode=disable"
```

### Migration Failures

```bash
# Check current migration status
./mycart migrate

# View PostgreSQL logs
tail -f /var/log/postgresql/postgresql-*.log

# Run migrations manually
psql -U mycart_user -h localhost mycart -f db/migrations/postgres/xxx.sql
```

### Performance Issues

```bash
# Check active connections
SELECT count(*) FROM pg_stat_activity;

# Find slow queries
SELECT pid, now() - query_start as duration, query 
FROM pg_stat_activity 
WHERE state = 'active' 
ORDER BY duration DESC;

# Check database size
SELECT pg_size_pretty(pg_database_size('mycart'));

# Vacuum and analyze
VACUUM ANALYZE;
```

## Architecture

### Database Abstraction Layer

```
Application Code
       ↓
queries.Base (backward compatible)
       ↓
database.Database interface
       ↓
PostgresAdapter / SQLiteAdapter
       ↓
sqlc-generated type-safe queries
       ↓
PostgreSQL / SQLite
```

### Connection Management

- **Retry logic**: Exponential backoff (5 attempts, 1-30s delays)
- **Health monitoring**: Background checks every 30 seconds
- **Connection pooling**: Configurable pool sizes
- **Timeout handling**: 5-second ping timeouts

### Migration System

- **Goose**: Used for schema migrations
- **Separate paths**: `db/migrations/sqlite/` and `db/migrations/postgres/`
- **Automatic execution**: Migrations run on startup
- **Version tracking**: Migration versions stored in database

## FAQ

**Q: Can I switch between SQLite and PostgreSQL without data loss?**

A: Yes! Use the migration commands:
- SQLite → PostgreSQL: `./mycart migrate-to-postgres`
- PostgreSQL → SQLite: `./mycart migrate-to-sqlite`

**Q: Which database should I use?**

A: 
- **SQLite**: Small to medium sites (<10,000 products, moderate traffic)
- **PostgreSQL**: Large sites, high traffic, multiple concurrent users

**Q: Does this support PostgreSQL 15/16?**

A: Yes, tested with PostgreSQL 12-16. Recommended: PostgreSQL 14+

**Q: Can I use PostgreSQL on Heroku, Railway, or Render?**

A: Yes! Just set the `DATABASE_URL` environment variable provided by your platform.

**Q: What about database backups?**

A: Both SQLite and PostgreSQL backups are supported. See the Backup section above.

**Q: Is the existing API compatible?**

A: Yes! The database abstraction maintains full backward compatibility with existing code.

## Support

For issues or questions:
- **GitHub Issues**: https://github.com/shurco/mycart/issues
- **Documentation**: See README.md and DOCKER.md
- **Community**: Join discussions in GitHub Discussions

## Credits

PostgreSQL support implemented with:
- [lib/pq](https://github.com/lib/pq): PostgreSQL driver
- [sqlc](https://github.com/sqlc-dev/sqlc): Type-safe SQL code generation
- [goose](https://github.com/pressly/goose): Database migrations
