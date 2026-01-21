#!/bin/bash
# pre-push hook: Ensure tests pass before pushing
#
# Installation:
#   ln -sf $(pwd)/scripts/pre-push-hook.sh .git/hooks/pre-push
#   (or for worktrees: link to the main repo .git/hooks/pre-push)

set -e

echo "🔍 Running pre-push validation..."

# Check if go is available
if ! command -v go &> /dev/null; then
    echo "⚠️  Go not found, skipping tests"
    exit 0
fi

# Allow skip with SKIP_TESTS env var
if [ "$SKIP_TESTS" = "1" ]; then
    echo "⚠️  SKIP_TESTS=1 set, bypassing validation"
    exit 0
fi

# Run build
echo "📦 Building..."
if ! go build -v ./... 2>&1 | tail -5; then
    echo ""
    echo "❌ Build failed! Cannot push."
    echo ""
    echo "Fix the build errors and try again."
    echo "Or skip with: SKIP_TESTS=1 git push"
    exit 1
fi

# Run tests (with timeout to prevent hanging)
echo "🧪 Running tests..."
if ! timeout 120s go test ./internal/... ./pkg/... 2>&1 | tail -10; then
    echo ""
    echo "❌ Tests failed! Cannot push."
    echo ""
    echo "Fix the failing tests and try again."
    echo "Run 'go test ./...' locally to see all failures."
    echo "Or skip with: SKIP_TESTS=1 git push"
    exit 1
fi

echo "✅ All checks passed!"
exit 0
