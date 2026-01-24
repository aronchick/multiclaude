# Socket API (Current Implementation)

<!-- socket-commands:
ping
status
stop
list_repos
add_repo
remove_repo
add_agent
remove_agent
list_agents
complete_agent
restart_agent
trigger_cleanup
repair_state
get_repo_config
update_repo_config
set_current_repo
get_current_repo
clear_current_repo
route_messages
task_history
spawn_agent
-->

The socket API is the only write-capable extension surface in multiclaude today. It is implemented in `internal/daemon/daemon.go` (`handleRequest`). This document tracks only the commands that exist in the code. Anything not listed here is **not implemented**.

## Protocol
- Transport: Unix domain socket at `~/.multiclaude/daemon.sock`
- Request type: JSON object `{ "command": "<name>", "args": { ... } }`
- Response type: `{ "success": true|false, "data": any, "error": string }`
- Client helper: `internal/socket.Client`

## Command Reference (source of truth)
Each command below matches a `case` in `handleRequest`.

| Command | Description | Args |
|---------|-------------|------|
| `ping` | Health check | none |
| `status` | Daemon status summary | none |
| `stop` | Stop the daemon | none |
| `list_repos` | List tracked repos (optionally rich info) | `rich` (bool, optional) |
| `add_repo` | Track a new repo | `path` (string) |
| `remove_repo` | Stop tracking a repo | `name` (string) |
| `add_agent` | Register an agent in state | `repo`, `name`, `type`, `worktree_path`, `tmux_window`, `session_id`, `pid` |
| `remove_agent` | Remove agent from state | `repo`, `name` |
| `list_agents` | List agents for a repo | `repo` |
| `complete_agent` | Mark agent ready for cleanup | `repo`, `name`, `summary`, `failure_reason` |
| `restart_agent` | Restart a persistent agent | `repo`, `name` |
| `trigger_cleanup` | Force cleanup cycle | none |
| `repair_state` | Run state repair routine | none |
| `get_repo_config` | Get merge-queue / pr-shepherd config | `repo` |
| `update_repo_config` | Update repo config | `repo`, `config` (JSON object) |
| `set_current_repo` | Persist current repo selection | `repo` |
| `get_current_repo` | Read current repo selection | none |
| `clear_current_repo` | Clear current repo selection | none |
| `route_messages` | Force message routing cycle | none |
| `task_history` | Return task history for a repo | `repo` |
| `spawn_agent` | Create a new agent worktree | `repo`, `type`, `task`, `name` (optional) |

## Minimal client examples

### Go
```go
package main

import (
    "fmt"

    "github.com/dlorenc/multiclaude/internal/socket"
)

func main() {
    client := socket.NewClient("/home/user/.multiclaude/daemon.sock")
    resp, err := client.Send(socket.Request{Command: "ping"})
    if err != nil {
        panic(err)
    }
    fmt.Printf("success=%v data=%v\n", resp.Success, resp.Data)
}
```

### Python
```python
import json
import socket

sock_path = "/home/user/.multiclaude/daemon.sock"
req = {"command": "status", "args": {}}

with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as s:
    s.connect(sock_path)
    s.sendall(json.dumps(req).encode("utf-8"))
    raw = s.recv(8192)
    resp = json.loads(raw.decode("utf-8"))
    print(resp)
```

## Updating this doc
- Add/remove commands **only** when the `handleRequest` switch changes.
- Keep the `socket-commands` marker above in sync; `go run ./cmd/verify-docs` enforces alignment.
- If you add arguments, update the table here with the real fields used by the handler.
