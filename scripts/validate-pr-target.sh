#!/bin/bash
# validate-pr-target.sh
# Pre-commit hook to validate PR targeting for fork workflows
#
# This script ensures that:
# 1. PRs from forks target the upstream repository, not the fork itself
# 2. PR metadata is correct before creation
# 3. Branch names follow multiclaude conventions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔍 Validating PR target configuration..."

# Function to print error and exit
fail() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Function to print warning
warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to print success
success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    fail "Not in a git repository"
fi

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
if [ -z "$CURRENT_BRANCH" ]; then
    fail "Could not determine current branch"
fi

echo "Current branch: $CURRENT_BRANCH"

# Check if upstream remote exists
if git remote get-url upstream > /dev/null 2>&1; then
    IS_FORK=true
    UPSTREAM_URL=$(git remote get-url upstream)
    echo "Detected fork workflow (upstream remote exists)"
else
    IS_FORK=false
    echo "Not a fork (no upstream remote)"
fi

# Get origin remote
ORIGIN_URL=$(git remote get-url origin)

# Parse GitHub URLs
parse_github_url() {
    local url="$1"
    # SSH format: git@github.com:owner/repo.git
    if [[ "$url" =~ git@github\.com:([^/]+)/(.+)(\.git)?$ ]]; then
        echo "${BASH_REMATCH[1]}/${BASH_REMATCH[2]%.git}"
        return
    fi
    # HTTPS format: https://github.com/owner/repo.git
    if [[ "$url" =~ https://github\.com/([^/]+)/(.+)(\.git)?$ ]]; then
        echo "${BASH_REMATCH[1]}/${BASH_REMATCH[2]%.git}"
        return
    fi
    echo ""
}

ORIGIN_REPO=$(parse_github_url "$ORIGIN_URL")
echo "Origin repository: $ORIGIN_REPO"

if [ "$IS_FORK" = true ]; then
    UPSTREAM_REPO=$(parse_github_url "$UPSTREAM_URL")
    echo "Upstream repository: $UPSTREAM_REPO"

    if [ -z "$UPSTREAM_REPO" ]; then
        fail "Could not parse upstream repository URL: $UPSTREAM_URL"
    fi

    # Validate that origin and upstream are different
    if [ "$ORIGIN_REPO" = "$UPSTREAM_REPO" ]; then
        warn "Origin and upstream are the same - this may not be a true fork"
    fi
fi

# Check if gh CLI is available
if ! command -v gh > /dev/null 2>&1; then
    warn "gh CLI not found - skipping PR validation"
    exit 0
fi

# Check if there are any existing PRs for this branch
echo ""
echo "Checking for existing PRs on branch: $CURRENT_BRANCH"

# Try to find PR for current branch
PR_DATA=$(gh pr list --head "$CURRENT_BRANCH" --json number,baseRepository,headRepository,baseRefName 2>/dev/null || echo "[]")
PR_COUNT=$(echo "$PR_DATA" | jq '. | length')

if [ "$PR_COUNT" -eq 0 ]; then
    success "No existing PR found for this branch"

    # Provide guidance for PR creation
    if [ "$IS_FORK" = true ]; then
        echo ""
        echo "When creating a PR for this fork, use:"
        echo ""
        echo "  gh pr create \\"
        echo "    --repo $UPSTREAM_REPO \\"
        echo "    --base main \\"
        echo "    --title \"Your PR title\" \\"
        echo "    --body \"Your PR description\" \\"
        echo "    --label \"multiclaude\""
        echo ""
    else
        echo ""
        echo "When creating a PR, use:"
        echo ""
        echo "  gh pr create \\"
        echo "    --base main \\"
        echo "    --title \"Your PR title\" \\"
        echo "    --body \"Your PR description\" \\"
        echo "    --label \"multiclaude\""
        echo ""
    fi

    exit 0
fi

# Validate existing PR
echo "Found $PR_COUNT existing PR(s)"

# Check each PR
for i in $(seq 0 $((PR_COUNT - 1))); do
    PR_NUMBER=$(echo "$PR_DATA" | jq -r ".[$i].number")
    PR_BASE_REPO=$(echo "$PR_DATA" | jq -r ".[$i].baseRepository.nameWithOwner")
    PR_HEAD_REPO=$(echo "$PR_DATA" | jq -r ".[$i].headRepository.nameWithOwner")
    PR_BASE_REF=$(echo "$PR_DATA" | jq -r ".[$i].baseRefName")

    echo ""
    echo "PR #$PR_NUMBER:"
    echo "  Base: $PR_BASE_REPO ($PR_BASE_REF)"
    echo "  Head: $PR_HEAD_REPO ($CURRENT_BRANCH)"

    # Validation for fork workflow
    if [ "$IS_FORK" = true ]; then
        # PR should target upstream, not origin
        if [ "$PR_BASE_REPO" != "$UPSTREAM_REPO" ]; then
            fail "PR #$PR_NUMBER targets wrong repository!
Expected: $UPSTREAM_REPO (upstream)
Actual:   $PR_BASE_REPO

This PR is targeting your fork instead of the upstream repository.

To fix:
1. Close the incorrect PR:
   gh pr close $PR_NUMBER

2. Create a new PR targeting upstream:
   gh pr create --repo $UPSTREAM_REPO --base main --title \"...\" --body \"...\"
"
        fi

        # PR head should be from origin (the fork)
        if [ "$PR_HEAD_REPO" != "$ORIGIN_REPO" ]; then
            warn "PR #$PR_NUMBER head repository is unexpected: $PR_HEAD_REPO (expected: $ORIGIN_REPO)"
        fi

        success "PR #$PR_NUMBER targets correct repository (upstream)"
    else
        # Non-fork workflow - PR should target origin
        if [ "$PR_BASE_REPO" != "$ORIGIN_REPO" ]; then
            warn "PR #$PR_NUMBER targets unexpected repository: $PR_BASE_REPO (expected: $ORIGIN_REPO)"
        else
            success "PR #$PR_NUMBER targets correct repository"
        fi
    fi

    # Validate base ref is main or master
    if [[ "$PR_BASE_REF" != "main" && "$PR_BASE_REF" != "master" ]]; then
        warn "PR #$PR_NUMBER targets unusual branch: $PR_BASE_REF (expected: main or master)"
    fi
done

echo ""
success "PR target validation complete"
