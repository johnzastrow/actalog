# Database Deployment Guide

ActaLog supports three database backends with Docker deployment:
- **SQLite** - Single-file database (default, best for single-server)
- **PostgreSQL** - Production-ready SQL database (recommended for multi-instance)
- **MariaDB/MySQL** - Alternative production SQL database

## Quick Start by Database

### SQLite (Default - Simplest)

**Best for:** Development, single-server production, small-scale deployments

```bash
cd docker

# Use default compose file
cp .env.example .env
vim .env  # Set GITHUB_OWNER, TAG, JWT_SECRET

# Start
docker compose up -d
```

**Pros:**
- No separate database container
- Zero configuration
- Automatic backups (single file)
- Perfect for embedded deployments

**Cons:**
- Single connection writer
- Not ideal for high-concurrency

---

### PostgreSQL (Recommended for Production)

**Best for:** Production deployments, high concurrency, data integrity

```bash
cd docker

# Use PostgreSQL compose file
cp .env.postgres .env
vim .env  # Set passwords and JWT_SECRET

# Start PostgreSQL + ActaLog
docker compose -f docker-compose.postgres.yml up -d

# View logs
docker compose -f docker-compose.postgres.yml logs -f

# Check health
curl http://localhost:8080/health
```

**Configuration (.env):**
```env
GITHUB_OWNER=yourusername
TAG=latest
DB_NAME=actalog
DB_USER=actalog
DB_PASSWORD=super_secure_password_here
JWT_SECRET=your_jwt_secret_here
```

**Pros:**
- Best for concurrent users
- ACID compliance
- Advanced features (JSON, full-text search)
- Battle-tested in production

**Cons:**
- Requires separate container
- More memory usage (~50MB)

**Connection String:**
```
postgresql://actalog:password@postgres:5432/actalog?sslmode=disable
```

---

### MariaDB (Alternative Production Option)

**Best for:** Teams familiar with MySQL, existing MySQL infrastructure

```bash
cd docker

# Use MariaDB compose file
cp .env.mariadb .env
vim .env  # Set passwords and JWT_SECRET

# Start MariaDB + ActaLog
docker compose -f docker-compose.mariadb.yml up -d

# View logs
docker compose -f docker-compose.mariadb.yml logs -f

# Check health
curl http://localhost:8080/health
```

**Configuration (.env):**
```env
GITHUB_OWNER=yourusername
TAG=latest
DB_NAME=actalog
DB_USER=actalog
DB_PASSWORD=super_secure_password_here
DB_ROOT_PASSWORD=different_root_password_here
JWT_SECRET=your_jwt_secret_here
```

**Pros:**
- MySQL-compatible
- Good performance
- Wide ecosystem support
- Familiar to many developers

**Cons:**
- Slightly more resource usage than PostgreSQL
- Less advanced JSON support

**Connection String:**
```
mysql://actalog:password@mariadb:3306/actalog
```

---

## Database Comparison

| Feature | SQLite | PostgreSQL | MariaDB |
|---------|--------|------------|---------|
| Setup Complexity | ⭐ Easiest | ⭐⭐ Moderate | ⭐⭐ Moderate |
| Concurrent Users | 1-10 | 100+ | 100+ |
| Memory Usage | ~5MB | ~50MB | ~60MB |
| Backup | Copy file | pg_dump | mysqldump |
| JSON Support | Limited | Excellent | Good |
| Full-text Search | Basic | Advanced | Good |
| Recommended For | Dev, Small | Production | Production |

---

## Migration Between Databases

### SQLite → PostgreSQL

```bash
# 1. Export from SQLite
sqlite3 actalog.db .dump > dump.sql

# 2. Convert to PostgreSQL format (manual)
# - Change AUTOINCREMENT to SERIAL
# - Fix datetime formats
# - Adjust syntax

# 3. Import to PostgreSQL
docker compose -f docker-compose.postgres.yml exec postgres \
  psql -U actalog -d actalog -f /tmp/dump.sql
```

### Using ActaLog Backup/Restore

ActaLog has built-in backup endpoints that work across databases:

