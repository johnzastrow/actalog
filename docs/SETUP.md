# Setup Guide

Quick setup instructions for ActaLog development.

## Prerequisites

- Go 1.21+
- Node.js 18+
- npm or yarn
- (Optional) Docker & Docker Compose
- (Optional) Make

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

### 3. Frontend Setup

```bash
# Navigate to web directory
cd web

# Install dependencies
npm install

# Start development server
npm run dev
```

Frontend will be available at http://localhost:3000

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

```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
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
