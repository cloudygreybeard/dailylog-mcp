# Multi-stage Dockerfile for DailyLog MCP
# Stage 1: Build using official Go image
FROM golang:1.24 AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o dailylog ./cmd/mcp-server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o dailyctl ./cmd/dailyctl

# Stage 2: Minimal runtime image (FROM scratch)
# Note: For HTTPS/TLS, we could use gcr.io/distroless/static instead
# which includes CA certificates, but scratch is the absolute minimum
FROM scratch

# Copy CA certificates from builder for HTTPS/TLS support
# (needed for GitHub API calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /build/dailylog /usr/local/bin/dailylog
COPY --from=builder /build/dailyctl /usr/local/bin/dailyctl

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/dailylog"]

