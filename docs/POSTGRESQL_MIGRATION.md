# PostgreSQL Migration Guide

**Version:** 0.8.0-beta
**Date:** 2025-11-22
**Driver Change:** lib/pq → pgx/v5

## Overview

ActaLog v0.8.0-beta migrates from the maintenance-mode `lib/pq` PostgreSQL driver to the actively-developed `pgx/v5` driver. This migration provides better performance, active support, and improved PostgreSQL-specific features.

## What Changed

### Dependencies
- **Removed:** `github.com/lib/pq v1.10.9`
- **Added:** `github.com/jackc/pgx/v5 v5.7.6`

### Configuration
New environment variables added for PostgreSQL schema and connection pooling:

```bash
# PostgreSQL-specific schema (default: public)
DB_SCHEMA=actalog

# Connection pooling (PostgreSQL and MySQL only)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

### Code Changes
1. **Driver Name:** Changed from `"postgres"` to `"pgx"` when opening connections
2. **DSN Format:** Changed from `host=x port=y` to `postgres://user:pass@host:port/db?params`
3. **Schema Support:** Added `search_path` parameter for PostgreSQL schema isolation
4. **Connection Pooling:** Configured `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`
5. **SQL Placeholders:** PostgreSQL now uses `$1, $2, $3` instead of `?`
6. **Boolean Values:** Uses `TRUE/FALSE` instead of `1/0` in SQL
7. **Timestamps:** Uses `CURRENT_TIMESTAMP` instead of `datetime('now')`
8. **LastInsertId:** PostgreSQL uses `RETURNING id` clause instead of `LastInsertId()`

## Migration Steps

### For Existing SQLite Users

No changes required! SQLite continues to work exactly as before. This migration only affects PostgreSQL users.

### For Existing PostgreSQL Users (lib/pq)

If you're currently running ActaLog with PostgreSQL using the old `lib/pq` driver:

1. **Backup your database:**
   ```bash
   pg_dump -h localhost -U actalog -d actalog > backup.sql
   ```

2. **Update your .env file:**
   ```bash
   # Add new configuration
   DB_SCHEMA=public                  # or your custom schema
   DB_MAX_OPEN_CONNS=25
   DB_MAX_IDLE_CONNS=5
   DB_CONN_MAX_LIFETIME=5m
   ```

3. **Update ActaLog:**
   ```bash
   git pull
   make build
   ```

4. **Restart the application:**
   ```bash
   make run
   ```

5. **Verify connection:**
   ```bash
   go run cmd/check-schema/main.go
   ```

### For New PostgreSQL Users

1. **Install PostgreSQL** (if not already installed):
   ```bash
   # Ubuntu/Debian
   sudo apt install postgresql postgresql-contrib

   # macOS
   brew install postgresql
   ```

2. **Create database and user:**
   ```sql
   sudo -u postgres psql
   CREATE DATABASE actalog;
   CREATE USER actalog WITH ENCRYPTED PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE actalog TO actalog;
   \q
   ```

3. **Configure .env:**
   ```bash
   DB_DRIVER=postgres
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=actalog
   DB_PASSWORD=your_password
   DB_NAME=actalog
   DB_SCHEMA=public
   DB_SSLMODE=disable                # or 'require' for production

   # Connection pooling
   DB_MAX_OPEN_CONNS=25
   DB_MAX_IDLE_CONNS=5
   DB_CONN_MAX_LIFETIME=5m
   ```

4. **Build and run:**
   ```bash
   make build
   make run
   ```

## Schema Isolation

PostgreSQL schemas provide namespace isolation. To use a custom schema:

1. **Create the schema:**
   ```sql
   CREATE SCHEMA actalog;
   GRANT ALL ON SCHEMA actalog TO actalog;
   ALTER USER actalog SET search_path TO actalog;
   ```

2. **Update .env:**
   ```bash
   DB_SCHEMA=actalog
   ```

## Connection Pooling

The new driver properly configures connection pooling for PostgreSQL:

- **MaxOpenConns (25):** Maximum simultaneous database connections
- **MaxIdleConns (5):** Keep 5 idle connections ready for reuse
- **ConnMaxLifetime (5m):** Recycle connections every 5 minutes

Adjust these values based on your workload:
- **High traffic:** Increase MaxOpenConns to 50-100
- **Low traffic:** Decrease to 10-15 to conserve resources
- **Long-running queries:** Increase ConnMaxLifetime to 15m-30m