```bash
# 1. Create backup on SQLite deployment
curl -X POST http://localhost:8080/api/admin/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"description":"Pre-migration backup"}'

# 2. Download backup
curl http://localhost:8080/api/admin/backups/1/download \
  -H "Authorization: Bearer $TOKEN" \
  -o backup.zip

# 3. Deploy with PostgreSQL
docker compose -f docker-compose.postgres.yml up -d

# 4. Restore backup
curl -X POST http://localhost:8080/api/admin/backups/restore \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@backup.zip"
```

---

## Database Maintenance

### PostgreSQL Maintenance

```bash
# Enter PostgreSQL shell
docker compose -f docker-compose.postgres.yml exec postgres \
  psql -U actalog -d actalog

# Vacuum (clean up)
VACUUM ANALYZE;

# Check database size
SELECT pg_size_pretty(pg_database_size('actalog'));

# List tables
\dt

# Backup
docker compose -f docker-compose.postgres.yml exec postgres \
  pg_dump -U actalog actalog > backup.sql

# Restore
docker compose -f docker-compose.postgres.yml exec -T postgres \
  psql -U actalog -d actalog < backup.sql
```

### MariaDB Maintenance

```bash
# Enter MariaDB shell
docker compose -f docker-compose.mariadb.yml exec mariadb \
  mysql -u actalog -p actalog

# Check database size
SELECT table_schema AS "Database",
  ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS "Size (MB)"
FROM information_schema.TABLES
WHERE table_schema = 'actalog'
GROUP BY table_schema;

# Optimize tables
OPTIMIZE TABLE users, workouts, workout_movements;

# Backup
docker compose -f docker-compose.mariadb.yml exec mariadb \
  mysqldump -u actalog -p actalog > backup.sql

# Restore
docker compose -f docker-compose.mariadb.yml exec -T mariadb \
  mysql -u actalog -p actalog < backup.sql
```

### SQLite Maintenance

```bash
# Backup (just copy the file)
docker compose cp actalog:/app/data/actalog.db ./backup.db

# Restore
docker compose cp ./backup.db actalog:/app/data/actalog.db
docker compose restart actalog

# Check database size
docker compose exec actalog ls -lh /app/data/actalog.db

# Vacuum (in container)
docker compose exec actalog sqlite3 /app/data/actalog.db "VACUUM;"
```

---

## External Database (Not in Docker)

You can use ActaLog with an existing database server running on:
- **Host machine** (accessed via `host.docker.internal` or host IP)
- **Another server** on your network (accessed via IP/hostname)
- **Cloud database** (RDS, Cloud SQL, etc.)

This is useful when you:
- Already have a database server running
- Want to share a database between multiple applications
- Need database features not available in containers
- Prefer to manage the database separately

### Option 1: Using Host Database Server

The Docker container can connect to a database server running on the host machine.

#### MariaDB/MySQL on Host

**1. Configure host database for remote connections:**

Edit MariaDB config (e.g., `/etc/mysql/mariadb.conf.d/50-server.cnf`):
```ini
[mysqld]
bind-address = 0.0.0.0    # Or specific IP
```

Restart MariaDB:
```bash
sudo systemctl restart mariadb
```

**2. Create database and user:**
```sql
CREATE DATABASE actalog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'actalog'@'%' IDENTIFIED BY 'yourpassword';
GRANT ALL PRIVILEGES ON actalog.* TO 'actalog'@'%';
FLUSH PRIVILEGES;
```

**3. Create docker-compose.yml for external host database:**

```yaml
# docker-compose.external-host.yml
version: '3.8'

services:
  actalog:
    image: ghcr.io/yourusername/actalog:latest
    container_name: actalog
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      # Connect to host database
      - DB_DRIVER=mysql
      - DB_HOST=host.docker.internal  # Docker Desktop (Mac/Windows)
      # - DB_HOST=172.17.0.1          # Linux (default bridge IP)
      # - DB_HOST=192.168.1.100       # Or use host's actual IP
      - DB_PORT=3306
      - DB_NAME=actalog
      - DB_USER=actalog
      - DB_PASSWORD=${DB_PASSWORD}

      # Application config
      - JWT_SECRET=${JWT_SECRET}
      - APP_URL=${APP_URL:-http://localhost:8080}
      - CORS_ORIGINS=${CORS_ORIGINS:-http://localhost:8080}

      # Optional: automatic seed import
      - ADMIN_EMAIL=${ADMIN_EMAIL:-}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD:-}

    volumes:
      - actalog-uploads:/app/uploads

    # No database container needed!

volumes:
  actalog-uploads:
```

