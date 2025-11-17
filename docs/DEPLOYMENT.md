# Setup Guide

Quick setup instructions for ActaLog development. ActaLog is a Progressive Web App (PWA) with offline capabilities.

## Prerequisites

- Go 1.21+
- Node.js 18+
- npm or yarn
- (Optional) Docker & Docker Compose
- (Optional) Make
- Modern browser (Chrome, Firefox, Safari, or Edge) for PWA features

## Quick Setup

### 1. Clone and Configure

```bash
git clone https://github.com/johnzastrow/actalog.git
cd actalog
cp .env.example .env
```

Edit `.env` with your settings (optional for local dev).

### 2. Backend Setup

```bash
# Install Go dependencies
go mod download

# Build the application
make build

# Run the server
make run
# Or: go run cmd/actalog/main.go
```

Backend will be available at http://localhost:8080

### 3. Frontend Setup (PWA)

```bash
# Navigate to web directory
cd web

# Install dependencies (includes PWA plugin)
npm install

# Start development server (PWA features enabled)
npm run dev
```

Frontend will be available at http://localhost:3000

**PWA Development Mode**:
- Service Worker is enabled in development (via `devOptions.enabled: true`)
- Manifest and offline features available at localhost
- No HTTPS required for localhost testing
- Check DevTools → Application → Service Workers to verify PWA status

## Windows Users

### Option 1: Use the Build Script (Recommended)

Windows users can use the provided `build.bat` script instead of Make:

```cmd
# Build the application
build.bat build

# Run the application
build.bat run

# Run tests
build.bat test

# Format code
build.bat fmt

# Clean build artifacts
build.bat clean

# Show help
build.bat help
```

### Option 2: Use Make with Git Bash or WSL

If you have Git Bash or WSL installed, you can use the Makefile commands as shown in the backend setup.

### Common Windows Issue: Access Denied

If you encounter an error like:
```
go: creating work dir: mkdir C:\WINDOWS\go-build...: Access is denied.
```

This is because Go tries to create its build cache in the Windows system directory. The Makefile and build.bat script automatically fix this by using the project's `.cache/` directory instead.

If using `go` commands directly, set these environment variables first:

```cmd
set GOCACHE=%CD%\.cache\go-build
set GOMODCACHE=%CD%\.cache\go-mod
set GOTMPDIR=%CD%\.cache\tmp
go build -o bin\actalog.exe cmd\actalog\main.go
```

Or in PowerShell:
```powershell
$env:GOCACHE="$PWD\.cache\go-build"
$env:GOMODCACHE="$PWD\.cache\go-mod"
$env:GOTMPDIR="$PWD\.cache\tmp"
go build -o bin\actalog.exe cmd\actalog\main.go
```

## PWA Development

### Testing PWA Features Locally

PWA features work on `http://localhost` without SSL:

1. **Service Worker Status**:
   - Open DevTools → Application → Service Workers
   - Verify service worker is registered
   - Check "Update on reload" for development

2. **Manifest Status**:
   - Open DevTools → Application → Manifest
   - Verify all fields are populated
   - Check icon availability

3. **Offline Testing**:
   - Open DevTools → Network
   - Check "Offline" to simulate no connection
   - Navigate the app to test offline functionality

4. **Cache Inspection**:
   - Open DevTools → Application → Cache Storage
   - View cached resources
   - Clear cache to test fresh install

### Generating PWA Icons

Icons are required for the app to install properly. Generate them from the logo:

```bash
# Option 1: Use online tool (easiest)
# Visit https://www.pwabuilder.com/imageGenerator
# Upload design/logo.png or design/logo.svg
# Download and extract to web/public/icons/

# Option 2: Use ImageMagick (if installed)
cd web/public/icons
# See icons/README.md for complete commands
convert ../../../design/logo.svg -resize 192x192 icon-192x192.png
convert ../../../design/logo.svg -resize 512x512 icon-512x512.png
# ... (see full list in web/public/icons/README.md)
```

### Testing on Mobile Devices

**Option 1: Same Network** (no SSL needed)
```bash
# Get your local IP address
# Windows: ipconfig
# Linux/Mac: ifconfig or ip addr

# Access from mobile: http://YOUR_IP:3000
# Example: http://192.168.1.100:3000
```

