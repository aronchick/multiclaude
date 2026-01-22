# Proposal: Optional Web Dashboard for Multi-Agent Monitoring

**Status:** RFC (Request for Comments)
**Created:** 2026-01-22
**Author:** Fork maintainer
**Target:** Upstream contribution discussion

---

## Executive Summary

This proposal argues for relaxing the "no web interfaces" restriction to allow an **optional, localhost-only web dashboard** for monitoring multi-agent orchestration. The dashboard would be:

- **Optional** (disabled by default, requires `--enable-dashboard` flag)
- **Local-only** (localhost:PORT, not exposed externally)
- **Read-only** (initially - just visualization, no control plane)
- **Supplementary** (terminal remains primary interface)
- **Minimal** (stdlib only, no framework dependencies)

**Core Philosophy Alignment:**
- ✅ Terminal remains the primary interface
- ✅ No public REST APIs for external consumption
- ✅ No complex web frameworks or dependencies
- ✅ Optional feature - doesn't affect users who prefer terminal-only
- ✅ Maintains daemon architecture - web is just another view

---

## The Problem: Multi-Agent Complexity is Inherently Visual

### Current Terminal-Only Limitations

**1. Monitoring Multiple Agents is Difficult**

With 5+ agents running simultaneously:
```bash
# Terminal approach requires multiple commands
$ multiclaude status                    # High-level overview
$ multiclaude agent list                # See all agents
$ multiclaude agent show supervisor     # Detailed view of one agent
$ multiclaude agent list-messages       # Check message queue
$ tmux attach -t mc-repo                # Attach to see actual work
```

**Problem:** You can only see one thing at a time. No holistic view.

**2. Understanding Message Flow is Hard**

```
supervisor → worker-1: "Start on issue #123"
worker-1 → supervisor: "Need clarification on scope"
supervisor → worker-2: "Review PR #456"
merge-queue → supervisor: "PR #456 ready"
worker-2 → merge-queue: "Tests passing"
```

**Terminal view:**
```bash
$ multiclaude agent list-messages
- Message 1/5: From supervisor to worker-1...
- Message 2/5: From worker-1 to supervisor...
- Message 3/5: From supervisor to worker-2...
# Can't see the flow, just sequential list
```

**Problem:** Message routing is spatial and temporal. Text lists don't capture this.

**3. Debugging Race Conditions and Timing Issues**

When an agent crashes or gets stuck:
```bash
$ multiclaude status
# Shows agent is "running" but is it actually working?

$ tmux attach -t mc-repo:worker-1
# See terminal output, but when did it last update?

$ multiclaude agent show worker-1
# Shows created_at timestamp, but no activity timeline
```

**Problem:** No visibility into agent lifecycle events, health trends, or activity patterns over time.

**4. Stakeholder Communication**

You're running 10 autonomous agents working on a large codebase:

**Manager asks:** "How's the multi-agent job going?"

**Current answer:**
```bash
$ multiclaude status
# Copy/paste terminal output into Slack
# Or: "Let me SSH in and check..."
```

**Problem:** Non-technical stakeholders can't easily monitor progress. You become a bottleneck.

---

## The Solution: Optional Localhost Web Dashboard

### What It Looks Like

**Dashboard Home (Localhost:8080):**

```
┌────────────────────────────────────────────────────────────────┐
│  multiclaude Dashboard                    🟢 Daemon Running    │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Repository: my-project     Agents: 7 running, 2 completed     │
│                                                                 │
│  ┌──────────────┬──────────┬──────────────┬─────────────────┐ │
│  │ Agent        │ Status   │ Last Active  │ Messages        │ │
│  ├──────────────┼──────────┼──────────────┼─────────────────┤ │
│  │ supervisor   │ 🟢 Running│ 2s ago      │ 3 pending       │ │
│  │ merge-queue  │ 🟢 Running│ 5s ago      │ 1 pending       │ │
│  │ worker-1     │ 🟢 Running│ 1s ago      │ 0 pending       │ │
│  │ worker-2     │ 🟡 Idle   │ 45s ago     │ 0 pending       │ │
│  │ worker-3     │ ✅ Done   │ 2m ago      │ -               │ │
│  └──────────────┴──────────┴──────────────┴─────────────────┘ │
│                                                                 │
│  Message Flow (Last 10 minutes):                               │
│                                                                 │
│    supervisor ──→ worker-1 ──→ merge-queue                    │
│         ↓                           ↓                          │
│    worker-2 ────────────────────→ supervisor                   │
│                                                                 │
│  System Health:                                                │
│  ├─ CPU: 23% (across all agents)                              │
│  ├─ Memory: 450MB / 8GB                                        │
│  └─ Messages/min: 12                                           │
│                                                                 │
│  [Refresh: Auto (5s)] [View Logs] [Message History]           │
└────────────────────────────────────────────────────────────────┘
```

