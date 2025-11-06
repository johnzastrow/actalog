# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o actalog ./cmd/actalog

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Create non-root user
RUN addgroup -g 1000 actalog && \
    adduser -D -u 1000 -G actalog actalog

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/actalog .

# Copy migrations (if they exist)
COPY --from=builder /app/migrations ./migrations 2>/dev/null || true

# Create directories for uploads and database
RUN mkdir -p /app/uploads /app/data && \
    chown -R actalog:actalog /app

# Switch to non-root user
USER actalog

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./actalog"]
