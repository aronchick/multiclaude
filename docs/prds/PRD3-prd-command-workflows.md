# PRD3: PRD Command Workflows

## Overview
Create a first-class workflow for generating PRDs via a command that matches existing multiclaude patterns. The goal is to let users rapidly capture an idea, generate a structured PRD, and then use it as a source for multi-agent work in multiclaude.

## Goals
- Provide a fast command path to create a new PRD from a brief prompt.
- Ensure PRDs follow the repo’s PRD template structure.
- Make the PRD discoverable and usable as input for spawning workers.

## Non-Goals
- Replacing manual PRD authoring for complex specs.
- Automatically implementing features without human review.
- Building a full UI; this is CLI-first.

## Background / Context
- Users have ideas that need quick capture into actionable PRDs.
- Agents need consistent, structured inputs to execute tasks correctly.

## Roadmap Alignment
- Aligns with the repo’s multi-agent workflow and PR discipline.
- Not listed as out-of-scope in `ROADMAP.md`.

## CI / Quality Gates
- No CI bypasses or relaxations.
- Tests added for new command parsing and PRD generation logic.

## Scope & PR Strategy
- Expected PR size: small feature (<800 lines).
- Likely split into 1-2 PRs: command handling, then templates/UX.

## User Stories
- As a user, I want to type a short command like `multiclaude prd create "idea"` and get a well-structured PRD draft.
- As a user, I want the PRD to land in `docs/prds/` with a sensible name.
- As an agent supervisor, I want PRDs to be consistent with the template so workers can execute reliably.

## Requirements
### Functional
- Provide a new CLI subcommand group consistent with existing patterns (e.g., `multiclaude prd create`, similar to `multiclaude agent <subcommand>`).
- Accept a short description prompt.
- Generate a PRD based on `docs/prds/PRD-TEMPLATE.md`.
- Name the file deterministically with a prefix and slug (e.g., `PRD4-<slug>.md`).
- Print the created file path and next steps.
- Support optional flags:
  - `--title` to override derived title
  - `--id` to set PRD number
  - `--output` to specify target folder
- Return non-zero exit on failure.

### Non-functional
- Does not overwrite existing files without `--force`.
- Uses ASCII-only output by default.
- Works in fork-only `docs/prds/` workflow.

## UX / Config
### User Experience
Example:
```
multiclaude prd create "Prevent agent shells from polluting history"
```
Output:
```
Created PRD: docs/prds/PRD4-shell-history-hygiene.md
Next: open the file and fill in Open Questions and Work Items.
```

### Config Surface
```
{
  "prd": {
    "default_dir": "docs/prds",
    "template_path": "docs/prds/PRD-TEMPLATE.md",
    "auto_increment": true
  }
}
```

## Technical Approach
### High-Level Design
- Add a new CLI command in `cmd/multiclaude` for PRD creation.
- Load the template file and inject title/overview scaffolding.
- Compute next PRD number by scanning `docs/prds/PRD*.md` unless `--id` provided.

### Data / State
- No new persistent state beyond the generated PRD files.

### Control Flow
1. Parse command and flags.
2. Resolve template path and output directory.
3. Compute PRD number and file name.
4. Render template with title and overview stub.
5. Write file (error if exists unless `--force`).

### Integrations
- File system access only.
- No network access required.

## Agent Execution Notes
- Workers should read `docs/prds/PRD-TEMPLATE.md` before modifying the command or template.
- Ask supervisor if naming or numbering policy is ambiguous.

## Edge Cases
- No `docs/prds` directory present: create it.
- Template file missing: error with guidance.
- Gaps in PRD numbering: choose next highest + 1.
- User supplies invalid characters in title: slugify safely.

## Security & Privacy
- No external data. No secrets.

## Metrics / Success Criteria
- PRD creation command succeeds in <1s.
- 0% file overwrite without explicit `--force`.
- PRDs are generated in the correct directory with consistent naming.

## Testing
- Unit tests for slug generation and PRD number detection.
- Integration test that creates a PRD in a temp directory.
- Ensure non-overwrite behavior.

## Rollout Plan
1. Implement command with minimal flags.
2. Add auto-increment and template resolution.
3. Add tests and update docs.

## Open Questions
- Should we allow multi-step interactive prompts?

## Work Items
- [ ] Add CLI command and flag parsing.
- [ ] Implement template loading and rendering.
- [ ] Add slug/ID generation logic.
- [ ] Add tests for PRD creation.
- [ ] Document the command in README or docs.
