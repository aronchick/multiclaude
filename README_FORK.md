# aronchick/multiclaude Fork

This is a fork of [dlorenc/multiclaude](https://github.com/dlorenc/multiclaude) with additional features and improvements.

## Fork Philosophy

**Upstream First**: We contribute everything possible back to upstream. This fork only contains:
1. Features waiting for upstream review
2. Features explicitly rejected by upstream (web UI, Slack, etc.)
3. Experimental features being tested before upstream contribution

## What's Different in This Fork?

### ✅ Merged to Fork (Pending Upstream Contribution)

These features are production-ready and will be contributed upstream:

1. **CI Guard Rails** - Local validation before pushing (Makefile, pre-commit hooks)
2. **Fork-Aware Workflows** - Auto-detect forks and guide agents on proper workflow
3. **Upstream Sync Enforcement** - Agent prompts for bidirectional sync
4. **Aggressive PR Pushing** - Encourage continuous small PRs
5. **Enhanced Task History** - Summaries and failure reasons
6. **Settings.json Copying** - Workers inherit global Claude config
7. **Automatic Worktree Sync** - Keep worktrees in sync with main
8. **Worktree Creation Fixes** - Handle existing branches correctly
9. **`multiclaude nuke` Command** - Reset to clean state
10. **Smart Error Detection** - Better error messages with suggestions
11. **Agent Restart Command** - Restart crashed agents

### 📋 Planned Fork-Only Features

These features are **not** in upstream's scope (per their ROADMAP.md):

1. **Slack Integration** - Real-time notifications to Slack
2. **Web Dashboard** - Browser-based monitoring across machines
3. **Multi-Machine Monitoring** - Centralized view of multiple instances

See [docs/FORK_FEATURES_ROADMAP.md](docs/FORK_FEATURES_ROADMAP.md) for details.

## Documentation

### Fork-Specific Docs
- [FORK_MAINTENANCE_STRATEGY.md](docs/FORK_MAINTENANCE_STRATEGY.md) - How we maintain the fork
- [FORK_FEATURES_ROADMAP.md](docs/FORK_FEATURES_ROADMAP.md) - Fork-only features
- [UPSTREAM_CONTRIBUTION_PLAN.md](docs/UPSTREAM_CONTRIBUTION_PLAN.md) - What we're contributing upstream

### Upstream Docs
- [README.md](README.md) - Main documentation (from upstream)
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guide
- [AGENTS.md](AGENTS.md) - Agent system documentation
- [CLAUDE.md](CLAUDE.md) - Claude Code integration guide

## Installation

Same as upstream:

```bash
go install github.com/aronchick/multiclaude/cmd/multiclaude@latest
```

Or build from source:

```bash
git clone https://github.com/aronchick/multiclaude.git
cd multiclaude
go build ./cmd/multiclaude
```

## Usage

Identical to upstream. See [README.md](README.md) for full documentation.

### Fork-Specific Features

Once implemented, fork-only features will be documented in:
- [docs/SLACK_INTEGRATION.md](docs/SLACK_INTEGRATION.md) (coming soon)
- [docs/WEB_DASHBOARD.md](docs/WEB_DASHBOARD.md) (coming soon)

## Contributing

### Contributing to Upstream

If your contribution fits upstream's scope (see [ROADMAP.md](ROADMAP.md)), please contribute directly to [dlorenc/multiclaude](https://github.com/dlorenc/multiclaude).

See [docs/UPSTREAM_CONTRIBUTION_PLAN.md](docs/UPSTREAM_CONTRIBUTION_PLAN.md) for our contribution workflow.

### Contributing Fork-Only Features

For features that don't fit upstream's scope:

1. Check [docs/FORK_FEATURES_ROADMAP.md](docs/FORK_FEATURES_ROADMAP.md) to see if it's planned
2. Open an issue to discuss the feature
3. Create a PR targeting this fork

## Syncing with Upstream

We regularly sync with upstream to stay current:

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

## Differences from Upstream

### Code Changes
- Additional packages in `internal/integrations/` (fork-only)
- Event system in daemon (for Slack/web dashboard)
- Fork-specific configuration in `.multiclaude/fork-config.json`

### Documentation
- Fork-specific docs in `docs/FORK_*.md`
- `README_FORK.md` (this file)

### No Breaking Changes
All fork features are **additive** - you can use this fork as a drop-in replacement for upstream.

## Why Fork?

Upstream has a clear, focused scope (see [ROADMAP.md](ROADMAP.md)):
- ✅ Local-first, terminal-native, Claude-only
- ❌ No web UIs, no external integrations, no remote deployment

We respect this scope but need additional features for our use case:
- **Slack integration** for team notifications
- **Web dashboard** for monitoring across machines
- **Multi-machine coordination** for DevOps workflows

Rather than pressure upstream to accept features outside their scope, we maintain this fork.

## Relationship with Upstream

We actively contribute back to upstream:
- Bug fixes
- Core improvements
- Developer experience enhancements
- Documentation improvements

See [docs/UPSTREAM_CONTRIBUTION_PLAN.md](docs/UPSTREAM_CONTRIBUTION_PLAN.md) for our contribution strategy.

## License

Same as upstream: Apache 2.0 (see [LICENSE](LICENSE))

## Credits

- **Upstream**: [dlorenc/multiclaude](https://github.com/dlorenc/multiclaude) by Dan Lorenc
- **Fork maintainer**: David Aronchick ([@aronchick](https://github.com/aronchick))

## Support

- **Upstream issues**: Report to [dlorenc/multiclaude](https://github.com/dlorenc/multiclaude/issues)
- **Fork-specific issues**: Report to [aronchick/multiclaude](https://github.com/aronchick/multiclaude/issues)

When in doubt, report to upstream first - they're the experts!

