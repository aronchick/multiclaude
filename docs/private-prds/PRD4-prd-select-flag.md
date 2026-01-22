# PRD4: PRD Selection Flag for `multiclaude work`

## Overview
Add a new flag to `multiclaude work` that lets users select an existing PRD as the task source. The flag should support choosing from PRDs that are not yet in progress and allow fuzzy selection by filename or title in the current repo.

## Goals
- Make it easy to start work from an existing PRD without retyping the task.
- Support fuzzy selection of PRDs by name or title.
- Prevent multiple workers from unknowingly starting the same PRD.

## Non-Goals
- Full PRD lifecycle management UI.
- Automatic PRD creation (covered by PRD3).
- Replacing freeform `multiclaude work "task"` usage.

## Background / Context
- We now have local PRDs in `docs/prds/` that should drive agent work.
- Users want to quickly pick a PRD from a list instead of copying text.

## Roadmap Alignment
- Aligns with structured agent workflows and consistent execution inputs.
- Not listed as out-of-scope in `ROADMAP.md`.

## CI / Quality Gates
- No weakening or bypassing CI.
- Add tests for PRD discovery and fuzzy matching.

## Scope & PR Strategy
- Expected PR size: small feature (<800 lines).
- Likely split into 1-2 PRs: flag parsing + PRD selection logic, then tests/docs.

## User Stories
- As a user, I want `multiclaude work --prd <name>` to start a worker based on a PRD.
- As a user, I want fuzzy matching so partial names work.
- As a user, I want to avoid duplicating work on a PRD already in progress.

## Requirements
### Functional
- Add a new flag to `multiclaude work`, e.g.:
  - `--prd <query>`: select a PRD by fuzzy match of filename or title.
  - `--prd-file <path>`: use a specific PRD file path in the repo.
- If multiple matches are found, prompt the user to choose.
- If no match is found, return a helpful error.
- Mark selected PRDs as “in progress” to prevent duplicate work.
- Allow override to force selection even if in progress.

### Non-functional
- No network access required.
- ASCII-only output by default.
- Does not modify upstream configuration.

## UX / Config
### User Experience
Examples:
```
multiclaude work --prd "shell history"
multiclaude work --prd-file docs/prds/PRD2-forked-ci-reliability.md
```
Output:
```
Selected PRD: PRD2: Forked Repo CI Reliability
Spawning worker: clever-fox
```

### Config Surface
```
{
  "prd": {
    "default_dir": "docs/prds",
    "in_progress_tag": "status:in-progress"
  }
}
```

## Technical Approach
### High-Level Design
- Extend `multiclaude work` command with PRD selection flags.
- Parse PRD files in `docs/prds/` and read title from first heading.
- Use fuzzy matching on filename and title.
- Track PRD “in progress” state via lightweight markers.

### Data / State
Options for in-progress tracking:
- Add a small front-matter block or tag line in PRD files.
- Or track in `.multiclaude/state.json` keyed by PRD file path.

### Control Flow
1. Parse flags and determine selection mode.
2. Discover PRD files in `docs/prds/`.
3. Fuzzy match by filename/title.
4. Confirm selection if multiple matches.
5. Mark PRD as in-progress.
6. Spawn worker using PRD content as task input.

### Integrations
- File system for PRD discovery.
- Existing worker spawn and messaging flow.

## Agent Execution Notes
- Workers should include PRD title in PR description.
- Ask supervisor if multiple PRDs overlap in scope.

## Edge Cases
- PRD files missing or empty.
- Ambiguous fuzzy matches.
- PRDs already in progress.
- PRD title differs from filename.

## Security & Privacy
- No external data or tokens required.
- Only local repo files are read.

## Metrics / Success Criteria
- >80% of workers spawned from PRDs use the selection flag.
- Reduced duplicated work on the same PRD.
- Low rate of user confusion/errors in PRD selection.

## Testing
- Unit tests for PRD discovery and fuzzy match logic.
- Integration test for `multiclaude work --prd`.
- Test behavior when PRD is already in progress.

## Rollout Plan
1. Implement PRD discovery and selection flags.
2. Add in-progress tracking.
3. Add tests and docs.

## Open Questions
- Should in-progress tracking live in PRD files or state.json?
- Should `multiclaude work` default to PRD selection when no task is provided?

## Work Items
- [ ] Add `--prd` and `--prd-file` flags to `multiclaude work`.
- [ ] Implement PRD discovery and fuzzy matching.
- [ ] Implement in-progress tracking and overrides.
- [ ] Add tests and update docs.