**Benefits:**
- 👀 **At-a-glance status** - See all agents without multiple commands
- 🔄 **Real-time updates** - WebSocket push instead of polling CLI
- 📊 **Visual message flow** - Understand agent communication patterns
- 📈 **Health metrics** - CPU, memory, message rates over time
- 🐛 **Debugging context** - Timeline of events, not just current state
- 🔗 **Shareable** - Send localhost URL via SSH tunnel for remote monitoring

---

## Addressing the "No Web Interface" Restriction

### Original Restriction (from project philosophy):

> **3. Web interfaces or dashboards**
> - No REST APIs for external consumption
> - No browser-based UIs
> - Terminal is the interface

### Why This Restriction Exists (Assumed Intent):

1. **Simplicity** - Avoid web framework complexity
2. **Unix Philosophy** - Terminal tools compose better
3. **Security** - No exposed HTTP endpoints
4. **Dependencies** - No Node.js, React, build toolchains
5. **Scope Creep** - "Just a dashboard" becomes a whole app

### How This Proposal Respects Those Concerns:

| Concern | Our Approach |
|---------|--------------|
| **Complexity** | Go stdlib only (`net/http`, `html/template`). No frameworks. |
| **Unix Philosophy** | Terminal remains primary. Dashboard is **optional visualization**. |
| **Security** | Localhost-only binding. No external exposure. Auth via daemon socket ownership. |
| **Dependencies** | Zero JavaScript frameworks. Vanilla JS + SSE/WebSocket. No build step. |
| **Scope Creep** | **Read-only** dashboard. No control plane (use CLI for commands). |

### Key Distinction: Visualization ≠ Control

**The dashboard is NOT:**
- ❌ A replacement for the CLI
- ❌ A REST API for external integration
- ❌ A public-facing web service
- ❌ A control plane (no "Start Agent" buttons)

**The dashboard IS:**
- ✅ A **real-time visualization** of daemon state
- ✅ An **optional debugging tool** for complex multi-agent scenarios
- ✅ A **localhost-only monitor** (127.0.0.1:8080)
- ✅ A **read-only view** (use `multiclaude` CLI for actions)

---

## Use Cases: When Web Dashboard Shines

### 1. **Large-Scale Orchestration (10+ Agents)**

**Scenario:** Running 15 workers on different issues simultaneously.

**Terminal Approach:**
```bash
# Check status
$ multiclaude status
# Output scrolls off screen with 15 agents

# Check specific agent
$ multiclaude agent show worker-7
# Lost context of what other 14 agents are doing

# Check messages
$ multiclaude agent list-messages
# 30 messages in queue, hard to see patterns
```

**Dashboard Approach:**
- Single screen showing all 15 agents
- Color-coded status (green=active, yellow=idle, red=error)
- Message flow visualization shows bottlenecks
- Click agent name → see details without losing context

### 2. **Debugging Stuck Workflows**

**Scenario:** Worker-3 hasn't responded in 5 minutes. Why?

**Terminal Approach:**
```bash
$ multiclaude agent show worker-3
Created: 2026-01-22 10:30:00
Status: running
# No activity timeline

$ tmux attach -t mc-repo:worker-3
# See terminal, but is Claude thinking or frozen?
```

**Dashboard Approach:**
- Timeline view shows worker-3's last activity was 5m ago
- Message history shows last message: supervisor → worker-3 at 10:35
- No response sent, indicates stuck agent
- Grafana-style activity chart shows sudden drop
- Quick diagnosis: Agent is frozen, needs restart

### 3. **Remote Monitoring (SSH Tunnel)**

**Scenario:** Running on remote server, want to monitor from laptop.

**Terminal Approach:**
```bash
# SSH in and run commands
$ ssh server.com
$ multiclaude status
# Or: set up complex tmux over SSH
```

**Dashboard Approach:**
```bash
# From laptop, create SSH tunnel
$ ssh -L 8080:localhost:8080 server.com

# Open browser to http://localhost:8080
# See live dashboard of remote agents
```

