# Configuring Automated Agents to Create PRs Against Forks

## Multiclaude Implementation

**If you're using multiclaude, fork detection is built-in and automatic.** When you initialize a repository with `multiclaude init`, multiclaude:

1. **Automatically detects** if the repository is a fork by checking for an `upstream` remote
2. **Stores fork metadata** in the daemon state (upstream owner/repo, fork owner/repo)
3. **Injects fork-aware instructions** into agent system prompts
4. **Validates PR targets** via CI workflows

**What this means for you:**
- Workers automatically know to create PRs against upstream (not your fork)
- Merge-queue validates PRs target the correct repository before merging
- CI enforces correct PR targeting
- Pre-commit hooks catch mistakes before PRs are created

**Configuration:** None required! Just ensure your repository has the `upstream` remote configured:

```bash
cd ~/.multiclaude/repos/your-repo
git remote -v

# Should show both origin (your fork) and upstream:
# origin    git@github.com:youruser/repo.git (fetch)
# upstream  git@github.com:original/repo.git (fetch)
```

If you don't have an upstream remote, add it:

```bash
git remote add upstream git@github.com:original-owner/original-repo.git
git fetch upstream
```

Then re-initialize the repository to update fork detection:

```bash
multiclaude remove <repo-name>
multiclaude init https://github.com/youruser/your-fork.git
```

---

## General Implementation Guide

The following sections provide a general guide for implementing fork-aware PR targeting in any automated agent system (not just multiclaude).

## Problem Statement

When automated agents (like multiclaude workers) operate in git worktrees cloned from a **fork** that tracks an **upstream** repository, the `gh pr create` command defaults to creating pull requests against the **upstream** repository instead of the fork. This results in:

- PRs opened against the wrong repository
- Manual cleanup (closing and recreating PRs)
- Wasted CI resources
- Confusion in PR history

## Why This Happens

### Git Remote Hierarchy

When you fork a repository and clone your fork:

```bash
# Standard fork workflow
git clone git@github.com:youruser/project.git
cd project
git remote add upstream git@github.com:original/project.git
```

The repository has two remotes:
- `origin` → your fork (`youruser/project`)
- `upstream` → original repository (`original/project`)

### GitHub CLI Default Behavior

The `gh pr create` command uses this logic:

1. Identifies the current branch's upstream tracking branch
2. Determines which remote that branch tracks
3. If the branch tracks `upstream/main`, creates PR against `upstream`
4. If the branch tracks `origin/main`, creates PR against `origin` (your fork)

**The Issue:** If your worktree branches are created from or track `upstream/main`, `gh` assumes you want to contribute to upstream, not your fork.

## Solution Architecture

### Overview

To ensure automated agents create PRs against your fork:

1. **Clone Configuration:** Clone from your fork, not upstream
2. **Remote Configuration:** Set `origin` to your fork, `upstream` to original
3. **Branch Tracking:** Create feature branches that track `origin`, not `upstream`
4. **GitHub CLI Configuration:** Explicitly set the default repository
5. **Agent Prompts:** Instruct agents to verify PR target before creation

## Implementation Guide

### Step 1: Repository Clone Configuration

When setting up your repository for automated agents:

```bash
# Clone YOUR fork, not the upstream
git clone git@github.com:youruser/project.git ~/.multiclaude/repos/project

# Add upstream as a secondary remote
cd ~/.multiclaude/repos/project
git remote add upstream git@github.com:original/project.git

# Verify remote configuration
git remote -v
# Should show:
# origin    git@github.com:youruser/project.git (fetch)
# origin    git@github.com:youruser/project.git (push)
# upstream  git@github.com:original/project.git (fetch)
# upstream  git@github.com:original/project.git (push)
```

**Key Principle:** `origin` must point to YOUR fork, not the upstream repository.

### Step 2: Worktree Branch Configuration

When creating worktrees (either manually or via automation):

```bash
# Create worktree with branch tracking YOUR fork
git worktree add ~/.multiclaude/wts/project/agent-1 -b agent-1-feature

# In the worktree, verify the branch tracks origin
cd ~/.multiclaude/wts/project/agent-1
git branch -vv
# Should show: agent-1-feature [origin/main] ...

# If it doesn't track origin, set it explicitly
git branch --set-upstream-to=origin/main
```

**Automation Implementation:**

In your worktree creation code:

```go
// Example: Ensure branch tracks the fork's main branch
func createWorktree(repoPath, worktreePath, branchName string) error {
    // Create worktree
    cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", branchName)
    cmd.Dir = repoPath
    if err := cmd.Run(); err != nil {
        return err
    }

    // Set upstream tracking to origin/main
    cmd = exec.Command("git", "branch", "--set-upstream-to=origin/main")
    cmd.Dir = worktreePath
    return cmd.Run()
}
```

