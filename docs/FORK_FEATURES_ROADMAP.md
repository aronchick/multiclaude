# Fork-Only Features Roadmap

This document outlines features specific to the `aronchick/multiclaude` fork that will **not** be contributed upstream due to upstream's explicit scope constraints.

## Why Fork-Only?

Upstream (`dlorenc/multiclaude`) explicitly rejects these in ROADMAP.md:
- ❌ Web interfaces or dashboards
- ❌ External integrations (Slack, Discord, etc.)
- ❌ Remote/hybrid deployment
- ❌ Multi-machine coordination

**Our fork adds these features while maintaining compatibility with upstream.**

---

## 1. Event Hooks System

### Status: ✅ Implemented (PR #51)

### Purpose
Hook-based notification system that enables integration with external systems like Slack, Discord, email, or custom monitoring tools.

### Use Cases
- Alert when CI fails on a PR
- Notify when workers get stuck
- Report when agents start/stop
- Escalate when merge queue enters emergency mode
- Custom integrations via user-provided scripts

### Architecture

```
┌─────────────┐
│   Daemon    │
│             │
│  ┌───────┐  │     ┌──────────────┐
│  │ Event │──┼────▶│ User Hook    │──▶ Slack/Discord/Email
│  │ Bus   │  │     │ Script       │
│  └───────┘  │     └──────────────┘
│             │
└─────────────┘
```

### Implementation Details

**Event Types** (9 total)
- `agent_started` - When an agent starts
- `agent_stopped` - When an agent stops
- `agent_idle` - When an agent is idle
- `pr_created` - When a PR is created
- `pr_merged` - When a PR is merged
- `task_assigned` - When a task is assigned
- `ci_failed` - When CI fails
- `worker_stuck` - When a worker is stuck
- `message_sent` - When an inter-agent message is sent

**Hook Configuration**
- Global hooks stored in `~/.multiclaude/state.json`
- Per-event hooks: `on_event`, `on_pr_created`, `on_agent_idle`, etc.
- Fire-and-forget execution with 30s timeout
- Zero dependencies: hooks are user-provided scripts

**Example: Slack Integration**
See `examples/hooks/slack-notify.sh` for complete implementation

### Files Created
- `internal/events/events.go` - Event types and event bus
- `internal/events/events_test.go` - Comprehensive tests
- `examples/hooks/slack-notify.sh` - Slack integration example
- `examples/hooks/README.md` - Hook documentation

### Related
- PR: #51
- Upstream Issue: https://github.com/dlorenc/multiclaude/issues/170

---

## 2. Web Dashboard

### Status: 📋 Planned

### Purpose
View status of all multiclaude instances across multiple machines in a web browser.

### Use Cases
- Monitor multiple repos from one place
- View agent activity across machines
- Check CI status at a glance
- Historical view of completed tasks

### Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Machine 1   │     │  Machine 2   │     │  Machine 3   │
│              │     │              │     │              │
│ ~/.multiclaude│    │ ~/.multiclaude│    │ ~/.multiclaude│
│   state.json │     │   state.json │     │   state.json │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │                    │                    │
       └────────────────────┼────────────────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │ multiclaude-web │
                   │  (read-only)    │
                   └────────┬────────┘
                            │
                            ▼
                      Web Browser
```

### Key Constraints
- **Read-only**: No control plane, only viewing
- **Optional**: Core multiclaude works without it
- **Separate binary**: `multiclaude-web` command
- **Local-first**: Can run on localhost or LAN

### Implementation Plan

**Phase 1: State Reader** (P0)
- Package: `internal/dashboard/reader.go`
- Read state.json from multiple paths
- Aggregate data from multiple machines
- Watch for file changes (live updates)

**Phase 2: Web Server** (P0)
- Binary: `cmd/multiclaude-web/main.go`
- Simple HTTP server (port 8080 default)
- REST API for state data
- Static file serving

**Phase 3: Frontend** (P1)
- Directory: `web/`
- Single-page app (vanilla JS or minimal framework)
- Real-time updates via SSE or WebSocket
- Responsive design for mobile

**Phase 4: Multi-Machine** (P2)
- SSH-based state collection
- Configuration for remote machines
- Aggregated view across infrastructure

### Configuration Example

```json
{
  "integrations": {
    "web_dashboard": {
      "enabled": true,
      "port": 8080,
      "bind": "127.0.0.1",
      "machines": [
        {
          "name": "local",
          "state_path": "~/.multiclaude/state.json"
        },
        {
          "name": "dev-server",
          "state_path": "ssh://dev.example.com/home/user/.multiclaude/state.json"
        }
      ],
      "refresh_interval": "5s"
    }
  }
}
```

### Usage

```bash
# Start web dashboard
multiclaude-web start

# Custom port
multiclaude-web start --port 3000

# Add remote machine
multiclaude-web add-machine dev-server ssh://dev.example.com/home/user/.multiclaude/state.json

# Open in browser
multiclaude-web open
```

### Files to Create
- `cmd/multiclaude-web/main.go` - Web server binary
- `internal/dashboard/reader.go` - State aggregation
- `internal/dashboard/server.go` - HTTP server
- `internal/dashboard/api.go` - REST API handlers
- `web/index.html` - Dashboard UI
- `web/app.js` - Frontend logic
- `web/styles.css` - Styling
- `docs/WEB_DASHBOARD.md` - Setup guide

---

## 3. Multi-Machine Monitoring

### Status: 📋 Planned (depends on Web Dashboard)

### Purpose
Centralized monitoring of multiclaude across development machines, CI servers, and production environments.

### Use Cases
- DevOps team monitoring multiple environments
- Track agent activity across infrastructure
- Aggregate metrics and logs
- Alert on anomalies

### Implementation
This builds on the Web Dashboard with:
- SSH-based state collection
- Centralized logging aggregation
- Metrics collection and visualization
- Alert rules for anomalies

---

## Implementation Priority

### P0 (Must Have)
1. Event system in daemon
2. Slack basic integration
3. Web dashboard state reader
4. Web dashboard basic UI

### P1 (Should Have)
1. Slack notification rules
2. Web dashboard live updates
3. Multi-machine configuration

### P2 (Nice to Have)
1. Slack quiet hours
2. Web dashboard historical view
3. Metrics and analytics
4. Alert rules

---

## Maintenance Strategy

### Keeping Fork-Only Code Isolated

1. **Separate packages**: All fork features in `internal/integrations/`
2. **Feature flags**: Disabled by default, opt-in via config
3. **No core changes**: Don't modify core daemon/agent logic
4. **Event-driven**: Use event bus to decouple from core

### Testing Fork Features

```bash
# Test only fork features
go test ./internal/integrations/... ./cmd/multiclaude-web/...

# Test core + fork
make check-all
```

### Documentation

- Each feature gets its own doc: `docs/SLACK_INTEGRATION.md`, `docs/WEB_DASHBOARD.md`
- Update `README_FORK.md` with fork-specific features
- Keep upstream docs clean (no fork features)

---

## Next Steps

1. ✅ Merge pending PRs to fork
2. ✅ Create fork maintenance strategy doc
3. 📋 Implement event system in daemon
4. 📋 Build Slack integration
5. 📋 Build web dashboard
6. 📋 Document fork features