### 4. **Demo/Presentation Mode**

**Scenario:** Showing multiclaude to team, investors, or conference audience.

**Terminal Approach:**
- Screen share terminal
- Run commands manually
- Output scrolls, hard to follow
- Not visually engaging

**Dashboard Approach:**
- Screen share browser at localhost:8080
- Real-time updates without typing commands
- Visual message flow is easier to understand
- Professional appearance for stakeholders

### 5. **Long-Running Jobs**

**Scenario:** 24-hour agent orchestration, periodic check-ins.

**Terminal Approach:**
```bash
# Check in hourly
$ multiclaude status
# Snapshot only, no historical context
```

**Dashboard Approach:**
- Leave browser tab open
- Glance at dashboard periodically
- See activity graph over last 24h
- Notice patterns (e.g., "workers idle during night hours")
- Historical view helps optimize agent scheduling

---

## Technical Implementation

### Architecture: Minimal and Optional

```
┌──────────────────────────────────────────────────────────┐
│                      Daemon Process                       │
│                                                           │
│  ┌─────────────────┐         ┌────────────────────────┐ │
│  │  Core Daemon    │         │  HTTP Server (opt)     │ │
│  │  (existing)     │◄────────│  - Localhost only      │ │
│  │                 │  Read   │  - SSE for updates     │ │
│  │  - State mgmt   │  State  │  - Static HTML+JS      │ │
│  │  - Agent loops  │         │  - Go stdlib only      │ │
│  │  - Unix socket  │         │  - Port: 8080 (config) │ │
│  └─────────────────┘         └────────────────────────┘ │
│                                        ▲                  │
└────────────────────────────────────────┼──────────────────┘
                                         │ HTTP (localhost)
                                         │
                                    ┌────┴─────┐
                                    │ Browser  │
                                    │ :8080    │
                                    └──────────┘
```

### Implementation Details

**1. Daemon Changes (Minimal)**

```go
// internal/daemon/dashboard.go (NEW FILE)
package daemon

import (
    "html/template"
    "net/http"
)

type Dashboard struct {
    daemon *Daemon
    server *http.Server
}

func NewDashboard(d *Daemon, port int) *Dashboard {
    return &Dashboard{
        daemon: d,
        server: &http.Server{
            Addr:    fmt.Sprintf("127.0.0.1:%d", port),
            Handler: http.NewServeMux(),
        },
    }
}

func (db *Dashboard) Start() error {
    mux := http.NewServeMux()

    // Serve static HTML
    mux.HandleFunc("/", db.handleIndex)

    // SSE endpoint for real-time updates
    mux.HandleFunc("/events", db.handleEvents)

    // JSON API for current state
    mux.HandleFunc("/api/status", db.handleStatus)

    db.server.Handler = mux
    return db.server.ListenAndServe()
}
```

**2. CLI Flag (Optional Enable)**

```go
// internal/cli/cli.go
func (c *CLI) runDaemon(args []string) error {
    // Existing flag parsing...

    enableDashboard := false
    dashboardPort := 8080

    // Parse flags
    if flagSet.Lookup("enable-dashboard") != nil {
        enableDashboard = flagSet.Lookup("enable-dashboard").Value.(bool)
    }

    // Start daemon (existing code)
    d, err := daemon.New(c.paths, GetVersion())
    if err != nil {
        return err
    }

    // Optional: Start dashboard
    if enableDashboard {
        dashboard := daemon.NewDashboard(d, dashboardPort)
        go dashboard.Start()
        fmt.Printf("Dashboard available at http://localhost:%d\n", dashboardPort)
    }

    // Continue with daemon.Run()...
}
```

**3. Frontend (Zero Build Step)**

```html
<!-- internal/dashboard/static/index.html (embedded) -->
<!DOCTYPE html>
<html>
<head>
    <title>multiclaude Dashboard</title>
    <style>
        /* Simple CSS, no frameworks */
        body { font-family: monospace; }
        .agent { padding: 10px; margin: 5px; border: 1px solid #ccc; }
        .running { background: #d4edda; }
        .idle { background: #fff3cd; }
    </style>
</head>
<body>
    <h1>multiclaude Dashboard</h1>
    <div id="agents"></div>

    <script>
        // Vanilla JS, no frameworks
        const evtSource = new EventSource('/events');

        evtSource.addEventListener('status', (e) => {
            const data = JSON.parse(e.data);
            updateAgents(data.agents);
        });

        function updateAgents(agents) {
            const container = document.getElementById('agents');
            container.innerHTML = agents.map(a => `
                <div class="agent ${a.status}">
                    <strong>${a.name}</strong> - ${a.status}
                </div>
            `).join('');
        }
    </script>
</body>
</html>
```

