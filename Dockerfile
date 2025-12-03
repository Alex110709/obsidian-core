# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod ./
# COPY go.sum ./ # go.sum might not exist yet if not fully tidy, but we ran tidy earlier.

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o obsidiand cmd/obsidiand/main.go

# Final Stage
FROM alpine:latest

WORKDIR /root/

# Install runtime dependencies including Tor
RUN apk --no-cache add ca-certificates tor

# Copy binary from builder
COPY --from=builder /app/obsidiand .

# Make binary executable and create Tor data directory
RUN chmod +x ./obsidiand && \
    mkdir -p /var/lib/tor && \
    chown -R tor:tor /var/lib/tor && \
    chmod 700 /var/lib/tor

# Set environment variables for Tor
ENV DATA_DIR=/var/lib

# Expose ports
EXPOSE 8333
EXPOSE 8545
EXPOSE 3333
EXPOSE 9050

# Run the binary
CMD ["./obsidiand"]