### Step 3: GitHub CLI Default Repository

Configure `gh` CLI to use your fork as the default:

```bash
# In each worktree, set the default repository
cd ~/.multiclaude/wts/project/agent-1
gh repo set-default youruser/project

# Verify the configuration
gh repo view
# Should show your fork, not upstream
```

**Automation Implementation:**

Run this during worktree initialization:

```go
func configureGHCLI(worktreePath, forkRepo string) error {
    cmd := exec.Command("gh", "repo", "set-default", forkRepo, "--yes")
    cmd.Dir = worktreePath
    return cmd.Run()
}

// Example usage
configureGHCLI("/path/to/worktree", "youruser/project")
```

**Alternative:** Set via environment variable:

```bash
# In agent startup script or environment
export GH_REPO="youruser/project"
```

### Step 4: Agent Prompt Configuration

Update your agent system prompts to include PR creation guidelines:

```markdown
## Pull Request Creation

When creating pull requests:

1. **Always verify the target repository:**
   ```bash
   gh repo view  # Should show: youruser/project
   ```

2. **Create PR with explicit base repository:**
   ```bash
   gh pr create \
     --repo youruser/project \
     --base main \
     --title "feat: Your feature" \
     --body "Description"
   ```

3. **Before creating the PR, confirm:**
   - Current repository: `gh repo view`
   - Branch tracking: `git branch -vv`
   - Remote configuration: `git remote -v`

4. **If creating PR fails with wrong repository:**
   ```bash
   # Reset the default repository
   gh repo set-default youruser/project --yes

   # Retry PR creation
   gh pr create --repo youruser/project ...
   ```
```

### Step 5: Validation Script

Create a validation script to verify configuration:

```bash
#!/bin/bash
# validate-fork-config.sh

set -e

WORKTREE_PATH="$1"
EXPECTED_FORK="$2"  # e.g., "youruser/project"

cd "$WORKTREE_PATH"

echo "Validating fork configuration for: $WORKTREE_PATH"

# Check git remotes
echo "Checking git remotes..."
ORIGIN=$(git remote get-url origin)
if [[ ! "$ORIGIN" =~ "$EXPECTED_FORK" ]]; then
    echo "❌ ERROR: origin does not point to fork"
    echo "   Expected: $EXPECTED_FORK"
    echo "   Got: $ORIGIN"
    exit 1
fi

# Check branch tracking
echo "Checking branch tracking..."
TRACKING=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || echo "none")
if [[ ! "$TRACKING" =~ ^origin/ ]]; then
    echo "⚠️  WARNING: Branch does not track origin"
    echo "   Tracking: $TRACKING"
fi

# Check gh CLI default
echo "Checking gh CLI default repository..."
GH_DEFAULT=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "none")
if [[ "$GH_DEFAULT" != "$EXPECTED_FORK" ]]; then
    echo "❌ ERROR: gh CLI default repository is wrong"
    echo "   Expected: $EXPECTED_FORK"
    echo "   Got: $GH_DEFAULT"
    exit 1
fi

echo "✅ Configuration validated successfully"
```

**Usage:**

```bash
./validate-fork-config.sh ~/.multiclaude/wts/project/agent-1 youruser/project
```

## Common Pitfalls and Solutions

### Pitfall 1: Syncing from Upstream Breaks Tracking

**Problem:** After running `git pull upstream main`, branches may start tracking upstream.

**Solution:**

```bash
# Always sync via rebase, not pull
git fetch upstream
git rebase upstream/main

# Ensure tracking remains on origin
git branch --set-upstream-to=origin/main
```

### Pitfall 2: Worktrees Created from Upstream Branch

**Problem:** Creating worktree from `upstream/main` sets wrong tracking.

**Solution:**

```bash
# Wrong:
git worktree add path -b feature upstream/main

# Correct:
git worktree add path -b feature origin/main
# OR
git worktree add path -b feature main  # (if main tracks origin)
```

### Pitfall 3: Multiple Agents Share Git Config

**Problem:** `gh` CLI config is global per user, not per worktree.

**Solution:** Use `--repo` flag explicitly in agent commands:

```bash
# Always specify the repository explicitly
gh pr create --repo youruser/project --base main ...
```

### Pitfall 4: Agent Doesn't Have Fork Awareness

**Problem:** Agent doesn't know it's working in a fork.

**Solution:** Pass fork information as environment variable:

```bash
export MULTICLAUDE_FORK_REPO="youruser/project"
export MULTICLAUDE_UPSTREAM_REPO="original/project"
```

