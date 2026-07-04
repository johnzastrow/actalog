# Multi-stage Dockerfile for ActaLog
# Builds both frontend and backend in a single optimized image

# ============================================================================
# Stage 1: Build Frontend (Vue.js)
# ============================================================================
FROM node:24-alpine@sha256:a0b9bf06e4e6193cf7a0f58816cc935ff8c2a908f81e6f1a95432d679c54fbfd AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package.json web/package-lock.json ./

# Pin npm to the version that generated package-lock.json. The node:24-alpine
# base image's bundled npm floats across rebuilds; a newer npm reshapes the
# rolldown wasm-fallback @emnapi optional deps and breaks `npm ci`. Keep this in
# lockstep with the "packageManager" field in web/package.json.
RUN npm i -g npm@11.16.0

# Install dependencies (including dev dependencies needed for build)
RUN npm ci

# Copy frontend source
COPY web/ ./

# Build frontend for production
RUN npm run build

# ============================================================================
# Stage 2: Build Backend (Go)
# ============================================================================
FROM golang:1.25-alpine@sha256:523c3effe300580ed375e43f43b1c9b091b68e935a7c3a92bfcc4e7ed55b18c2 AS backend-builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
# Allow Go to auto-download required toolchain version
ENV GOTOOLCHAIN=auto
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY configs/ ./configs/
COPY docs/ ./docs/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build the application
# CGO is required for SQLite support
# GOTOOLCHAIN=auto allows automatic toolchain download if needed
RUN CGO_ENABLED=1 GOOS=linux GOTOOLCHAIN=auto go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o /app/bin/actalog \
    ./cmd/actalog

# ============================================================================
# Stage 3: Runtime Image
# ============================================================================
FROM alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite \
    sqlite-libs \
    tzdata

# Create non-root user
RUN addgroup -g 1000 actalog && \
    adduser -D -u 1000 -G actalog actalog

WORKDIR /app

# Copy backend binary from builder
COPY --from=backend-builder /app/bin/actalog /app/actalog

# Copy frontend build from builder
COPY --from=frontend-builder /app/web/dist /app/web/dist

# Copy seed CSV files (movements and WODs)
COPY seeds/ /app/seeds/

# Copy seed initialization script and entrypoint
COPY docker/scripts/init-seeds.sh /app/scripts/init-seeds.sh
COPY docker/scripts/entrypoint.sh /app/scripts/entrypoint.sh

# Create directories for uploads, database, and backups
RUN mkdir -p /app/uploads /app/data /app/backups /app/scripts && \
    chmod +x /app/scripts/init-seeds.sh /app/scripts/entrypoint.sh && \
    chown -R actalog:actalog /app

# Switch to non-root user
USER actalog

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default environment variables
ENV DB_DRIVER=sqlite3 \
    DB_NAME=/app/data/actalog.db \
    PORT=8080 \
    UPLOADS_DIR=/app/uploads \
    FRONTEND_DIR=/app/web/dist

# Run the application via entrypoint (handles optional seed import)
CMD ["/app/scripts/entrypoint.sh"]
