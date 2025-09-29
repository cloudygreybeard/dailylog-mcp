# DailyLog MCP - Development Automation
# Model Context Protocol server for daily activity logging

.PHONY: help setup build test clean install uninstall dev.iterate dev.clean-iterate info.status version.sync version.validate version.bump-patch version.bump-minor version.bump-major

# Default target
help:
	@echo "DailyLog MCP - Development Commands"
	@echo "==================================="
	@echo ""
	@echo "Setup & Build:"
	@echo "  setup           Initialize development environment"
	@echo "  build           Build all components"
	@echo "  build.server    Build MCP server for current platform"
	@echo "  build.dailyctl  Build dailyctl CLI for current platform"
	@echo "  build.all       Build all architectures (requires GoReleaser)"
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev.iterate     Full build and install cycle"
	@echo "  dev.clean-iterate  Clean uninstall -> build -> install cycle"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run all tests"
	@echo "  test.unit       Run unit tests"
	@echo "  test.integration Run integration tests"
	@echo "  test.coverage   Run tests with coverage"
	@echo ""
	@echo "Installation:"
	@echo "  install         Install MCP server and dailyctl locally"
	@echo "  install.server  Install MCP server binary only"
	@echo "  install.dailyctl Install dailyctl binary only"
	@echo "  uninstall       Remove all binaries"
	@echo ""
	@echo "Release:"
	@echo "  release         Create release with GoReleaser"
	@echo "  release.snapshot Create snapshot release"
	@echo "  release.validate Run release validation"
	@echo ""
	@echo "Version Management:"
	@echo "  version.sync    Sync all versions from git tag"
	@echo "  version.validate Check version consistency across components"
	@echo "  version.bump-patch Bump patch version with git tag"
	@echo "  version.bump-minor Bump minor version with git tag"
	@echo "  version.bump-major Bump major version with git tag"
	@echo ""
	@echo "Information:"
	@echo "  info.status     Show project status"

# Get version from git tag for build ldflags
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Setup development environment
setup:
	@echo "[SETUP] Initializing DailyLog MCP development environment..."
	@mkdir -p build/bin hack docs
	@echo "[SETUP] Created directory structure"
	@if [ ! -f .gitignore ]; then \
		echo "# DailyLog MCP" > .gitignore; \
		echo "build/" >> .gitignore; \
		echo "*.log" >> .gitignore; \
		echo "*.tmp" >> .gitignore; \
		echo "*.temp" >> .gitignore; \
		echo "coverage.out" >> .gitignore; \
		echo "coverage.html" >> .gitignore; \
		echo ".DS_Store" >> .gitignore; \
		echo "config.yaml" >> .gitignore; \
		echo ".env" >> .gitignore; \
		echo "[SETUP] Updated .gitignore"; \
	fi
	@echo "[SUCCESS] Development environment ready"

# Build all components
build: build.server build.dailyctl

# Build MCP server
build.server:
	@echo "[BUILD] Building MCP server..."
	@mkdir -p build/bin
	@cd cmd/mcp-server && go build $(LDFLAGS) -o ../../build/bin/dailylog .
	@echo "[SUCCESS] MCP server built: build/bin/dailylog"

# Build dailyctl CLI
build.dailyctl:
	@echo "[BUILD] Building dailyctl CLI..."
	@mkdir -p build/bin
	@cd cmd/dailyctl && go build $(LDFLAGS) -o ../../build/bin/dailyctl .
	@echo "[SUCCESS] dailyctl CLI built: build/bin/dailyctl"

# Build all architectures (requires GoReleaser)
build.all:
	@echo "[BUILD] Building for all architectures..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser build --snapshot --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Multi-architecture builds completed"

# Development iteration cycle
dev.iterate: build install
	@echo "[DEV] Development iteration complete!"

# Clean development iteration cycle (uninstall -> build -> install)
dev.clean-iterate: uninstall build install
	@echo "[DEV] Clean development iteration complete!"

# Install MCP server and dailyctl locally
install: install.server install.dailyctl

install.server:
	@echo "[INSTALL] Installing DailyLog MCP server..."
	@if [ -f build/bin/dailylog ]; then \
		echo "[INSTALL] Installing MCP server to /usr/local/bin/dailylog (requires sudo)..."; \
		sudo cp build/bin/dailylog /usr/local/bin/dailylog && echo "[SUCCESS] MCP server installed to /usr/local/bin/dailylog"; \
	else \
		echo "[INFO] MCP server not built - run 'make build.server' first"; \
	fi