Update agent prompts to reference these variables.

## Testing the Configuration

### Manual Test

From within a worktree:

```bash
# 1. Verify remotes
git remote -v | grep origin
# Should show: youruser/project

# 2. Verify gh default
gh repo view
# Should show: youruser/project

# 3. Create test branch
git checkout -b test-pr-targeting

# 4. Make dummy commit
echo "test" > test.txt
git add test.txt
git commit -m "test: PR targeting verification"

# 5. Create PR (dry-run)
gh pr create --title "Test PR" --body "Testing" --dry-run
# Should show: Creating pull request for youruser/project

# 6. Cleanup
git checkout main
git branch -D test-pr-targeting
```

### Automated Test

```bash
#!/bin/bash
# test-pr-targeting.sh

WORKTREE="$1"
FORK="$2"

cd "$WORKTREE"

# Create test branch
git checkout -b test-targeting-$$
echo "test" > test-$$.txt
git add .
git commit -m "test: targeting"

# Check where PR would go
PR_TARGET=$(gh pr create --title "Test" --body "Test" --dry-run 2>&1 | grep -oP 'Creating pull request for \K[^ ]+')

# Cleanup
git checkout main
git branch -D test-targeting-$$
rm -f test-$$.txt

# Validate
if [[ "$PR_TARGET" == "$FORK" ]]; then
    echo "✅ PR targeting correct: $PR_TARGET"
    exit 0
else
    echo "❌ PR targeting wrong: $PR_TARGET (expected: $FORK)"
    exit 1
fi
```

## Integration Checklist

When implementing this in your automation system:

- [ ] Repository cloned from fork (not upstream)
- [ ] `origin` remote points to fork
- [ ] `upstream` remote points to original repository (if needed)
- [ ] Worktrees created with branches tracking `origin`
- [ ] `gh repo set-default` run during worktree initialization
- [ ] Agent prompts include PR creation guidelines
- [ ] Validation script runs before agent starts work
- [ ] Environment variables set for fork awareness
- [ ] PR creation commands use `--repo` flag explicitly
- [ ] Test suite validates PR targeting

## Monitoring and Debugging

### Log PR Creation Attempts

Wrap `gh pr create` in logging:

```bash
#!/bin/bash
# create-pr-with-logging.sh

PR_TITLE="$1"
PR_BODY="$2"

# Log current configuration
echo "[PR Creation] Repository: $(gh repo view --json nameWithOwner -q .nameWithOwner)" >> pr-creation.log
echo "[PR Creation] Branch: $(git branch --show-current)" >> pr-creation.log
echo "[PR Creation] Tracking: $(git rev-parse --abbrev-ref --symbolic-full-name @{u})" >> pr-creation.log

# Create PR
gh pr create --repo "$MULTICLAUDE_FORK_REPO" --title "$PR_TITLE" --body "$PR_BODY" 2>&1 | tee -a pr-creation.log
```

### Alert on Wrong Repository

Add validation before PR creation:

```bash
CURRENT_REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
if [[ "$CURRENT_REPO" != "$EXPECTED_FORK" ]]; then
    echo "ERROR: About to create PR against wrong repository!"
    echo "Current: $CURRENT_REPO"
    echo "Expected: $EXPECTED_FORK"
    exit 1
fi
```

## Summary

**Key Principles:**

1. **Clone from fork, not upstream** - `origin` must be your fork
2. **Track origin, not upstream** - Feature branches track `origin/main`
3. **Set gh default explicitly** - Run `gh repo set-default` in each worktree
4. **Always specify --repo** - Don't rely on implicit defaults in automation
5. **Validate before PR creation** - Check configuration before creating PRs

**Quick Reference:**

```bash
# Setup (one time)
git clone git@github.com:youruser/project.git
cd project
git remote add upstream git@github.com:original/project.git

# Worktree creation (per agent)
git worktree add path -b branch-name origin/main
cd path
gh repo set-default youruser/project --yes

# PR creation (in agent)
gh pr create --repo youruser/project --base main --title "..." --body "..."
```

Following these guidelines ensures automated agents consistently create PRs against your fork, avoiding the upstream targeting issue.

---

## Multiclaude-Specific Implementation

Multiclaude implements all of the above principles automatically with multiple layers of enforcement:

### Layer 1: Fork Detection at Initialization

When you run `multiclaude init <github-url>`:

1. **Repository is cloned** to `~/.multiclaude/repos/<repo-name>/`
2. **Fork detection runs** (`internal/worktree/worktree.go:DetectFork()`)
   - Checks for `upstream` remote existence
   - Parses both `origin` and `upstream` URLs
   - Extracts owner/repo for both
