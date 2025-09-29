# DailyLog MCP

A Model Context Protocol (MCP) server implementation for daily activity logging with GitHub storage.

## Overview

DailyLog MCP provides daily activity logging through MCP integration in AI assistants and command-line tools. Activities, status updates, notes, and summaries are stored in private GitHub repositories using structured JSON format.

**Status**: Functional MCP server with automated setup and configuration tools.

## Features

- **MCP Integration**: Protocol implementation for Cursor IDE and VS Code
- **Activity Tracking**: Log activities, status updates, notes, summaries with metadata
- **Search**: Text search with filtering by type, tags, status, dates
- **Summaries**: Generate daily, weekly, monthly summaries
- **GitHub Storage**: File-based storage in private GitHub repositories
- **CLI Tools**: Command-line interface for all operations
- **Structured Data**: JSON-based storage with timestamps, tags, and metadata
- **Cross-platform**: Support for macOS, Linux, and Windows

## Architecture

### Components

- **MCP Server** (`dailylog`): JSON-RPC 2.0 server implementing the Model Context Protocol
- **CLI Tool** (`dailyctl`): Human-friendly command-line interface for daily logging
- **GitHub Storage**: Private repository backend with structured JSON files
- **AI Integration**: Framework for AI-assisted features (summarization, insights)

### MCP Tools

The MCP server implements daily logging tools:

**Core Logging:**
- `dailylog_entry` - Create new daily log entries (activities, status updates, notes, summaries)
- `dailylog_get_entries` - Retrieve entries for specific dates or ranges
- `dailylog_search` - Search through logs by text, tags, status, or criteria
- `dailylog_summarize` - Generate summaries for daily, weekly, monthly periods
- `dailylog_ai_assist` - AI assistance for wording, tags, status analysis, insights

## Demo

![DailyLog MCP Demo](docs/dailylog-demo.svg)

*Interactive demo showing daily logging workflow for SRE teams - logging activities, searching entries, and generating standup reports.*

### Data Structure

Daily logs are stored in structured JSON format:

```json
{
  "date": "2025-09-29",
  "entries": [
    {
      "id": "entry_1727612345000",
      "timestamp": "2025-09-29T10:30:00Z",
      "type": "activity",
      "title": "Team standup meeting",
      "description": "Weekly team sync and planning session",
      "tags": ["work", "meeting", "team"],
      "status": 8,
      "priority": 3,
      "duration": 30,
      "location": "Office Conference Room A",
      "metadata": {
        "project": "daily-logging",
        "participants": "5"
      }
    }
  ],
  "day_summary": "Productive day with good team collaboration",
  "status_average": 7.5,
  "total_entries": 8,
  "created_at": "2025-09-29T08:00:00Z",
  "updated_at": "2025-09-29T18:30:00Z"
}
```

**Manual Installation:**
```bash
# Download latest release
curl -L https://github.com/cloudygreybeard/dailylog-mcp/releases/latest/download/dailylog_Linux_x86_64.tar.gz | tar xz
sudo mv dailylog dailyctl /usr/local/bin/
```

**Build from Source:**
```bash
git clone https://github.com/cloudygreybeard/dailylog-mcp.git
cd dailylog-mcp
make setup.complete  # Complete setup with environment and MCP configuration
```

### Setup

1. Create a private GitHub repository for daily logs
2. Generate a GitHub Personal Access Token with `repo` permissions
3. Configure environment variables:

```bash
export DAILYLOG_GITHUB_REPO="cloudygreybeard/daily-logs"
export DAILYLOG_GITHUB_TOKEN="ghp_your_github_token_here"
export DAILYLOG_GITHUB_PATH="logs"
```

### MCP Configuration

**Cursor IDE Configuration:**
```json
{
  "mcpServers": {
    "dailylog": {
      "command": "dailylog",
      "args": [],
      "env": {
        "DAILYLOG_GITHUB_REPO": "cloudygreybeard/daily-logs",
        "DAILYLOG_GITHUB_TOKEN": "ghp_your_token_here"
      }
    }
  }
}
```

## Usage Examples

### MCP Integration (Cursor/VS Code)

Use DailyLog in AI conversations:

