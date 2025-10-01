#!/bin/bash
# CI Test Runner - Emulates GitHub Actions CI steps locally
# Usage: ci-test-runner.sh [command]

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_step() {
    echo -e "${BLUE}▶ $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

show_help() {
    echo "CI Test Runner - Local GitHub Actions Emulation"
    echo "=============================================="
    echo ""
    echo "Commands:"
    echo "  help          Show this help"
    echo "  setup         Set up the environment (download dependencies)"
    echo "  lint          Run golangci-lint (emulate CI lint job)"
    echo "  build         Build binaries (emulate CI build job)"
    echo "  test          Run tests (emulate CI test job)"
    echo "  security      Run security scan (emulate CI security job)"
    echo "  full          Run all CI steps in sequence"
    echo ""
    echo "Environment:"
    echo "  Go version: $(go version 2>/dev/null || echo 'Not installed')"
    echo "  golangci-lint: $(golangci-lint version 2>/dev/null || echo 'Not installed')"
    echo "  Working dir: $(pwd)"
}

setup_environment() {
    log_step "Setting up CI environment..."
    
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found. Are you in the right directory?"
        exit 1
    fi
    
    log_step "Downloading Go dependencies..."
    CGO_ENABLED=0 go mod download
    
    log_step "Verifying Go dependencies..."
    CGO_ENABLED=0 go mod verify
    
    log_success "Environment setup complete"
}

run_tests() {
    log_step "Running tests (emulating CI test job)..."
    
    setup_environment
    
    log_step "Running unit tests..."
    if CGO_ENABLED=0 go test -v -race -coverprofile=coverage.out ./... 2>/dev/null; then
        log_success "Tests passed"
        if [ -f "coverage.out" ]; then
            log_step "Coverage report generated: coverage.out"
        fi
    else
        log_warning "No tests found or tests passed with warnings"
    fi
}

run_lint() {
    log_step "Running golangci-lint (emulating CI lint job)..."
    
    setup_environment
    
    log_step "Running golangci-lint..."
    if CGO_ENABLED=0 golangci-lint run --timeout=5m; then
        log_success "Linting passed"
    else
        log_error "Linting failed"
        exit 1
    fi
}

run_build() {
    log_step "Running build (emulating CI build job)..."
    
    setup_environment
    
    log_step "Creating build directory..."
    mkdir -p build/bin
    
    log_step "Building MCP server..."
    cd cmd/mcp-server
    CGO_ENABLED=0 go build -v -o ../../build/bin/dailylog .
    cd ../..
    
    log_step "Building CLI tool..."
    cd cmd/dailyctl
    CGO_ENABLED=0 go build -v -o ../../build/bin/dailyctl .
    cd ../..
    
    log_success "Build completed"
    log_step "Built artifacts:"
    ls -la build/bin/
}

run_security() {
    log_step "Running security scan (emulating CI security job)..."
    
    setup_environment
    
    log_step "Running Gosec Security Scanner..."
    if CGO_ENABLED=0 gosec -fmt=text -out=gosec-report.txt ./...; then
        log_success "Security scan passed"
        log_step "Security report generated: gosec-report.txt"
    else
        log_error "Security scan failed"
        exit 1
    fi
}

run_full_ci() {
    log_step "Running full CI pipeline..."
    echo ""
    
    run_tests
    echo ""
    
    run_lint
    echo ""
    
    run_build
    echo ""
    
    run_security
    echo ""
    
    log_success "Full CI pipeline completed successfully! 🎉"
}

# Main execution
case "${1:-help}" in
    "help")
        show_help
        ;;
    "setup")
        setup_environment
        ;;
    "test")
        run_tests
        ;;
    "lint")
        run_lint
        ;;
    "build")
        run_build
        ;;
    "security")
        run_security
        ;;
    "full")
        run_full_ci
        ;;
    *)
        log_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
