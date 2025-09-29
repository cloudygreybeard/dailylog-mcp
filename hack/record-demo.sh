#!/bin/bash
# Script to record and render DailyLog MCP demo

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"
SVG_FILE="$DOCS_DIR/dailylog-demo.svg"

echo "Recording DailyLog MCP Demo with svg-term-cli..."
echo ""
echo "This will demonstrate daily logging workflow:"
echo "- Logging activities, status updates, and notes"
echo "- Searching through entries"  
echo "- Generating summaries"
echo ""

# Check if running interactively
if [ -t 0 ]; then
    echo "Press Enter to start recording..."
    read
else
    echo "Running in non-interactive mode, starting immediately..."
fi

echo ""
echo "Recording with svg-term-cli..."

# Install tools if needed
if ! command -v asciinema >/dev/null 2>&1; then
    echo "Installing asciinema..."
    brew install asciinema
fi

if ! command -v svg-term >/dev/null 2>&1; then
    echo "Installing svg-term-cli..."
    npm install -g svg-term-cli
fi

# Step 1: Record with asciinema in v2 format for svg-term compatibility
CAST_FILE="$DOCS_DIR/dailylog-demo.cast"
echo "Step 1: Recording with asciinema (v2 format)..."
asciinema rec "$CAST_FILE" \
    --command "$SCRIPT_DIR/demo-script.sh" \
    --idle-time-limit 3 \
    --overwrite \
    --output-format asciicast-v2 \
    --stdin

echo "Step 2: Converting to SVG with svg-term..."
# Step 2: Pipe the .cast file to svg-term
svg-term \
    --out "$SVG_FILE" \
    --window \
    < "$CAST_FILE"

echo ""
echo "Recording complete!"

# Check if the SVG was created successfully
if [ -f "$SVG_FILE" ]; then
    echo "SVG file: $SVG_FILE"
    echo "File size: $(du -h "$SVG_FILE" | cut -f1)"
    echo "Cast file: $CAST_FILE ($(du -h "$CAST_FILE" | cut -f1))"
else
    echo "SVG file not created: $SVG_FILE"
fi

echo ""
echo "Demo creation successful!"
echo "Cast file: $(du -h "$CAST_FILE" | cut -f1)"
echo "SVG file: $(du -h "$SVG_FILE" | cut -f1)" 
echo ""
echo "To include in README, add:"
echo "![DailyLog MCP Demo](docs/dailylog-demo.svg)"
echo ""
echo "Tip: Open the SVG in a browser to see the animated demo with macOS window styling"



