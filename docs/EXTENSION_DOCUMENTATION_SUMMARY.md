# Extension Documentation Summary

This document provides a quick reference to multiclaude's extension points and their documentation.

## Extension Points

multiclaude provides two primary extension surfaces for external tools and integrations:

### 1. State File (Read-Only)

**File:** `~/.multiclaude/state.json`

**Use Cases:**
- Monitoring agent status
- Building dashboards and visualizations
- Metrics collection (Prometheus, Grafana)
- Status displays and notifications

**Documentation:** [`docs/extending/STATE_FILE_INTEGRATION.md`](extending/STATE_FILE_INTEGRATION.md)

**Key Features:**
- Atomic writes (never corrupt)
- No locking required
- Standard JSON format
- Safe for concurrent reads

### 2. Socket API (Read-Write)

**Socket:** `~/.multiclaude/daemon.sock`

**Use Cases:**
- Custom CLIs and automation scripts
- Programmatic control of agents
- Repository and agent management
- Triggering operations (cleanup, message routing)

**Documentation:** [`docs/extending/SOCKET_API.md`](extending/SOCKET_API.md)

**Key Features:**
- Full programmatic control
- JSON request/response protocol
- Command reference with examples
- Client libraries in Go, Python, Bash

## Keeping Extension Docs Updated

When modifying multiclaude internals that affect these extension surfaces:

1. **State Schema Changes** (`internal/state/state.go`):
   - Update [`STATE_FILE_INTEGRATION.md`](extending/STATE_FILE_INTEGRATION.md)
   - Update schema markers and examples
   - Run `go run ./cmd/verify-docs`

2. **Socket Command Changes** (`internal/daemon/daemon.go` - `handleRequest`):
   - Update [`SOCKET_API.md`](extending/SOCKET_API.md)
   - Add/update command reference entries
   - Run `go run ./cmd/verify-docs`

3. **CLI Changes** (`internal/cli/cli.go`):
   - Run `go generate ./pkg/config` to update generated docs
   - Verify extension doc references remain valid

## Verification

Always run verification before committing changes to extension surfaces:

```bash
go run ./cmd/verify-docs
```

This ensures:
- State schema documentation matches code
- Socket commands are documented
- File references are valid

## Out of Scope

Per ROADMAP.md, the following are **not** implemented:
- Web UIs (no reference implementation)
- Event hooks system
- Notification systems
- Plugin architecture

External tools should use the state file (read-only) or socket API (read-write) for all integrations.