**Option 2: Port Forwarding** (Android only)
```bash
# Connect Android device via USB
# Enable USB debugging
# Chrome DevTools → Remote Devices → Port Forwarding
# Forward localhost:3000 to device
```

**Option 3: ngrok** (requires HTTPS for full PWA testing)
```bash
# Install ngrok: https://ngrok.com/
npx ngrok http 3000
# Use the https URL provided
```

### PWA Update Testing

Test the update flow:

1. Make changes to the app code
2. Rebuild: `npm run build`
3. The service worker will detect new version
4. User sees update prompt (configured in `main.js`)

### Lighthouse PWA Audit

Run Lighthouse to verify PWA compliance:

1. Open Chrome DevTools → Lighthouse
2. Select "Progressive Web App" category
3. Run audit
4. Target score: 90+

## Docker Setup (Alternative)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Development Tools

### Install Development Tools

```bash
make install-tools
```

This installs:
- `air` - Live reload for Go
- `goimports` - Import formatter
- `golangci-lint` - Linter

### Running Tests

```bash
# Backend tests
make test

# Unit tests only
make test-unit

# Frontend tests
cd web
npm test
```

### Code Quality

```bash
# Format Go code
make fmt

# Run linters
make lint

# Format frontend code
cd web
npm run format
npm run lint
```

## Database Setup

ActaLog supports three database systems: SQLite, PostgreSQL, and MySQL/MariaDB. See [DATABASE_SUPPORT.md](DATABASE_SUPPORT.md) for detailed multi-database configuration.

### How to Tell the App Which Database to Use

The application uses the **`DB_DRIVER` environment variable** in your `.env` file to determine which database system to connect to.

**Quick Setup:**
1. Copy the example configuration:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and set `DB_DRIVER` to one of:
   - `sqlite3` - File-based database (default for development)
   - `postgres` - PostgreSQL (recommended for production)
   - `mysql` - MySQL or MariaDB (production alternative)

3. Configure the corresponding database connection settings (see examples below)

4. Run the application:
   ```bash
   make run
   ```

The app will automatically connect to the configured database and run migrations on startup.

### SQLite (Default for Development)

No setup required. Database file will be created automatically at `actalog.db`.

### PostgreSQL (Production)

1. Install PostgreSQL
2. Create database:
   ```sql
   CREATE DATABASE actalog;
   CREATE USER actalog WITH ENCRYPTED PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE actalog TO actalog;
   ```
3. Update `.env`:
   ```env
   DB_DRIVER=postgres
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=actalog
   DB_PASSWORD=your_password
   DB_NAME=actalog
   DB_SSLMODE=disable
   ```

### MySQL/MariaDB (Production)

1. Install MySQL or MariaDB
2. Create database:
   ```sql
   CREATE DATABASE actalog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'actalog'@'localhost' IDENTIFIED BY 'your_password';
   GRANT ALL PRIVILEGES ON actalog.* TO 'actalog'@'localhost';
   FLUSH PRIVILEGES;
   ```
3. Update `.env`:
   ```env
   DB_DRIVER=mysql
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=actalog
   DB_PASSWORD=your_password
   DB_NAME=actalog
   ```

### Running Migrations

```bash
# Create a new migration
make migrate-create name=create_users_table

# Migrations will be in the migrations/ directory
# Implement your schema changes in the .up.sql and .down.sql files
```

## Environment Variables

Key environment variables:

```env
# Application
APP_ENV=development
LOG_LEVEL=info

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_DRIVER=sqlite
DB_NAME=actalog.db

# Security (CHANGE IN PRODUCTION!)
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRATION=24h

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

### Frontend Build Issues

If you encounter npm or build issues, try these steps in order:

**Option 1: Quick reinstall (safest)**
```bash
cd web
npm install
```

**Option 2: Clean cache and reinstall**
```bash
cd web
npm cache clean --force
npm install
```

**Option 3: Complete cleanup (for corrupted dependencies)**
```bash
cd web
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

