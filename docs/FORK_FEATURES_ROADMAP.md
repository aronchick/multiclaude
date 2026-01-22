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

### Status: ✅ Implemented (PR TBD)

### Purpose
Read-only web dashboard for local observability of multiclaude state.

### Use Cases
- Quick visual feedback without leaving browser
- Monitor agent status at a glance
- View repository and agent information
- Check system health

### Architecture

```
┌─────────────────────────────────────┐
│         multiclaude daemon          │
│                                     │
│  ┌──────────────┐  ┌─────────────┐ │
│  │ Unix Socket  │  │ HTTP Server │ │
│  │ (CLI IPC)    │  │ (Dashboard) │ │
│  └──────────────┘  └──────┬──────┘ │
│                            │        │
└────────────────────────────┼────────┘
                             │
                             ▼
                    http://127.0.0.1:8080
                             │
                             ▼
                       Web Browser
```

### Key Constraints
- **Read-only**: No write operations - pure observability
- **Localhost-only**: Binds to 127.0.0.1:8080 (not 0.0.0.0)
- **Terminal-native remains primary**: Dashboard supplements CLI
- **Zero new dependencies**: Uses Go stdlib only
- **Opt-in**: Disabled by default, enabled with `multiclaude dashboard`
- **Minimal**: ~500 lines total (server + HTML + tests)

### Implementation Details

**HTTP Server** (`internal/dashboard/server.go`)
- Integrated with daemon lifecycle
- Serves embedded HTML template
- Provides JSON API endpoints
- Auto-refresh every 5 seconds

**Dashboard UI** (`internal/dashboard/templates/index.html`)
- Single-page application
- Vanilla JavaScript (no frameworks)
- Dark theme matching GitHub design
- Shows: stats, repositories, agents, activity

**API Endpoints**
- `GET /` - Dashboard HTML
- `GET /api/status` - Overall system status
- `GET /api/repos` - List repositories
- `GET /api/repos/{repo}/agents` - List agents
- `GET /api/repos/{repo}/messages` - List messages
- `GET /api/repos/{repo}/history` - Task history
- `GET /api/repos/{repo}/activity` - Activity feed

### Usage

```bash
# Start dashboard
multiclaude dashboard

# Stop dashboard
multiclaude dashboard --stop

# Open in browser
open http://127.0.0.1:8080
```

### Files Created
- `internal/dashboard/server.go` - HTTP server implementation
- `internal/dashboard/templates/index.html` - Dashboard UI (embedded)
- `internal/dashboard/server_test.go` - Comprehensive tests
- `docs/DASHBOARD.md` - Complete documentation
- `docs/DASHBOARD_API.md` - API specification

### Related
- Upstream Issue: https://github.com/dlorenc/multiclaude/issues/169
- Documentation: [docs/DASHBOARD.md](DASHBOARD.md)
- API Spec: [docs/DASHBOARD_API.md](DASHBOARD_API.md)

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
3. ✅ Implement event system in daemon (PR #51)
4. ✅ Build web dashboard (PR TBD)
5. 📋 Create PRs for both fork-only features
6. 📋 Test both features in production
7. 📋 Consider multi-machine monitoring (future)

