#!/bin/bash
# check-pr-target.sh - Prevents PRs from being created against upstream
#
# This script is called by Claude Code hooks (PreToolUse) to intercept
# `gh pr create` commands and ensure they never target the upstream
# repository (dlorenc/multiclaude).
#
# Exit codes:
#   0 - Allow the command
#   2 - Block the command (error message in stderr)

set -e

# Read the hook input from stdin
INPUT=$(cat)

# Extract the command from the JSON input
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# If not a bash command or empty, allow it
if [ -z "$COMMAND" ]; then
    exit 0
fi

# Check if this is a gh pr create command
if ! echo "$COMMAND" | grep -qE 'gh\s+(pr\s+create|pr\s+--repo|--repo.*pr\s+create)'; then
    exit 0
fi

# Blocked upstream repositories (case-insensitive check)
UPSTREAM_PATTERNS=(
    "dlorenc/multiclaude"
    "DLORENC/MULTICLAUDE"
    "dlorenc/Multiclaude"
)

# Check if the command explicitly targets upstream via --repo or -R
for pattern in "${UPSTREAM_PATTERNS[@]}"; do
    if echo "$COMMAND" | grep -qiE "(--repo[= ]|--repo$|-R[= ]|-R$).*${pattern}"; then
        echo "ERROR: Cannot create PR against upstream repository (dlorenc/multiclaude)." >&2
        echo "" >&2
        echo "PRs must target the fork: aronchick/multiclaude" >&2
        echo "" >&2
        echo "Fix: Remove --repo flag or use: --repo aronchick/multiclaude" >&2
        exit 2
    fi
done

# Check for upstream in the command even without explicit --repo flag
# (catches cases like `gh pr create -R dlorenc/multiclaude`)
for pattern in "${UPSTREAM_PATTERNS[@]}"; do
    if echo "$COMMAND" | grep -qi "$pattern"; then
        echo "ERROR: Cannot create PR against upstream repository (dlorenc/multiclaude)." >&2
        echo "" >&2
        echo "PRs must target the fork: aronchick/multiclaude" >&2
        exit 2
    fi
done

# Command is allowed
exit 0