**After cleanup, verify the build:**
```bash
npm run dev    # Test development server
# or
npm run build  # Test production build
```

### Database Connection Issues

1. Check database is running
2. Verify credentials in `.env`
3. Check firewall settings
4. For PostgreSQL, ensure `pg_hba.conf` allows connections

## IDE Setup

### VS Code

Recommended extensions:
- Go (golang.go)
- Vue Language Features (Volar)
- ESLint
- Prettier

### GoLand / WebStorm

Project should work out of the box with default settings.

## Next Steps

After setup:
1. Review [Architecture Documentation](ARCHITECTURE.md)
2. Read [Database Schema](DATABASE_SCHEMA.md)
3. Check [Requirements](REQUIIREMENTS.md) for features to implement
4. Start coding!

## Getting Help

- Check the [README](../README.md) for more details
- Review documentation in `docs/`
- Open an issue on GitHub



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

# Some notes on hosting

## Deploying ActaLog
ActaLog can be hosted in various environments, including local servers, cloud platforms, and containerized setups. Below are some general guidelines for deploying ActaLog effectively.

### Deployment Considerations
1. **Environment**: Choose an appropriate environment based on your needs. ActaLog can run on Linux, Windows, or macOS. For production environments, Linux servers are often preferred for their stability and performance.
2. **Dependencies**: Ensure that all necessary dependencies are installed. ActaLog requires Node.js and a database (e.g., PostgreSQL, MySQL) to function correctly.
3. **Configuration**: Configure ActaLog according to your requirements. This includes setting up
4. database connections, environment variables, and any other application-specific settings.
5. **Security**: Implement security best practices, such as using HTTPS, setting up firewalls, and regularly updating software to protect against vulnerabilities.
6. **Scaling**: Consider how you will scale ActaLog as your user base grows. This may involve load balancing, database optimization, and using container orchestration tools like Kubernetes.
7. **Monitoring**: Set up monitoring and logging to keep track of application performance and errors. Tools like Prometheus, Grafana, or ELK Stack can be useful for this purpose.
8. **Backup**: Regularly back up your database and application data to prevent data loss in case of failures.
9. **Documentation**: Refer to the official ActaLog documentation for detailed deployment instructions and best practices.
10. **Support**: Join the ActaLog community or seek professional support if you encounter issues during deployment.
11. By following these guidelines, you can ensure a smooth deployment of ActaLog that meets your operational needs.

### Deploying with Docker
ActaLog can be easily deployed using Docker, which simplifies the setup process and ensures consistency across different environments. Below are the steps to deploy ActaLog using Docker:
1. **Install Docker**: Ensure that Docker is installed on your server. You can follow the official Docker installation guide for your operating system.
2. **Pull the ActaLog Docker Image**: Use the following command to pull the latest ActaLog Docker image from Docker Hub:
   ```bash
    docker pull actalog/actalog:latest
    ```
3. **Create a Docker Network**: Create a Docker network to allow communication between the ActaLog container and the database container:
4.   ```bash
    docker network create actalog-network
    ```
5. **Run the Database Container**: Start a database container (e.g., PostgreSQL) and connect it to the Docker network:
6.  ```bash
    docker run -d --name actalog-db --network actalog-network -e POSTGRES_USER=actalog -e POSTGRES_PASSWORD=yourpassword -e POSTGRES_DB=actalog_db postgres:latest
    ```
7. **Run the ActaLog Container**: Start the ActaLog container, linking it to the database container:
8.  ```bash
    docker run -d --name actalog --network actalog-network -p 3000:3000 -e DB_HOST=actalog-db -e DB_USER=actalog -e DB_PASSWORD=yourpassword -e DB_NAME=actalog_db actalog/actalog:latest
    ```
9. **Access ActaLog**: Once the containers are running, you can access ActaLog by navigating to `http://your-server-ip:3000` in your web browser.
10. **Persist Data**: To ensure that your database data persists across container restarts,
    consider using Docker volumes to store the database data outside of the container.
11. **Monitor and Manage**: Use Docker commands to monitor and manage your ActaLog and database containers as needed.


