# Extensibility (Current State)

Multiclaude can be extended today through two supported surfaces:

| Extension Point | Capabilities | Docs |
|-----------------|--------------|------|
| **State file** (`~/.multiclaude/state.json`) | Read-only monitoring and analytics | [`docs/extending/STATE_FILE_INTEGRATION.md`](extending/STATE_FILE_INTEGRATION.md) |
| **Socket API** (`~/.multiclaude/daemon.sock`) | Programmatic control via the daemon | [`docs/extending/SOCKET_API.md`](extending/SOCKET_API.md) |

The following are **planned, not implemented**: Event hooks, first-party web UI. They are explicitly marked as [PLANNED] to prevent hallucinations.

## Quick-start patterns
- Build dashboards or metrics exporters by polling the state file (no daemon interaction needed).
- Build automation or alternative CLIs by sending JSON requests to the socket API commands that exist in `internal/daemon/daemon.go`.

## Guardrails for contributors
- Never document an API/command/event that is not present in the code. If a feature is planned, label it [PLANNED].
- When modifying `internal/state/state.go` or `handleRequest` in `internal/daemon/daemon.go`, update the corresponding extension docs and rerun `go run ./cmd/verify-docs`.
- CI will run `verify-docs`; drift between code and docs will fail the build.