**4. Create .env file:**
```env
GITHUB_OWNER=yourusername
TAG=latest
DB_PASSWORD=yourpassword
JWT_SECRET=your_secure_random_string_here
APP_URL=https://al.fluidgrid.site
CORS_ORIGINS=https://al.fluidgrid.site
```

**5. Start ActaLog:**
```bash
docker compose -f docker-compose.external-host.yml up -d
```

**6. Verify connection:**
```bash
# Check logs for successful database connection
docker compose logs -f

# Should see:
# Database initialized successfully
# Server listening on :8080
```

#### PostgreSQL on Host

Same process, but with PostgreSQL-specific configuration:

**1. Configure PostgreSQL for remote connections:**

Edit `postgresql.conf`:
```ini
listen_addresses = '*'
```

Edit `pg_hba.conf`:
```
# Allow connections from Docker containers
host    actalog    actalog    172.17.0.0/16    md5
```

Restart PostgreSQL:
```bash
sudo systemctl restart postgresql
```

**2. Create database and user:**
```sql
CREATE DATABASE actalog;
CREATE USER actalog WITH PASSWORD 'yourpassword';
GRANT ALL PRIVILEGES ON DATABASE actalog TO actalog;
```

**3. Update docker-compose.yml:**
```yaml
environment:
  - DB_DRIVER=postgres
  - DB_HOST=host.docker.internal  # Or host IP
  - DB_PORT=5432
  - DB_NAME=actalog
  - DB_USER=actalog
  - DB_PASSWORD=${DB_PASSWORD}
  - DB_SSLMODE=disable  # Or 'require' for production
```

### Option 2: Using Database on Another Server

If your database is on a different machine on the network:

```yaml
# docker-compose.external-network.yml
version: '3.8'

services:
  actalog:
    image: ghcr.io/yourusername/actalog:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      # MariaDB on network
      - DB_DRIVER=mysql
      - DB_HOST=192.168.1.100      # Database server IP
      - DB_PORT=3306
      - DB_NAME=actalog
      - DB_USER=actalog
      - DB_PASSWORD=${DB_PASSWORD}

      # OR PostgreSQL on network
      # - DB_DRIVER=postgres
      # - DB_HOST=db.example.com    # Can use hostname
      # - DB_PORT=5432

      # App config
      - JWT_SECRET=${JWT_SECRET}
      - APP_URL=${APP_URL}
      - CORS_ORIGINS=${CORS_ORIGINS}

    volumes:
      - actalog-uploads:/app/uploads

volumes:
  actalog-uploads:
```

Ensure the database server:
- Allows connections from the Docker host IP
- Has appropriate firewall rules (port 3306/5432 open)
- Has database and user created with proper permissions

### Troubleshooting External Databases

#### "Connection refused" errors

This error typically means:
1. The database isn't listening on the right interface
2. The container can't reach the host network
3. Firewall is blocking the connection

**Step 1: Check database is listening on all interfaces (not just localhost)**

For MariaDB/MySQL:
```bash
# Check current bind-address setting
sudo grep -E "bind-address|skip-networking" /etc/mysql/mariadb.conf.d/*.cnf

# If it shows bind-address = 127.0.0.1, change to 0.0.0.0
sudo vim /etc/mysql/mariadb.conf.d/50-server.cnf
# Set: bind-address = 0.0.0.0

# Restart MariaDB
sudo systemctl restart mariadb
```

For PostgreSQL:
```bash
# Check postgresql.conf
sudo grep listen_addresses /etc/postgresql/*/main/postgresql.conf
# Should be: listen_addresses = '*'

# Check pg_hba.conf allows Docker network
sudo grep "172.17" /etc/postgresql/*/main/pg_hba.conf
# Should have: host actalog actalog 172.17.0.0/16 md5

sudo systemctl restart postgresql
```

