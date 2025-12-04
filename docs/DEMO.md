# DailyLog MCP Demo Creation Guide

This document explains how to create and update the animated demo shown in the README.

## Overview

The demo is created using:
1. **asciinema** - Records terminal sessions to `.cast` files
2. **svg-term-cli** - Converts `.cast` files to animated SVG

The final SVG is embedded in the README and displays an animated terminal session showing DailyLog MCP usage.

## Quick Start

```bash
# Record a new demo (interactive)
make demo.record

# Render existing .cast file to SVG
make demo.render

# Clean up demo files
make demo.clean
```

## Prerequisites

Install the required tools:

```bash
# Install asciinema (macOS)
brew install asciinema

# Install svg-term-cli (Node.js)
npm install -g svg-term-cli
```

## Recording Process

### Step 1: Record with asciinema

The `demo.record` target runs `hack/record-demo.sh`, which:

1. Records the terminal session using `asciinema rec`
2. Executes `hack/demo-script.sh` (the demo script)
3. Saves output to `docs/dailylog-demo.cast`

The recording uses:
- **Format**: asciicast v2 (compatible with svg-term)
- **Idle time limit**: 3 seconds
- **Command**: `./hack/demo-script.sh`

### Step 2: Convert to SVG

The script automatically converts the `.cast` file to SVG using `svg-term-cli`:

```bash
svg-term --out docs/dailylog-demo.svg --window < docs/dailylog-demo.cast
```

The `--window` flag adds macOS-style window decorations to the SVG.

### Manual Process

If you need to record manually:

```bash
# 1. Record the demo
asciinema rec docs/dailylog-demo.cast \
  --command ./hack/demo-script.sh \
  --idle-time-limit 3 \
  --overwrite \
  --output-format asciicast-v2

# 2. Convert to SVG
svg-term --out docs/dailylog-demo.svg --window < docs/dailylog-demo.cast
```

## Demo Script

The demo script (`hack/demo-script.sh`) demonstrates:

- Logging activities with tags and status
- Logging infrastructure work
- Logging deployments
- Logging planned work
- Retrieving entries for a specific date
- Searching entries by tags
- Generating standup reports

### Customizing the Demo

Edit `hack/demo-script.sh` to change the demo content. The script uses:
- Color codes for terminal output
- Simulated typing speed (configurable via `TYPE_SPEED`)
- Pauses between commands for readability

## File Locations

- **Cast file**: `docs/dailylog-demo.cast` (asciinema recording)
- **SVG file**: `docs/dailylog-demo.svg` (final animated demo)
- **Demo script**: `hack/demo-script.sh` (the script being recorded)
- **Recording script**: `hack/record-demo.sh` (automation script)

## Updating the Demo

1. Edit `hack/demo-script.sh` if you want to change the demo content
2. Run `make demo.record` to create a new recording
3. Review the generated SVG in a browser
4. Commit both `.cast` and `.svg` files

## Tips

- **Test the script first**: Run `./hack/demo-script.sh` manually to verify it works
- **Check SVG size**: Large SVGs can slow down README rendering
- **Browser preview**: Open `docs/dailylog-demo.svg` in a browser to preview
- **Re-record easily**: The `--overwrite` flag replaces existing recordings

## Troubleshooting

### asciinema not found
```bash
brew install asciinema
```

### svg-term-cli not found
```bash
npm install -g svg-term-cli
```

### SVG not animating
- Ensure you're using asciicast v2 format
- Check that the SVG is opened in a browser (not a static viewer)
- Verify the `.cast` file is valid JSON

### Demo script errors
- Make sure `dailyctl` is installed and in PATH
- Verify environment variables are set (if needed)
- Test the script manually before recording

## Make Targets

| Target | Description |
|--------|-------------|
| `demo.record` | Record new demo and convert to SVG |
| `demo.render` | Convert existing .cast file to SVG |
| `demo.clean` | Remove demo files (.cast and .svg) |

## References

- [asciinema documentation](https://asciinema.org/docs)
- [svg-term-cli on npm](https://www.npmjs.com/package/svg-term-cli)
- [asciicast format v2](https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md)