3. **Fork metadata is stored** in daemon state (`~/.multiclaude/state.json`)
   ```json
   {
     "repos": {
       "multiclaude": {
         "github_url": "https://github.com/youruser/multiclaude.git",
         "is_fork": true,
         "upstream_owner": "dlorenc",
         "upstream_repo": "multiclaude",
         "fork_owner": "youruser",
         "fork_repo": "multiclaude"
       }
     }
   }
   ```

**Code:**
- `internal/worktree/worktree.go` - `DetectFork()`, `parseGitHubURL()`
- `internal/state/state.go` - `Repository` struct with fork fields
- `internal/cli/cli.go` - `initRepo()` calls fork detection
- `internal/daemon/daemon.go` - `handleAddRepo()` stores fork info

### Layer 2: Agent System Prompts

Worker and merge-queue agents receive fork-aware instructions embedded in their system prompts:

**Worker Prompt** (`internal/prompts/worker.md`):
- **Fork detection section** with bash commands to check `git remote -v`
- **Fork workflow** - explicit instructions to push to `origin` and create PR against `upstream`
- **Non-fork workflow** - standard PR creation for main repositories
- **Pre-creation checklist** - validation steps before `gh pr create`
- **Common mistakes** - examples of wrong vs. correct commands

**Merge-Queue Prompt** (`internal/prompts/merge-queue.md`):
- **Fork repository validation** - checks PR base repository matches upstream
- **Automatic label enforcement** - adds `multiclaude` label if missing
- **Wrong-target detection** - comments on PRs targeting fork instead of upstream
- **Halt merging** - prevents merging PRs with wrong targets

**Code:**
- `internal/prompts/worker.md` - Lines 19-109 (Fork-Aware Pull Request Creation)
- `internal/prompts/merge-queue.md` - Lines 26-89 (Fork Repository PR Validation)

### Layer 3: Pre-Commit Hook Validation

A validation script checks PR targeting before commits:

**Script:** `scripts/validate-pr-target.sh`
- Detects fork status by checking for `upstream` remote
- Parses origin and upstream URLs
- Validates existing PRs target the correct repository
- Provides actionable error messages with fix commands
- Outputs guidance for creating new PRs

**Usage:**
```bash
# Manual run
./scripts/validate-pr-target.sh

# Pre-commit integration
pre-commit run validate-pr-target --hook-stage manual
```

**Code:**
- `scripts/validate-pr-target.sh` - Standalone validation script
- `.pre-commit-hooks.yaml` - Pre-commit framework integration
- `docs/pre-commit-config-example.yaml` - Example configuration

### Layer 4: CI Enforcement

GitHub Actions workflow validates all PRs automatically:

**Workflow:** `.github/workflows/validate-pr-target.yml`

**Triggers:** On PR open, edit, synchronize, reopened

**Checks:**
1. **Fork detection** - Determines if PR is from a fork
2. **Repository target validation** - Ensures fork PRs target upstream (not the fork)
3. **Base branch validation** - Warns if targeting non-main/master branches
4. **Label enforcement** - Auto-adds `multiclaude` label to agent PRs
5. **Failure comments** - Posts actionable fix instructions on failed validation

**Actions:**
- ✅ Pass: PR targets correct repository
- ❌ Fail: PR targets fork instead of upstream (with comment explaining fix)
- ⚠️ Warn: Unusual base branch or missing label (auto-fixed)

**Code:**
- `.github/workflows/validate-pr-target.yml` - Full CI validation workflow

### Layer 5: State-Driven Prompt Generation

Fork information flows from state to agent prompts dynamically:

**Future Enhancement:**
The fork detection state can be used to generate customized prompt sections:

```go
// internal/prompts/prompts.go (future enhancement)
func GetPromptWithForkInfo(repoPath string, agentType AgentType, forkInfo *state.ForkInfo) string {
    basePrompt := GetDefaultPrompt(agentType)

    if forkInfo != nil && forkInfo.IsFork {
        // Inject fork-specific variables into prompt
        forkSection := fmt.Sprintf(`
## Repository Fork Information

This repository is a fork:
- **Your fork:** %s/%s (push here)
- **Upstream:** %s/%s (PR target)

When creating PRs, always use:
gh pr create --repo %s/%s --base main
        `, forkInfo.ForkOwner, forkInfo.ForkRepo,
           forkInfo.UpstreamOwner, forkInfo.UpstreamRepo,
           forkInfo.UpstreamOwner, forkInfo.UpstreamRepo)

        basePrompt = forkSection + "\n\n" + basePrompt
    }

    return basePrompt
}
```

