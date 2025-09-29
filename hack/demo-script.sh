#!/bin/bash
# DailyLog MCP Usage Demo Script
# Demonstrates daily logging workflow with realistic scenarios

set -e

# Colors for demo output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

# Demo settings
TYPE_SPEED=${TYPE_SPEED:-0.017}

# Helper function to simulate typing
type_command() {
    local cmd="$1"
    local speed="${2:-$TYPE_SPEED}"
    
    printf "${BLUE}$ ${NC}"
    for (( i=0; i<${#cmd}; i++ )); do
        printf "${cmd:$i:1}"
        sleep "$speed"
    done
    printf "\n"
}

# Helper function to show mock output
show_output() {
    printf "${GRAY}$1${NC}\n"
}

# Helper function to pause between commands
demo_pause() {
    sleep "${1:-1.0}"
}

# Helper function for commentary
comment() {
    printf "${YELLOW}# $1${NC}\n"
    demo_pause 1
}

# Clear screen and start demo
clear

comment "DailyLog MCP - Developer Standup Logging Demo"
demo_pause 2

comment "Let's log yesterday's development activities for standup..."
demo_pause 1

comment "First, log a code review activity"
type_command "dailyctl log activity \"Reviewed PR #142 for API refactoring\" --tags code-review,backend --status 8 --duration 45"
show_output "✓ Activity logged: Reviewed PR #142 for API refactoring (duration: 45min)"
show_output "  Tags: code-review, backend"
show_output "  Stored: 2025-07-21T14:30:00Z"
demo_pause 2

comment "Log an infrastructure fix"
type_command "dailyctl log activity \"Fixed Redis connection pooling issue in staging\" --tags infrastructure,bugfix --status 7 --duration 90"
show_output "✓ Activity logged: Fixed Redis connection pooling issue in staging (duration: 90min)"
show_output "  Tags: infrastructure, bugfix"
show_output "  Stored: 2025-07-21T16:00:00Z"
demo_pause 2

comment "Log a deployment"
type_command "dailyctl log activity \"Deployed user-service v2.1.3 to production\" --tags deployment,production --status 9 --duration 30"
show_output "✓ Activity logged: Deployed user-service v2.1.3 to production (duration: 30min)"
show_output "  Tags: deployment, production"
show_output "  Stored: 2025-07-21T17:30:00Z"
demo_pause 2

comment "Log today's planned work"
type_command "dailyctl log activity \"Implement rate limiting for user API endpoints\" --tags development,api --status 8 --planned"
show_output "✓ Activity logged: Implement rate limiting for user API endpoints (planned)"
show_output "  Tags: development, api"
show_output "  Stored: 2025-07-21T09:00:00Z"
demo_pause 2

comment "Let's see yesterday's completed work"
type_command "dailyctl get yesterday"
show_output "Daily Log for 2025-07-21"
show_output "========================="
show_output ""
show_output "TIME     TYPE      TAGS            TITLE                           STATUS   PRIORITY"
show_output "-------------------------------------------------------------------------------------"
show_output "14:30:00 activity  code-review     Reviewed PR #142 for API       8/10     "
show_output "                   backend         refactoring                              "
show_output "16:00:00 activity  infrastructure  Fixed Redis connection         7/10     "
show_output "                   bugfix          pooling issue in staging                "
show_output "17:30:00 activity  deployment      Deployed user-service v2.1.3   9/10     "
show_output "                   production      to production                           "
show_output ""
show_output "Statistics:"
show_output "  Average status: 8.0/10"
show_output "  Total entries: 3"
demo_pause 3

comment "Now let's check deployments specifically"
type_command "dailyctl search --tags deployment --status-min 8"
show_output "Search Results (1 entries found):"
show_output "=============================================="
show_output ""
show_output "2025-07-21 17:30  Deployed user-service v2.1.3 to production"
show_output "  Tags: deployment, production"
show_output "  Duration: 30min"
show_output ""
show_output "2025-07-21 15:45  Deployed auth-service v1.2.1 to staging"
show_output "  Tags: deployment, staging"
show_output "  Duration: 20min"
demo_pause 3

comment "Generate YAML-formatted standup report"
type_command "dailyctl standup --format slack-yaml"
show_output "Standup Report (2025-07-21)"
show_output "============================"
show_output ""
show_output "\`\`\`yaml"
show_output "Y: # Yesterday (Jul 21)"
show_output "  - Reviewed PR #142 for API refactoring (status: 8/10)"
show_output "  - Fixed Redis connection pooling issue in staging (status: 7/10)"
show_output "  - Deployed user-service v2.1.3 to production (status: 9/10)"
show_output ""
show_output "T: # Today (Jul 22)"
show_output "  - Implement rate limiting for user API endpoints (priority: 3/5)"
show_output "  - Code review session"
show_output "  - Sprint planning meeting"
show_output "\`\`\`"
demo_pause 3

comment "Copy standup report to clipboard for Slack"
type_command "dailyctl standup --format slack-yaml --copy"
show_output "Report copied to clipboard!"
show_output ""
show_output "Standup Report (2025-07-21)"
show_output "============================"
show_output ""
show_output "\`\`\`yaml"
show_output "Y: # Yesterday (Jul 21)"
show_output "  - Reviewed PR #142 for API refactoring (status: 8/10)"
show_output "  - Fixed Redis connection pooling issue in staging (status: 7/10)"
show_output "  - Deployed user-service v2.1.3 to production (status: 9/10)"
show_output ""
show_output "T: # Today (Jul 22)"
show_output "  - Implement rate limiting for user API endpoints"
show_output "  - Code review session"
show_output "\`\`\`"
demo_pause 4

comment "Demo complete! DailyLog MCP helps SRE teams track daily activities efficiently."
demo_pause 2

# Keep the script running for a moment to ensure asciinema captures everything
read -n 1 -s -r



