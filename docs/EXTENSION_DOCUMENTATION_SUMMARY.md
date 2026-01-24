# Extension Documentation Summary

This repo documents only the extension points that are implemented today. Anything marked [PLANNED] is intentionally not available to avoid hallucinated APIs.

## Files
- [`docs/EXTENSIBILITY.md`](EXTENSIBILITY.md): Overview and guardrails.
- [`docs/extending/STATE_FILE_INTEGRATION.md`](extending/STATE_FILE_INTEGRATION.md): Read-only state file schema.
- [`docs/extending/SOCKET_API.md`](extending/SOCKET_API.md): Live socket command surface (`handleRequest`).
- [`docs/extending/EVENT_HOOKS.md`](extending/EVENT_HOOKS.md): [PLANNED] placeholder until event emitters exist.
- [`docs/extending/WEB_UI_DEVELOPMENT.md`](extending/WEB_UI_DEVELOPMENT.md): [PLANNED] note for downstream dashboards.

## Verification
- Run `go run ./cmd/verify-docs` after changing the state schema or socket commands. CI will block doc/code drift.
- The verification tool extracts data from `internal/state/state.go` and `internal/daemon/daemon.go` via AST; the docs must list the same fields/commands.

## Downstream guidance
- Build monitors by reading `~/.multiclaude/state.json`.
- Build control-plane tools by sending JSON requests to the documented socket commands.
- Do not rely on events or web UI until implemented upstream.
