# ActaLog Docker Deployment Guide

This guide covers building, publishing, and deploying ActaLog using Docker containers with GitHub Container Registry (ghcr.io).

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Building Images](#building-images)
- [Publishing to GitHub Container Registry](#publishing-to-github-container-registry)
- [Automated Builds with GitHub Actions](#automated-builds-with-github-actions)
- [Rapid Testing with Development Builds](#rapid-testing-with-development-builds)
- [Production Deployment](#production-deployment)
- [Creating Stable Releases](#creating-stable-releases)
- [Troubleshooting](#troubleshooting)

---

## Overview

ActaLog uses a multi-stage Docker build that creates an optimized production image containing both the Go backend and Vue.js frontend. Images are published to GitHub Container Registry for easy distribution and deployment.

**Image Location:** `ghcr.io/OWNER/REPO:TAG`

**Key Features:**
- Multi-stage build (frontend, backend, runtime)
- Optimized Alpine-based runtime (< 50MB)
- Multi-architecture support (amd64, arm64)
- Automated builds on every push
- Build number tagging for rapid testing
- Semantic versioning for stable releases

---

## Prerequisites

### Local Development
- Docker 20.10+ with BuildKit
- Docker Compose 2.0+
- Git
- GitHub account with repository access

### GitHub Container Registry
- GitHub Personal Access Token (PAT) with `write:packages` permission
- Repository must be public OR you need `read:packages` on private repos

---

## Quick Start

### 1. Pull and Run Latest Image

```bash
# Pull latest image from GitHub Container Registry
docker pull ghcr.io/OWNER/REPO:latest

# Run with default SQLite database
docker run -d \
  --name actalog \
  -p 8080:8080 \
  -v actalog-data:/app/data \
  -v actalog-uploads:/app/uploads \
  -e JWT_SECRET=your_secure_secret_here \
  ghcr.io/OWNER/REPO:latest

# Access the application
open http://localhost:8080
```

### 2. Using Docker Compose (Recommended)

```bash
cd docker

# Copy and configure environment
cp .env.example .env
# Edit .env with your configuration

# Start the application
docker compose up -d

# View logs
docker compose logs -f

# Stop the application
docker compose down
```

---

## Building Images

### Using Build Script (Recommended)

The build script automatically extracts version information and creates properly tagged images.

```bash
# Build with default 'dev' tag
./docker/scripts/build.sh

# Build with custom tag
./docker/scripts/build.sh v0.9.0

# Build for specific platform
DOCKER_PLATFORM=linux/arm64 ./docker/scripts/build.sh

# Build for multiple platforms (requires buildx)
DOCKER_PLATFORM=linux/amd64,linux/arm64 ./docker/scripts/build.sh
```

**What the script does:**
1. Extracts version and build number from `pkg/version/version.go`
2. Builds multi-stage Docker image
3. Tags image with:
   - Custom tag (or 'dev')
   - Build number tag (`build-62`)
4. Loads image into local Docker daemon
5. Displays pull and push instructions

### Manual Build

```bash
# From project root
docker build \
  -f docker/Dockerfile \
  -t ghcr.io/OWNER/REPO:dev \
  .

# Multi-platform build
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f docker/Dockerfile \
  -t ghcr.io/OWNER/REPO:dev \
  --push \
  .
```

---

## Publishing to GitHub Container Registry

### 1. Authenticate with GitHub

```bash
# Option 1: Using GitHub CLI (recommended)
gh auth login

# Option 2: Using Personal Access Token
export GITHUB_TOKEN=ghp_yourpersonalaccesstoken
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# Option 3: Interactive login
docker login ghcr.io
# Username: YOUR_GITHUB_USERNAME
# Password: YOUR_PERSONAL_ACCESS_TOKEN
```

### 2. Push Image Using Script

```bash
# Push image with specific tag
./docker/scripts/push.sh dev

# Push stable release
./docker/scripts/push.sh v0.9.0
```

**What the script does:**
1. Checks GitHub Container Registry authentication
2. Pushes image with specified tag
3. Pushes build-specific tag (`build-62`)
4. Displays pull command and package URL

### 3. Manual Push

```bash
# Push specific tag
docker push ghcr.io/OWNER/REPO:dev

# Push multiple tags
docker push ghcr.io/OWNER/REPO:v0.9.0
docker push ghcr.io/OWNER/REPO:latest
```

### 4. Make Package Public (Optional)

After first push, packages are private by default.

1. Go to `https://github.com/OWNER/REPO/packages`
2. Click on `actalog` package
3. Go to **Package settings**
4. Scroll to **Danger Zone**
5. Click **Change visibility** ’ **Public**

---

## Automated Builds with GitHub Actions

### How It Works

The GitHub Actions workflow (`.github/workflows/docker-build.yml`) automatically builds and pushes Docker images on:

- **Push to `main` or `develop` branches** ’ builds and pushes
- **Push tags matching `v*`** ’ builds release with semantic versioning
- **Pull requests to `main`** ’ builds only (no push)
- **Manual workflow dispatch** ’ build with custom tag

### Automatic Tagging Strategy

Images are automatically tagged based on the event:

| Event | Tags Created | Example |
|-------|-------------|---------|
| Push to `main` | `latest`, `main-SHA`, `build-N` | `latest`, `main-a1b2c3d`, `build-42` |
| Push to `develop` | `develop`, `develop-SHA`, `build-N` | `develop`, `develop-x9y8z7w`, `build-43` |
| Tag `v1.2.3` | `v1.2.3`, `v1.2`, `v1`, `build-N` | `v1.2.3`, `v1.2`, `v1`, `build-44` |
| PR #123 | `pr-123` | `pr-123` |
| Manual trigger | Custom tag, `build-N` | `manual`, `build-45` |

### Required Secrets

**No secrets required!** The workflow uses the built-in `GITHUB_TOKEN` which has automatic access to GitHub Container Registry.

### Viewing Build Status

1. Go to **Actions** tab in GitHub
2. Click on **Build and Push Docker Image** workflow
3. View build summary with:
   - Build metadata
   - All tags created
   - Pull command

---

## Rapid Testing with Development Builds

### Problem: Testing Remote Deployments Quickly

You need to test changes on a remote server without creating formal releases.

### Solution: Build Number Tagging

Every build automatically gets a unique `build-N` tag based on GitHub Actions run number.

**Workflow:**

```bash
# 1. Push code to main/develop branch
git add .
git commit -m "feat: add new feature"
git push origin main

# 2. GitHub Actions automatically builds and tags:
#    - ghcr.io/OWNER/REPO:main-abc1234
#    - ghcr.io/OWNER/REPO:build-62
#    - ghcr.io/OWNER/REPO:latest

# 3. On remote server, pull latest build
ssh user@remote-server
docker pull ghcr.io/OWNER/REPO:build-62

# 4. Update docker-compose.yml or restart container
export TAG=build-62
docker compose up -d

# 5. Verify deployment
docker ps
curl http://localhost:8080/api/version
```

### Fast Iteration Cycle

```bash
# Local development
make build                    # Increments build number
git add . && git commit -m "..." && git push

# Watch GitHub Actions (2-5 minutes)
gh run watch

# Deploy on remote
ssh user@server "docker pull ghcr.io/OWNER/REPO:build-63 && docker compose up -d"

# Test
curl http://server:8080/api/version
# {"version":"0.9.0-beta","build":63,...}
```

### Automated Remote Deployment Script

Create `scripts/deploy-dev.sh`:

```bash
#!/bin/bash
set -e

SERVER="user@your-server.com"
BUILD_NUM=$(grep -E "^\s*Build\s*=\s*[0-9]+" pkg/version/version.go | awk '{print $3}')

echo "Deploying build ${BUILD_NUM} to ${SERVER}..."

# SSH and pull latest build
ssh $SERVER << EOF
  cd /opt/actalog
  export TAG=build-${BUILD_NUM}
  docker pull ghcr.io/OWNER/REPO:\$TAG
  docker compose up -d
  docker ps
  curl -s http://localhost:8080/api/version | jq .
EOF

echo "Deployment complete!"
```

**Usage:**
```bash
# Make changes
vim internal/handler/some_handler.go

# Build, commit, push
make build
git add . && git commit -m "fix: bug fix" && git push

# Wait for GitHub Actions, then deploy
gh run watch && ./scripts/deploy-dev.sh
```

---

## Production Deployment

### Using Docker Compose (Recommended)

**1. Create deployment directory:**
```bash
ssh user@production-server
mkdir -p /opt/actalog
cd /opt/actalog
```

**2. Copy docker-compose.yml and .env:**
```bash
# Download from repository
curl -O https://raw.githubusercontent.com/OWNER/REPO/main/docker/docker-compose.yml
curl -O https://raw.githubusercontent.com/OWNER/REPO/main/docker/.env.example
mv .env.example .env
```

**3. Configure environment:**
```bash
vim .env
```

**Required configuration:**
```env
GITHUB_OWNER=your-github-username
TAG=v0.9.0                    # Use stable release tag

# Change this!
JWT_SECRET=GENERATE_SECURE_RANDOM_STRING_HERE

# Optional: PostgreSQL instead of SQLite
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_NAME=actalog
DB_USER=actalog
DB_PASSWORD=secure_password
```

**4. Start services:**
```bash
docker compose up -d
```

**5. Verify deployment:**
```bash
docker compose ps
docker compose logs -f
curl http://localhost:8080/health
curl http://localhost:8080/api/version
```

### Using Standalone Docker

```bash
docker run -d \
  --name actalog \
  --restart unless-stopped \
  -p 8080:8080 \
  -v actalog-data:/app/data \
  -v actalog-uploads:/app/uploads \
  -e DB_DRIVER=sqlite3 \
  -e DB_NAME=/app/data/actalog.db \
  -e JWT_SECRET=your_secure_secret_here \
  -e CORS_ORIGINS=https://yourdomain.com \
  ghcr.io/OWNER/REPO:v0.9.0
```

### Production with PostgreSQL

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  actalog:
    image: ghcr.io/OWNER/REPO:v0.9.0
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - DB_DRIVER=postgres
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=actalog
      - DB_USER=actalog
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - actalog-uploads:/app/uploads
    depends_on:
      - postgres
    networks:
      - actalog-network

  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_DB=actalog
      - POSTGRES_USER=actalog
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - actalog-network

volumes:
  actalog-uploads:
  postgres-data:

networks:
  actalog-network:
```

---

## Creating Stable Releases

### Release Process

Stable releases use semantic versioning tags (`v1.2.3`) and are automatically built by GitHub Actions.

**1. Update version in code:**

Edit `pkg/version/version.go`:
```go
const (
    Major = 0
    Minor = 9
    Patch = 0
    PreRelease = "beta"  // or "" for stable
    Build = 62  // Auto-incremented
)
```

Edit `web/package.json`:
```json
{
  "version": "0.9.0"
}
```

**2. Update changelog:**

Edit `docs/CHANGELOG.md`:
```markdown
## [0.9.0-beta] - 2025-01-23

### Added
- Feature X
- Feature Y

### Fixed
- Bug Z
```

**3. Commit and push changes:**
```bash
git add pkg/version/version.go web/package.json docs/CHANGELOG.md
git commit -m "chore: bump version to v0.9.0-beta"
git push origin main
```

**4. Create and push Git tag:**
```bash
# Create annotated tag
git tag -a v0.9.0-beta -m "Release v0.9.0-beta

- Offline PWA support
- Docker deployment
- 1RM calculation
- Performance improvements"

# Push tag to GitHub
git push origin v0.9.0-beta
```

**5. GitHub Actions automatically:**
- Detects the tag
- Builds multi-platform images
- Tags images with:
  - `v0.9.0-beta`
  - `v0.9`
  - `v0`
  - `build-62`
- Pushes to GitHub Container Registry
- Creates build summary

**6. Create GitHub Release (optional):**

```bash
# Using GitHub CLI
gh release create v0.9.0-beta \
  --title "ActaLog v0.9.0-beta" \
  --notes-file docs/CHANGELOG.md \
  --prerelease

# Or manually at: https://github.com/OWNER/REPO/releases/new
```

**7. Verify release:**

```bash
# Pull release image
docker pull ghcr.io/OWNER/REPO:v0.9.0-beta

# Verify version
docker run --rm ghcr.io/OWNER/REPO:v0.9.0-beta /app/actalog --version
# ActaLog v0.9.0-beta+build.62
```

### Release Channels

| Channel | Tag Pattern | Use Case | Stability |
|---------|------------|----------|-----------|
| `latest` | Auto from `main` | Latest stable | Production-ready |
| `develop` | Auto from `develop` | Bleeding edge | Testing |
| `v1.2.3` | Git tag | Specific version | Production |
| `v1.2` | Auto from tag | Minor version | Production |
| `v1` | Auto from tag | Major version | Production |
| `build-N` | Every build | Rapid testing | Development |

### For End Users

**Stable Release (Recommended):**
```bash
docker pull ghcr.io/OWNER/REPO:v0.9.0
```

**Latest Stable:**
```bash
docker pull ghcr.io/OWNER/REPO:latest
```

**Specific Minor Version (auto-updates patches):**
```bash
docker pull ghcr.io/OWNER/REPO:v0.9
```

**Beta/Pre-release:**
```bash
docker pull ghcr.io/OWNER/REPO:v0.9.0-beta
```

---

## Troubleshooting

### Build Failures

**Problem:** Docker build fails with "npm ci" errors

**Solution:**
```bash
# Clear npm cache in Dockerfile or rebuild without cache
docker build --no-cache -f docker/Dockerfile .
```

**Problem:** Go build fails with "missing module"

**Solution:**
```bash
# Ensure go.mod and go.sum are up to date
go mod tidy
git add go.mod go.sum
git commit -m "chore: update go modules"
```

### Push Failures

**Problem:** "unauthorized: authentication required"

**Solution:**
```bash
# Re-authenticate
docker logout ghcr.io
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin
```

**Problem:** "denied: permission_denied"

**Solution:**
- Verify your GitHub token has `write:packages` permission
- Ensure you have push access to the repository
- Check package visibility settings

### Runtime Issues

**Problem:** Container starts but `/health` returns 404

**Solution:**
```bash
# Check logs
docker logs actalog

# Verify environment variables
docker exec actalog env | grep DB_

# Check if migrations ran
docker exec actalog ls -la /app/migrations
```

**Problem:** Database connection errors

**Solution:**
```bash
# For SQLite
docker exec actalog ls -la /app/data/

# For PostgreSQL
docker exec actalog ping postgres
docker compose logs postgres
```

### Image Pull Issues

**Problem:** "pull access denied"

**Solution:**
```bash
# For private repositories, authenticate first
docker login ghcr.io

# Or make package public (see Publishing section)
```

**Problem:** "manifest unknown"

**Solution:**
```bash
# Verify tag exists
gh api /user/packages/container/actalog/versions

# Or check at: https://github.com/OWNER/REPO/pkgs/container/actalog
```

---

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [GitHub Container Registry Docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [GitHub Actions Docker Build](https://docs.docker.com/build/ci/github-actions/)
- [Docker Multi-platform Builds](https://docs.docker.com/build/building/multi-platform/)

---

## Summary Commands Cheat Sheet

```bash
# Build locally
./docker/scripts/build.sh v0.9.0

# Push to registry
./docker/scripts/push.sh v0.9.0

# Pull specific build for testing
docker pull ghcr.io/OWNER/REPO:build-62

# Deploy with Docker Compose
cd docker && docker compose up -d

# Create stable release
git tag -a v0.9.0 -m "Release v0.9.0"
git push origin v0.9.0

# Pull stable release
docker pull ghcr.io/OWNER/REPO:v0.9.0
```
