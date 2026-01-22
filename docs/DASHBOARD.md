# Web Dashboard

## Overview

The MultiClaude web dashboard is a **read-only, localhost-only** web interface for monitoring the state of your MultiClaude daemon and agents. It provides real-time visibility into repositories, agents, tasks, and activity without leaving your browser.

**Key Point:** This is a **fork-only feature** that will not be contributed upstream. The upstream project explicitly rejects web interfaces per [ROADMAP.md](../ROADMAP.md#out-of-scope).

## Design Principles

The dashboard adheres to strict constraints to maintain the spirit of MultiClaude:

1. **Read-Only**: No write operations - pure observability only
2. **Localhost-Only**: Binds to `127.0.0.1:8080` (not `0.0.0.0`) for security
3. **Terminal-Native Remains Primary**: Dashboard supplements CLI, doesn't replace it
4. **Zero New Dependencies**: Uses Go stdlib only (no external web frameworks)
5. **Opt-In**: Disabled by default, enabled with `multiclaude dashboard`
6. **Minimal**: ~500 lines total (server + HTML + tests)
7. **Auto-Refresh**: Updates every 5 seconds automatically

## Usage

### Starting the Dashboard

```bash
# Start the dashboard HTTP server
multiclaude dashboard

# Output:
# ✓ Dashboard started successfully!
#
#   Open in browser: http://127.0.0.1:8080
#
#   The dashboard will auto-refresh every 5 seconds.
#   To stop the dashboard, run: multiclaude dashboard --stop
#
#   Note: This is a fork-only feature and will not be contributed upstream.
```

### Accessing the Dashboard

Open your browser to: **http://127.0.0.1:8080**

The dashboard will automatically refresh every 5 seconds to show the latest state.

### Stopping the Dashboard

```bash
# Stop the dashboard HTTP server
multiclaude dashboard --stop

# Output:
# Dashboard stopped successfully
```

### Checking Dashboard Status

```bash
# Check if dashboard is running
multiclaude status

# The status output includes daemon info and will show if dashboard is running
```

## What You'll See

The dashboard displays four main sections:

### 1. Statistics Overview
- **Total Agents**: Count of all agents across all repositories
- **Active Agents**: Agents currently working on tasks
- **Idle Agents**: Agents waiting for work
- **Repositories**: Number of tracked repositories

### 2. Repositories
Grid view of all tracked repositories showing:
- Repository name
- GitHub URL
- Number of agents

### 3. Agents
Grid view of all agents across all repositories showing:
- Agent name and type (supervisor, worker, review, merge-queue)
- Status badge (active, idle, stuck)
- Current task (if any)
- Repository association

### 4. Activity Feed
Chronological feed of recent activity showing:
- Completed tasks with status
- Inter-agent messages
- Timestamps and agent names
- Limited to 20 most recent items

## Architecture

### Components

1. **HTTP Server** (`internal/dashboard/server.go`)
   - Integrated with daemon lifecycle
   - Serves embedded HTML template
   - Provides JSON REST API endpoints
   - Graceful shutdown on daemon stop

2. **Dashboard UI** (`internal/dashboard/templates/index.html`)
   - Single-page application
   - Vanilla JavaScript (no frameworks)
   - Dark theme matching GitHub design
   - Client-side auto-refresh

3. **CLI Command** (`internal/cli/cli.go`)
   - `multiclaude dashboard` - Start dashboard
   - `multiclaude dashboard --stop` - Stop dashboard
   - Socket-based communication with daemon

4. **Daemon Integration** (`internal/daemon/daemon.go`)
   - `start_dashboard` command handler
   - `stop_dashboard` command handler
   - `dashboard_status` command handler
   - Lifecycle management

### Data Flow

```
Browser → HTTP GET /api/status
         ↓
Dashboard Server (port 8080)
         ↓
State Manager (in-memory state)
         ↓
JSON Response → Browser
```

All data is read from the daemon's in-memory state. No database or persistent storage is used.

## API Endpoints

The dashboard exposes a REST API for programmatic access. See [DASHBOARD_API.md](DASHBOARD_API.md) for complete API documentation.

**Quick Reference:**
- `GET /` - Dashboard HTML page
- `GET /api/status` - Overall system status
- `GET /api/repos` - List all repositories
- `GET /api/repos/{repo}/agents` - List agents for a repository
- `GET /api/repos/{repo}/messages` - List messages for a repository
- `GET /api/repos/{repo}/history` - Task history for a repository
- `GET /api/repos/{repo}/activity` - Activity feed for a repository

## Security Considerations

### Localhost-Only Binding

The dashboard binds to `127.0.0.1:8080`, **not** `0.0.0.0:8080`. This means:
- ✅ Accessible from the local machine only
- ❌ Not accessible from other machines on the network
- ✅ No firewall configuration needed
- ✅ No authentication needed (local-only access)

### Read-Only Operations

The dashboard provides **zero write operations**:
- Cannot create/delete agents
- Cannot modify repository configuration
- Cannot send messages between agents
- Cannot trigger any daemon actions

All write operations must be performed via the CLI.

### XSS Protection

The dashboard HTML template includes XSS protection:
- All user-provided content is escaped via `escapeHtml()`
- No `innerHTML` usage with untrusted data
- Content-Type headers properly set

## Why Fork-Only?

This feature will **not** be contributed upstream because:

1. **Upstream Policy**: The [ROADMAP.md](../ROADMAP.md) explicitly states:
   > **Out of Scope:**
   > 3. Web interfaces or dashboards
   >    - No REST APIs for external consumption
   >    - No browser-based UIs
   >    - Terminal is the interface

2. **Design Philosophy**: MultiClaude is intentionally terminal-native to keep the codebase simple and focused.

3. **Fork Strategy**: This fork adds optional observability features while respecting upstream's design decisions. See [FORK_MAINTENANCE_STRATEGY.md](FORK_MAINTENANCE_STRATEGY.md) for details.

## Testing

### Running Tests

```bash
# Run dashboard tests
go test ./internal/dashboard/...

# Run with coverage
go test -cover ./internal/dashboard/...

# Run integration tests
go test ./test/... -run Dashboard
```

### Manual Testing

1. Start the daemon: `multiclaude start`
2. Add a repository: `multiclaude add-repo test-repo https://github.com/user/repo`
3. Start the dashboard: `multiclaude dashboard`
4. Open browser to http://127.0.0.1:8080
5. Verify all sections display correctly
6. Add an agent and verify it appears in the dashboard
7. Stop the dashboard: `multiclaude dashboard --stop`

## Future Enhancements

Potential improvements (all maintaining read-only, localhost-only constraints):

1. **Enhanced Filtering**
   - Filter agents by status (active/idle/stuck)
   - Filter activity by type (tasks/messages)
   - Search functionality

2. **Detailed Views**
   - Click agent to see full details
   - Click repository to see detailed stats
   - Message thread visualization

3. **Performance Metrics**
   - Task completion times
   - Agent utilization graphs
   - Message delivery latency

4. **Export Functionality**
   - Export activity log as JSON/CSV
   - Download task history
   - Generate reports

5. **Customization**
   - Configurable refresh interval
   - Theme selection (light/dark)
   - Layout preferences

All enhancements must maintain the core principles: read-only, localhost-only, zero dependencies, minimal code.

## Related Documentation

- [DASHBOARD_API.md](DASHBOARD_API.md) - Complete REST API specification
- [FORK_FEATURES_ROADMAP.md](FORK_FEATURES_ROADMAP.md) - All fork-specific features
- [FORK_MAINTENANCE_STRATEGY.md](FORK_MAINTENANCE_STRATEGY.md) - How we maintain the fork
- [ROADMAP.md](../ROADMAP.md) - Upstream roadmap and out-of-scope items
- [SPEC.md](../SPEC.md) - MultiClaude architecture and design
- [AGENTS.md](../AGENTS.md) - Agent types and communication

## Troubleshooting

### Dashboard Won't Start

**Error:** "Dashboard is already running"
- **Solution:** Stop the existing instance: `multiclaude dashboard --stop`

**Error:** "Failed to start dashboard: address already in use"
- **Solution:** Another process is using port 8080. Stop it or change the port in `server.go`

### Dashboard Shows No Data

**Issue:** Dashboard loads but shows "No repositories" or "No agents"
- **Check:** Verify daemon is running: `multiclaude status`
- **Check:** Verify repositories are added: `multiclaude list-repos`
- **Check:** Check browser console for API errors

### Auto-Refresh Not Working

**Issue:** Dashboard doesn't update automatically
- **Check:** Browser console for JavaScript errors
- **Check:** Network tab shows API calls every 5 seconds
- **Solution:** Hard refresh the page (Ctrl+Shift+R or Cmd+Shift+R)

### Cannot Access from Another Machine

**This is by design.** The dashboard binds to `127.0.0.1` for security.

If you need remote access (not recommended):
1. Use SSH port forwarding: `ssh -L 8080:localhost:8080 user@machine`
2. Access via `http://localhost:8080` on your local machine

