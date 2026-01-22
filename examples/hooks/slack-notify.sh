#!/usr/bin/env bash
#
# Example Slack notification hook for multiclaude
#
# This script receives event data via stdin as JSON and posts notifications to Slack.
#
# Setup:
# 1. Create a Slack webhook URL: https://api.slack.com/messaging/webhooks
# 2. Set SLACK_WEBHOOK_URL environment variable or edit this script
# 3. Configure multiclaude to use this hook:
#    multiclaude hooks set on_event /path/to/slack-notify.sh
#
# Event-specific hooks (optional):
#    multiclaude hooks set on_pr_created /path/to/slack-notify.sh
#    multiclaude hooks set on_agent_idle /path/to/slack-notify.sh
#    multiclaude hooks set on_ci_failed /path/to/slack-notify.sh
#
# Environment variables:
#   SLACK_WEBHOOK_URL - Required. Your Slack webhook URL
#   SLACK_CHANNEL     - Optional. Override default channel (e.g., #multiclaude)
#   SLACK_USERNAME    - Optional. Bot username (default: multiclaude)
#   SLACK_ICON_EMOJI  - Optional. Bot icon (default: :robot_face:)

set -euo pipefail

# Configuration
WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"
CHANNEL="${SLACK_CHANNEL:-}"
USERNAME="${SLACK_USERNAME:-multiclaude}"
ICON_EMOJI="${SLACK_ICON_EMOJI:-:robot_face:}"

# Validate webhook URL
if [ -z "$WEBHOOK_URL" ]; then
    echo "Error: SLACK_WEBHOOK_URL environment variable not set" >&2
    exit 1
fi

# Read event JSON from stdin
EVENT_JSON=$(cat)

# Parse event fields using jq (or fallback to basic parsing)
if command -v jq &> /dev/null; then
    EVENT_TYPE=$(echo "$EVENT_JSON" | jq -r '.type')
    REPO_NAME=$(echo "$EVENT_JSON" | jq -r '.repo_name // ""')
    AGENT_NAME=$(echo "$EVENT_JSON" | jq -r '.agent_name // ""')
    TIMESTAMP=$(echo "$EVENT_JSON" | jq -r '.timestamp')
else
    # Fallback: basic grep/sed parsing (less robust)
    EVENT_TYPE=$(echo "$EVENT_JSON" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)
    REPO_NAME=$(echo "$EVENT_JSON" | grep -o '"repo_name":"[^"]*"' | cut -d'"' -f4 || echo "")
    AGENT_NAME=$(echo "$EVENT_JSON" | grep -o '"agent_name":"[^"]*"' | cut -d'"' -f4 || echo "")
    TIMESTAMP=$(echo "$EVENT_JSON" | grep -o '"timestamp":"[^"]*"' | cut -d'"' -f4)
fi

# Build message based on event type
case "$EVENT_TYPE" in
    agent_started)
        TASK=$(echo "$EVENT_JSON" | jq -r '.data.task // ""' 2>/dev/null || echo "")
        if [ -n "$TASK" ]; then
            MESSAGE="🚀 Agent *$AGENT_NAME* started in *$REPO_NAME*\nTask: $TASK"
        else
            MESSAGE="🚀 Agent *$AGENT_NAME* started in *$REPO_NAME*"
        fi
        COLOR="good"
        ;;
    
    agent_stopped)
        REASON=$(echo "$EVENT_JSON" | jq -r '.data.reason // "unknown"' 2>/dev/null || echo "unknown")
        MESSAGE="🛑 Agent *$AGENT_NAME* stopped in *$REPO_NAME*\nReason: $REASON"
        COLOR="warning"
        ;;
    
    pr_created)
        PR_NUMBER=$(echo "$EVENT_JSON" | jq -r '.data.pr_number // ""' 2>/dev/null || echo "")
        PR_TITLE=$(echo "$EVENT_JSON" | jq -r '.data.title // ""' 2>/dev/null || echo "")
        MESSAGE="📝 PR #$PR_NUMBER created in *$REPO_NAME*\n$PR_TITLE"
        COLOR="good"
        ;;
    
    pr_merged)
        PR_NUMBER=$(echo "$EVENT_JSON" | jq -r '.data.pr_number // ""' 2>/dev/null || echo "")
        MESSAGE="✅ PR #$PR_NUMBER merged in *$REPO_NAME*"
        COLOR="good"
        ;;
    
    ci_failed)
        PR_NUMBER=$(echo "$EVENT_JSON" | jq -r '.data.pr_number // ""' 2>/dev/null || echo "")
        JOB_NAME=$(echo "$EVENT_JSON" | jq -r '.data.job_name // ""' 2>/dev/null || echo "")
        MESSAGE="❌ CI failed in *$REPO_NAME*\nPR: #$PR_NUMBER\nJob: $JOB_NAME"
        COLOR="danger"
        ;;
    
    agent_idle)
        DURATION=$(echo "$EVENT_JSON" | jq -r '.data.duration_minutes // ""' 2>/dev/null || echo "")
        MESSAGE="💤 Agent *$AGENT_NAME* idle in *$REPO_NAME* for ${DURATION}m"
        COLOR="warning"
        ;;
    
    worker_stuck)
        DURATION=$(echo "$EVENT_JSON" | jq -r '.data.duration_minutes // ""' 2>/dev/null || echo "")
        MESSAGE="⚠️ Worker *$AGENT_NAME* stuck in *$REPO_NAME* for ${DURATION}m"
        COLOR="danger"
        ;;
    
    task_assigned)
        TASK=$(echo "$EVENT_JSON" | jq -r '.data.task // ""' 2>/dev/null || echo "")
        MESSAGE="📋 Task assigned to *$AGENT_NAME* in *$REPO_NAME*\n$TASK"
        COLOR="#439FE0"
        ;;
    
    *)
        MESSAGE="ℹ️ Event *$EVENT_TYPE* in *$REPO_NAME*"
        COLOR="#808080"
        ;;
esac

# Build Slack payload
PAYLOAD=$(cat <<EOF
{
    "username": "$USERNAME",
    "icon_emoji": "$ICON_EMOJI",
    "attachments": [{
        "color": "$COLOR",
        "text": "$MESSAGE",
        "footer": "multiclaude",
        "ts": $(date -d "$TIMESTAMP" +%s 2>/dev/null || date +%s)
    }]
}
EOF
)

# Add channel override if specified
if [ -n "$CHANNEL" ]; then
    PAYLOAD=$(echo "$PAYLOAD" | jq --arg ch "$CHANNEL" '. + {channel: $ch}')
fi

# Send to Slack
curl -X POST -H 'Content-type: application/json' --data "$PAYLOAD" "$WEBHOOK_URL" --silent --show-error

exit 0

