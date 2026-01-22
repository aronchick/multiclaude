# Web Dashboard (FORK-ONLY)

**⚠️ This is a fork-only feature that upstream explicitly rejects.**

The multiclaude web dashboard provides a read-only web interface for monitoring your multiclaude agents and repositories.

## Why Fork-Only?

Upstream (`dlorenc/multiclaude`) explicitly rejects web interfaces and dashboards per their [ROADMAP.md](../ROADMAP.md):

> **Out of Scope (Do Not Implement)**
>
> 3. **Web interfaces or dashboards**
>    - No REST APIs for external consumption
>    - No browser-based UIs
>    - Terminal is the interface

This fork (`aronchick/multiclaude`) adds web dashboard functionality while maintaining compatibility with upstream.

## Features

- **Real-time monitoring** - Live updates via Server-Sent Events
- **Multi-repository view** - See all your tracked repos in one place
- **Agent status** - View active agents and their tasks
- **Task history** - Browse completed tasks and PR status
- **Read-only** - No control plane, only viewing
- **Local-first** - Runs on localhost by default

## Installation

Build the web dashboard binary:

```bash
go build ./cmd/multiclaude-web
```

Or install it to your `$GOPATH/bin`:

```bash
go install ./cmd/multiclaude-web
```

## Quick Start

1. Make sure multiclaude is initialized and running:
   ```bash
   multiclaude init https://github.com/user/repo
   multiclaude start
   ```

2. Start the web dashboard:
   ```bash
   multiclaude-web
   ```

3. Open your browser to http://localhost:8080

## Usage

### Basic Usage

Start the dashboard with default settings (localhost:8080):

```bash
multiclaude-web
```

### Custom Port

Run on a different port:

```bash
multiclaude-web --port 3000
```

### Listen on All Interfaces

By default, the dashboard only binds to localhost. To make it accessible from other machines on your network:

```bash
multiclaude-web --bind 0.0.0.0
```

**⚠️ Warning:** Only use `--bind 0.0.0.0` on trusted networks. The dashboard has no authentication.

### Custom State File

If your multiclaude state file is in a non-default location:

```bash
multiclaude-web --state /path/to/state.json
```

## API Endpoints

The dashboard exposes a REST API:

| Endpoint | Description |
|----------|-------------|
| `GET /api/state` | Full aggregated state |
| `GET /api/repos` | List all repositories |
| `GET /api/repos/{name}` | Repository details |
| `GET /api/repos/{name}/agents` | Repository agents |
| `GET /api/repos/{name}/history` | Task history |
| `GET /api/events` | Server-Sent Events (live updates) |

### Example API Usage

```bash
# Get all repositories
curl http://localhost:8080/api/repos

# Get agents for a specific repo
curl http://localhost:8080/api/repos/myrepo/agents

# Get task history (limit to 10)
curl http://localhost:8080/api/repos/myrepo/history?limit=10

# Stream live updates
curl -N http://localhost:8080/api/events
```

## Architecture

```
┌──────────────┐
│  ~/.multiclaude │
│   state.json │
└──────┬───────┘
       │ (watches)
       │
┌──────▼───────┐
│ multiclaude-web │
│  (read-only)    │
└────────┬────────┘
         │
         ▼
   Web Browser
```

### Components

- **StateReader** (`internal/dashboard/reader.go`) - Reads and watches state files
- **APIHandler** (`internal/dashboard/api.go`) - REST API endpoints
- **Server** (`internal/dashboard/server.go`) - HTTP server setup
- **Frontend** (`internal/dashboard/web/`) - HTML/CSS/JS single-page app

## Multi-Machine Monitoring (Future)

The architecture supports monitoring multiple machines, though this is not yet implemented:

```bash
# Planned for future
multiclaude-web \
  --state ~/.multiclaude/state.json \
  --remote ssh://dev.example.com/home/user/.multiclaude/state.json \
  --remote ssh://ci.example.com/home/user/.multiclaude/state.json
```

## Development

### Project Structure

```
cmd/multiclaude-web/         - Web server binary
  main.go                    - Entry point

internal/dashboard/          - Dashboard implementation
  reader.go                  - State file reader
  api.go                     - REST API handlers
  server.go                  - HTTP server
  web/                       - Frontend assets
    index.html               - Main HTML
    app.js                   - JavaScript
    styles.css               - CSS

docs/WEB_DASHBOARD.md        - This file
```

### Running in Development

```bash
# Build and run
go build ./cmd/multiclaude-web && ./multiclaude-web

# Run tests (when added)
go test ./internal/dashboard/...
```

### Modifying the Frontend

The frontend files in `internal/dashboard/web/` are embedded into the binary using Go's `embed` directive. After modifying HTML/CSS/JS:

1. Rebuild the binary:
   ```bash
   go build ./cmd/multiclaude-web
   ```

2. The changes will be included in the next binary

## Security Considerations

**⚠️ Important Security Notes:**

1. **No Authentication** - The dashboard has no authentication mechanism
2. **Read-Only** - Cannot control agents, but can see state data
3. **Local-Only by Default** - Binds to 127.0.0.1 (localhost) by default
4. **Network Exposure** - Use `--bind 0.0.0.0` only on trusted networks

### Recommended Usage

- **Development**: Run on localhost (default)
- **Team/LAN**: Use `--bind 0.0.0.0` with firewall rules
- **Remote**: Use SSH tunneling instead of exposing publicly

### SSH Tunnel Example

To access the dashboard remotely without exposing it:

```bash
# On remote machine
multiclaude-web

# On local machine
ssh -L 8080:localhost:8080 user@remote.example.com

# Then browse to http://localhost:8080
```

## Troubleshooting

### Dashboard won't start

**Problem:** `Error: state file not found`

**Solution:** Make sure multiclaude is initialized:
```bash
multiclaude init https://github.com/user/repo
```

### Port already in use

**Problem:** `Error: listen tcp :8080: bind: address already in use`

**Solution:** Use a different port:
```bash
multiclaude-web --port 3000
```

### Live updates not working

**Problem:** Dashboard shows "Live" indicator but updates don't appear

**Solution:**
1. Check that the state file is being updated (run `multiclaude list`)
2. Check browser console for SSE errors
3. Restart the dashboard

### Empty dashboard

**Problem:** Dashboard loads but shows no repositories

**Solution:**
1. Verify multiclaude has repositories: `multiclaude list`
2. Check the state file path: `~/.multiclaude/state.json`
3. Try the custom state path: `multiclaude-web --state /path/to/state.json`

## Contributing

This feature is fork-only and will not be contributed upstream. Changes should:

1. Be marked as `[fork-only]` in commit messages
2. Not modify core multiclaude functionality
3. Remain optional (core multiclaude works without it)
4. Be documented in this file

## Related Documentation

- [Fork Features Roadmap](../docs/FORK_FEATURES_ROADMAP.md) - Overview of fork-only features
- [Upstream Workflow](../docs/UPSTREAM_WORKFLOW.md) - Managing fork/upstream relationship
- [CLAUDE.md](../CLAUDE.md) - Project overview and architecture

## License

Same as multiclaude (see [LICENSE](../LICENSE))
