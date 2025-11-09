# Deployment Guide

This guide covers deploying ActaLog as a Progressive Web App (PWA) to production.

## Prerequisites

- Server with Linux or Windows Server
- Domain name (for HTTPS)
- Go 1.21+ installed on server
- Node.js 18+ installed on server
- (Recommended) Nginx or Apache
- (Required) SSL certificate (Let's Encrypt recommended)

## ⚠️ HTTPS Requirement

**PWA features require HTTPS in production**. Service workers will not register over HTTP (except localhost).

### Getting SSL Certificate

**Option 1: Let's Encrypt (Free, Recommended)**
```bash
# Install certbot
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx

# Get certificate (Nginx)
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Auto-renewal is configured automatically
sudo certbot renew --dry-run
```

**Option 2: Commercial SSL**
- Purchase from provider (Namecheap, GoDaddy, etc.)
- Follow provider's installation instructions

## Deployment Options

### Option 1: Traditional Server Deployment (Recommended)

#### 1. Build the Application

**Backend**:
```bash
# On your development machine or server
cd /path/to/actalog

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/actalog cmd/actalog/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/actalog.exe cmd/actalog/main.go
```

**Frontend (PWA)**:
```bash
cd web

# Install dependencies
npm install

# Build for production (includes PWA assets)
npm run build

# Output will be in web/dist/
# This includes:
# - Minified JS/CSS
# - Service worker (sw.js)
# - Web app manifest (manifest.webmanifest)
# - Precached assets
```

#### 2. Server Setup

**Directory Structure**:
```
/opt/actalog/
├── actalog              # Backend binary
├── web/dist/           # Frontend build (PWA)
├── .env                # Configuration
└── actalog.db          # SQLite database (or use PostgreSQL)
```

**Copy Files**:
```bash
# Create directory
sudo mkdir -p /opt/actalog

# Copy backend binary
sudo cp bin/actalog /opt/actalog/

# Copy frontend build
sudo cp -r web/dist /opt/actalog/web/

# Copy environment file
sudo cp .env.production /opt/actalog/.env

# Set permissions
sudo chmod +x /opt/actalog/actalog
```

#### 3. Configure Environment

Create `/opt/actalog/.env`:
```env
# Application
APP_ENV=production
LOG_LEVEL=info

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database (PostgreSQL recommended for production)
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=actalog
DB_PASSWORD=secure_password_here
DB_NAME=actalog
DB_SSLMODE=require

# Security (CHANGE THESE!)
JWT_SECRET=your-very-secure-random-secret-key-here
JWT_EXPIRATION=24h

# CORS (your domain)
CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

#### 4. Setup Systemd Service (Linux)

Create `/etc/systemd/system/actalog.service`:
```ini
[Unit]
Description=ActaLog PWA Server
After=network.target postgresql.service

[Service]
Type=simple
User=actalog
WorkingDirectory=/opt/actalog
ExecStart=/opt/actalog/actalog
Restart=on-failure
RestartSec=5s

# Environment
Environment="APP_ENV=production"
EnvironmentFile=/opt/actalog/.env

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/actalog

[Install]
WantedBy=multi-user.target
```

**Enable and start**:
```bash
# Create user
sudo useradd -r -s /bin/false actalog
sudo chown -R actalog:actalog /opt/actalog

# Enable service
sudo systemctl daemon-reload
sudo systemctl enable actalog
sudo systemctl start actalog

# Check status
sudo systemctl status actalog

# View logs
sudo journalctl -u actalog -f
```

#### 5. Configure Nginx Reverse Proxy

Create `/etc/nginx/sites-available/actalog`:
```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server (PWA requires HTTPS)
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL certificates (Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Root directory for PWA
    root /opt/actalog/web/dist;
    index index.html;

    # Compression (important for PWA performance)
    gzip on;
    gzip_vary on;
    gzip_min_length 1000;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Service worker - MUST be at root scope with no cache
    location = /sw.js {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
        try_files $uri =404;
    }

    # Manifest file
    location = /manifest.webmanifest {
        add_header Content-Type "application/manifest+json";
        add_header Cache-Control "public, max-age=3600";
        try_files $uri =404;
    }

    # API requests to Go backend
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Health check
    location /health {
        proxy_pass http://localhost:8080;
        access_log off;
    }

    # SPA fallback - serve index.html for all other routes
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

**Enable site**:
```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/actalog /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Option 2: Docker Deployment

#### 1. Build Docker Images

**Backend Dockerfile** (already in project):
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o actalog cmd/actalog/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/actalog .
EXPOSE 8080
CMD ["./actalog"]
```

**Frontend Dockerfile**:
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### 2. Docker Compose with HTTPS

**docker-compose.prod.yml**:
```yaml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: Dockerfile
    env_file:
      - .env.production
    depends_on:
      - db
    networks:
      - actalog-network

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    depends_on:
      - backend
    networks:
      - actalog-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.prod.conf:/etc/nginx/nginx.conf
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - frontend
      - backend
    networks:
      - actalog-network

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: actalog
      POSTGRES_USER: actalog
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - actalog-network

volumes:
  postgres_data:

networks:
  actalog-network:
    driver: bridge
```

**Deploy**:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## PWA-Specific Deployment Checklist

### Before Deployment

- [ ] Generate all PWA icons (72px - 512px)
- [ ] Update manifest with correct domain in `start_url`
- [ ] Set correct `theme_color` and `background_color`
- [ ] Test service worker registration locally
- [ ] Run Lighthouse PWA audit (target 90+)
- [ ] Verify offline functionality works

### After Deployment

- [ ] Verify HTTPS is working correctly
- [ ] Test service worker registration in production
- [ ] Check manifest in DevTools (Application → Manifest)
- [ ] Test "Add to Home Screen" on mobile device
- [ ] Verify offline mode works in production
- [ ] Test update flow (deploy new version)
- [ ] Run Lighthouse audit on production URL

### Testing PWA Install

**Desktop (Chrome/Edge)**:
1. Visit https://yourdomain.com
2. Look for install icon in address bar
3. Click to install
4. App opens in standalone window

**Mobile (Android)**:
1. Visit https://yourdomain.com
2. Tap browser menu → "Add to Home screen"
3. App icon appears on home screen
4. Opens in fullscreen mode

**Mobile (iOS)**:
1. Visit https://yourdomain.com in Safari
2. Tap Share button
3. Tap "Add to Home Screen"
4. App icon appears on home screen

## Database Setup

### PostgreSQL (Recommended for Production)

```bash
# Install PostgreSQL
sudo apt-get install postgresql postgresql-contrib

# Create database and user
sudo -u postgres psql
postgres=# CREATE DATABASE actalog;
postgres=# CREATE USER actalog WITH ENCRYPTED PASSWORD 'secure_password';
postgres=# GRANT ALL PRIVILEGES ON DATABASE actalog TO actalog;
postgres=# \q

# Run migrations (if using migration tool)
# Or the app will auto-migrate on first run
```

### SQLite (Small Deployments)

```env
DB_DRIVER=sqlite3
DB_NAME=/opt/actalog/actalog.db
```

## Monitoring and Maintenance

### Health Checks

```bash
# Check app is running
curl https://yourdomain.com/health

# Check service worker is served correctly
curl -I https://yourdomain.com/sw.js
# Should have Cache-Control: no-cache
```

### Logs

```bash
# Application logs
sudo journalctl -u actalog -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Backups

**Database**:
```bash
# PostgreSQL backup
pg_dump -U actalog actalog > actalog_backup_$(date +%Y%m%d).sql

# Restore
psql -U actalog actalog < actalog_backup_20250101.sql

# Automate with cron
0 2 * * * /usr/bin/pg_dump -U actalog actalog > /backups/actalog_$(date +\%Y\%m\%d).sql
```

**Application Data**:
```bash
# Backup entire application
tar -czf actalog_backup_$(date +%Y%m%d).tar.gz /opt/actalog
```

## Updates and Rollbacks

### Deploying Updates

```bash
# Build new version
npm run build

# Stop service
sudo systemctl stop actalog

# Backup current version
sudo cp -r /opt/actalog /opt/actalog.backup

# Update files
sudo cp -r web/dist/* /opt/actalog/web/dist/
sudo cp bin/actalog /opt/actalog/

# Start service
sudo systemctl start actalog

# Service worker will auto-update for users
```

### Rollback

```bash
# Stop service
sudo systemctl stop actalog

# Restore backup
sudo cp -r /opt/actalog.backup/* /opt/actalog/

# Start service
sudo systemctl start actalog
```

## Security Best Practices

1. **HTTPS Only**: Never deploy PWA over HTTP in production
2. **Strong JWT Secret**: Use cryptographically secure random string (64+ characters)
3. **Database Passwords**: Use strong, unique passwords
4. **CORS Configuration**: Whitelist only your domains
5. **Regular Updates**: Keep dependencies updated
6. **Firewall**: Only expose ports 80, 443, and SSH
7. **Rate Limiting**: Implement rate limiting on API endpoints
8. **SQL Injection**: Use parameterized queries (already implemented)

## Troubleshooting

### Service Worker Not Registering

- Verify HTTPS is working
- Check console for errors
- Clear browser cache and reload
- Verify sw.js is served with correct headers (no-cache)

### Install Prompt Not Showing

- Verify all PWA requirements met (HTTPS, manifest, icons, service worker)
- Check Lighthouse PWA audit
- Some browsers have install criteria (e.g., user engagement)

### Offline Mode Not Working

- Check service worker is active (DevTools → Application → Service Workers)
- Verify cache strategies in vite.config.js
- Test with DevTools offline mode
- Check network tab for cached responses

### Performance Issues

- Enable gzip compression
- Use CDN for static assets
- Optimize images
- Enable HTTP/2
- Check database query performance

## Scaling

For high-traffic deployments:

1. **Load Balancer**: Multiple backend instances behind load balancer
2. **Database**: PostgreSQL with connection pooling
3. **CDN**: CloudFlare, AWS CloudFront for static assets
4. **Caching**: Redis for session storage and API caching
5. **Horizontal Scaling**: Multiple app instances
6. **Database Replication**: Primary-replica setup

## Version

- **Current Deployment Version**: 0.2.0
- **Last Updated**: 2025-11-08
