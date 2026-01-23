# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**multiclaude** is a lightweight orchestrator for running multiple Claude Code agents on GitHub repositories. Each agent runs in its own tmux window with an isolated git worktree, enabling parallel autonomous work on a shared codebase.

### The Brownian Ratchet Philosophy

This project embraces controlled chaos: multiple agents work simultaneously, potentially duplicating effort or creating conflicts. **CI is the ratchet** - if tests pass, the code goes in. Progress is permanent.

**Core Beliefs (hardcoded, not configurable):**
- CI is King: Never weaken CI to make work pass
- Forward Progress > Perfection: Partial working solutions beat perfect incomplete ones
- Chaos is Expected: Redundant work is cheaper than blocked work
- Humans Approve, Agents Execute: Agents create PRs but don't bypass review

## Quick Reference

```bash
# Build & Install
go build ./cmd/multiclaude         # Build binary
go install ./cmd/multiclaude       # Install to $GOPATH/bin

# CI Guard Rails (run before pushing)
make pre-commit                    # Fast checks: build + unit tests + verify docs
make check-all                     # Full CI: all checks that GitHub CI runs
make install-hooks                 # Install git pre-commit hook

# Test
go test ./...                      # All tests
go test ./internal/daemon          # Single package
go test -v ./test/...              # E2E tests (requires tmux)
go test ./internal/state -run TestSave  # Single test

# Development
go generate ./pkg/config           # Regenerate CLI docs for prompts
MULTICLAUDE_TEST_MODE=1 go test ./test/...  # Skip Claude startup
```

## 🚨 PRE-PUSH CHECKLIST 🚨

**STOP! Before you push ANY changes, complete this mandatory checklist:**

### 1. Run Local Tests (MANDATORY)

```bash
# REQUIRED before every push
make pre-commit  # Fast checks: build + unit tests
make check-all   # Full CI validation (recommended for significant changes)

# Or run tests directly
go test ./...    # MUST exit with code 0
```

### 2. Type-Specific Validation

**If you modified daemon code (`internal/daemon/`):**
```bash
# Test state persistence and crash recovery
go test ./internal/state/...
go test ./test/ -run Recovery

# Check for race conditions
go test -race ./internal/daemon/...
```

**If you modified agent prompts (`internal/prompts/*.md`):**
```bash
# Rebuild to embed new prompts
go build ./cmd/multiclaude

# Test with a real worker to verify prompt changes
multiclaude work "test task" --repo <test-repo>
multiclaude attach <worker-name> --read-only  # Verify behavior
```

**If you modified CLI commands (`internal/cli/cli.go`):**
```bash
# Regenerate CLI documentation
go generate ./pkg/config

# Test the command manually
multiclaude <your-command> --help
multiclaude <your-command> <args>  # Verify it works
```

**If you modified extension points (`internal/state/`, `internal/events/`, socket API):**
```bash
# Update extension documentation
# See "For LLMs: Keeping Extension Docs Updated" section

# Verify external tools still work
grep -r "YourChangedType" docs/extending/
```

### 3. Pre-Push Verification (DO THIS BEFORE `git push`)

```bash
# 1. Ensure all tests pass
make check-all  # OR: go test ./...

# 2. Verify build succeeds
go build ./cmd/multiclaude

# 3. Check git status for unintended changes
git status
git diff --staged

# 4. Verify commit message is clear
git log -1 --pretty=%B
```

**If ANY of these fail, DO NOT PUSH. Fix the issues first.**

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI (cmd/multiclaude)                    │
└────────────────────────────────┬────────────────────────────────┘
                                 │ Unix Socket
┌────────────────────────────────▼────────────────────────────────┐
│                          Daemon (internal/daemon)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │ Health   │  │ Message  │  │ Wake/    │  │ Socket   │        │
│  │ Check    │  │ Router   │  │ Nudge    │  │ Server   │        │
│  │ (2min)   │  │ (2min)   │  │ (2min)   │  │          │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
└────────────────────────────────┬────────────────────────────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────────┐
    │                            │                                │
