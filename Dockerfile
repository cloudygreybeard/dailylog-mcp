# Dockerfile for DailyLog MCP
# This Dockerfile is used by GoReleaser, which provides pre-built binaries
# in the build context. GoReleaser builds statically linked binaries with
# CGO_ENABLED=0, so we can use a minimal base image.

# Use Google's distroless static image for minimal size with CA certificates
# This includes CA certs needed for HTTPS/TLS (GitHub API calls)
FROM gcr.io/distroless/static:nonroot

# Copy pre-built binaries from GoReleaser build context
# GoReleaser provides these binaries in the Docker build context
COPY dailylog /usr/local/bin/dailylog
COPY dailyctl /usr/local/bin/dailyctl

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/dailylog"]

# Use non-root user (provided by distroless/static:nonroot)
USER nonroot:nonroot