### Deploying with Docker Compose
Using Docker Compose simplifies the deployment process by allowing you to define and manage multi-container applications with a single configuration file. Below are the steps to deploy ActaLog using Docker Compose:
1. **Install Docker and Docker Compose**: Ensure that both Docker and Docker Compose are installed
2. on your server. You can follow the official installation guides for your operating system.
3. **Create a Docker Compose File**: Create a `docker-compose.yml` file with the following content:

    --- fix ---

### Deploying from Source
To deploy ActaLog from source, follow these steps:

   **See SETUP.md for initial setup instructions.**

    
### Reverse Proxy Setup
Setting up a reverse proxy is essential for securely hosting ActaLog in a production environment. A reverse proxy acts as an intermediary between clients and the ActaLog application, providing benefits such as load balancing, SSL termination, and improved security.
Here are the general steps to set up a reverse proxy for ActaLog:
    1.  **Choose a Reverse Proxy**: Select a reverse proxy server such as Nginx, Apache, or Caddy based on your preferences and requirements.
    2.  **Install the Reverse Proxy**: Install the chosen reverse proxy server on
    3.  your server following the official installation instructions.
    4.  **Configure the Reverse Proxy**: Set up the reverse proxy to forward
    5.  requests to the ActaLog application. This typically involves creating a configuration file that specifies the server name, port, and proxy settings.
    6.  **Enable HTTPS**: For secure communication, configure SSL/TLS certificates
    7.  for your reverse proxy. You can use Let's Encrypt for free SSL certificates.
    8.  **Test the Configuration**: After setting up the reverse proxy, test
    9.  the configuration to ensure that requests are correctly forwarded to ActaLog and that HTTPS is functioning properly.
    10. **Monitor and Maintain**: Regularly monitor the reverse proxy logs and
    


## Example Reverse Proxy Configuration

When hosting ActaLog or similar applications behind a reverse proxy, it's important to set up your DNS and proxy server correctly to ensure smooth operation and accessibility.

Given a zone file such as the following from Linode:

```
; mydomainname.site [XXXXXXX]
$TTL 86400
@  IN  SOA  ns1.linode.com. myemail.gmail.com. 2021000021 14400 14400 1209600 86400
@    NS  ns1.linode.com.
@    NS  ns2.linode.com.
@    NS  ns3.linode.com.
@    NS  ns4.linode.com.
@    NS  ns5.linode.com.
@      MX  10  mail.mydomainname.site.
@      A  the.public.ip.address
al    30  A  the.public.ip.address  ; example of a subdomain for ActaLog. This would run internally on port 3000 but mapped to 80/443 externally
apo    30  A  the.public.ip.address ; example of a subdomain for another container perhaps running on the same server and port 9443, but mapped to 443 externally
mail      A  the.public.ip.address
recipe      A  the.public.ip.address ; example of a subdomain for a recipe app
www      AAAA  2600:3c02::f03c:95ff:feda:027e ; example of an IPv6 address for www
```

I like to use Caddy as a proxy server because it has automatic HTTPS via Let's Encrypt and is easy to configure.

Here is an example Caddyfile for the above zone file:

```
mydomainname.site, www.mydomainname.site {  
    reverse_proxy localhost:8080  # assuming ActaLog is running on port 8080 internally
}
al.mydomainname.site {
    reverse_proxy localhost:3000  # ActaLog running on port 3000 internally
        log {
                output file /var/log/caddy/al.access.log {
                        roll_size 1MB # Create new file when size exceeds 10MB
                        roll_keep 5 # Keep at most 5 rolled files
                        #            roll_keep_days 14 # Delete files older than 14 days
                }
        }
}
apo.mydomainname.site {
    reverse_proxy localhost:9443  # another app running on port 9443 internally
}
recipe.mydomainname.site {
    reverse_proxy localhost:5000  # recipe app running on port 5000 internally
}
```

* Make sure to adjust the internal ports according to where your applications are running. With this setup, Caddy will handle incoming requests and route them to the appropriate internal service based on the subdomain.
* Remember to open the necessary ports (80 and 443) on your server's firewall to allow HTTP and HTTPS traffic.
* Also, ensure that your DNS settings are correctly pointing to your server's public IP address for each of the subdomains you wish to use.