install.dailyctl:
	@echo "[INSTALL] Installing dailyctl CLI..."
	@if [ -f build/bin/dailyctl ]; then \
		echo "[INSTALL] Installing dailyctl to /usr/local/bin/dailyctl (requires sudo)..."; \
		sudo cp build/bin/dailyctl /usr/local/bin/dailyctl && echo "[SUCCESS] dailyctl installed to /usr/local/bin/dailyctl"; \
	else \
		echo "[INFO] dailyctl not built - run 'make build.dailyctl' first"; \
	fi

# Uninstall binaries
uninstall:
	@echo "[UNINSTALL] Removing DailyLog MCP binaries..."
	@if [ -f /usr/local/bin/dailylog ]; then \
		echo "[UNINSTALL] Removing MCP server from /usr/local/bin/dailylog (requires sudo)..."; \
		sudo rm -f /usr/local/bin/dailylog && echo "[SUCCESS] MCP server removed"; \
	else \
		echo "[INFO] MCP server not installed at /usr/local/bin/dailylog"; \
	fi
	@if [ -f /usr/local/bin/dailyctl ]; then \
		echo "[UNINSTALL] Removing dailyctl from /usr/local/bin/dailyctl (requires sudo)..."; \
		sudo rm -f /usr/local/bin/dailyctl && echo "[SUCCESS] dailyctl removed"; \
	else \
		echo "[INFO] dailyctl not installed at /usr/local/bin/dailyctl"; \
	fi

# Testing
test: test.unit test.integration

test.unit:
	@echo "[TEST] Running unit tests..."
	@go test -v ./...

test.integration:
	@echo "[TEST] Running integration tests..."
	@if [ -f hack/test-integration.sh ]; then ./hack/test-integration.sh; else echo "[INFO] No integration tests configured"; fi

test.coverage:
	@echo "[TEST] Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "[INFO] Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "[CLEAN] Removing build artifacts..."
	@rm -rf build/ 2>/dev/null || true
	@rm -f coverage.out coverage.html 2>/dev/null || true
	@echo "[SUCCESS] Build artifacts cleaned"

# Release with GoReleaser
release:
	@echo "[RELEASE] Creating release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Release completed"

# Release snapshot for testing
release.snapshot:
	@echo "[RELEASE] Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "[ERROR] GoReleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@echo "[SUCCESS] Snapshot release completed"

release.validate:
	@echo "[VALIDATE] Running release validation..."
	@if [ -f hack/validate-release.sh ]; then ./hack/validate-release.sh; else echo "[INFO] No release validation configured"; fi

# Information commands
info.status:
	@echo "DailyLog MCP Project Status"
	@echo "==========================="
	@echo ""
	@echo "Architecture: Model Context Protocol server for daily activity logging"
	@echo "Components:"
	@if [ -f build/bin/dailylog ]; then echo "  ✓ MCP server (built)"; else echo "  ✗ MCP server (not built)"; fi
	@if [ -f build/bin/dailyctl ]; then echo "  ✓ dailyctl CLI (built)"; else echo "  ✗ dailyctl CLI (not built)"; fi
	@echo ""
	@echo "Installation Status:"
	@if [ -f /usr/local/bin/dailylog ]; then \
		echo "  ✓ MCP server: /usr/local/bin/dailylog"; \
		echo "  Version: $$(dailylog --version 2>/dev/null | head -1 || echo 'Unknown')"; \
	else \
		echo "  ✗ MCP server not installed"; \
	fi
	@if [ -f /usr/local/bin/dailyctl ]; then \
		echo "  ✓ dailyctl CLI: /usr/local/bin/dailyctl"; \
		echo "  Version: $$(dailyctl version 2>/dev/null | head -1 || echo 'Unknown')"; \
	else \
		echo "  ✗ dailyctl CLI not installed"; \
	fi
	@echo ""
	@echo "Configuration:"
	@if [ -n "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "  GitHub Repo: $$DAILYLOG_GITHUB_REPO"; \
	else \
		echo "  ✗ GitHub repository not configured (set DAILYLOG_GITHUB_REPO)"; \
	fi
	@if [ -n "$$DAILYLOG_GITHUB_TOKEN" ]; then \
		echo "  ✓ GitHub token configured"; \
	else \
		echo "  ✗ GitHub token not configured (set DAILYLOG_GITHUB_TOKEN)"; \
	fi

# Git-based version management
version.sync:
	@./hack/version sync

version.validate:
	@./hack/version validate

version.bump-patch:
	@./hack/version bump patch

version.bump-minor:
	@./hack/version bump minor

version.bump-major:
	@./hack/version bump major
