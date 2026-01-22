# PRD5: Daemon REST API with Webhooks and Contracts

## Overview
Provide a RESTful interface for the multiclaude daemon so external tools can read state, trigger actions, and subscribe to webhook events. This enables integrations, dashboards, and automation layers to sit on top of multiclaude without direct process or filesystem access.

## Goals
- Expose a stable, documented REST API for daemon operations.
- Provide webhook events for key lifecycle and CI/PR states.
- Define clear contracts (request/response schemas and event payloads).
- Keep the API safe by default and opt-in for local use.

## Non-Goals
- Building a public cloud service.
- Replacing the CLI or tmux UX.
- Supporting unauthenticated remote access.

## Background / Context
- Integrations need a supported interface rather than scraping logs or files.
- Webhooks allow external systems to react to agent progress and CI state.
- Contracts reduce breaking changes and improve compatibility.

## Roadmap Alignment
- Aligns with extensibility and multi-tool workflows.
- Not listed as out-of-scope in `ROADMAP.md`.

## CI / Quality Gates
- No weakening or bypassing CI.
- Add tests for API endpoints and webhook payloads.

## Scope & PR Strategy
- Expected PR size: medium (<1500 lines).
- Likely split into 2-3 PRs: API skeleton, endpoint implementations, webhooks/contracts/tests.
- Downstream-only work: none expected.

## User Stories
- As a user, I want to query daemon status via HTTP.
- As a tool builder, I want to receive webhook events for agent lifecycle changes.
- As an integrator, I want stable payload contracts so I can build safely.

## Requirements
### Functional
- Expose an HTTP server in the daemon (opt-in via config).
- Provide REST endpoints for:
  - Health/status
  - Repos and agents list
  - Spawn/stop worker
  - List messages
  - Trigger refresh/status actions
- Provide webhook delivery for:
  - Agent lifecycle: spawned, completed, cleaned
  - PR events: created, updated, merged (when available)
  - CI events: checks pass/fail (when available)
- Version the API (e.g., `/api/v1`).
- Provide a contract file for schemas and events (JSON Schema or OpenAPI).
- Publish OpenAPI as the primary contract.
- Provide a JSON Schema endpoint for payloads (see UX / Config).

### Non-functional
- Default bind to localhost only.
- Require explicit enablement in config.
- Minimal latency impact on daemon loop.
- Backward compatibility within a major version.
- JSON Schema must be auto-generated from source and never manually edited.

## UX / Config
### User Experience
- Local tools can query `http://localhost:<port>/api/v1/status`.
- Webhook consumers receive JSON payloads with consistent schema.
- Contracts available at:
  - `GET /api/v1/openapi.json`
  - `GET /api/v1/schemas/<name>.json`

### Config Surface
```
{
  "daemon_api": {
    "enabled": false,
    "bind": "127.0.0.1",
    "port": 7878,
    "webhooks": [
      {
        "url": "http://localhost:3000/webhooks/multiclaude",
        "secret": "optional-shared-secret",
        "events": ["agent.spawned", "agent.completed"]
      }
    ]
  }
}
```

## Technical Approach
### High-Level Design
- Add an HTTP server to the daemon process.
- Use a router with explicit versioned routes.
- Add a webhook dispatcher with retries and backoff.
- Define OpenAPI for REST responses and webhook payloads.
- Generate JSON Schemas for payloads from OpenAPI during CI and embed them in the binary.
- Store generated schemas as CI artifacts for inspection.

### Data / State
- Reuse existing daemon state from `state.json`.
- No new persistent data beyond config.

### Control Flow
1. Load config and start HTTP server if enabled.
2. Handle requests by reading daemon state and invoking actions.
3. On relevant daemon events, emit webhook payloads asynchronously.
4. Record delivery attempts and basic metrics (in memory).

### Integrations
- Optional GitHub API for PR/CI event enrichment (best-effort).
- Webhook delivery via HTTP POST with signature header (if secret provided).

## Agent Execution Notes
- Workers should not call the API directly unless instructed.
- Ask supervisor before changing endpoint contracts.

## Edge Cases
- Webhook endpoint unavailable: retry with exponential backoff.
- Large payloads: enforce size limits.
- Conflicting actions: reject or serialize.

## Security & Privacy
- Localhost-only by default.
- Optional shared-secret signing for webhooks.
- Avoid leaking secrets or tokens in payloads.

## Metrics / Success Criteria
- API enabled with no daemon stability regressions.
- Webhook delivery success rate >95% for local endpoints.
- External tools can reliably read state and trigger actions.

## Testing
- Unit tests for route handlers and schema validation.
- Integration tests for webhook delivery (success/failure).
- E2E test for enabling API and querying status.
- CI check that fails if generated schemas are out of date.

## Rollout Plan
1. Add API config and skeleton server.
2. Implement core endpoints and schemas.
3. Add schema generation in CI and embed outputs.
4. Add webhook support and tests.
5. Document usage and contracts.

## Open Questions
- How should auth be handled beyond localhost?
- Which events are required vs optional?

## Work Items
- [ ] Add daemon API config and server lifecycle.
- [ ] Implement core REST endpoints.
- [ ] Add schema contracts and versioning.
- [ ] Implement webhook dispatcher with retries.
- [ ] Add tests and docs.
