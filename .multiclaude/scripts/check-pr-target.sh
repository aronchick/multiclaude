#!/bin/bash
# check-pr-target.sh - Prevents PRs from being created against fork main
#
# This script is called by Claude Code hooks (PreToolUse) to intercept
# `gh pr create` commands and ensure they never target the fork's main
# (aronchick/multiclaude). PRs should go to upstream (dlorenc/multiclaude).
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

# Blocked fork repository (case-insensitive check)
FORK_PATTERNS=(
    "aronchick/multiclaude"
    "ARONCHICK/MULTICLAUDE"
    "Aronchick/Multiclaude"
)

# Check if the command explicitly targets the fork via --repo or -R
for pattern in "${FORK_PATTERNS[@]}"; do
    if echo "$COMMAND" | grep -qiE "(--repo[= ]|--repo$|-R[= ]|-R$).*${pattern}"; then
        echo "ERROR: Cannot create PR against fork repository (aronchick/multiclaude)." >&2
        echo "" >&2
        echo "PRs should target upstream: dlorenc/multiclaude" >&2
        echo "" >&2
        echo "Fix: Use --repo dlorenc/multiclaude or remove --repo to use default" >&2
        exit 2
    fi
done

# Command is allowed
exit 0