## Testing

### Test SQLite (backward compatibility):
```bash
DB_DRIVER=sqlite3 DB_NAME=test.db go run cmd/check-schema/main.go
```

### Test PostgreSQL:
```bash
DB_DRIVER=postgres \
DB_HOST=localhost \
DB_PORT=5432 \
DB_NAME=actalog \
DB_USER=actalog \
DB_PASSWORD=your_password \
DB_SCHEMA=public \
DB_SSLMODE=disable \
go run cmd/check-schema/main.go
```

### Test MariaDB/MySQL:
```bash
DB_DRIVER=mysql \
DB_HOST=localhost \
DB_PORT=3306 \
DB_NAME=actalog \
DB_USER=actalog \
DB_PASSWORD=your_password \
go run cmd/check-schema/main.go
```

Expected output (all databases):
```
=== Database Tables ===
  - audit_logs
  - movements
  - refresh_tokens
  - schema_migrations
  - user_settings
  - user_workout_movements
  - user_workout_wods
  - user_workouts
  - users
  - wods
  - workout_movements
  - workout_wods
  - workouts

=== Applied Migrations ===
  ✓ 0.5.0 - Baseline schema (2025-11-19)
  ✓ 0.5.1 - Add missing tables
```

## Troubleshooting

### Connection refused
**Error:** `connection refused`

**Solution:** Check PostgreSQL is running:
```bash
# Ubuntu/Debian
sudo systemctl status postgresql

# macOS
brew services list
```

### Authentication failed
**Error:** `password authentication failed`

**Solution:** Verify credentials in .env match PostgreSQL user:
```sql
ALTER USER actalog WITH PASSWORD 'new_password';
```

### Schema not found
**Error:** `schema "actalog" does not exist`

**Solution:** Create the schema:
```sql
CREATE SCHEMA actalog;
GRANT ALL ON SCHEMA actalog TO actalog;
```

### Connection pool exhausted
**Error:** `pq: sorry, too many clients already`

**Solution:** Increase PostgreSQL max_connections or decrease DB_MAX_OPEN_CONNS:
```sql
-- Check current limit
SHOW max_connections;

-- Increase (requires restart)
ALTER SYSTEM SET max_connections = 100;
```

## Test Results

All three supported databases have been tested and verified working with v0.8.0-beta:

### ✅ SQLite (sqlite3)
- **Status:** Fully compatible (backward compatible)
- **Tables:** 14 tables created successfully
- **Migrations:** Both baseline migrations applied
- **Seeding:** All standard data seeded (movements, WODs, templates)
- **Use Case:** Development, single-user deployments, embedded systems

### ✅ PostgreSQL (pgx/v5)
- **Status:** Fully compatible with new driver
- **Host:** 192.168.1.143
- **Schema:** actalog
- **Tables:** 13 tables created successfully
- **Migrations:** Both baseline migrations applied
- **Seeding:** All standard data seeded
- **Use Case:** Production multi-user deployments, high concurrency

### ✅ MariaDB/MySQL (mysql)
- **Status:** Fully compatible
- **Host:** 192.168.1.234
- **Tables:** 13 tables created successfully
- **Migrations:** Both baseline migrations applied
- **Seeding:** All standard data seeded
- **Use Case:** Production deployments, shared hosting environments

## Performance Improvements

Benefits of pgx/v5 over lib/pq:

1. **Better Performance:** 10-30% faster for most workloads
2. **Active Development:** Regular updates and security patches
3. **Native PostgreSQL Features:** Better support for LISTEN/NOTIFY, COPY, etc.
4. **Binary Protocol:** More efficient data transfer
5. **Context Support:** Better cancellation and timeout handling

## Rollback Instructions

If you need to rollback to lib/pq:

1. **Checkout previous version:**
   ```bash
   git checkout v0.7.6-beta
   ```

2. **Rebuild:**
   ```bash
   make build
   ```

3. **No database changes needed** - schemas are identical

## Additional Resources

- [pgx Documentation](https://github.com/jackc/pgx)
- [PostgreSQL Connection Pooling](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [ActaLog Database Schema](./DATABASE_SCHEMA.md)
- [Configuration Reference](../.env.example)

## Support

If you encounter issues:
1. Check logs: `journalctl -u actalog -f` (if running as service)
2. Test connection: `go run cmd/check-schema/main.go`
3. Report issues: https://github.com/johnzastrow/actalog/issues