**Key Implementation Constraints:**

- ✅ **No external dependencies** - `net/http` only (Go stdlib)
- ✅ **No build step** - HTML/JS served directly, no webpack/vite
- ✅ **Embedded assets** - Use `//go:embed` to bundle HTML into binary
- ✅ **Localhost-only** - Bind to `127.0.0.1`, never `0.0.0.0`
- ✅ **Read-only** - No POST endpoints, only GET and SSE
- ✅ **< 500 LOC** - Keep it minimal (dashboard.go + template)

---

## Security Considerations

### Threat Model

**Threat:** Unauthorized access to dashboard reveals sensitive repo information.

**Mitigation:**

1. **Localhost-only binding** - Not accessible from network
   ```go
   server.Addr = "127.0.0.1:8080" // NOT "0.0.0.0:8080"
   ```

2. **Same-machine security** - If attacker has localhost access, they already have:
   - Access to `multiclaude` CLI (same permissions)
   - Access to state.json file
   - Access to tmux sessions
   - Access to git repos

   Dashboard doesn't increase attack surface.

3. **Optional auth (future)** - Simple token-based auth:
   ```bash
   $ multiclaude start --enable-dashboard --dashboard-token=secret123
   # Dashboard requires ?token=secret123 in URL
   ```

4. **No sensitive data leaks** - Dashboard shows same data as `multiclaude status`
   - No API keys or credentials
   - No full git history
   - No file contents
   - Just agent state and message queue

---

## Migration Path: Gradual Rollout

### Phase 1: Read-Only Dashboard (MVP)
- ✅ View agent status
- ✅ View message queue
- ✅ Real-time updates via SSE
- ✅ Localhost-only
- ✅ Disabled by default

**Scope:** ~500 LOC, 1-2 weeks development

### Phase 2: Enhanced Visualization
- ✅ Message flow diagram (D3.js or similar lightweight)
- ✅ Activity timeline (last 1hr/24hr)
- ✅ System metrics (CPU, memory, message rate)

**Scope:** +300 LOC, 1 week

### Phase 3: Historical Data (Optional)
- ⬜ Store metrics in SQLite
- ⬜ Charts for long-running jobs
- ⬜ Export data as JSON/CSV

**Scope:** +500 LOC, 2 weeks

**Gates:** Each phase requires:
- Upstream approval
- No new dependencies
- Maintains "disabled by default"
- Zero regression for terminal-only users

---

## Alternatives Considered

### 1. **Rich Terminal UI (TUI) with Bubble Tea**

**Pros:**
- Stays in terminal
- No web browser needed
- Aligns with "terminal is the interface"

**Cons:**
- Limited visualization (text grid only)
- No remote viewing (SSH forwarding harder)
- Harder to implement real-time multi-pane updates
- New dependency (charmbracelet/bubbletea)

**Verdict:** Good for simple cases, but doesn't solve complex multi-agent visualization.

### 2. **Export Metrics to External Tools (Prometheus/Grafana)**

**Pros:**
- Professional-grade dashboards
- Existing ecosystem
- No custom UI code

**Cons:**
- Requires running 2 extra services (Prometheus + Grafana)
- Massive dependency overhead
- Overkill for most users
- Doesn't help with message flow visualization

**Verdict:** Too heavy. Better for production deployments, not local development.

### 3. **JSON Output + User Builds Own Viz**

```bash
$ multiclaude status --json | jq '...'
```

**Pros:**
- Unix philosophy (compose tools)
- No built-in UI needed

**Cons:**
- Every user reinvents the wheel
- No standard visualization
- Doesn't solve real-time updates
- Barrier to entry for non-power-users

**Verdict:** Works for experts, but leaves most users without good tooling.

### 4. **Status Quo: Terminal Only**

**Pros:**
- No development effort
- No new code to maintain
- Keeps project simple

**Cons:**
- Doesn't address multi-agent complexity
- Users build ad-hoc solutions (tmux scripts, polling loops)
- Missed opportunity for better UX
- Limits multiclaude adoption for complex workflows

**Verdict:** Safe, but doesn't scale with project ambition.

---

## Success Metrics

If we build this, how do we know it's valuable?