┌───▼───┐  ┌───────────┐  ┌─────▼─────┐  ┌──────────┐  ┌────────┐
│super- │  │merge-     │  │workspace  │  │worker-N  │  │review  │
│visor  │  │queue      │  │           │  │          │  │        │
└───────┘  └───────────┘  └───────────┘  └──────────┘  └────────┘
    │           │              │              │             │
    └───────────┴──────────────┴──────────────┴─────────────┘
              tmux session: mc-<repo>  (one window per agent)
```

### Package Responsibilities

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `cmd/multiclaude` | Entry point | `main()` |
| `internal/cli` | All CLI commands | `CLI`, `Command` |
| `internal/daemon` | Background process | `Daemon`, daemon loops |
| `internal/state` | Persistence | `State`, `Agent`, `Repository` |
| `internal/messages` | Inter-agent IPC | `Manager`, `Message` |
| `internal/prompts` | Agent system prompts | Embedded `*.md` files, `GetSlashCommandsPrompt()` |
| `internal/prompts/commands` | Slash command templates | `GenerateCommandsDir()`, embedded `*.md` (legacy) |
| `internal/hooks` | Claude hooks config | `CopyConfig()` |
| `internal/worktree` | Git worktree ops | `Manager`, `WorktreeInfo` |
| `internal/tmux` | Internal tmux client | `Client` (internal use) |
| `internal/socket` | Unix socket IPC | `Server`, `Client`, `Request` |
| `internal/errors` | User-friendly errors | `CLIError`, error constructors |
| `internal/names` | Worker name generation | `Generate()` (adjective-animal) |
| `pkg/config` | Path configuration | `Paths`, `NewTestPaths()` |
| `pkg/tmux` | **Public** tmux library | `Client` (multiline support) |
| `pkg/claude` | **Public** Claude runner | `Runner`, `Config` |

### Data Flow

1. **CLI** parses args → sends `Request` via Unix socket
2. **Daemon** handles request → updates `state.json` → manages tmux
3. **Agents** run in tmux windows with embedded prompts and per-agent slash commands (via `CLAUDE_CONFIG_DIR`)
4. **Messages** flow via filesystem JSON files, routed by daemon
5. **Health checks** (every 2 min) attempt self-healing restoration before cleanup of dead agents

## Key Files to Understand

| File | What It Does |
|------|--------------|
| `internal/cli/cli.go` | **Large file** (~3700 lines) with all CLI commands |
| `internal/daemon/daemon.go` | Daemon implementation with all loops |
| `internal/state/state.go` | State struct with mutex-protected operations |
| `internal/prompts/*.md` | Agent system prompts (embedded at compile) |
| `pkg/tmux/client.go` | Public tmux library with `SendKeysLiteralWithEnter` |

## Patterns and Conventions

### Error Handling

Use structured errors from `internal/errors` for user-facing messages:

```go
// Good: User gets helpful message + suggestion
return errors.DaemonNotRunning()  // "daemon is not running" + "Try: multiclaude start"

// Good: Wrap with context
return errors.GitOperationFailed("clone", err)

// Avoid: Raw errors lose context for users
return fmt.Errorf("clone failed: %w", err)
```

### State Mutations

Always use atomic writes for crash safety:

```go
// internal/state/state.go pattern
func (s *State) saveUnlocked() error {
    data, _ := json.MarshalIndent(s, "", "  ")
    tmpPath := s.path + ".tmp"
    os.WriteFile(tmpPath, data, 0644)  // Write temp
    os.Rename(tmpPath, s.path)          // Atomic rename
}
```

### Tmux Text Input

Use `SendKeysLiteralWithEnter` for atomic text + Enter (prevents race conditions):

```go
// Good: Atomic operation
tmux.SendKeysLiteralWithEnter(session, window, message)

// Avoid: Race condition between text and Enter
tmux.SendKeysLiteral(session, window, message)
tmux.SendEnter(session, window)  // Enter might be lost!
```

### Agent Context Detection

Agents infer their context from working directory:

```go
// internal/cli/cli.go:2385
func (c *CLI) inferRepoFromCwd() (string, error) {
    // Checks if cwd is under ~/.multiclaude/wts/<repo>/ or repos/<repo>/
}
```

## Testing

### Test Categories

| Directory | What | Requirements |
|-----------|------|--------------|
| `internal/*/` | Unit tests | None |
| `test/` | E2E integration | tmux installed |
| `test/recovery_test.go` | Crash recovery | tmux installed |

### Test Mode

```bash
# Skip actual Claude startup in tests
MULTICLAUDE_TEST_MODE=1 go test ./test/...
```

### Writing Tests

```go
// Create isolated test environment using the helper
tmpDir, _ := os.MkdirTemp("", "multiclaude-test-*")
paths := config.NewTestPaths(tmpDir)  // Sets up all paths correctly
defer os.RemoveAll(tmpDir)

