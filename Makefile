# DailyLog MCP - Development Automation
# Model Context Protocol server for daily activity logging

.PHONY: help setup build test clean install uninstall dev.iterate dev.clean-iterate info.status version.sync version.validate version.bump-patch version.bump-minor version.bump-major config.cursor config.vscode config.show setup.env setup.github-repo setup.complete validate.config validate.github build.cli install.cli build.ci

# Default target
help:
	@echo "DailyLog MCP - Development Commands"
	@echo "==================================="
	@echo ""
	@echo "Setup & Build:"
	@echo "  setup           Initialize development environment"
	@echo "  build           Build all components for current platform"
	@echo "  build.server    Build MCP server for current platform"
	@echo "  build.cli       Build dailyctl CLI for current platform"
	@echo "  build.ci        Build for all platforms/architectures (CI-style via GoReleaser)"
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
	@echo "  install.cli     Install dailyctl binary only"
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
	@echo "Configuration:"
	@echo "  config.cursor   Configure MCP server for Cursor IDE"
	@echo "  config.vscode   Configure MCP server for VS Code"
	@echo "  config.show     Show current MCP configuration"
	@echo ""
	@echo "Setup:"
	@echo "  setup.env       Set up environment variables"
	@echo "  setup.github-repo Create backing store GitHub repository"
	@echo "  setup.complete  Complete setup workflow (env + repo + config)"
	@echo ""
	@echo "Validation:"
	@echo "  validate.config Validate current configuration"
	@echo "  validate.github Test GitHub repository connectivity"
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
build: build.server build.cli

# Build MCP server
build.server:
	@echo "[BUILD] Building MCP server..."
	@mkdir -p build/bin
	@cd cmd/mcp-server && go build $(LDFLAGS) -o ../../build/bin/dailylog .
	@echo "[SUCCESS] MCP server built: build/bin/dailylog"

# Build dailyctl CLI
build.cli:
	@echo "[BUILD] Building dailyctl CLI..."
	@mkdir -p build/bin
	@cd cmd/dailyctl && go build $(LDFLAGS) -o ../../build/bin/dailyctl .
	@echo "[SUCCESS] dailyctl CLI built: build/bin/dailyctl"

# Build all architectures (CI-style via GoReleaser)
build.ci:
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
install: install.server install.cli

install.server:
	@echo "[INSTALL] Installing DailyLog MCP server..."
	@if [ -f build/bin/dailylog ]; then \
		echo "[INSTALL] Installing MCP server to /usr/local/bin/dailylog (requires sudo)..."; \
		sudo cp build/bin/dailylog /usr/local/bin/dailylog && echo "[SUCCESS] MCP server installed to /usr/local/bin/dailylog"; \
	else \
		echo "[INFO] MCP server not built - run 'make build.server' first"; \
	fi

install.cli:
	@echo "[INSTALL] Installing dailyctl CLI..."
	@if [ -f build/bin/dailyctl ]; then \
		echo "[INSTALL] Installing dailyctl to /usr/local/bin/dailyctl (requires sudo)..."; \
		sudo cp build/bin/dailyctl /usr/local/bin/dailyctl && echo "[SUCCESS] dailyctl installed to /usr/local/bin/dailyctl"; \
	else \
		echo "[INFO] dailyctl not built - run 'make build.cli' first"; \
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

