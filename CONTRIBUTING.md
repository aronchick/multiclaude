# Contributing to multiclaude

## Philosophy

multiclaude values **forward progress over perfection**. This means:

- Three okay PRs beat one perfect PR
- Small, frequent commits beat large, infrequent ones
- Working code now beats perfect code later
- CI is the arbiter: if tests pass, the code can ship

## Recommended Workflow: Upstream Changes One at a Time

When you've been working on a branch and want to upstream your changes incrementally:

### 1. Identify Logical Chunks

Review your changes and identify pieces that can stand alone:

```bash
# See all commits on your branch
git log main..HEAD --oneline

# See what files changed
git diff main --stat
```

Good chunk candidates:
- A single bug fix
- One new function with its tests
- A refactor that doesn't change behavior
- Documentation updates

### 2. Create Focused PRs

**Option A: Clean commit history (each commit is a logical unit)**

Create a branch for each commit you want to upstream:

```bash
# From main, cherry-pick the commit you want
git checkout main
git pull
git checkout -b fix/typo-in-readme
git cherry-pick <commit-hash>
git push -u origin fix/typo-in-readme
gh pr create
```

**Option B: Messy history (need to reorganize)**

Use interactive rebase to split or combine commits:

```bash
# Create a new branch for the chunk you want
git checkout main
git pull
git checkout -b feature/add-validation

# Cherry-pick relevant commits, or use:
git checkout your-working-branch -- path/to/specific/file.go

# Make a clean commit
git add -p  # Stage specific hunks
git commit -m "Add input validation to user handler"
```

**Option C: Stacked PRs (changes build on each other)**

When changes depend on each other, create a stack:

```bash
# First PR: base functionality
git checkout main && git pull
git checkout -b feature/base
# ... make changes ...
git push -u origin feature/base
gh pr create --base main

# Second PR: builds on first
git checkout -b feature/extension
# ... make changes ...
git push -u origin feature/extension
gh pr create --base feature/base  # Note: base is first PR's branch
```

Update the stack as PRs merge:
```bash
# After feature/base merges to main
git checkout feature/extension
git rebase main
git push --force-with-lease
# Update PR base to main via GitHub UI or:
gh pr edit --base main
```

### 3. Keep Working While PRs Review

Don't wait for PRs to merge. Keep your working branch moving:

```bash
# Your working branch continues from where you are
# PRs will merge independently via the merge queue
```

If a PR needs changes:
```bash
git checkout fix/typo-in-readme
# Make fixes
git commit --amend  # or new commit
git push --force-with-lease
```

### 4. Sync After Merges

After your PRs merge, sync your working branch:

```bash
git checkout your-working-branch
git fetch origin
git rebase origin/main
# Resolve any conflicts (your changes may already be in main now)
```

## PR Guidelines

### Size

- Prefer small, focused PRs (under 200 lines when possible)
- Split large changes into a series of incremental PRs
- Each PR should be independently reviewable and mergeable

### Commits

- Each commit should compile and pass tests
- Write clear commit messages explaining "why" not just "what"
- Squash WIP commits before opening PR

### CI

- All tests must pass
- Never skip or weaken CI to make your PR pass
- If tests are flaky, fix the flakiness (or report it)

**Run CI checks locally before pushing:**

```bash
# Quick check (runs build, unit tests, verify docs)
make pre-commit

# Full CI validation (includes E2E tests and coverage)
make check-all
```

**Install git pre-commit hook:**

```bash
make install-hooks
```

This installs a git hook that automatically runs `make pre-commit` before each commit, catching CI issues locally. To skip temporarily: `git commit --no-verify`

**Individual CI checks:**

```bash
make build        # Build all packages
make unit-tests   # Run unit tests
make e2e-tests    # Run E2E tests (slower)
make verify-docs  # Check generated docs are up to date
make coverage     # Coverage report
```

### Review

- PRs from workers are monitored by the merge queue
- Human PRs follow normal GitHub review process
- Respond to review feedback promptly

## Quick Reference

```bash
# See what you have to upstream
git log main..HEAD --oneline

# Create a PR for a single commit
git checkout main && git pull
git checkout -b pr/description
git cherry-pick <hash>
git push -u origin pr/description
gh pr create

# Create a PR for specific files
git checkout main && git pull
git checkout -b pr/description
git checkout your-branch -- path/to/file
git commit -m "Description"
git push -u origin pr/description
gh pr create

# Sync your branch after merges
git fetch origin && git rebase origin/main
```

## Contributing from a Fork

External contributors should work from a fork:

### Setup

```bash
# Fork the repo on GitHub, then:
git clone https://github.com/YOUR-USERNAME/multiclaude.git
cd multiclaude
git remote add upstream https://github.com/dlorenc/multiclaude.git
```

### Workflow

1. **Work in your fork** - Develop on a feature branch in your fork
2. **Test locally** - Run `go test ./...` to verify everything passes
3. **Create upstream PR** - Open a PR from your fork to the main repo

```bash
# Keep your fork updated
git fetch upstream
git checkout main
git merge upstream/main
git push origin main

# Create a feature branch
git checkout -b feature/my-change

# ... make changes ...

# Push to your fork
git push -u origin feature/my-change

# Create PR to upstream (via GitHub UI or gh CLI)
gh pr create --repo dlorenc/multiclaude
```

### Upstreaming Multiple Changes

When you have several changes to contribute:

```bash
# See what you have that upstream doesn't
git log upstream/main..HEAD --oneline

# For each logical change, create a separate branch and PR:
git checkout upstream/main
git checkout -b fix/first-thing
git cherry-pick <hash>
git push -u origin fix/first-thing
gh pr create --repo dlorenc/multiclaude

git checkout upstream/main
git checkout -b fix/second-thing
git cherry-pick <hash>
git push -u origin fix/second-thing
gh pr create --repo dlorenc/multiclaude
```

This keeps PRs small, focused, and independently mergeable.

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- For questions about the codebase, see `CLAUDE.md`
