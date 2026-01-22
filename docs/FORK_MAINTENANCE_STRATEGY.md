# Fork Maintenance Strategy

This document defines how we maintain the `aronchick/multiclaude` fork while contributing back to `dlorenc/multiclaude` upstream.

## Philosophy

**Upstream First, Fork Second**: We contribute everything possible to upstream. The fork only contains features that upstream explicitly rejects or are specific to our use case.

## Upstream Scope (What Dan/dlorenc Accepts)

Based on ROADMAP.md and DESIGN.md, upstream accepts:

✅ **Core functionality improvements**
- Bug fixes
- Performance improvements
- Better error messages
- Agent lifecycle improvements
- Worktree management
- CI/testing improvements

✅ **Developer experience**
- Better CLI commands
- Improved documentation
- Local development tools (Makefile, pre-commit hooks)
- Crash recovery

✅ **Agent intelligence**
- Better prompts
- Fork detection and workflow guidance
- Upstream sync automation
- PR scope enforcement

## Upstream Rejects (Fork-Only Features)

Based on ROADMAP.md "Out of Scope" section:

❌ **Web interfaces or dashboards**
- No REST APIs for external consumption
- No browser-based UIs
- Terminal is the interface

❌ **Remote/hybrid deployment**
- No cloud coordination
- No distributed orchestration

❌ **Multi-provider support**
- Claude-only, no OpenAI/Gemini/etc.

❌ **External integrations**
- No Slack, Discord, etc. (not terminal-native)
- No issue tracker integrations beyond GitHub

## Fork-Specific Features to Maintain

### 1. Slack Integration (Planned)

**Location**: `internal/integrations/slack/`

**Purpose**: Send notifications about agent activity, PR status, CI failures

**Maintenance Strategy**:
- Keep as a separate package that doesn't touch core code
- Use daemon hooks/events to trigger notifications
- Document clearly as fork-only feature
- Disable by default (opt-in via config)

**Implementation Approach**:
```go
// internal/integrations/slack/client.go
// Fork-only: Slack notifications for multiclaude events
```

### 2. Web Dashboard (Planned)

**Location**: `cmd/multiclaude-web/` and `web/`

**Purpose**: View status of all multiclaude instances across multiple machines

**Maintenance Strategy**:
- Separate binary (`multiclaude-web`) that reads state.json
- Read-only view of daemon state
- No control plane (stays terminal-native for core)
- Optional component, not required for core functionality

**Implementation Approach**:
```
cmd/multiclaude-web/     # Separate web server binary
web/                     # Static assets
  ├── index.html
  ├── dashboard.js
  └── styles.css
```

### 3. Fork-Specific Configuration

**Location**: `.multiclaude/fork-config.json`

**Purpose**: Configuration for fork-only features

**Example**:
```json
{
  "integrations": {
    "slack": {
      "enabled": false,
      "webhook_url": "",
      "notify_on": ["pr_created", "ci_failed", "worker_stuck"]
    },
    "web_dashboard": {
      "enabled": false,
      "port": 8080,
      "read_only": true
    }
  }
}
```

## Merge Strategy

### Merging to Fork Main

1. **Test locally first**: `make check-all`
2. **Merge to fork/main**: All PRs go here first
3. **Evaluate for upstream**: Does it fit upstream scope?
   - ✅ Yes → Create upstream PR from feature branch
   - ❌ No → Keep in fork, document as fork-only

### Syncing from Upstream

```bash
# Every week or when upstream has significant changes
git fetch upstream
git checkout main
git merge upstream/main
git push origin main

# Handle conflicts carefully - preserve fork-only features
```

### Creating Upstream PRs

```bash
# Work on feature branch
git checkout -b feat/better-error-messages

# Make changes, test locally
make check-all

# Push to fork
git push origin feat/better-error-messages

# Create PR to UPSTREAM (not fork)
gh pr create --repo dlorenc/multiclaude \
  --title "feat: Better error messages for git failures" \
  --body "Improves error messages with specific suggestions..."
```

## File Organization

### Core Files (Sync with Upstream)
- `cmd/multiclaude/` - Main CLI
- `internal/` - All core packages
- `pkg/` - Public libraries
- `test/` - Tests
- `docs/` - Documentation (except fork-specific)

### Fork-Only Files (Never Upstream)
- `internal/integrations/` - External integrations
- `cmd/multiclaude-web/` - Web dashboard
- `web/` - Web assets
- `docs/FORK_MAINTENANCE_STRATEGY.md` - This file
- `.multiclaude/fork-config.json` - Fork configuration

## Contribution Workflow

### For Core Features (Upstream-Bound)

1. Create feature branch: `feat/feature-name`
2. Develop and test locally
3. Create PR to **fork** first
4. Merge to fork/main after review
5. Create PR to **upstream** from same branch
6. If upstream accepts: ✅ Done
7. If upstream rejects: Document why, keep in fork

### For Fork-Only Features

1. Create feature branch: `fork/feature-name` (prefix with `fork/`)
2. Develop in isolation from core
3. Create PR to **fork** only
4. Document as fork-only in PR description
5. Never create upstream PR

## Testing Fork-Only Features

```bash
# Test core + fork features
make check-all

# Test only core (what upstream tests)
go test ./cmd/multiclaude/... ./internal/... ./pkg/... ./test/...

# Test only fork features
go test ./internal/integrations/... ./cmd/multiclaude-web/...
```

## Documentation Strategy

### Upstream Documentation
- README.md - Core features only
- CONTRIBUTING.md - Upstream contribution guide
- AGENTS.md, CLAUDE.md, DESIGN.md - Core architecture

### Fork Documentation
- docs/FORK_MAINTENANCE_STRATEGY.md - This file
- docs/SLACK_INTEGRATION.md - Slack setup (fork-only)
- docs/WEB_DASHBOARD.md - Dashboard setup (fork-only)
- README_FORK.md - Fork-specific features and setup

## Version Management

We don't maintain separate version numbers. Instead:

- Fork tracks upstream main branch
- Fork-only features are additive (don't break core)
- Use git tags for fork releases: `fork-v1.0.0`

## When Upstream Rejects a PR

1. **Understand why**: Read feedback carefully
2. **Document decision**: Add to this file
3. **Keep in fork**: If feature is valuable to us
4. **Maintain separately**: Ensure it doesn't conflict with upstream changes
5. **Revisit periodically**: Upstream scope may change

## Current Status

### Merged to Fork ✅
- Enhanced task history (#21)
- Settings.json copying (#22)
- Automatic worktree sync (#23)
- Worktree creation fixes (#24)
- `multiclaude nuke` command (#25)
- Smart error detection (#26)
- Agent restart command (#27)

### Pending Fork Merge 🔄
- CI guard rails (#46)
- Fork-aware workflows (#47)
- Upstream sync enforcement (#48)
- Aggressive PR pushing (#49)
- Dual-layer CI design (#50)

### Planned Fork-Only Features 📋
- Slack integration
- Web dashboard
- Multi-machine monitoring

### Contributed to Upstream 🎯
- (Track upstream PRs here as we create them)

