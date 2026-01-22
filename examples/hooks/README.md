# multiclaude Event Hooks Examples

This directory contains example hook scripts for multiclaude's event notification system.

## Overview

multiclaude emits events at key lifecycle points (agent started, PR created, CI failed, etc.) and can execute user-provided hook scripts when these events occur. This enables integration with external systems like Slack, Discord, email, or custom monitoring tools.

## Philosophy

- **Hook-based, not built-in**: Notifications belong in user-controlled hooks, not in the daemon
- **Fire-and-forget**: No retries, no delivery guarantees (hooks timeout after 30s)
- **Zero dependencies**: Core only emits events; notification logic is in user scripts
- **Unix philosophy**: multiclaude emits JSON events, users compose the rest

## Available Hooks

| Hook | Event Type | When It Fires |
|------|------------|---------------|
| `on_event` | All events | Catch-all for any event |
| `on_agent_started` | `agent_started` | When an agent starts |
| `on_agent_stopped` | `agent_stopped` | When an agent stops |
| `on_agent_idle` | `agent_idle` | When an agent is idle |
| `on_pr_created` | `pr_created` | When a PR is created |
| `on_pr_merged` | `pr_merged` | When a PR is merged |
| `on_task_assigned` | `task_assigned` | When a task is assigned |
| `on_ci_failed` | `ci_failed` | When CI fails |
| `on_worker_stuck` | `worker_stuck` | When a worker is stuck |
| `on_message_sent` | `message_sent` | When an inter-agent message is sent |

## Event JSON Format

All hooks receive event data via stdin as JSON:

```json
{
  "type": "agent_started",
  "timestamp": "2024-01-15T10:30:00Z",
  "repo_name": "my-repo",
  "agent_name": "clever-fox",
  "data": {
    "task": "Implement auth feature",
    "agent_type": "worker"
  }
}
```

## Example Scripts

### Slack Notifications

See [`slack-notify.sh`](./slack-notify.sh) for a complete Slack integration example.

**Setup:**

```bash
# 1. Create Slack webhook: https://api.slack.com/messaging/webhooks
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# 2. Configure multiclaude to use the hook
multiclaude hooks set on_event /path/to/examples/hooks/slack-notify.sh

# Or configure specific events
multiclaude hooks set on_ci_failed /path/to/examples/hooks/slack-notify.sh
multiclaude hooks set on_pr_created /path/to/examples/hooks/slack-notify.sh
```

### Custom Hook Template

```bash
#!/usr/bin/env bash
set -euo pipefail

# Read event JSON from stdin
EVENT_JSON=$(cat)

# Parse event type
EVENT_TYPE=$(echo "$EVENT_JSON" | jq -r '.type')

# Handle specific events
case "$EVENT_TYPE" in
    agent_started)
        # Your custom logic here
        echo "Agent started!" | mail -s "multiclaude alert" you@example.com
        ;;
    ci_failed)
        # Your custom logic here
        curl -X POST https://your-monitoring-system.com/alert \
            -H "Content-Type: application/json" \
            -d "$EVENT_JSON"
        ;;
esac

exit 0
```

## Configuration

### View Current Configuration

```bash
multiclaude hooks list
```

### Set a Hook

```bash
multiclaude hooks set on_event /path/to/your-hook.sh
multiclaude hooks set on_pr_created /path/to/pr-hook.sh
```

### Clear a Hook

```bash
multiclaude hooks clear on_event
```

### Clear All Hooks

```bash
multiclaude hooks clear-all
```

## Testing Hooks

Test your hook script manually:

```bash
# Create test event JSON
echo '{
  "type": "agent_started",
  "timestamp": "2024-01-15T10:30:00Z",
  "repo_name": "test-repo",
  "agent_name": "test-agent",
  "data": {"task": "Test task"}
}' | /path/to/your-hook.sh
```

## Best Practices

1. **Keep hooks fast**: Hooks timeout after 30 seconds
2. **Handle failures gracefully**: Hooks should not crash on missing data
3. **Use environment variables**: Don't hardcode secrets in scripts
4. **Test thoroughly**: Use the manual testing approach above
5. **Log errors**: Write errors to stderr for debugging

## Security Notes

- Hook scripts run with the same permissions as the multiclaude daemon
- Never commit webhook URLs or secrets to version control
- Use environment variables for sensitive configuration
- Validate and sanitize event data before using it in commands

## Fork-Only Feature

⚠️ **This is a fork-only feature** that should not be contributed upstream without explicit approval from the upstream maintainers. See `docs/FORK_MAINTENANCE_STRATEGY.md` for details.

The upstream project (dlorenc/multiclaude) explicitly rejects external integrations in favor of a terminal-native, local-first approach. This feature is maintained in the aronchick/multiclaude fork to support users who need Slack/Discord/email notifications.