**Step 2: Verify database is listening**
```bash
# Check if database is listening on all interfaces
sudo netstat -tlnp | grep 3306   # MySQL/MariaDB
sudo netstat -tlnp | grep 5432   # PostgreSQL

# Should show 0.0.0.0:3306 (not 127.0.0.1:3306)
```

**Step 3: Test connectivity from container**
```bash
# Test from within container
docker exec -it actalog sh -c "nc -zv host.docker.internal 3306"

# Or install client tools
docker exec -it actalog sh
apk add mariadb-client
mysql -h host.docker.internal -u actalog -p actalog
```

#### "host.docker.internal" not resolving (Linux)

On Linux, `host.docker.internal` doesn't work by default unlike Mac/Windows Docker Desktop.

**Recommended: Add extra_hosts to docker-compose.yml**
```yaml
services:
  actalog:
    image: ghcr.io/johnzastrow/actalog:latest
    extra_hosts:
      - "host.docker.internal:host-gateway"  # This makes it work on Linux
    environment:
      - DB_DRIVER=mysql
      - DB_HOST=host.docker.internal
      - DB_PORT=3306
      - DB_NAME=acta
      - DB_USER=acta
      - DB_PASSWORD=${DB_PASSWORD}
```

**Alternative 1: Use Docker bridge IP**
```yaml
environment:
  - DB_HOST=172.17.0.1  # Default Docker bridge gateway
```

**Alternative 2: Use actual host IP**
```yaml
environment:
  - DB_HOST=192.168.1.100  # Replace with your host's actual IP
```

#### Database user lacks permission from Docker network

The database user may only have permission to connect from localhost, not from the Docker network.

**MariaDB/MySQL - Grant permissions:**
```sql
-- For Docker bridge network (more secure)
GRANT ALL PRIVILEGES ON acta.* TO 'acta'@'172.17.%' IDENTIFIED BY 'your_password';
FLUSH PRIVILEGES;

-- Or allow from any host (less secure, but easier)
GRANT ALL PRIVILEGES ON acta.* TO 'acta'@'%' IDENTIFIED BY 'your_password';
FLUSH PRIVILEGES;

-- Check grants
SHOW GRANTS FOR 'acta'@'%';
```

**PostgreSQL - Update pg_hba.conf:**
```bash
# Add line to pg_hba.conf
host    acta    acta    172.17.0.0/16    md5

# Reload PostgreSQL
sudo systemctl reload postgresql
```

#### Firewall blocking Docker network (connection times out)

If the connection times out (not refused), the firewall may be blocking Docker's network:

```bash
# Check UFW status
sudo ufw status

# Allow MariaDB from Docker network
sudo ufw allow from 172.17.0.0/16 to any port 3306
sudo ufw reload

# For PostgreSQL
sudo ufw allow from 172.17.0.0/16 to any port 5432
sudo ufw reload
```

#### Permission denied errors

**MariaDB/MySQL:**
```sql
-- Check user grants
SHOW GRANTS FOR 'actalog'@'%';

-- Recreate user if needed
DROP USER 'actalog'@'%';
CREATE USER 'actalog'@'%' IDENTIFIED BY 'yourpassword';
GRANT ALL PRIVILEGES ON actalog.* TO 'actalog'@'%';
FLUSH PRIVILEGES;
```

**PostgreSQL:**
```sql
-- Check user permissions
\du actalog

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE actalog TO actalog;
GRANT ALL ON SCHEMA public TO actalog;
```

### When to Use External vs Containerized Database

| Use External Database When | Use Containerized Database When |
|---------------------------|--------------------------------|
| Already have DB server running | Starting fresh deployment |
| Multiple apps share database | Single application |
| Need specific DB version/config | Want zero-configuration setup |
| Manage backups separately | Want Docker-managed volumes |
| Database on dedicated hardware | All-in-one Docker deployment |

---

## Troubleshooting

### "Connection refused" errors

**PostgreSQL:**
```bash
# Check if PostgreSQL is ready
docker compose -f docker-compose.postgres.yml exec postgres pg_isready

# Check logs
docker compose -f docker-compose.postgres.yml logs postgres
```

**MariaDB:**
```bash
# Check if MariaDB is ready
docker compose -f docker-compose.mariadb.yml exec mariadb \
  healthcheck.sh --connect

# Check logs
docker compose -f docker-compose.mariadb.yml logs mariadb
```

