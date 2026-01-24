# PRD2: Forked Repo CI Reliability

## Overview
Forked repos are a core part of multiclaude workflows, but we still see frequent failures: forks drift from upstream, branches don’t push correctly, upstream CI fails after PRs are opened, and the core repo gets polluted with broken PRs. This PRD defines a reliability initiative to harden the forked-repo CI workflow, reduce drift, and prevent bad PRs from reaching upstream.

## Goals
- Keep forks in sync with upstream before any work begins.
- Prevent PR creation when fork CI is failing or fork is out of date.
- Ensure branches are created and pushed reliably.
- Reduce noise and broken PRs hitting upstream CI.

## Non-Goals
- Changing upstream CI configuration directly.
- Rewriting git workflows outside multiclaude’s control.
- Eliminating all CI failures (only those caused by fork drift or bad automation).

## Background / Context
- Forks drift when upstream moves quickly and local automation lags.
- Broken PRs waste CI cycles and increase maintainer overhead.
- We need a strict gate to keep the core repo clean.

## Roadmap Alignment
- Aligns with CI-first and upstream contribution workflows.
- Not listed as out-of-scope in `ROADMAP.md`.

## CI / Quality Gates
- No weakening or bypassing CI.
- Add tests for sync strategy, remote detection, and gating logic.

## Scope & PR Strategy
- Expected PR size: small feature (<800 lines) or medium (<1500 lines).
- Likely split into 2 PRs: preflight sync + branch push validation, then CI gate/reporting.
- No downstream-only changes anticipated.

## User Stories
- As a user, I want multiclaude to keep my fork in sync so PRs don’t fail due to stale base.
- As a maintainer, I want only clean, passing PRs to reach upstream.
- As a user, I want clear diagnostics when fork sync or CI checks block a PR.

## Requirements
### Functional
- Preflight sync check runs before any worker starts.
- Fork base is fast-forwarded (or rebased) onto upstream default branch before creating a branch.
- Branch creation and push are verified; failure blocks PR creation.
- PR creation is blocked if fork CI fails or is missing required checks.
- When upstream CI fails due to known sync issues, the system should auto-resync and retry.
- Clear status reporting to user and to merge-queue.

### Non-functional
- No extra credentials stored beyond existing GitHub auth.
- Minimal added latency for “happy path” runs.
- Safe defaults: block PR creation rather than create broken PRs.

## UX / Config
### User Experience
- On failure, show a clear status summary with next steps.
- On success, behavior remains unchanged except for improved reliability.

### Config Surface
```
{
  "forks": {
    "sync_before_work": true,
    "block_pr_on_ci_failure": true,
    "required_checks": ["build", "test"],
    "sync_strategy": "rebase"
  }
}
```

## Technical Approach
### High-Level Design
- Add a preflight sync step before worker spawn.
- Add a fork CI gate before PR creation.
- Add branch push verification with retries.

### Data / State
- No new persistent state beyond existing config.
- Cache last sync timestamp in memory (optional).

### Control Flow
1. Detect upstream remote and default branch.
2. Fetch upstream default branch.
3. Sync fork default branch (rebase/merge).
4. Push fork default branch to origin.
5. Create worker branch and verify remote tracking.
6. Run fork CI gate before PR creation.

### Integrations
- GitHub CLI for checks and PR creation.
- Git remotes for upstream/origin.

## Agent Execution Notes
- Workers should block if preflight sync fails and notify supervisor.
- Merge-queue should not merge PRs that bypass fork CI gate.

## Edge Cases
- Fork missing upstream remote: add or repair automatically.
- Fork with protected branches: use a safe sync strategy or warn.
- Diverged fork with conflicts: block and request human intervention.
- Repo uses non-`main` default branch: detect dynamically.

## Security & Privacy
- No new tokens or scopes beyond existing GitHub CLI auth.
- Avoid leaking fork-specific info into upstream logs.

## Metrics / Success Criteria
- 90%+ reduction in upstream CI failures attributable to fork drift.
- 95%+ of PRs created only after fork checks pass.
- Fewer manual interventions from merge-queue for fork-related issues.

## Testing
- Unit tests for sync strategy and remote detection.
- Integration tests:
  - Fork out-of-date, ensure sync happens.
  - CI failure on fork blocks PR creation.
  - Missing upstream remote triggers repair.
- E2E test in `MULTICLAUDE_TEST_MODE=1` with mocked GitHub API.

## Rollout Plan
1. Implement fork sync preflight and branch push verification.
2. Add fork CI gate before PR creation.
3. Add reporting and config toggles.
4. Monitor metrics and adjust defaults.

## Open Questions
- Should merge-queue enforce fork CI gate or should workers block earlier?
- How to handle repos with required checks that only run in upstream?
- Should auto-sync happen on a schedule as well as preflight?

## Work Items
- [ ] Add preflight fork sync step.
- [ ] Add branch push verification with retries.
- [ ] Implement fork CI gate before PR creation.
- [ ] Add reporting output for sync/CI status.
- [ ] Add tests and docs.

