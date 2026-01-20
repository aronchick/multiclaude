# ADR: Activity Tracking Approach

**Status:** Decided
**Date:** 2026-01-20
**Author:** lively-elephant (worker agent)

## Context

The question was raised: should multiclaude write a static Markdown file to each tracked repository that gets continually updated to show what multiclaude has done in that repo?

This document analyzes the current approach and explains why adding an auto-updating file is not recommended.

## Current Tracking Mechanisms

Multiclaude already has several tracking mechanisms:

### 1. State File (`~/.multiclaude/state.json`)

Each repository maintains a `task_history` array with entries containing:
- Worker name
- Task description
- Git branch
- PR URL and number (if created)
- Status (pending, merged, closed, no-pr, unknown)
- Created and completed timestamps

### 2. History Command (`multiclaude history`)

```
$ multiclaude history
Task History for 'multiclaude' (last 10):

NAME             STATUS  PR    COMPLETED    TASK
-------------------------------------------------------------------------------
proud-otter      open    #159  22 mins ago  There are two PRs that were clos...
witty-bear       closed  #157  1 hour ago   Push an empty commit to PR #157 ...
proud-otter      merged  #154  3 hours ago  The history command isn't bad bu...
```

### 3. Agent Output Logs (`~/.multiclaude/output/<repo>/workers/`)

Detailed logs of agent activity, useful for debugging and review.

### 4. GitHub Pull Requests

The canonical record of all agent work. Each worker creates a PR for its changes, which:
- Shows exactly what was changed
- Has full discussion and review history
- Is visible to anyone with repo access
- Persists forever in GitHub

## Decision

**Do not implement an auto-updating Markdown file in repositories.**

## Rationale

### Arguments Against (decisive)

1. **Git commit noise**: Frequent commits to update a tracking file would pollute the commit history. Every time an agent completes work, the tracking file would need a commit, creating noise unrelated to actual code changes.

2. **Merge conflicts**: Multiple agents working simultaneously would frequently conflict when updating the same tracking file. This is especially problematic given multiclaude's "brownian ratchet" philosophy where chaos is expected.

3. **Duplication**: The information would duplicate what's already available via `multiclaude history` and GitHub's PR list.

4. **Violates "terminal-native" principle**: The ROADMAP explicitly states "Terminal is the interface, not files in the repo." A tracking file shifts the interface away from the terminal.

5. **Violates "simple" principle**: The ROADMAP states "Prefer deleting code over adding complexity." This feature adds complexity for minimal benefit.

6. **PRs ARE the record**: In the brownian ratchet philosophy, "CI is King" and PRs are the canonical record of agent work. Adding another tracking mechanism is redundant.

7. **Requires pushing to remote**: Writing to the repo means committing, which implies pushing - this could conflict with the local-first principle and create unexpected pushes.

### Arguments For (insufficient)

1. **Visibility**: Anyone viewing the repo could see multiclaude activity. However, they can already see this via the PR list.

2. **Persistence**: The file would survive if `~/.multiclaude` state is lost. However, PRs also survive and are the authoritative record.

## Alternatives Considered

### Export Command (Recommended for future)

Instead of auto-updating, a command like `multiclaude history --export` could generate a Markdown summary on-demand. This:
- Avoids commit noise (user decides when to commit)
- Avoids merge conflicts (not auto-updated)
- Provides the visibility benefit when needed

This is noted as a potential enhancement but not part of this decision.

### Enhanced History Command (Already on roadmap)

ROADMAP P1 includes "Task history: Track what workers have done and their outcomes" - the current `multiclaude history` command already addresses this.

## Consequences

- No new file tracking is implemented
- Existing `multiclaude history` command remains the primary way to view activity
- GitHub PRs remain the canonical record of agent work
- The codebase stays simple