```bash
# Log an activity
@dailylog dailylog_entry {
  "type": "activity",
  "title": "Morning workout",
  "description": "30-minute cardio session",
  "tags": ["health", "exercise"],
  "status": 8,
  "duration": 30
}

# Get today's entries
@dailylog dailylog_get_entries {
  "date": "2025-09-29",
  "include_stats": true
}

# Search for exercise activities
@dailylog dailylog_search {
  "query": "exercise",
  "tags": ["health"],
  "status_min": 7
}

# Generate a weekly summary
@dailylog dailylog_summarize {
  "type": "week",
  "use_ai": true
}
```

### CLI Usage

**Log Activities:**
```bash
# Log different types of entries
dailyctl log activity "Team meeting" --tags work,meeting --status 8 --duration 60
dailyctl log status "Feeling energetic" --status 9
dailyctl log note "Remember to call dentist" --priority 3
dailyctl log summary "Productive day overall"
```

**Retrieve Entries:**
```bash
# Get entries
dailyctl get today
dailyctl get yesterday
dailyctl get week
dailyctl get month
dailyctl get date 2025-09-29
dailyctl get --date-start 2025-09-01 --date-end 2025-09-30 --stats
```

**Search Logs:**
```bash
# Search examples
dailyctl search --query "exercise"
dailyctl search --tags work,meeting
dailyctl search --status-min 8 --status-max 10
dailyctl search --type activity --date-start 2025-09-01
```

**Generate Summaries:**
```bash
# Summary examples
dailyctl summarize day
dailyctl summarize week --ai
dailyctl summarize month --save
dailyctl summarize custom --date-start 2025-09-01 --date-end 2025-09-30
```

## Storage Structure

Your GitHub repository will be organized as:

```
daily-logs/
├── logs/
│   ├── 2025/
│   │   ├── 09/
│   │   │   ├── 2025-09-29.json
│   │   │   ├── 2025-09-30.json
│   │   │   └── ...
│   │   └── 10/
│   └── 2024/
└── README.md
```

## Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/cloudygreybeard/dailylog-mcp.git
cd dailylog-mcp

# Build all components
make build

# Install locally
make install

# Run tests
make test

# Development iteration
make dev.iterate
```

### Project Structure

```
dailylog-mcp/
├── cmd/
│   ├── mcp-server/          # MCP server implementation
│   └── dailyctl/            # CLI tool
├── internal/
│   ├── storage/             # Storage interfaces and models
│   ├── providers/           # GitHub storage provider
│   └── ai/                  # AI integration (future)
├── docs/                    # Documentation
│   ├── examples/            # Configuration examples
│   └── INSTALLATION.md     # Installation guide
├── .github/workflows/       # CI/CD automation
├── Makefile                 # Development automation
├── .goreleaser.yaml        # Release configuration
└── go.mod
```

### Available Make Targets

```bash
# Setup and Configuration
make setup.complete    # Complete setup workflow (env + repo + config)
make setup.env         # Set up environment variables
make setup.github-repo # Create backing store GitHub repository
make config.cursor     # Configure MCP server for Cursor IDE
make config.vscode     # Configure MCP server for VS Code

# Build and Install  
make build             # Build all components for current platform
make build.ci          # Build for all platforms (CI-style via GoReleaser)
make install           # Install binaries locally
make clean             # Clean build artifacts

# Validation and Status
make validate.config   # Validate current configuration
make validate.github   # Test GitHub repository connectivity
make info.status       # Show project status
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Make changes with tests
4. Commit changes using conventional commit format
5. Push to branch: `git push origin feature/new-feature`
6. Open a Pull Request

### Code Standards
- Follow Go conventions and formatting
- Use descriptive variable and function names
- Include tests for new functionality
- Update documentation for user-facing changes

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) file for details.

## Roadmap

### Current State (v1.0)
- MCP server implementation with 5 tools
- CLI tool with subcommands
- GitHub storage backend with structured JSON
- Cross-platform builds and releases
- Documentation and examples

### Planned Features (v1.1)
- AI integration for summaries
- Backup and sync options
- Template system for common entry types
- Calendar system integration
- Data export and migration tools

### Future Development (v2.0)
- Web interface for log visualization
- Mobile application
- Team collaboration features
- Analytics and insights
- Plugin system for extensibility

## Support

- Documentation: Available in the [docs](docs/) directory
- Issues: Report bugs on [GitHub Issues](https://github.com/cloudygreybeard/dailylog-mcp/issues)
- Discussions: Available in [GitHub Discussions](https://github.com/cloudygreybeard/dailylog-mcp/discussions)