## Validating Caddyfile Syntax

### Using caddy validate

You can validate the syntax of a Caddyfile using the command line with the `caddy validate` or `caddy adapt` commands.

The caddy validate command loads and provisions the configuration, checking for any errors that might arise during the loading and provisioning stages, without actually starting the server.

To use it:

```bash
caddy validate --config /path/to/Caddyfile # often /etc/caddy/Caddyfile
```

* If the command succeeds, it will exit with no output and a status code of 0, indicating a valid configuration.
* If there are any syntax errors or structural issues, it will print an error message to the console and exit with a non-zero status code.
* If your file is named just Caddyfile and is in your current directory, you can simply run `caddy validate` without the --config flag.

### Using caddy adapt

* The caddy adapt command is another way to check your Caddyfile. It converts the Caddyfile into Caddy's native JSON format. This process will catch most syntax errors.
* To use it:

```bash
caddy adapt --config /path/to/Caddyfile # often /etc/caddy/Caddyfile
```

* If successful, it will output the resulting JSON configuration to standard output (stdout).
* If it fails, it means there is a syntax error in your Caddyfile.
* Both commands are excellent ways to perform a "dry run" of your configuration and ensure it is correct before deploying it to a production environment. For more details, consult the Caddy Documentation on the command-line interface.

## Common Reverse Proxy Configuration Mistakes

When configuring a reverse proxy for applications like ActaLog, several common mistakes can lead to performance issues, functional problems, or security vulnerabilities. Here are some of the most frequent errors to watch out for:

### Performance & Operational Mistakes

* **Not enabling keepalive connections**: By default, a new connection is often opened for every request to the backend server, which is inefficient. Enabling keepalive connections to upstream servers reuses existing connections, significantly improving performance.

HTTP keep-alive is enabled by default in Caddy, but you can configure its behavior in the reverse_proxy block of your Caddyfile or JSON config. You can customize settings like the idle timeout and the interval for probing connections.

#### Caddyfile

```caddyfile
your_site {
    reverse_proxy localhost:8080 {
        health_uri /health
        health_interval 1m
        health_timeout 5s
    }
    # To adjust the HTTP Keep-Alive settings, add a transport block.
    # This is a basic example. The full configuration options are in the Caddy JSON structure.
    transport http {
        keep_alive {
            # Optional: Enable/disable keep-alive (default is true)
            # enabled true 

            # Optional: Set probe interval (default 30s)
            # probe_interval 60s

            # Optional: Set idle timeout (default 5 minutes)
            # idle_timeout 5m
        }
    }
}
```

#### JSON config

```json
{
  "apps": {
    "http": {
      "servers": {
        "my_server": {
          "listen": [":443"],
          "routes": [
            {
              "handle": [
                {
                  "handler": "reverse_proxy",
                  "upstreams": [{"address": "localhost:8080"}],
                  "transport": {
                    "http": {
                      "keep_alive": {
                        "enabled": true,
                        "probe_interval": "30s",
                        "idle_timeout": "5m"
                      }
                    }
                  }
                }
              ]
            }
          ]
        }
      }
    }
  }
}
```

* **Default DNS resolution settings**: Many proxies (like Nginx) resolve domain names only once at startup. If your backend server's IP address changes, the proxy will keep sending traffic to the old IP unless you manually reload the configuration or use the resolver directive for dynamic resolution.
* **Improper timeout settings**: Timeouts (e.g., proxy_read_timeout) are often left at their defaults. If a backend server is slow to respond, users may encounter 504 Gateway Timeout errors. These settings should be adjusted to match the expected behavior and load of the backend application.

You can adjust Caddy's timeout settings in the Caddyfile or JSON config by using the `timeouts` directive, which can set a default for all timeouts or be configured individually for read, header, write, and idle times. For a reverse_proxy in Caddyfile, you can use the `response_header_timeout` and ``transport timeouts`. Timeouts can be set in duration formats like 30s, 1m, or 5m, or set to 0 or none to disable them.

#### In the Caddyfile

**Global timeouts**

Use the timeouts global option to set a default for all timeouts across your Caddyfile.
caddyfile