### "database locked" with SQLite

This happens with high concurrency. Consider migrating to PostgreSQL:
```bash
docker compose -f docker-compose.postgres.yml up -d
```

### Lost database password

**PostgreSQL:**
```bash
# Reset password
docker compose -f docker-compose.postgres.yml exec postgres \
  psql -U actalog -c "ALTER USER actalog PASSWORD 'new_password';"
```

**MariaDB:**
```bash
# Reset password
docker compose -f docker-compose.mariadb.yml exec mariadb \
  mysql -u root -p -e "SET PASSWORD FOR 'actalog'@'%' = PASSWORD('new_password');"
```

---

## Performance Tuning

### PostgreSQL

Add to compose file under `postgres.command`:
```yaml
command:
  - postgres
  - -c
  - max_connections=200
  - -c
  - shared_buffers=256MB
  - -c
  - effective_cache_size=1GB
```

### MariaDB

Add to compose file under `mariadb.command`:
```yaml
command:
  - --max-connections=200
  - --innodb-buffer-pool-size=256M
  - --innodb-log-file-size=64M
```

---

## Recommendation by Use Case

| Use Case | Database | Reason |
|----------|----------|--------|
| Personal use | SQLite | Simple, zero-config |
| Small team (< 10) | SQLite or PostgreSQL | Low overhead |
| Production team | PostgreSQL | Best concurrency |
| High traffic | PostgreSQL | Proven at scale |
| Existing MySQL infra | MariaDB | Easy integration |
| Edge/embedded | SQLite | Single file, portable |
| Multi-region | PostgreSQL | Replication support |

---

## Seed Data Import

ActaLog includes comprehensive seed data:
- **182 movements** (all CrossFit movements including Girl/Hero WOD movements)
- **314 WODs** (all benchmark Girl and Hero WODs)

### Automatic Import (Recommended)

To automatically import seed data on first deployment:

1. **Set admin credentials in `.env`** before starting:
   ```env
   ADMIN_EMAIL=admin@example.com
   ADMIN_PASSWORD=YourSecurePassword123
   ```

2. **Start the application:**
   ```bash
   docker compose up -d
   ```

3. **Register your first user** with the same email/password from `.env`

4. **Seeds import automatically** on first startup if credentials are set

The import script:
- Runs only once (creates marker file `/app/data/.seeds_imported`)
- Skips if admin credentials are not set
- Logs progress to container logs: `docker compose logs -f`

### Manual Import (Alternative)

If you prefer to import manually or didn't set credentials:

1. **Via Web UI** (easiest):
   - Login as admin
   - Navigate to Import page
   - Upload `/app/seeds/movements.csv`
   - Upload `/app/seeds/wods.csv`

2. **Via API**:
   ```bash
   # Get admin token
   TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"YourPassword"}' \
     | jq -r '.access_token')

   # Import movements
   docker compose cp actalog:/app/seeds/movements.csv ./movements.csv
   curl -X POST http://localhost:8080/api/import/movements/confirm \
     -H "Authorization: Bearer $TOKEN" \
     -F "file=@movements.csv" \
     -F "skip_duplicates=true"

   # Import WODs
   docker compose cp actalog:/app/seeds/wods.csv ./wods.csv
   curl -X POST http://localhost:8080/api/import/wods/confirm \
     -H "Authorization: Bearer $TOKEN" \
     -F "file=@wods.csv" \
     -F "skip_duplicates=true"
   ```

### Checking Import Status

```bash
# Check if seeds were imported
docker compose exec actalog ls -la /app/data/.seeds_imported

# View import logs
docker compose logs actalog | grep -i seed
```

---

## Quick Commands

```bash
# SQLite deployment
docker compose up -d

# PostgreSQL deployment
docker compose -f docker-compose.postgres.yml up -d

# MariaDB deployment
docker compose -f docker-compose.mariadb.yml up -d

# Stop any deployment
docker compose down  # or add -f <file>

# View logs
docker compose logs -f

# Backup database
docker compose exec actalog wget -O- http://localhost:8080/api/admin/backups/1/download
```
