# Upstream Contribution Plan

This document tracks what we should contribute back to `dlorenc/multiclaude` upstream.

## Recently Merged to Fork (Candidates for Upstream)

### ✅ Ready to Contribute

These are **core improvements** that fit upstream's scope and should be contributed:

#### 1. CI Guard Rails (#46)
**What**: Makefile + pre-commit hooks for local CI validation
**Why upstream wants it**: Prevents broken commits, improves developer experience
**Files**: `Makefile`, `.git/hooks/pre-commit`, docs updates
**Upstream PR**: Create from `feat/ci-guard-rails` branch

#### 2. Fork-Aware Workflows (#47)
**What**: Auto-detect forks and provide workflow guidance to agents
**Why upstream wants it**: Helps contributors working in forks
**Files**: `internal/prompts/prompts.go`, tests
**Upstream PR**: Create from `feat/fork-aware-workflows` branch

#### 3. Upstream Sync Enforcement (#48)
**What**: Agent prompts for bidirectional sync and PR scope enforcement
**Why upstream wants it**: Better contribution workflow, focused PRs
**Files**: `internal/prompts/*.md`, `docs/UPSTREAM_WORKFLOW.md`
**Upstream PR**: Create from `feat/upstream-sync-enforcement` branch

#### 4. Aggressive PR Pushing (#49)
**What**: Prompts encouraging continuous small PRs
**Why upstream wants it**: Aligns with Brownian Ratchet philosophy
**Files**: `internal/prompts/worker.md`, `internal/prompts/supervisor.md`
**Upstream PR**: Create from `feat/aggressive-pr-pushing` branch

### ⚠️ Needs Discussion

#### 5. Dual-Layer CI Design (#50)
**What**: Design proposal for fork/upstream CI validation
**Why it might be rejected**: Complex, fork-specific
**Action**: Share as discussion/issue first, not PR

### ✅ Already Contributed (Merged to Fork, Ready for Upstream)

These were merged to fork main and should be contributed:

1. **Enhanced task history** (#21) - Better task tracking with summaries
2. **Settings.json copying** (#22) - Workers inherit global Claude config
3. **Automatic worktree sync** (#23) - Keep worktrees in sync with main
4. **Worktree creation fixes** (#24) - Handle existing branches correctly
5. **`multiclaude nuke` command** (#25) - Reset to clean state
6. **Smart error detection** (#26) - Better error messages with suggestions
7. **Agent restart command** (#27) - Restart crashed agents

## Contribution Workflow

### Step 1: Prepare the Branch

```bash
# Make sure we're up to date with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create a clean branch from the feature
git checkout -b upstream/ci-guard-rails feat/ci-guard-rails

# Rebase on upstream main to ensure clean history
git rebase upstream/main

# Push to fork
git push origin upstream/ci-guard-rails
```

### Step 2: Create Upstream PR

```bash
# Create PR targeting upstream
gh pr create \
  --repo dlorenc/multiclaude \
  --base main \
  --head aronchick:upstream/ci-guard-rails \
  --title "feat: Add CI guard rails for local validation" \
  --body "$(cat .github/pr-templates/upstream.md)"
```

### Step 3: Track Upstream PR

Add to this document:

```markdown
### Pending Upstream Review

- [ ] #XX - CI guard rails - https://github.com/dlorenc/multiclaude/pull/XX
- [ ] #XX - Fork-aware workflows - https://github.com/dlorenc/multiclaude/pull/XX
```

### Step 4: Handle Feedback

If upstream requests changes:
1. Make changes on the `upstream/*` branch
2. Push updates
3. Merge back to fork if needed

If upstream rejects:
1. Document why in `FORK_MAINTENANCE_STRATEGY.md`
2. Keep feature in fork
3. Ensure it doesn't conflict with future upstream changes

## PR Template for Upstream

Create `.github/pr-templates/upstream.md`:

```markdown
## Summary

[Brief description of what this PR does]

## Motivation

[Why this change is valuable for multiclaude]

## Changes

- [List of changes]

## Testing

- [x] All tests pass: `go test ./...`
- [x] E2E tests pass: `go test ./test/...`
- [x] Manual testing: [describe what you tested]

## Compatibility

- [x] No breaking changes
- [x] Backward compatible with existing setups
- [x] Documentation updated

## Related

- Fixes #XX (if applicable)
- Related to #XX (if applicable)

---

This PR comes from the `aronchick/multiclaude` fork where it has been tested in production.
```

## Current Status

### Merged to Fork ✅
- Enhanced task history (#21)
- Settings.json copying (#22)
- Automatic worktree sync (#23)
- Worktree creation fixes (#24)
- `multiclaude nuke` command (#25)
- Smart error detection (#26)
- Agent restart command (#27)
- CI guard rails (#46)
- Fork-aware workflows (#47)
- Upstream sync enforcement (#48)
- Aggressive PR pushing (#49)
- Dual-layer CI design (#50)

### Ready for Upstream Contribution 🎯
- [ ] CI guard rails (#46)
- [ ] Fork-aware workflows (#47)
- [ ] Upstream sync enforcement (#48)
- [ ] Aggressive PR pushing (#49)
- [ ] Enhanced task history (#21)
- [ ] Settings.json copying (#22)
- [ ] Automatic worktree sync (#23)
- [ ] Worktree creation fixes (#24)
- [ ] `multiclaude nuke` command (#25)
- [ ] Smart error detection (#26)
- [ ] Agent restart command (#27)

### Pending Upstream Review 🔄
(None yet - create PRs above)

### Rejected by Upstream ❌
(None yet)

### Fork-Only Features 🔒
(See `FORK_FEATURES_ROADMAP.md`)
- Slack integration (planned)
- Web dashboard (planned)
- Multi-machine monitoring (planned)

## Next Actions

1. **Create upstream PRs** for the 11 ready features above
2. **Group related changes** - consider combining some PRs:
   - Task history + agent restart (both about agent lifecycle)
   - Worktree sync + creation fixes (both about worktrees)
   - CI guard rails + error detection (both about developer experience)
3. **Start with smallest PRs** - easier to review and merge
4. **Build relationship** - engage with Dan/upstream maintainers

## Success Metrics

- **Acceptance rate**: % of PRs merged upstream
- **Time to merge**: How long PRs sit in review
- **Feedback quality**: Are we aligned with upstream vision?
- **Fork divergence**: How many commits ahead of upstream?

Target: 80%+ acceptance rate, <1 week to merge, <10 commits ahead of upstream

