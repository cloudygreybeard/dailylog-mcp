# Dockerfile to emulate GitHub Actions Ubuntu 24.04 environment
# for local CI testing and debugging

FROM ubuntu:24.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies matching GitHub Actions environment
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.24 (matching our CI configuration)
RUN curl -L https://go.dev/dl/go1.24.7.linux-amd64.tar.gz -o go1.24.7.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz \
    && rm go1.24.7.linux-amd64.tar.gz

# Set up Go environment
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV GOCACHE="/tmp/go-cache"
ENV GOMODCACHE="/tmp/go-mod-cache"

# Install golangci-lint (same version as CI)
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v1.64.8

# Install gosec security scanner (matching GitHub Actions)
RUN curl -sSfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b /usr/local/bin v2.22.9

# Create workspace directory
WORKDIR /workspace

# Set up entrypoint for testing
COPY hack/ci-test-runner.sh /usr/local/bin/ci-test-runner.sh
RUN chmod +x /usr/local/bin/ci-test-runner.sh

ENTRYPOINT ["/usr/local/bin/ci-test-runner.sh"]
CMD ["help"]