// Use NewWithPaths for testing
cli := cli.NewWithPaths(paths, "claude")
```

## CI Monitoring and Failure Recovery

### MANDATORY: Wait for and Verify CI

**After EVERY push, you MUST:**

1. **Immediately start monitoring CI:**
   ```bash
   # Start watching CI run (blocks until complete)
   gh run watch

   # Or for a specific PR
   gh pr checks <PR-number>
   ```

2. **Do not move on to other work until CI is green**
   - If CI fails, fix it immediately
   - Do not push new changes while CI is red
   - Do not consider the task complete until CI passes

3. **If CI fails:**
   ```bash
   # View failure details
   gh run view --log-failed

   # For PR-specific failures
   gh pr checks <PR-number>
   gh pr view <PR-number>
   ```

### Common CI Failures and How to Prevent Them

Learn from common failure patterns to catch issues BEFORE pushing:

#### 1. Test Failures Only in CI

**Symptom:** Tests pass locally but fail in CI

**Prevention:**
```bash
# Always run full test suite before pushing
make check-all  # Runs exactly what CI runs

# Check for race conditions (CI may catch these)
go test -race ./...

# Run E2E tests that require tmux
go test ./test/...
```

**Fix:**
- Investigate environment differences (paths, dependencies)
- Check for timing issues or race conditions
- Add retries or proper synchronization

#### 2. State Management Race Conditions

**Symptom:** Flaky tests, state corruption, concurrent access errors

**Prevention:**
```bash
# Test with race detector
go test -race ./internal/state/...
go test -race ./internal/daemon/...

# Verify atomic writes in state.saveUnlocked()
# Check mutex protection on all state access
```

**Fix:**
- Always use mutex locks for state access
- Verify atomic file writes (write temp → rename)
- Add proper synchronization primitives

#### 3. Tmux Test Failures

**Symptom:** E2E tests fail with tmux errors

**Prevention:**
```bash
# Ensure tmux is running
tmux list-sessions

# Use test mode to skip Claude startup
MULTICLAUDE_TEST_MODE=1 go test ./test/...

# Cleanup test sessions after failures
tmux kill-session -t multiclaude-test-* 2>/dev/null || true
```

**Fix:**
- Check tmux version compatibility
- Ensure tests cleanup sessions properly
- Verify SendKeysLiteralWithEnter usage (not separate SendKeys + Enter)

#### 4. Documentation Out of Sync

**Symptom:** Generated docs don't match code

**Prevention:**
```bash
# Regenerate docs before committing
go generate ./pkg/config