**Set all timeouts to 1 minute**
timeouts 1m

**Set specific timeouts**

```
timeouts {
    read 30s
    write 20s
    idle 5m
}
```

**Per-site timeouts**

You can also set timeouts for individual sites.
caddyfile

```
site.com {
    timeouts {
        read_header 1m
        read_body 0 # Disable read_body timeout
    }
}
```

**reverse_proxy timeouts**

For reverse proxy configurations, you can specify additional timeouts within the reverse_proxy block:

* `response_header_timeout`: Sets a timeout for when the backend does not write any response headers.
* `transport`: Sets a timeout for an individual API call to the backend.
*

```caddyfile
reverse_proxy example.com {
    header_up X-Real-IP {remote_host}
    response_header_timeout 10s
    transport 20s
}
```

#### In JSON configuration

**Global timeouts**
You can set global timeouts in the apps.http.servers block.

```json
{
    "apps": {
        "http": {
            "servers": {
                "myserver": {
                    "routes": [
                        {
                            "handle": [
                                {
                                    "handler": "reverse_proxy",
                                    "upstreams": [
                                        {"dial": "localhost:8080"}
                                    ]
                                }
                            ]
                        }
                    ],
                    "timeouts": {
                        "read_header": "1m",
                        "read_body": "1m",
                        "write": "1m",
                        "idle": "5m"
                    }
                }
            }
        }
    }
}
```

**Per-handler timeouts**

Some handlers may support their own timeouts. For example, `reverse_proxy` can take `response_header_timeout` and transport under the response field.

```json
{
    "apps": {
        "http": {
            "servers": {
                "myserver": {
                    "routes": [
                        {
                            "handle": [
                                {
                                    "handler": "reverse_proxy",
                                    "upstreams": [
                                        {"dial": "localhost:8080"}
                                    ],
                                    "response": {
                                        "header_timeout": "10s",
                                        "transport_timeout": "20s"
                                    }
                                }
                            ]
                        }
                    ]
                }
            }
        }
    }
}
```

* **Inadequate resource limits**: Failing to increase system resource limits, such as the maximum number of open file descriptors, can lead to service disruptions under heavy load.

### Functional Configuration Mistakes

* **Not forwarding the correct client IP/Host headers**: Backend applications often need the original client's IP address for logging, analytics, or security purposes. Without correctly setting headers like X-Real-IP or X-Forwarded-For, the backend will only see the proxy's IP address.
* **Mishandling redirects (absolute URLs)**: If the backend application sends an absolute URL in a redirect (e.g., in a 302 response) that is different from the external URL the client used, the client may be redirected to an internal, inaccessible URL.
* **Improper WebSocket proxying**: WebSocket connections require specific headers (Upgrade and Connection) to be handled correctly by the proxy; otherwise, real-time communication features may fail.

### Security Mistakes

* **Assuming a reverse proxy is a WAF**: A basic reverse proxy provides an abstraction layer but does not automatically provide Web Application Firewall (WAF), Intrusion Prevention System (IPS), or Intrusion Detection System (IDS) features. Additional security measures or a dedicated WAF are needed for robust protection.
* **Exposing internal services or paths**: Misconfigured routing rules, virtual hosts, or aliases can inadvertently expose internal applications or sensitive files to the public internet.
* **Leaving default credentials/settings**: Not changing default configurations or credentials for the proxy management interface or underlying systems leaves them vulnerable to attackers who can use publicly known default values.
* **Outdated software**: Failing to apply security patches to the reverse proxy software itself is a critical mistake that leaves known vulnerabilities open for exploitation.
* **Host Header manipulation vulnerabilities**: If not properly configured, an attacker can manipulate the Host header to access unintended backend servers or exploit misconfigured virtual hosts.
* **Improper SSL/TLS configuration**: Misconfiguring SSL/TLS settings can lead to security vulnerabilities, such as man-in-the-middle attacks or insecure data transmission.

## Conclusion

Careful configuration and regular review of your reverse proxy settings are essential to ensure optimal performance, functionality
, and security for applications like ActaLog. Always test your configuration in a staging environment before deploying it to production.



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
