#!/bin/bash
# check-pr-target.sh - Validates PRs follow the dual-layer workflow
#
# This script is called by Claude Code hooks (PreToolUse) to intercept
# `gh pr create` commands and ensure they follow best practices for
# the dual-layer Brownian ratchet workflow:
#   1. Changes should be tested in the fork (aronchick/multiclaude) first
#   2. After fork CI passes, targeted PRs can go to upstream (dlorenc/multiclaude)
#
# Exit codes:
#   0 - Allow the command
#   1 - Warning (show message but allow)
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

# Upstream repository patterns
UPSTREAM_PATTERNS=(
    "dlorenc/multiclaude"
    "DLORENC/MULTICLAUDE"
    "dlorenc/Multiclaude"
)

# Check if targeting upstream
TARGETS_UPSTREAM=false
for pattern in "${UPSTREAM_PATTERNS[@]}"; do
    if echo "$COMMAND" | grep -qiE "(--repo[= ]|--repo$|-R[= ]|-R$).*${pattern}"; then
        TARGETS_UPSTREAM=true
        break
    fi
    if echo "$COMMAND" | grep -qi "$pattern"; then
        TARGETS_UPSTREAM=true
        break
    fi
done

# If not targeting upstream, allow it
if [ "$TARGETS_UPSTREAM" = false ]; then
    exit 0
fi

# Targeting upstream - check if changes have been validated in fork
echo "⚠️  UPSTREAM PR DETECTED" >&2
echo "" >&2
echo "You're creating a PR to upstream (dlorenc/multiclaude)." >&2
echo "" >&2
echo "📋 REQUIRED CHECKLIST before proceeding:" >&2
echo "  ✓ Changes tested in fork (aronchick/multiclaude) first" >&2
echo "  ✓ Fork CI passed (all tests green)" >&2
echo "  ✓ PR is targeted and focused (not too broad)" >&2
echo "  ✓ Changes work properly in the fork" >&2
echo "" >&2
echo "If you've completed the checklist, this PR is allowed." >&2
echo "If not, please test in the fork first: --repo aronchick/multiclaude" >&2
echo "" >&2

# Allow the command (exit 0) - trust the agent/human has followed the process
# The checklist reminder is enough; we don't enforce it mechanically
exit 0
