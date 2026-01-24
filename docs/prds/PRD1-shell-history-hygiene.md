# PRD1: Shell History Hygiene for multiclaude

## Overview
multiclaude spawns supervisor/worker/review agents that execute many shell commands. Those commands currently end up in the user's shell history (bash/zsh) and in external history tools like Atuin. This pollutes history, makes interactive recall unusable, and can leak sensitive commands into searchable history. This PRD ensures multiclaude-launched agents do not write to shell history by default, while preserving normal history behavior for the user's own interactive shell.

## Goals
- Prevent multiclaude agent commands from being recorded in bash/zsh history.
- Best-effort prevent commands from being recorded in tools like Atuin.
- Make behavior consistent across all agent types (supervisor, worker, merge-queue, review, workspace).
- Provide an opt-out for users who want to keep history.
- Avoid regressions to interactive user shells.

## Non-Goals
- Modifying or sanitizing the user’s existing history.
- Universal guarantees for all shells and all third-party tools.
- Changing how history works for users outside multiclaude-managed processes.

## Background / Context
- multiclaude agents run many automated commands.
- Shell history pollution reduces usability and can surface sensitive commands.
- Users need predictable history behavior and the ability to opt out.

## Roadmap Alignment
- Aligns with core workflow reliability and user experience goals.
- Not listed as out-of-scope in `ROADMAP.md`.

## CI / Quality Gates
- No weakening or bypassing CI.
- Add tests for config parsing and environment injection.

## Scope & PR Strategy
- Expected PR size: small feature (<800 lines).
- Likely split into 1-2 PRs: config + env injection, then tests/docs.
- No downstream-only changes anticipated.

## User Stories
- As a user, I want my shell history to remain usable even if multiclaude runs for hours.
- As a user, I want automated agent commands not to be searchable in Atuin.
- As a power user, I want to opt in to keeping history for debugging.

## Requirements
### Functional
- Default behavior: history is disabled for multiclaude-managed shells.
- Applies to all agent types, including workspace if launched by multiclaude daemon.
- Users can opt out per-repo via config.
- Users can opt out globally via config.
- Log a single startup note in agent logs indicating history state.

### Non-functional
- Must not modify user shell config files (`.zshrc`, `.bashrc`).
- Must not break command execution or interactive features.
- Minimal performance overhead.

## UX / Config
### User Experience
- Agents should run normally without changing user workflows.
- If history is disabled, it should be transparent aside from a single log line.

### Config Surface
Add settings (exact location TBD):
- Global: `~/.multiclaude/config.json`
- Repo: `.multiclaude/config.json`

```
{
  "shell_history": {
    "enabled": false
  }
}
```
Notes:
- `enabled=false` means disable history for multiclaude shells (default).
- `enabled=true` restores normal history behavior.

## Technical Approach
### High-Level Design
- Set history-related environment variables before spawning agent shells.
- Ensure variables propagate to all child processes.

### Data / State
- No new persistent state beyond config.

### Control Flow
1. Read config (global then repo).
2. Determine history enabled/disabled.
3. Inject env vars at tmux window or process spawn.
4. Log history state once on agent startup.

### Integrations
- Shells (bash, zsh) and optional history tools like Atuin.

### Bash
- `HISTFILE=/dev/null`
- `HISTSIZE=0`
- `HISTFILESIZE=0`
- `set +o history` (if we can inject pre-shell command)

### Zsh
- `HISTFILE=/dev/null`
- `HISTSIZE=0`
- `SAVEHIST=0`
- `setopt NO_HIST_BEEP` optional (avoid noise)

### Atuin (Best-effort)
- `ATUIN_HISTORY=0`

## Agent Execution Notes
- Workers should verify `ROADMAP.md` and this PRD before changes.
- Ask supervisor if shell detection or config precedence is unclear.

## Edge Cases
- Users run a different shell: best-effort via `HISTFILE=/dev/null` and `HISTSIZE=0`.
- Users run interactive shells manually inside agent tmux window: still history-disabled unless opted out.
- Tools that write their own history files independently: out of scope unless they honor env vars.

## Security & Privacy
- Reduces risk of sensitive commands being saved in global history.

## Metrics / Success Criteria
- No new entries from multiclaude agents appear in bash/zsh history by default.
- Atuin history does not include multiclaude agent commands (best-effort).
- No reported regressions of command execution in agents.

## Testing
- Unit tests for config parsing defaults and overrides.
- Integration test:
  - Spawn agent shell with history disabled.
  - Run a command.
  - Verify no new entries in temp HISTFILE.
- Manual validation: run with Atuin installed and verify no entries.

## Rollout Plan
1. Implement environment overrides with default `enabled=false`.
2. Add config flags and tests.
3. Document in README/ARCHITECTURE or relevant config docs.
4. Release and observe feedback.

## Open Questions
- Where should config live and how should precedence work (global vs repo)?
- Should workspace agents inherit history-disabled by default?
- Should we add a CLI flag for temporary override?

## Work Items
- [ ] Add config parsing for `shell_history.enabled`.
- [ ] Inject history-related env vars when spawning agent shells.
- [ ] Add tests for config and env injection.
- [ ] Document behavior in repo docs.

