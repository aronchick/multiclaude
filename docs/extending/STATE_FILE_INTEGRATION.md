# State File Integration (Read-Only)

<!-- state-struct: State repos current_repo -->
<!-- state-struct: Repository github_url tmux_session agents task_history merge_queue_config pr_shepherd_config fork_config target_branch -->
<!-- state-struct: Agent type worktree_path tmux_window session_id pid task summary failure_reason created_at last_nudge ready_for_cleanup -->
<!-- state-struct: TaskHistoryEntry name task branch pr_url pr_number status summary failure_reason created_at completed_at -->
<!-- state-struct: MergeQueueConfig enabled track_mode -->
<!-- state-struct: PRShepherdConfig enabled track_mode -->
<!-- state-struct: ForkConfig is_fork upstream_url upstream_owner upstream_repo force_fork_mode -->

The daemon persists state to `~/.multiclaude/state.json` and writes it atomically. This file is safe for external tools to **read only**. Write access belongs to the daemon.

## Schema (from `internal/state/state.go`)
```json
{
  "repos": {
    "<repo-name>": {
      "github_url": "https://github.com/owner/repo",
      "tmux_session": "mc-repo",
      "agents": {
        "<agent-name>": {
          "type": "supervisor|worker|merge-queue|pr-shepherd|workspace|review|generic-persistent",
          "worktree_path": "/path/to/worktree",
          "tmux_window": "window-name",
          "session_id": "uuid",
          "pid": 12345,
          "task": "task description",
          "summary": "optional summary",
          "failure_reason": "optional failure",
          "created_at": "2025-01-01T00:00:00Z",
          "last_nudge": "2025-01-01T00:00:00Z",
          "ready_for_cleanup": false
        }
      },
      "task_history": [
        {
          "name": "clever-fox",
          "task": "...",
          "branch": "work/clever-fox",
          "pr_url": "https://github.com/...",
          "pr_number": 42,
          "status": "open|merged|closed|no-pr|failed|unknown",
          "summary": "optional summary",
          "failure_reason": "optional failure",
          "created_at": "2025-01-01T00:00:00Z",
          "completed_at": "2025-01-02T00:00:00Z"
        }
      ],
      "merge_queue_config": {
        "enabled": true,
        "track_mode": "all|author|assigned"
      },
      "pr_shepherd_config": {
        "enabled": true,
        "track_mode": "all|author|assigned"
      },
      "fork_config": {
        "is_fork": true,
        "upstream_url": "https://github.com/upstream/repo",
        "upstream_owner": "upstream",
        "upstream_repo": "repo",
        "force_fork_mode": false
      },
      "target_branch": "main"
    }
  },
  "current_repo": "<repo-name>"
}
```

## Reading the state file

### Go
```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/dlorenc/multiclaude/internal/state"
)

func main() {
    data, err := os.ReadFile("/home/user/.multiclaude/state.json")
    if err != nil {
        panic(err)
    }

    var st state.State
    if err := json.Unmarshal(data, &st); err != nil {
        panic(err)
    }

    for name := range st.Repos {
        fmt.Println("repo", name)
    }
}
```

### Python
```python
import json
from pathlib import Path

state_path = Path.home() / ".multiclaude" / "state.json"
state = json.loads(state_path.read_text())
for repo, data in state.get("repos", {}).items():
    print("repo", repo, "agents", list(data.get("agents", {}).keys()))
```

## Updating this doc
- Keep the `state-struct` markers above in sync with `internal/state/state.go`.
- Do **not** add fields here unless they exist in the structs.
- Run `go run ./cmd/verify-docs` after schema changes; CI will block if docs drift.