# Verify docs are updated
git status  # Should show docs/cli-reference.md changes
```

**Fix:**
- Run `go generate ./pkg/config` and commit results
- Update extension docs if you changed internal APIs

### CI Debugging Checklist

When CI fails, work through this checklist:

1. **Read the full error message**
   ```bash
   gh run view --log-failed
   ```

2. **Reproduce locally**
   - Can you trigger the same error with `make check-all`?
   - Try with `-race` flag: `go test -race ./...`
   - Run specific failing test: `go test ./pkg -run TestName -v`

3. **Check recent changes**
   ```bash
   git diff main...HEAD
   ```
   - Did you modify state management, daemon loops, or tests?
   - Did you update CLI commands without regenerating docs?

4. **Verify test isolation**
   - Do tests cleanup properly (temp dirs, tmux sessions)?
   - Do tests interfere with each other when run in parallel?

5. **Fix and re-test locally BEFORE pushing fix**
   ```bash
   # MUST pass before pushing
   make check-all

   git add -p  # Stage only relevant fixes
   git commit -m "Fix: <specific issue>"
   git push

   # WAIT and monitor again
   gh run watch
   ```

## Agent System

See `AGENTS.md` for detailed agent documentation including:
- Agent types and their roles
- Message routing implementation
- Prompt system and customization
- Agent lifecycle management
- Adding new agent types

## Extensibility

Multiclaude is designed for extension **without modifying the core binary**. External tools can integrate via:

### Extension Points

| Extension Point | Use Cases | Documentation |
|----------------|-----------|---------------|
| **State File** | Monitoring, dashboards, analytics | [`docs/extending/STATE_FILE_INTEGRATION.md`](docs/extending/STATE_FILE_INTEGRATION.md) |
| **Event Hooks** | Notifications, webhooks, alerting | [`docs/extending/EVENT_HOOKS.md`](docs/extending/EVENT_HOOKS.md) |
| **Socket API** | Custom CLIs, automation, control planes | [`docs/extending/SOCKET_API.md`](docs/extending/SOCKET_API.md) |
| **Web UIs** | Visual monitoring dashboards | [`docs/extending/WEB_UI_DEVELOPMENT.md`](docs/extending/WEB_UI_DEVELOPMENT.md) |

**Start here:** [`docs/EXTENSIBILITY.md`](docs/EXTENSIBILITY.md) - Complete extension guide

### For LLMs: Keeping Extension Docs Updated

**CRITICAL:** When modifying multiclaude core, check if extension documentation needs updates:

1. **State Schema Changes** (`internal/state/state.go`)
   - Update: [`docs/extending/STATE_FILE_INTEGRATION.md`](docs/extending/STATE_FILE_INTEGRATION.md)
   - Update schema reference section
   - Update all code examples showing state structure
   - Run: `go run cmd/verify-docs/main.go` (when implemented)

2. **Event Type Changes** (`internal/events/events.go`)
   - Update: [`docs/extending/EVENT_HOOKS.md`](docs/extending/EVENT_HOOKS.md)
   - Update event type table
   - Update event JSON format examples
   - Add new event examples if new types added

3. **Socket Command Changes** (`internal/daemon/daemon.go`)
   - Update: [`docs/extending/SOCKET_API.md`](docs/extending/SOCKET_API.md)
   - Add/update command reference entries
   - Add code examples for new commands
   - Update client library examples if needed

4. **Runtime Directory Changes** (`pkg/config/config.go`)
   - Update: All extension docs that reference file paths
   - Update the "Runtime Directories" section below
   - Update [`docs/EXTENSIBILITY.md`](docs/EXTENSIBILITY.md) file layout

5. **New Extension Points**
   - Create new guide in `docs/extending/`
   - Add entry to [`docs/EXTENSIBILITY.md`](docs/EXTENSIBILITY.md)
   - Add to this section in `CLAUDE.md`

**Pattern:** After any internal/* or pkg/* changes, search extension docs for outdated references:
```bash
# Find docs that might need updating
grep -r "internal/state" docs/extending/
grep -r "EventType" docs/extending/
grep -r "socket.Request" docs/extending/
```

## Upstream Workflow (For Forks)

If this repository is a fork, see these documents for comprehensive guidance:

**Workflow & Process:**
- `docs/UPSTREAM_WORKFLOW.md` - Bidirectional sync, PR scope enforcement, contribution cadence
- `docs/UPSTREAM_CONTRIBUTION_PLAN.md` - Current status of upstream contributions

**Maintenance & Strategy:**
- `docs/FORK_MAINTENANCE_STRATEGY.md` - Branching conventions, label usage, merge strategy
- `docs/FORK_FEATURES_ROADMAP.md` - Features that stay in fork (not upstream)

**Key principles:**
- **Upstream First**: Contribute everything possible to upstream
- **Focused PRs**: One task = one PR (no exceptions)
- **Proper Labeling**: Use `upstream-ready`, `fork-only`, `upstream-pending` labels
- **Branch Naming**: `upstream/*` for upstream contributions, `fork/*` for fork-only features

## Contributing Checklist

When modifying agent behavior:
- [ ] Update the relevant prompt in `internal/prompts/*.md`
- [ ] Run `go generate ./pkg/config` if CLI changed
- [ ] Test with tmux: `go test ./test/...`
- [ ] Check state persistence: `go test ./internal/state/...`

When adding CLI commands:
- [ ] Add to `registerCommands()` in `internal/cli/cli.go`
- [ ] Use `internal/errors` for user-facing errors
- [ ] Add help text with `Usage` field
- [ ] Regenerate docs: `go generate ./pkg/config`

When modifying daemon loops:
- [ ] Consider interaction with health check (2 min cycle)
- [ ] Test crash recovery: `go test ./test/ -run Recovery`
- [ ] Verify state atomicity with concurrent access tests

When modifying extension points (state, events, socket API):
- [ ] Update relevant extension documentation in `docs/extending/`
- [ ] Update code examples in docs to match new behavior
- [ ] Run documentation verification (when implemented): `go run cmd/verify-docs/main.go`
- [ ] Check that external tools still work (e.g., `cmd/multiclaude-web`)

When modifying any code:
- [ ] Run `make pre-commit` BEFORE pushing (build + tests)
- [ ] Monitor CI immediately after pushing: `gh run watch`
- [ ] Verify all CI checks pass before considering work done
- [ ] If CI fails, fix immediately and re-test locally before re-pushing

## Quality Gates - MANDATORY

**🚨 WORK IS NOT COMPLETE UNTIL ALL CHECKS ARE GREEN 🚨**

### Three-Phase Quality Gate Process

This aligns with our "CI is King" philosophy: we ensure tests pass, we don't weaken tests to make them pass.

**Phase 1: Pre-Push (BEFORE `git push`)**

1. **Local tests MUST pass:**
   ```bash
   make pre-commit  # MUST return exit code 0
   # OR for full validation:
   make check-all   # Recommended for significant changes
   ```
   If this fails, DO NOT push. Fix issues first.

2. **Type-specific validation MUST pass:**
   - Daemon changes: Test state persistence and crash recovery
   - Prompt changes: Rebuild and test with real worker
   - CLI changes: Regenerate docs with `go generate ./pkg/config`
   - Extension changes: Update docs in `docs/extending/`
   - See "PRE-PUSH CHECKLIST" section above for details

3. **Build MUST succeed:**
   ```bash
   go build ./cmd/multiclaude  # MUST complete without errors
   ```

**Phase 2: Post-Push (IMMEDIATELY after `git push`)**

4. **Start CI monitoring (DO NOT walk away):**
   ```bash
   # Required immediately after pushing
   gh run watch
   ```

   **You MUST wait for CI to complete. Do not:**
   - Move on to other work
   - End your session
   - Consider the task complete
   - Push additional changes

5. **CI MUST be green - ALL checks must pass:**
   ```bash
   # Verify all checks pass
   gh pr checks  # All should show ✅
   ```

   **Required passing checks:**
   - ✅ All Go tests pass (`go test ./...`)
   - ✅ Build succeeds on all platforms
   - ✅ No race conditions detected
   - ✅ Documentation is up to date
   - ✅ E2E tests pass (if modified relevant code)

**Phase 3: Failure Recovery (If ANY check fails)**

6. **If CI fails, immediate action required:**
   ```bash
   # View failure details
   gh run view --log-failed

   # Fix the issue (see "CI Monitoring and Failure Recovery" section)

   # Re-run local tests to verify fix
   make check-all  # MUST pass

   # Push fix
   git add -p
   git commit -m "Fix: <specific issue>"
   git push

   # WAIT and monitor again
   gh run watch
   ```

7. **Work is ONLY complete when:**
   - [ ] All local tests pass (`make check-all` or `go test ./...` returns 0)
   - [ ] Build succeeds locally
   - [ ] All CI checks are green (verified with `gh pr checks`)
   - [ ] No CI runs are in "failed" or "pending" state
   - [ ] You have personally verified all checks passed

**Failing any of these checks means the work is incomplete. Fix all failures before considering work done.**

### What "Green CI" Means

CI is NOT green if:
- ❌ Any check shows a red X
- ❌ Any check is still running (yellow circle)
- ❌ Workflow shows "cancelled" or "skipped"
- ❌ You didn't wait for checks to complete

CI IS green when:
- ✅ All checks show green checkmarks
- ✅ No checks are running or pending
- ✅ PR shows "All checks have passed"
- ✅ You have verified this personally with `gh pr checks`

### Philosophy: CI is King

**"CI is King: Never weaken CI to make work pass"**

This means:
- If tests fail, fix the code, don't skip tests
- If CI catches a race condition, add proper synchronization
- If CI is flaky, fix the flakiness, don't ignore it
- Tests exist to protect the codebase - respect them

**Remember: You are responsible for ensuring your changes work in CI, not just locally.**

## Runtime Directories

```
~/.multiclaude/
├── daemon.pid              # Daemon PID (lock file)
├── daemon.sock             # Unix socket for CLI<->daemon
├── daemon.log              # Daemon logs (rotated at 10MB)
├── state.json              # All state (repos, agents, config)
├── prompts/                # Generated prompt files for agents
├── repos/<repo>/           # Cloned repositories
├── wts/<repo>/<agent>/     # Git worktrees (one per agent)
├── messages/<repo>/<agent>/ # Message JSON files
├── output/<repo>/          # Agent output logs
│   └── workers/            # Worker-specific logs
└── claude-config/<repo>/<agent>/ # Per-agent CLAUDE_CONFIG_DIR
    └── commands/           # Slash command files (*.md)
```

## Common Operations

### Debug a stuck agent

```bash
# Attach to see what it's doing
multiclaude attach <agent-name> --read-only

# Check its messages
multiclaude agent list-messages  # (from agent's tmux window)

# Manually nudge via daemon logs
tail -f ~/.multiclaude/daemon.log
```

### Repair inconsistent state

```bash
# Local repair (no daemon)
multiclaude repair

# Daemon-side repair
multiclaude cleanup --dry-run  # See what would be cleaned
multiclaude cleanup            # Actually clean up
```

### Test prompt changes

```bash
# Prompts are embedded at compile time
vim internal/prompts/worker.md
go build ./cmd/multiclaude
# New workers will use updated prompt
```

### Complete workflow for making changes

**Standard development workflow:**

```bash
# 1. Make your changes
vim internal/daemon/daemon.go

# 2. PRE-PUSH: Test locally (MANDATORY)
make pre-commit  # Fast checks
# OR for significant changes:
make check-all   # Full CI validation

# 3. PRE-PUSH: Verify build
go build ./cmd/multiclaude

# 4. Commit and push
git add -p
git commit -m "Clear description of change"
git push

# 5. POST-PUSH: Monitor (DO NOT walk away)
gh run watch  # REQUIRED - wait for completion

# 6. POST-PUSH: Verify all checks passed
gh pr checks  # All should show ✅

# 7. If any check failed: Fix immediately
gh run view --log-failed  # Debug the failure
# Fix the issue
make check-all  # Verify fix locally
git add -p && git commit -m "Fix: <issue>" && git push
gh run watch  # Monitor again
```

**Work is complete when:**
- ✅ Local tests pass (`make check-all`)
- ✅ Build succeeds
- ✅ All CI checks are green
- ✅ You have personally verified with `gh pr checks`

**Example: Modifying agent prompts**

```bash
# 1. Edit prompt
vim internal/prompts/worker.md

# 2. Rebuild (prompts are embedded at compile time)
go build ./cmd/multiclaude

# 3. Test with real worker
multiclaude work "test the new prompt behavior" --repo test-repo
multiclaude attach <worker-name> --read-only  # Verify behavior

# 4. Run tests
make pre-commit

# 5. Commit and push
git add internal/prompts/worker.md
git commit -m "Update worker prompt to clarify X"
git push

# 6. Monitor CI
gh run watch  # WAIT for green
```

**Example: Adding CLI command**

```bash
# 1. Add command to internal/cli/cli.go
vim internal/cli/cli.go

# 2. Regenerate docs
go generate ./pkg/config

# 3. Test the command
go build ./cmd/multiclaude
./multiclaude <new-command> --help
./multiclaude <new-command> <test-args>

# 4. Run tests
make check-all

# 5. Commit (include generated docs!)
git add internal/cli/cli.go docs/cli-reference.md
git commit -m "Add <command> to do X"
git push

# 6. Monitor CI
gh run watch
```
