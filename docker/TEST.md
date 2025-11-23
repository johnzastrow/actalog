# Testing Docker Build Locally

## Fix Docker Permissions (One-time setup)

You need to add your user to the docker group:

```bash
sudo usermod -aG docker $USER
newgrp docker  # Or logout/login to apply group changes
```

Verify Docker access:
```bash
docker ps
# Should not show permission denied
```

## Test 1: Build the Image

```bash
cd /home/jcz/Github/actionlog

# Build using helper script
./docker/scripts/build.sh test

# Or build manually
docker build -f docker/Dockerfile -t ghcr.io/yourusername/actalog:test .
```

**Expected output:**
- Frontend build completes successfully
- Backend build completes successfully  
- Final image size ~40-60MB
- Tags: `ghcr.io/yourusername/actalog:test` and `ghcr.io/yourusername/actalog:build-62`

## Test 2: Run the Container

```bash
# Run with SQLite (default)
docker run -d \
  --name actalog-test \
  -p 9080:8080 \
  -e JWT_SECRET=test_secret_12345 \
  -e CORS_ORIGINS=http://localhost:9080 \
  ghcr.io/yourusername/actalog:test

# View logs
docker logs -f actalog-test
```

**Expected in logs:**
```
Starting ActaLog...
Database driver: sqlite3
Running migrations...
Migrations completed successfully
Server listening on :8080
```

## Test 3: Verify Health

```bash
# Health check
curl http://localhost:9080/health
# Expected: {"status":"healthy"}

# Version endpoint
curl http://localhost:9080/api/version
# Expected: {"version":"0.9.0-beta","build":62,...}

# Frontend (open in browser)
open http://localhost:9080
# Should show ActaLog login page
```

## Test 4: Docker Compose

```bash
cd /home/jcz/Github/actionlog/docker

# Create .env
cp .env.example .env

# Edit .env - set these:
# GITHUB_OWNER=yourusername
# TAG=test
# JWT_SECRET=your_secure_secret_here

# Start services
docker compose up -d

# View logs
docker compose logs -f actalog

# Test health
curl http://localhost:8080/health

# Stop services
docker compose down
```

## Test 5: Multi-platform Build (Optional)

This tests that the image can build for both amd64 and arm64:

```bash
# Create buildx builder if not exists
docker buildx create --name multiplatform --use

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f docker/Dockerfile \
  -t ghcr.io/yourusername/actalog:multiplatform \
  --load \
  .
```

## Cleanup

```bash
# Stop and remove test container
docker stop actalog-test
docker rm actalog-test

# Remove test images
docker rmi ghcr.io/yourusername/actalog:test
docker rmi ghcr.io/yourusername/actalog:build-62

# Docker compose cleanup
cd docker
docker compose down -v  # -v removes volumes too
```

## Common Issues

### Issue: "permission denied" when building

**Solution:**
```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Issue: Frontend build fails with npm errors

**Solution:** The Dockerfile uses `npm ci --only=production`. Make sure `web/package-lock.json` is up to date:
```bash
cd web
npm install
git add package-lock.json
```

### Issue: Backend build fails with "missing module"

**Solution:**
```bash
go mod tidy
git add go.mod go.sum
```

### Issue: Container starts but health check fails

**Solution:** Check logs:
```bash
docker logs actalog-test
```

Common causes:
- Database migration errors
- Missing JWT_SECRET environment variable
- Port already in use

### Issue: Can't access frontend in browser

**Solution:**
- Check CORS_ORIGINS includes your access URL
- Verify port mapping: `-p 9080:8080` (host:container)
- Check if port is already in use: `lsof -i:9080`

## Success Criteria

All tests pass if:

1. ✅ Docker build completes without errors
2. ✅ Image size is reasonable (~40-60MB)
3. ✅ Container starts and shows "Server listening on :8080" in logs
4. ✅ Health endpoint returns `{"status":"healthy"}`
5. ✅ Version endpoint returns correct version
6. ✅ Frontend loads in browser (login page visible)
7. ✅ Can register and login successfully
8. ✅ Docker Compose deployment works

## Next Steps After Successful Testing

1. **Update placeholders:**
   - Replace `yourusername` in docker-compose.yml
   - Update GITHUB_OWNER in .env.example

2. **Push to GitHub:**
   ```bash
   git add docker/ .dockerignore .github/workflows/docker-build.yml
   git commit -m "feat: add Docker deployment with GitHub Container Registry"
   git push
   ```

3. **GitHub Actions will automatically:**
   - Build multi-platform image
   - Tag with `main-SHA`, `build-N`, `latest`
   - Push to ghcr.io

4. **Create first release:**
   ```bash
   git tag -a v0.9.0-beta -m "Docker deployment release"
   git push origin v0.9.0-beta
   ```
