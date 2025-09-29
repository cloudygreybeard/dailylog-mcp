# DailyLog MCP Installation Guide

## Overview

Installation and configuration guide for DailyLog MCP server and CLI tools with GitHub storage backend.

## Prerequisites

- Go 1.24 or later (for building from source)
- GitHub account with a private repository for log storage
- GitHub personal access token with repository permissions
- Cursor IDE or VS Code (for MCP integration)

## Installation Methods

### Option 1: Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/cloudygreybeard/dailylog-mcp/releases):

```bash
# Download and extract for your platform
curl -L https://github.com/cloudygreybeard/dailylog-mcp/releases/latest/download/dailylog_Linux_x86_64.tar.gz | tar xz

# Move binaries to PATH
sudo mv dailylog /usr/local/bin/
sudo mv dailyctl /usr/local/bin/

# Verify installation
dailylog --version
dailyctl version
```

### Option 2: Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap cloudygreybeard/tap

# Install DailyLog
brew install dailylog

# Verify installation
dailylog --version
dailyctl version
```

### Option 3: Go Install

```bash
# Install MCP server
go install github.com/cloudygreybeard/dailylog-mcp/cmd/mcp-server@latest

# Install CLI tool
go install github.com/cloudygreybeard/dailylog-mcp/cmd/dailyctl@latest

# Note: Binaries will be named mcp-server and dailyctl
# You may want to create aliases or symlinks
```

### Option 4: Build from Source

```bash
# Clone the repository
git clone https://github.com/cloudygreybeard/dailylog-mcp.git
cd dailylog-mcp

# Build all components
make build

# Install locally
make install

# Verify installation
dailylog --version
dailyctl version
```

## GitHub Repository Setup

### 1. Create a Private Repository

Create a new private repository on GitHub for storing your daily logs:

```bash
# Example repository name: daily-logs
# URL: https://github.com/cloudygreybeard/daily-logs
```

### 2. Generate GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name like "DailyLog MCP Access"
4. Select the following scopes:
   - `repo` (Full control of private repositories)
   - `contents:write` (Write access to repository contents)
5. Click "Generate token" and copy the token

**Important**: Store this token securely as it won't be shown again.

### 3. Configure Environment Variables

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
# GitHub repository configuration
export DAILYLOG_GITHUB_REPO="cloudygreybeard/daily-logs"
export DAILYLOG_GITHUB_TOKEN="ghp_your_github_token_here"
export DAILYLOG_GITHUB_PATH="logs"  # Optional: subdirectory in repo
```

Apply the changes:
```bash
source ~/.bashrc  # or ~/.zshrc
```

## MCP Configuration

### Cursor IDE Configuration

1. Open Cursor IDE
2. Go to Settings → Extensions → MCP
3. Add the following configuration to your MCP settings:

```json
{
  "mcpServers": {
    "dailylog": {
      "command": "dailylog",
      "args": [],
      "env": {
        "DAILYLOG_GITHUB_REPO": "cloudygreybeard/daily-logs",
        "DAILYLOG_GITHUB_TOKEN": "ghp_your_github_token_here",
        "DAILYLOG_GITHUB_PATH": "logs"
      }
    }
  }
}
```

Alternative: Use a configuration file:

```bash
# Create MCP config directory
mkdir -p ~/.config/cursor/mcp

# Copy the example configuration
cp docs/examples/cursor-mcp-config.json ~/.config/cursor/mcp/config.json

# Edit with your values
nano ~/.config/cursor/mcp/config.json
```

### VS Code Configuration

1. Install the MCP extension for VS Code
2. Add the following to your VS Code settings.json:

```json
{
  "mcp.servers": {
    "dailylog": {
      "command": "dailylog",
      "args": [],
      "env": {
        "DAILYLOG_GITHUB_REPO": "cloudygreybeard/daily-logs",
        "DAILYLOG_GITHUB_TOKEN": "ghp_your_github_token_here",
        "DAILYLOG_GITHUB_PATH": "logs"
      }
    }
  }
}
```

## CLI Configuration

### Create Configuration File

```bash
# Create config directory
mkdir -p ~/.config/dailyctl

# Create configuration file
cat > ~/.config/dailyctl/config.yaml << EOF
github:
  repo: "cloudygreybeard/daily-logs"
  token: "ghp_your_github_token_here"
  path: "logs"

output:
  format: "table"

defaults:
  timezone: "America/New_York"
EOF
```

### Alternative: Command-line Flags

You can also specify configuration via command-line flags:

```bash
dailyctl log activity "Team standup" \
  --github-repo "cloudygreybeard/daily-logs" \
  --github-token "ghp_your_token_here" \
  --tags work,meeting \
  --status 8
```

## Verification

### Test MCP Server

```bash
# Test server connectivity
dailylog --help

# Test GitHub connection (this will attempt to read from your repo)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | dailylog
```

### Test CLI Tool

```bash
# Test basic functionality
dailyctl version

# Create a test entry
dailyctl log activity "Installation test" --tags setup --status 9

# View today's entries
dailyctl get today

# Search for the test entry
dailyctl search --query "Installation test"
```

### Test MCP Integration

In Cursor or VS Code:

1. Open a new chat/conversation
2. Type: `@dailylog dailylog_entry`
3. The MCP tools should be available and auto-complete

Example MCP usage:
```
@dailylog dailylog_entry {
  "type": "activity",
  "title": "MCP Integration Test",
  "description": "Testing the MCP integration",
  "tags": ["test", "mcp"],
  "status": 9
}
```

## Troubleshooting

### Common Issues

1. **"GitHub token not configured"**
   - Verify `DAILYLOG_GITHUB_TOKEN` environment variable is set
   - Check that the token has correct permissions

2. **"GitHub repository not found"**
   - Verify `DAILYLOG_GITHUB_REPO` format is "owner/repo"
   - Ensure the repository exists and is accessible

3. **"Permission denied" errors**
   - Check that your GitHub token has `repo` scope
   - Verify repository permissions

4. **MCP server not found**
   - Ensure `dailylog` binary is in your PATH
   - Try specifying full path in MCP configuration

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# CLI debug mode
dailyctl --verbose log activity "Debug test"

# MCP server debug (check stderr output)
DAILYLOG_DEBUG=true dailylog
```

### Health Check

```bash
# Check system status
dailyctl version
dailylog --version

# Test GitHub connectivity
curl -H "Authorization: token $DAILYLOG_GITHUB_TOKEN" \
  https://api.github.com/repos/$DAILYLOG_GITHUB_REPO
```

## Next Steps

1. **Create your first log entry**: Use the CLI or MCP integration to create your first daily log entry
2. **Set up automation**: Configure recurring reminders or integrate with your workflow
3. **Customize**: Adjust configuration files and create aliases for common commands
4. **Backup**: Consider setting up automated backups of your GitHub repository

## Security Best Practices

1. **Protect your GitHub token**: Never commit tokens to version control
2. **Use environment variables**: Store sensitive configuration in environment variables
3. **Limit token scope**: Only grant necessary permissions to your GitHub token
4. **Regular rotation**: Rotate your GitHub token periodically
5. **Private repository**: Ensure your daily logs repository is private

## Support

- **Documentation**: Available in the [docs](docs/) directory
- **Issues**: Report bugs on [GitHub Issues](https://github.com/cloudygreybeard/dailylog-mcp/issues)
- **Discussions**: Available in [GitHub Discussions](https://github.com/cloudygreybeard/dailylog-mcp/discussions)