# Configuration management
config.cursor:
	@echo "[CONFIG] Setting up Cursor MCP configuration..."
	@mkdir -p ~/.config/cursor/mcp 2>/dev/null || mkdir -p ~/Library/Application\ Support/Cursor/User/globalStorage/mcp 2>/dev/null || true
	@if [ -z "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_REPO environment variable not set"; \
		echo "Please run 'make setup.env' first"; \
		exit 1; \
	fi
	@if [ -z "$$DAILYLOG_GITHUB_TOKEN" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_TOKEN environment variable not set"; \
		echo "Please run 'make setup.env' first"; \
		exit 1; \
	fi
	@cat > /tmp/cursor-mcp-config.json << EOF && \
	{ \
	  "mcpServers": { \
	    "dailylog": { \
	      "command": "/usr/local/bin/dailylog", \
	      "args": [], \
	      "env": { \
	        "DAILYLOG_GITHUB_REPO": "$$DAILYLOG_GITHUB_REPO", \
	        "DAILYLOG_GITHUB_TOKEN": "$$DAILYLOG_GITHUB_TOKEN", \
	        "DAILYLOG_GITHUB_PATH": "$${DAILYLOG_GITHUB_PATH:-logs}" \
	      } \
	    } \
	  } \
	} \
	EOF
	@echo "[CONFIG] Generated Cursor MCP configuration:"
	@cat /tmp/cursor-mcp-config.json
	@echo ""
	@echo "[CONFIG] To apply this configuration:"
	@echo "1. Copy the JSON above to your Cursor MCP settings"
	@echo "2. Or place it in: ~/.config/cursor/mcp/config.json"
	@echo "3. Restart Cursor IDE"
	@echo ""
	@echo "[SUCCESS] Cursor configuration ready"

config.vscode:
	@echo "[CONFIG] Setting up VS Code MCP configuration..."
	@if [ -z "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_REPO environment variable not set"; \
		echo "Please run 'make setup.env' first"; \
		exit 1; \
	fi
	@if [ -z "$$DAILYLOG_GITHUB_TOKEN" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_TOKEN environment variable not set"; \
		echo "Please run 'make setup.env' first"; \
		exit 1; \
	fi
	@cat > /tmp/vscode-settings.json << EOF && \
	{ \
	  "mcp.servers": { \
	    "dailylog": { \
	      "command": "/usr/local/bin/dailylog", \
	      "args": [], \
	      "env": { \
	        "DAILYLOG_GITHUB_REPO": "$$DAILYLOG_GITHUB_REPO", \
	        "DAILYLOG_GITHUB_TOKEN": "$$DAILYLOG_GITHUB_TOKEN", \
	        "DAILYLOG_GITHUB_PATH": "$${DAILYLOG_GITHUB_PATH:-logs}" \
	      } \
	    } \
	  } \
	} \
	EOF
	@echo "[CONFIG] Generated VS Code MCP configuration:"
	@cat /tmp/vscode-settings.json
	@echo ""
	@echo "[CONFIG] To apply this configuration:"
	@echo "1. Open VS Code Settings (JSON)"
	@echo "2. Add the above configuration to your settings.json"
	@echo "3. Restart VS Code"
	@echo ""
	@echo "[SUCCESS] VS Code configuration ready"

config.show:
	@echo "[CONFIG] Current MCP Configuration"
	@echo "=================================="
	@echo ""
	@echo "Environment Variables:"
	@if [ -n "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "  ✓ DAILYLOG_GITHUB_REPO: $$DAILYLOG_GITHUB_REPO"; \
	else \
		echo "  ✗ DAILYLOG_GITHUB_REPO: Not set"; \
	fi
	@if [ -n "$$DAILYLOG_GITHUB_TOKEN" ]; then \
		echo "  ✓ DAILYLOG_GITHUB_TOKEN: ••••••••$${DAILYLOG_GITHUB_TOKEN##*????????????}"; \
	else \
		echo "  ✗ DAILYLOG_GITHUB_TOKEN: Not set"; \
	fi
	@if [ -n "$$DAILYLOG_GITHUB_PATH" ]; then \
		echo "  ✓ DAILYLOG_GITHUB_PATH: $$DAILYLOG_GITHUB_PATH"; \
	else \
		echo "  ○ DAILYLOG_GITHUB_PATH: Using default (logs)"; \
	fi
	@echo ""
	@echo "Binary Status:"
	@if [ -f /usr/local/bin/dailylog ]; then \
		echo "  ✓ MCP Server: /usr/local/bin/dailylog"; \
		echo "  Version: $$(dailylog --version 2>/dev/null | head -1 || echo 'Unknown')"; \
	else \
		echo "  ✗ MCP Server: Not installed"; \
	fi
	@if [ -f /usr/local/bin/dailyctl ]; then \
		echo "  ✓ CLI Tool: /usr/local/bin/dailyctl"; \
		echo "  Version: $$(dailyctl version 2>/dev/null | head -1 || echo 'Unknown')"; \
	else \
		echo "  ✗ CLI Tool: Not installed"; \
	fi

# Setup automation
setup.env:
	@echo "[SETUP] Environment Variable Configuration"
	@echo "========================================="
	@echo ""
	@echo "DailyLog MCP requires the following environment variables:"
	@echo ""
	@echo "1. DAILYLOG_GITHUB_REPO - Your private GitHub repository"
	@echo "   Example: cloudygreybeard/daily-logs"
	@echo ""
	@echo "2. DAILYLOG_GITHUB_TOKEN - GitHub Personal Access Token"
	@echo "   Required permissions: repo (full repository access)"
	@echo ""
	@echo "3. DAILYLOG_GITHUB_PATH - Path within repository (optional)"
	@echo "   Default: logs"
	@echo ""
	@read -p "Enter your GitHub repository (owner/repo): " repo && \
	read -p "Enter your GitHub token: " token && \
	read -p "Enter GitHub path [logs]: " path && \
	path=$${path:-logs} && \
	echo "" && \
	echo "Add these to your shell profile (~/.bashrc, ~/.zshrc, etc.):" && \
	echo "" && \
	echo "export DAILYLOG_GITHUB_REPO=\"$$repo\"" && \
	echo "export DAILYLOG_GITHUB_TOKEN=\"$$token\"" && \
	echo "export DAILYLOG_GITHUB_PATH=\"$$path\"" && \
	echo "" && \
	echo "Then run: source ~/.bashrc (or ~/.zshrc)" && \
	echo ""

setup.github-repo:
	@echo "[SETUP] GitHub Repository Setup"
	@echo "==============================="
	@echo ""
	@if [ -z "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_REPO environment variable not set"; \
		echo "Please run 'make setup.env' first"; \
		exit 1; \
	fi
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "[ERROR] GitHub CLI (gh) not found"; \
		echo "Please install: brew install gh"; \
		exit 1; \
	fi
	@echo "Creating private repository: $$DAILYLOG_GITHUB_REPO"
	@echo "Description: Private repository for daily activity logs"
	@echo ""
	@if gh repo view $$DAILYLOG_GITHUB_REPO >/dev/null 2>&1; then \
		echo "[INFO] Repository $$DAILYLOG_GITHUB_REPO already exists"; \
	else \
		echo "[CREATE] Creating repository..."; \
		gh repo create $$DAILYLOG_GITHUB_REPO --private --description "Private repository for daily activity logs" --clone=false; \
		echo "[SUCCESS] Repository created: https://github.com/$$DAILYLOG_GITHUB_REPO"; \
	fi
	@echo ""
	@echo "[SETUP] Initializing repository structure..."
	@mkdir -p /tmp/dailylog-setup && cd /tmp/dailylog-setup && \
	git init && \
	mkdir -p logs && \
	echo "# Daily Activity Logs" > README.md && \
	echo "" >> README.md && \
	echo "This repository stores daily activity logs generated by DailyLog MCP." >> README.md && \
	echo "" >> README.md && \
	echo "## Structure" >> README.md && \
	echo "" >> README.md && \
	echo "- \`logs/YYYY/MM/YYYY-MM-DD.json\` - Daily log files" >> README.md && \
	echo "" >> README.md && \
	echo "## Privacy" >> README.md && \
	echo "" >> README.md && \
	echo "This is a private repository containing personal daily activity data." >> README.md && \
	echo "logs/" > .gitignore && \
	echo "*.log" >> .gitignore && \
	echo "*.tmp" >> .gitignore && \
	git add . && \
	git commit -m "Initial commit: Daily logs repository structure" && \
	git remote add origin https://github.com/$$DAILYLOG_GITHUB_REPO.git && \
	git push -u origin main && \
	cd - && rm -rf /tmp/dailylog-setup
	@echo "[SUCCESS] Repository $$DAILYLOG_GITHUB_REPO is ready for daily logs"

setup.complete: setup.env setup.github-repo build install config.cursor
	@echo ""
	@echo "[COMPLETE] DailyLog MCP Setup Complete!"
	@echo "======================================"
	@echo ""
	@echo "✓ Environment variables configured"
	@echo "✓ GitHub repository created and initialized"
	@echo "✓ MCP server and CLI built and installed"
	@echo "✓ Cursor MCP configuration generated"
	@echo ""
	@echo "Next steps:"
	@echo "1. Restart Cursor IDE"
	@echo "2. Test with: @dailylog dailylog_entry"
	@echo "3. Or use CLI: dailyctl log activity \"Setup complete\" --mood 9"
	@echo ""
	@echo "For help: make help"

# Validation
validate.config:
	@echo "[VALIDATE] Configuration Validation"
	@echo "==================================="
	@echo ""
	@if [ -z "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "✗ DAILYLOG_GITHUB_REPO not set"; \
		exit 1; \
	else \
		echo "✓ DAILYLOG_GITHUB_REPO: $$DAILYLOG_GITHUB_REPO"; \
	fi
	@if [ -z "$$DAILYLOG_GITHUB_TOKEN" ]; then \
		echo "✗ DAILYLOG_GITHUB_TOKEN not set"; \
		exit 1; \
	else \
		echo "✓ DAILYLOG_GITHUB_TOKEN: Set"; \
	fi
	@if [ ! -f /usr/local/bin/dailylog ]; then \
		echo "✗ MCP server not installed"; \
		echo "  Run: make install"; \
		exit 1; \
	else \
		echo "✓ MCP server: /usr/local/bin/dailylog"; \
	fi
	@if [ ! -f /usr/local/bin/dailyctl ]; then \
		echo "✗ CLI tool not installed"; \
		echo "  Run: make install"; \
		exit 1; \
	else \
		echo "✓ CLI tool: /usr/local/bin/dailyctl"; \
	fi
	@echo ""
	@echo "[SUCCESS] Configuration is valid"

validate.github:
	@echo "[VALIDATE] GitHub Repository Connectivity"
	@echo "========================================"
	@echo ""
	@if [ -z "$$DAILYLOG_GITHUB_REPO" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_REPO not set"; \
		exit 1; \
	fi
	@if [ -z "$$DAILYLOG_GITHUB_TOKEN" ]; then \
		echo "[ERROR] DAILYLOG_GITHUB_TOKEN not set"; \
		exit 1; \
	fi
	@echo "Testing connection to: $$DAILYLOG_GITHUB_REPO"
	@if curl -s -H "Authorization: token $$DAILYLOG_GITHUB_TOKEN" \
		"https://api.github.com/repos/$$DAILYLOG_GITHUB_REPO" >/dev/null 2>&1; then \
		echo "✓ Repository accessible"; \
		echo "✓ Token has valid permissions"; \
	else \
		echo "✗ Cannot access repository"; \
		echo "  Check repository name and token permissions"; \
		exit 1; \
	fi
	@echo ""
	@echo "[SUCCESS] GitHub connectivity validated"

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
