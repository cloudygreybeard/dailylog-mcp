#!/bin/bash
# Local CI Testing Script using Podman
# Emulates GitHub Actions Ubuntu 24.04 environment

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONTAINER_NAME="dailylog-ci-test"
IMAGE_NAME="dailylog-ci-test:latest"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_step() {
    echo -e "${BLUE}▶ $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

show_help() {
    echo "Local CI Testing with Podman"
    echo "============================"
    echo ""
    echo "Usage: $0 [command] [ci-args...]"
    echo ""
    echo "Commands:"
    echo "  build         Build the CI test container"
    echo "  run [args]    Run CI tests (forwards args to ci-test-runner.sh)"
    echo "  shell         Open interactive shell in container"
    echo "  clean         Remove container and image"
    echo "  rebuild       Clean + build"
    echo ""
    echo "Examples:"
    echo "  $0 build                    # Build the container"
    echo "  $0 run lint                 # Run only linting"
    echo "  $0 run full                 # Run full CI pipeline"
    echo "  $0 shell                    # Interactive debugging"
    echo ""
    echo "This emulates GitHub Actions Ubuntu 24.04 + Go 1.24 environment"
}

build_container() {
    log_step "Building CI test container..."
    
    cd "$PROJECT_ROOT"
    podman build \
        -f hack/ci-test.Dockerfile \
        -t "$IMAGE_NAME" \
        .
    
    log_success "Container built: $IMAGE_NAME"
}

run_ci_test() {
    local ci_args="$*"
    
    log_step "Running CI test in container..."
    log_step "Command: ci-test-runner.sh $ci_args"
    
    cd "$PROJECT_ROOT"
    
    # Remove existing container if it exists
    podman rm -f "$CONTAINER_NAME" 2>/dev/null || true
    
    # Run the container with the project mounted
    podman run \
        --name "$CONTAINER_NAME" \
        --rm \
        -v "$PROJECT_ROOT:/workspace:Z" \
        -w /workspace \
        "$IMAGE_NAME" \
        $ci_args
}

run_shell() {
    log_step "Opening interactive shell in CI container..."
    
    cd "$PROJECT_ROOT"
    
    # Remove existing container if it exists
    podman rm -f "$CONTAINER_NAME" 2>/dev/null || true
    
    # Run interactive shell
    podman run \
        --name "$CONTAINER_NAME" \
        --rm \
        -it \
        -v "$PROJECT_ROOT:/workspace:Z" \
        -w /workspace \
        --entrypoint /bin/bash \
        "$IMAGE_NAME"
}

clean_container() {
    log_step "Cleaning up CI test container and image..."
    
    # Remove container
    podman rm -f "$CONTAINER_NAME" 2>/dev/null || true
    
    # Remove image
    podman rmi -f "$IMAGE_NAME" 2>/dev/null || true
    
    log_success "Cleanup completed"
}

# Check if podman is available
if ! command -v podman >/dev/null 2>&1; then
    echo "❌ Podman not found. Please install podman first."
    exit 1
fi

# Parse command
case "${1:-help}" in
    "build")
        build_container
        ;;
    "run")
        shift
        run_ci_test "$@"
        ;;
    "shell")
        run_shell
        ;;
    "clean")
        clean_container
        ;;
    "rebuild")
        clean_container
        build_container
        ;;
    "help"|"--help"|"-h")
        show_help
        ;;
    *)
        echo "❌ Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