### Quantitative Metrics:
- **Adoption rate:** % of users who enable `--enable-dashboard`
- **Usage patterns:** Time spent on dashboard vs. CLI commands
- **Issue reduction:** Fewer "how do I monitor agents?" support requests
- **Performance:** Dashboard adds <5% CPU overhead when enabled

### Qualitative Metrics:
- **User feedback:** Survey users who enable dashboard
- **Demo effectiveness:** Conference/blog demos are easier to follow
- **Debugging stories:** "Dashboard helped me find bug X"
- **Stakeholder satisfaction:** Non-technical users can monitor progress

### Exit Criteria (When to Remove):
If after 6 months:
- < 5% adoption rate
- Frequent bug reports
- Maintenance burden too high
- No positive feedback

Then we deprecate and remove the dashboard.

---

## FAQ

### Q: Doesn't this violate the "terminal is the interface" philosophy?

**A:** No. The terminal remains the **primary** interface. Dashboard is **supplementary visualization**. All actions still require CLI:
- Start agents: `multiclaude worker create`
- Stop agents: `multiclaude worker complete`
- Send messages: `multiclaude agent message`

Dashboard is read-only. Think of it like `htop` vs `top` - both valid, different UX.

### Q: Why not use tmux layouts for visualization?

**A:** Tmux is great for viewing individual agent terminals, but:
- Can't show message flow graphs
- No metrics/charts
- Limited to text grid
- Hard to see all agents at once (layout limitations)

Dashboard complements tmux, doesn't replace it.

### Q: What about users without browsers?

**A:** Dashboard is optional. If you SSH into a server without X11 forwarding, just use CLI. No regression.

### Q: Will this bloat the binary size?

**A:** Minimal impact:
- HTML/CSS/JS embedded: ~10KB (gzipped)
- Go code: ~500 LOC ≈ ~20KB compiled
- Total: < 50KB added to binary (~1% of current size)

### Q: What if I want to disable it completely?

**A:** Build flag option:
```bash
go build -tags nodashboard ./cmd/multiclaude
```

Compiles out all dashboard code at build time.

### Q: Could this be a plugin instead?

**A:** Possible, but:
- More complex architecture (plugin system overhead)
- Harder for users to discover/install
- Dashboard needs deep daemon integration (state access)

Built-in but optional is simpler.

---

## Call to Action

### For Upstream Maintainers:

**Question:** Does this proposal align with multiclaude's philosophy?

**If yes:**
- Approve Phase 1 (MVP) implementation
- Provide feedback on technical approach
- Set expectations for contribution process

**If no:**
- What specific concerns remain?
- What would make this acceptable?
- Is there a compromise approach?

### For Fork Users:

**The dashboard already exists in fork-only/web-dashboard branch.**

If upstream rejects this proposal, the fork will maintain it separately. But we believe this feature benefits everyone and should be upstream.

---

## Appendix: Reference Implementation

The `fork-only/web-dashboard` branch contains a working implementation:

**Files:**
- `internal/daemon/server.go` - HTTP server (NewServer, routes)
- `internal/daemon/handlers.go` - Dashboard handlers
- `internal/cli/cli.go` - Added `dashboard` command
- `web/` - Static HTML/CSS/JS assets (embedded)

**Demo:**
```bash
git checkout fork-only/web-dashboard
go build ./cmd/multiclaude
multiclaude start --enable-dashboard
# Open http://localhost:8080
```

**Metrics:**
- Binary size: +42KB
- Performance: <2% CPU overhead
- Dependencies: 0 new
- Lines of code: ~450 (including HTML)

---

## Conclusion

Multi-agent orchestration is inherently **visual and spatial**. The terminal is excellent for commands and control, but **real-time multi-agent monitoring** needs better tooling.

An **optional, localhost-only web dashboard** provides:
- At-a-glance status of all agents
- Visual message flow understanding
- Real-time debugging capabilities
- Better stakeholder communication
- Remote monitoring support

All while respecting the project's core philosophy:
- Terminal remains primary interface
- No external APIs for consumption
- Minimal implementation (stdlib only)
- Optional feature (disabled by default)
- No control plane (read-only visualization)

**We believe multiclaude's ambition (Brownian ratchet, autonomous agents, complex workflows) deserves tooling that matches its complexity.**

The web dashboard makes multiclaude more accessible, debuggable, and professional - without sacrificing its Unix philosophy roots.

---

**Ready to discuss?** Open an issue or comment on this proposal.