### How It All Works Together

```
┌─────────────────────────────────────────────────────────────┐
│ 1. multiclaude init <fork-url>                              │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Clone Repository   │
         └────────┬───────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Detect Fork        │  (Check for upstream remote)
         │ Parse URLs         │
         └────────┬───────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Store in State     │  (~/.multiclaude/state.json)
         └────────┬───────────┘
                  │
                  ▼
         ┌────────────────────────────────────────────┐
         │ Generate Agent Prompts                     │
         │ - Embed fork detection instructions        │
         │ - Include upstream repo information        │
         └────────┬───────────────────────────────────┘
                  │
                  ▼
         ┌────────────────────────────────────────────┐
         │ Worker Creates PR                          │
         │ - Checks git remote -v                     │
         │ - Pushes to origin                         │
         │ - Creates PR to upstream                   │
         └────────┬───────────────────────────────────┘
                  │
                  ▼
         ┌────────────────────────────────────────────┐
         │ Pre-commit Hook Validates                  │
         │ - Runs scripts/validate-pr-target.sh       │
         │ - Checks PR targets upstream               │
         └────────┬───────────────────────────────────┘
                  │
                  ▼
         ┌────────────────────────────────────────────┐
         │ CI Workflow Enforces                       │
         │ - GitHub Actions validate-pr-target.yml    │
         │ - Auto-adds labels                         │
         │ - Comments on failures                     │
         └────────┬───────────────────────────────────┘
                  │
                  ▼
         ┌────────────────────────────────────────────┐
         │ Merge-Queue Validates Before Merge         │
         │ - Checks PR base == upstream               │
         │ - Halts merge if wrong target              │
         └────────────────────────────────────────────┘
```

### Verification

After setting up a fork with multiclaude, verify the implementation:

```bash
# 1. Check state.json contains fork info
cat ~/.multiclaude/state.json | jq '.repos.multiclaude | {is_fork, upstream_owner, upstream_repo}'

# 2. Verify worker prompt includes fork instructions
grep -A 10 "Fork-Aware Pull Request Creation" ~/.multiclaude/prompts/worker.md

# 3. Test validation script
./scripts/validate-pr-target.sh

# 4. Check CI workflow is active
ls -l .github/workflows/validate-pr-target.yml
```

### Troubleshooting

**Problem:** Fork not detected

**Solution:**
```bash
# Ensure upstream remote exists
cd ~/.multiclaude/repos/<repo-name>
git remote add upstream git@github.com:original-owner/original-repo.git
git fetch upstream

# Re-initialize to re-detect
multiclaude remove <repo-name>
multiclaude init https://github.com/youruser/your-fork.git
```

**Problem:** PRs still targeting wrong repository

**Solution:**
```bash
# Check agent prompt has fork instructions
multiclaude attach worker-123
# Look for "Fork-Aware Pull Request Creation" section

# Check state.json fork metadata
cat ~/.multiclaude/state.json | jq '.repos.<repo>'

# Manually verify in worktree
cd ~/.multiclaude/wts/<repo>/<worker>/
git remote -v
gh repo view
```

### Key Files Reference

| File | Purpose | What It Does |
|------|---------|--------------|
| `internal/worktree/worktree.go:DetectFork()` | Fork detection logic | Checks for upstream remote, parses URLs |
| `internal/state/state.go:Repository` | State storage | Stores `IsFork`, upstream/fork owner/repo |
| `internal/cli/cli.go:initRepo()` | Initialization | Calls DetectFork() after clone |
| `internal/daemon/daemon.go:handleAddRepo()` | Daemon state | Receives and stores fork metadata |
| `internal/prompts/worker.md` | Worker instructions | Fork-aware PR creation guide |
| `internal/prompts/merge-queue.md` | Merge validation | Fork PR target verification |
| `scripts/validate-pr-target.sh` | Pre-commit validation | Standalone PR target checker |
| `.github/workflows/validate-pr-target.yml` | CI enforcement | GitHub Actions validation |

---

## Adapting This Approach

To implement similar fork-awareness in your own automation system:

1. **Detect fork status** - Check for dual remotes (origin + upstream)
2. **Store fork metadata** - Save upstream/fork information in persistent state
3. **Inject into prompts** - Provide fork-aware instructions to agents
4. **Validate at multiple layers**:
   - Pre-commit hooks (developer workflow)
   - CI checks (enforcement)
   - Merge gates (final validation)
5. **Make it visible** - Log fork status, show in UI, include in agent context

The key is **defense in depth**: don't rely on a single validation point, but build multiple layers that catch mistakes at different stages.
