# Build Stage
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y git && \
    rm -rf /var/lib/apt/lists/*

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Run tests
RUN go test ./...

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o obsidiand cmd/obsidiand/main.go

# Final Stage
FROM alpine:latest

# Create non-root user
RUN addgroup -g 1000 obsidian && \
    adduser -D -u 1000 -G obsidian obsidian

WORKDIR /home/obsidian

# Install runtime dependencies including Tor
RUN apk --no-cache add ca-certificates tor tini && \
    mkdir -p /var/lib/tor /home/obsidian/data /home/obsidian/logs && \
    chown -R tor:tor /var/lib/tor && \
    chmod 700 /var/lib/tor && \
    chown -R obsidian:obsidian /home/obsidian

# Copy binary from builder
COPY --from=builder /app/obsidiand /usr/local/bin/obsidiand

# Make binary executable
RUN chmod +x /usr/local/bin/obsidiand

# Switch to non-root user
USER obsidian

# Set environment variables
ENV DATA_DIR=/home/obsidian/data \
    LOG_FILE=/home/obsidian/logs/obsidian.log \
    LOG_LEVEL=info \
    NETWORK=mainnet

# Expose ports
EXPOSE 8333 8545 3333 9050

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8545/health || exit 1

# Use tini as init system for proper signal handling
ENTRYPOINT ["/sbin/tini", "--"]

# Run the binary
CMD ["/usr/local/bin/obsidiand"]
