# PRD Template

## Overview
Brief, non-technical description of the problem and why it matters.

## Goals
- Primary outcomes this PRD must achieve.
- Keep to measurable, user-meaningful results.

## Non-Goals
- Explicitly list what is out of scope.

## Background / Context
- Current behavior and pain points.
- Why now? Any incidents or regressions driving this.

## Roadmap Alignment
- Link to relevant items in `ROADMAP.md`.
- Explicitly confirm this is not out-of-scope.
- If this is a stretch goal, note priority and owner.

## CI / Quality Gates
- Required tests and checks that must pass.
- Any new tests or coverage expectations.
- Explicitly state: do not weaken or bypass CI.

## Scope & PR Strategy
- Expected PR size range (per `docs/UPSTREAM_WORKFLOW.md`).
- Number of PRs anticipated and how to slice work.
- Any "downstream-only" work that should not go upstream.

## User Stories
- As a <type>, I want <capability> so that <benefit>.
- Keep 3-6 focused stories.

## Requirements
### Functional
- Behaviors the system must support.
- Include default behaviors and failure modes.

### Non-functional
- Performance, reliability, security, compliance, compatibility.

## UX / Config
### User Experience
- UI/CLI behavior, error messages, and user flow impacts.

### Config Surface
```
{
  "feature": {
    "enabled": true,
    "options": []
  }
}
```
- Document default values and precedence (global vs repo).

## Technical Approach
### High-Level Design
- Architecture changes and where they live in the codebase.

### Data / State
- New fields, files, or migrations.

### Control Flow
- Key steps and decision points.

### Integrations
- External systems and APIs touched.

## Agent Execution Notes
- Worker task breakdown (one task = one PR).
- Points where agents must ask the supervisor for decisions.
- Required docs to read before changes (e.g., `ROADMAP.md`, `AGENTS.md`).

## Edge Cases
- Rare or failure scenarios and how they are handled.

## Security & Privacy
- Data handling, redaction, access, and auth changes.

## Metrics / Success Criteria
- How we know it worked.
- Quantitative targets if possible.

## Testing
- Unit tests to add or change.
- Integration / E2E scenarios.
- Manual validation steps.

## Rollout Plan
1. Implementation steps in order.
2. Feature flag or staged rollout if needed.
3. Monitoring and rollback plan.

## Open Questions
- Unknowns that need a decision or research.

## Work Items
- [ ] Concrete tasks another agent can execute.
- [ ] Include file paths or commands where possible.
