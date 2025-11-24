# Docker Deployment Guide

This guide covers Docker deployment for ActaLog, including the single-port production architecture, multi-stage build process, and deployment options.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Quick Start](#quick-start)
- [Building Images](#building-images)
- [Deployment Options](#deployment-options)
- [Configuration](#configuration)
- [Volumes and Data Persistence](#volumes-and-data-persistence)
- [Database Options](#database-options)
- [Troubleshooting](#troubleshooting)

## Architecture Overview

### Single-Port Production Architecture

ActaLog uses a **production-optimized single-port architecture** where the Go backend serves both the API and the frontend static files from port 8080.

**Key Benefits:**
- Simpler deployment (only one port to expose)
- No CORS configuration needed
- Lower resource usage (no Node.js in production)
- Industry-standard pattern for production SPAs
- Easier reverse proxy configuration

**How It Works:**

```
Port 8080 → Go Application
    ├── /api/*      → Backend API endpoints
    ├── /uploads/*  → User-uploaded files
    └── /*          → Frontend static files (with SPA routing)
```

**Route Priority:**
1. Health check: `/health` → API health status
2. API routes: `/api/*` → Backend handlers
3. Uploads: `/uploads/*` → Static file server
4. Frontend: `/*` → Serves static files or `index.html` for SPA routing

**Implementation:** See `cmd/actalog/main.go:251-262` for frontend directory configuration and lines 418-436 for static file serving logic.

### Multi-Stage Build Process

The Dockerfile uses a multi-stage build to create an optimized production image:

```dockerfile
# Stage 1: Build Frontend (Node.js)
FROM node:20-alpine AS frontend-builder
# ... builds /app/web/dist

# Stage 2: Build Backend (Go)
FROM golang:1.23-alpine AS backend-builder
# ... builds /app/bin/actalog

# Stage 3: Runtime (Alpine)
FROM alpine:latest
# Copies both frontend and backend
```

**Why Multi-Stage?**
- Smaller final image (~50MB vs 1GB+)
- No build tools in production image
- Security: minimal attack surface
- Separate build caching for frontend and backend

## Quick Start

### Pull and Run (Simplest)

```bash
# Pull the latest image from GitHub Container Registry
docker pull ghcr.io/johnzastrow/actalog:latest

# Run with SQLite (single-container deployment)
docker run -d \
  -p 8080:8080 \
  -v actalog-data:/app/data \
  -v actalog-uploads:/app/uploads \
  --name actalog \
  ghcr.io/johnzastrow/actalog:latest

# Access the application
open http://localhost:8080
```

### Using Docker Compose (Recommended for Production)

```bash
# SQLite (development/single-user)
docker-compose -f docker/docker-compose.sqlite.yml up -d

# PostgreSQL (production recommended)
docker-compose -f docker/docker-compose.postgresql.yml up -d

# MariaDB (alternative production database)
docker-compose -f docker/docker-compose.mariadb.yml up -d
```

## Building Images

### Using Build Script (Recommended)

```bash
# Build and tag image
./docker/scripts/build.sh v0.10.0

# Push to GitHub Container Registry
./docker/scripts/push.sh v0.10.0
```

**What the script does:**
1. Checks for uncommitted changes (fails if dirty)
2. Runs multi-stage Docker build
3. Tags with version and 'latest'
4. Optionally pushes to ghcr.io

### Manual Build

```bash
# Build from root of repository
docker build -t actalog:local -f docker/Dockerfile .

# Run locally built image
docker run -p 8080:8080 -v actalog-data:/app/data actalog:local
```

### Build Arguments

The Dockerfile supports build-time configuration:

```bash
docker build \
  --build-arg GOTOOLCHAIN=auto \
  -t actalog:custom \
  -f docker/Dockerfile .
```

## Deployment Options

### Option 1: Single Container with SQLite

**Best for:** Development, testing, single-user deployments

```bash
docker run -d \
  -p 8080:8080 \
  -e DB_DRIVER=sqlite3 \
  -e DB_NAME=/app/data/actalog.db \
  -v actalog-data:/app/data \
  -v actalog-uploads:/app/uploads \
  --name actalog \
  --restart unless-stopped \
  ghcr.io/johnzastrow/actalog:latest
```

**Advantages:**
- Simplest deployment (one container)
- No external database needed
- Easy backups (copy volume)

**Limitations:**
- No horizontal scaling
- Single point of failure

### Option 2: Multi-Container with PostgreSQL

**Best for:** Production, multi-user, high-availability

```bash
# Using docker-compose
docker-compose -f docker/docker-compose.postgresql.yml up -d
```

**docker-compose.postgresql.yml:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: actalog
      POSTGRES_PASSWORD: changeme
      POSTGRES_DB: actalog
    volumes:
      - postgres-data:/var/lib/postgresql/data
    restart: unless-stopped

  actalog:
    image: ghcr.io/johnzastrow/actalog:latest
    ports:
      - "8080:8080"
    environment:
      DB_DRIVER: postgres
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: actalog
      DB_PASSWORD: changeme
      DB_NAME: actalog
      DB_SSLMODE: disable
    volumes:
      - actalog-uploads:/app/uploads
    depends_on:
      - postgres
    restart: unless-stopped

volumes:
  postgres-data:
  actalog-uploads:
```

**Advantages:**
- Production-ready database
- Better performance at scale
- Proper transaction support
- Can scale horizontally

### Option 3: Behind Reverse Proxy (Nginx/Caddy)

**Best for:** Production with TLS, multiple services

```nginx
# /etc/nginx/sites-available/actalog
server {
    listen 80;
    server_name actalog.example.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name actalog.example.com;

    ssl_certificate /etc/letsencrypt/live/actalog.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/actalog.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Increase timeouts for large uploads
        client_max_body_size 50M;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }
}
```

**Caddy (simpler alternative):**

```caddyfile
actalog.example.com {
    reverse_proxy localhost:8080
}
```

## Configuration

### Environment Variables

All configuration is done via environment variables:

**Database:**
```bash
DB_DRIVER=sqlite3                    # sqlite3, postgres, mysql
DB_NAME=/app/data/actalog.db        # For SQLite
DB_HOST=postgres                     # For PostgreSQL/MySQL
DB_PORT=5432                         # Database port
DB_USER=actalog                      # Database user
DB_PASSWORD=changeme                 # Database password
DB_NAME=actalog                      # Database name
DB_SSLMODE=disable                   # PostgreSQL SSL mode
DB_SCHEMA=public                     # PostgreSQL schema
```

**Application:**
```bash
PORT=8080                            # Application port
APP_URL=https://actalog.example.com  # Base URL for emails
CORS_ORIGINS=https://actalog.example.com  # CORS allowed origins
ALLOW_REGISTRATION=true              # Allow new user registration
```

**Authentication:**
```bash
JWT_SECRET=changeme-in-production    # JWT signing secret (REQUIRED)
JWT_EXPIRATION=24h                   # Access token lifetime
JWT_REFRESH_DURATION=168h            # Refresh token lifetime (7 days)
```

**Email (Optional):**
```bash
EMAIL_ENABLED=true
EMAIL_FROM=noreply@actalog.example.com
EMAIL_SMTP_HOST=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USER=your-email@gmail.com
EMAIL_SMTP_PASSWORD=your-app-password
EMAIL_REQUIRE_VERIFICATION=false
```

**Security:**
```bash
MAX_LOGIN_ATTEMPTS=5                 # Max failed login attempts
ACCOUNT_LOCKOUT_DURATION=30m         # Account lock duration
```

**Frontend (Build-time):**
```bash
FRONTEND_DIR=/app/web/dist           # Frontend static files location
```

### Docker Environment File

Create `.env` file for docker-compose:

```bash
# .env
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=actalog
DB_PASSWORD=your-secure-password-here
DB_NAME=actalog

JWT_SECRET=your-jwt-secret-here-min-32-chars

EMAIL_ENABLED=true
EMAIL_FROM=noreply@actalog.example.com
EMAIL_SMTP_HOST=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USER=your-email@gmail.com
EMAIL_SMTP_PASSWORD=your-app-password

APP_URL=https://actalog.example.com
CORS_ORIGINS=https://actalog.example.com
```

Then reference in docker-compose.yml:

```yaml
services:
  actalog:
    env_file: .env
```

## Volumes and Data Persistence

### Required Volumes

1. **`/app/data`** - Database files (for SQLite) and application data
   ```bash
   -v actalog-data:/app/data
   ```

2. **`/app/uploads`** - User-uploaded files (avatars, etc.)
   ```bash
   -v actalog-uploads:/app/uploads
   ```

### Backup Strategy

#### SQLite Backup

```bash
# Create backup
docker exec actalog sqlite3 /app/data/actalog.db ".backup '/app/data/backup.db'"

# Copy backup to host
docker cp actalog:/app/data/backup.db ./actalog-backup-$(date +%Y%m%d).db

# Or use volume mount
docker run --rm \
  -v actalog-data:/data \
  -v $(pwd)/backups:/backups \
  alpine cp /data/actalog.db /backups/actalog-$(date +%Y%m%d).db
```

#### PostgreSQL Backup

```bash
# Dump database
docker exec postgres pg_dump -U actalog actalog > backup.sql

# Restore database
docker exec -i postgres psql -U actalog actalog < backup.sql
```

#### Uploads Backup

```bash
# Tar and compress uploads
docker run --rm \
  -v actalog-uploads:/uploads \
  -v $(pwd)/backups:/backups \
  alpine tar czf /backups/uploads-$(date +%Y%m%d).tar.gz -C /uploads .
```

## Database Options

### SQLite (Default)

**Configuration:**
```bash
-e DB_DRIVER=sqlite3
-e DB_NAME=/app/data/actalog.db
```

**Pros:**
- Zero configuration
- Fast for single-user
- Easy backups

**Cons:**
- No concurrent writes
- Limited scalability

### PostgreSQL (Recommended for Production)

**Configuration:**
```bash
-e DB_DRIVER=postgres
-e DB_HOST=postgres
-e DB_PORT=5432
-e DB_USER=actalog
-e DB_PASSWORD=changeme
-e DB_NAME=actalog
-e DB_SSLMODE=disable
-e DB_SCHEMA=public
```

**Pros:**
- Production-grade
- Horizontal scaling
- ACID compliance
- Advanced features

### MariaDB/MySQL

**Configuration:**
```bash
-e DB_DRIVER=mysql
-e DB_HOST=mariadb
-e DB_PORT=3306
-e DB_USER=actalog
-e DB_PASSWORD=changeme
-e DB_NAME=actalog
```

**Pros:**
- Widely supported
- Good performance
- Familiar to many admins

## Troubleshooting

### Container Won't Start

**Check logs:**
```bash
docker logs actalog
```

**Common issues:**
- Missing JWT_SECRET environment variable
- Database connection failure
- Port 8080 already in use

### Frontend Not Serving

**Verify frontend files exist:**
```bash
docker exec actalog ls -la /app/web/dist
```

**Check logs for:**
```
Serving frontend from: /app/web/dist
```

**If missing:**
- Image might not have been built correctly
- Rebuild with: `./docker/scripts/build.sh`

### Database Connection Errors

**SQLite:**
```bash
# Check if database file exists
docker exec actalog ls -la /app/data/

# Check permissions
docker exec actalog ls -la /app/data/actalog.db
```

**PostgreSQL/MariaDB:**
```bash
# Test database connection
docker exec actalog nc -zv postgres 5432

# Check database logs
docker logs postgres
```

### 404 Errors for Frontend Routes

This is expected behavior! The Go backend serves `index.html` for all non-existent paths to support Vue Router's history mode.

**If still getting 404s:**
1. Check that `/` returns the frontend
2. Verify frontend files exist: `docker exec actalog ls /app/web/dist`
3. Check logs for static file serving configuration

### Permission Denied Errors

The container runs as non-root user `actalog` (UID 1000):

```bash
# If you need to fix permissions on volumes
docker run --rm -v actalog-data:/data alpine chown -R 1000:1000 /data
docker run --rm -v actalog-uploads:/uploads alpine chown -R 1000:1000 /uploads
```

### High Memory Usage

**Check resource usage:**
```bash
docker stats actalog
```

**Limit resources:**
```bash
docker run -d \
  --memory="512m" \
  --cpus="1.0" \
  -p 8080:8080 \
  ghcr.io/johnzastrow/actalog:latest
```

### Seed Data Not Loading

The application automatically imports seed data (movements and WODs) on first run.

**Check if seeds were imported:**
```bash
docker exec actalog ls -la /app/data/.seeds_imported
```

**If marker file doesn't exist:**
```bash
# Manually run seed import
docker exec actalog /app/scripts/init-seeds.sh
```

## Advanced Topics

### Custom Seed Data

Mount your own seed files:

```bash
docker run -d \
  -p 8080:8080 \
  -v ./custom-movements.csv:/app/seeds/movements.csv:ro \
  -v ./custom-wods.csv:/app/seeds/wods.csv:ro \
  ghcr.io/johnzastrow/actalog:latest
```

### Health Checks

The image includes a built-in health check:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

**Check container health:**
```bash
docker inspect --format='{{.State.Health.Status}}' actalog
```

### Multi-Instance Deployment

For horizontal scaling:

1. Use shared database (PostgreSQL/MariaDB)
2. Use shared uploads storage (NFS/S3)
3. Run multiple containers behind load balancer

```bash
# Instance 1
docker run -d --name actalog-1 -p 8081:8080 \
  -e DB_DRIVER=postgres -e DB_HOST=postgres-server \
  -v shared-uploads:/app/uploads \
  ghcr.io/johnzastrow/actalog:latest

# Instance 2
docker run -d --name actalog-2 -p 8082:8080 \
  -e DB_DRIVER=postgres -e DB_HOST=postgres-server \
  -v shared-uploads:/app/uploads \
  ghcr.io/johnzastrow/actalog:latest

# Load balancer forwards to both instances
```

## Security Best Practices

1. **Change default JWT secret:**
   ```bash
   JWT_SECRET=$(openssl rand -base64 32)
   ```

2. **Use environment files for secrets:**
   - Never commit `.env` files
   - Use Docker secrets or external secret management

3. **Enable HTTPS:**
   - Use reverse proxy (Nginx/Caddy) with TLS
   - Required for PWA features

4. **Restrict CORS:**
   ```bash
   CORS_ORIGINS=https://yourdomain.com
   ```

5. **Regular updates:**
   ```bash
   docker pull ghcr.io/johnzastrow/actalog:latest
   docker-compose down && docker-compose up -d
   ```

6. **Network isolation:**
   ```yaml
   services:
     postgres:
       networks:
         - backend
     actalog:
       networks:
         - backend
         - frontend
   ```

## Production Checklist

- [ ] PostgreSQL or MariaDB configured (not SQLite)
- [ ] Custom JWT_SECRET set (min 32 characters)
- [ ] CORS_ORIGINS configured for your domain
- [ ] TLS/HTTPS enabled via reverse proxy
- [ ] Volumes configured for data persistence
- [ ] Backup strategy implemented
- [ ] Health checks enabled
- [ ] Resource limits set
- [ ] Logging configured
- [ ] Monitoring in place

## Support

For issues or questions:
- GitHub Issues: https://github.com/johnzastrow/actalog/issues
- Documentation: See `docs/` directory
- Architecture: See `docs/ARCHITECTURE.md`
- Development: See `CLAUDE.md`
