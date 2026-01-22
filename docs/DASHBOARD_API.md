# Dashboard REST API Specification

## Overview

The MultiClaude dashboard exposes a REST API for programmatic access to daemon state. All endpoints are **read-only** and return JSON responses.

**Base URL:** `http://127.0.0.1:8080`

**Design Principles:**
- Read-only operations only
- Localhost-only access (127.0.0.1)
- No authentication required (local access)
- JSON responses with proper Content-Type headers
- Standard HTTP status codes

## Endpoints

### 1. Dashboard HTML Page

**Endpoint:** `GET /`

**Description:** Serves the dashboard HTML page.

**Response:**
- **Content-Type:** `text/html; charset=utf-8`
- **Status Code:** `200 OK`

**Example:**
```bash
curl http://127.0.0.1:8080/
```

---

### 2. System Status

**Endpoint:** `GET /api/status`

**Description:** Returns overall system status including agent counts and repository count.

**Response Schema:**
```json
{
  "total_agents": 5,
  "active_agents": 3,
  "idle_agents": 2,
  "repos": 2
}
```

**Fields:**
- `total_agents` (integer): Total number of agents across all repositories
- `active_agents` (integer): Number of agents currently working on tasks
- `idle_agents` (integer): Number of agents waiting for work
- `repos` (integer): Number of tracked repositories

**Example:**
```bash
curl http://127.0.0.1:8080/api/status

# Response:
{
  "total_agents": 5,
  "active_agents": 3,
  "idle_agents": 2,
  "repos": 2
}
```

---

### 3. List Repositories

**Endpoint:** `GET /api/repos`

**Description:** Returns list of all tracked repositories.

**Response Schema:**
```json
[
  {
    "name": "my-repo",
    "github_url": "https://github.com/user/my-repo",
    "tmux_session": "multiclaude-my-repo",
    "agent_count": 3
  }
]
```

**Fields:**
- `name` (string): Repository name (unique identifier)
- `github_url` (string): GitHub repository URL
- `tmux_session` (string): Associated tmux session name
- `agent_count` (integer): Number of agents in this repository

**Example:**
```bash
curl http://127.0.0.1:8080/api/repos

# Response:
[
  {
    "name": "my-repo",
    "github_url": "https://github.com/user/my-repo",
    "tmux_session": "multiclaude-my-repo",
    "agent_count": 3
  },
  {
    "name": "another-repo",
    "github_url": "https://github.com/user/another-repo",
    "tmux_session": "multiclaude-another-repo",
    "agent_count": 2
  }
]
```

---

### 4. List Repository Agents

**Endpoint:** `GET /api/repos/{repo}/agents`

**Description:** Returns list of all agents for a specific repository.

**Path Parameters:**
- `repo` (string): Repository name

**Response Schema:**
```json
[
  {
    "name": "supervisor",
    "type": "supervisor",
    "status": "idle",
    "task": "",
    "worktree_path": "/home/user/.multiclaude/repos/my-repo",
    "tmux_window": "supervisor",
    "created_at": "2026-01-22T10:00:00Z",
    "last_nudge": "2026-01-22T10:30:00Z"
  }
]
```

**Fields:**
- `name` (string): Agent name (unique within repository)
- `type` (string): Agent type - `supervisor`, `worker`, `review`, or `merge-queue`
- `status` (string): Agent status - `active`, `idle`, or `stuck`
  - `active`: Agent has a task assigned
  - `idle`: Agent has no task
  - `stuck`: Agent hasn't been nudged in >10 minutes
- `task` (string): Current task description (empty if idle)
- `worktree_path` (string): Path to agent's git worktree
- `tmux_window` (string): Tmux window name
- `created_at` (string): ISO 8601 timestamp of agent creation
- `last_nudge` (string): ISO 8601 timestamp of last nudge

**Example:**
```bash
curl http://127.0.0.1:8080/api/repos/my-repo/agents

# Response:
[
  {
    "name": "supervisor",
    "type": "supervisor",
    "status": "idle",
    "task": "",
    "worktree_path": "/home/user/.multiclaude/repos/my-repo",
    "tmux_window": "supervisor",
    "created_at": "2026-01-22T10:00:00Z",
    "last_nudge": "2026-01-22T10:30:00Z"
  },
  {
    "name": "happy-platypus",
    "type": "worker",
    "status": "active",
    "task": "Add authentication tests",
    "worktree_path": "/home/user/.multiclaude/wts/my-repo/happy-platypus",
    "tmux_window": "happy-platypus",
    "created_at": "2026-01-22T10:05:00Z",
    "last_nudge": "2026-01-22T10:35:00Z"
  }
]
```

**Error Responses:**
- `404 Not Found`: Repository does not exist

---

### 5. List Repository Messages

**Endpoint:** `GET /api/repos/{repo}/messages`

**Description:** Returns all messages for agents in a specific repository.

**Path Parameters:**
- `repo` (string): Repository name

**Response Schema:**
```json
[
  {
    "id": "msg-abc123def456",
    "from": "supervisor",
    "to": "happy-platypus",
    "timestamp": "2026-01-22T10:15:00Z",
    "body": "Please review PR #42",
    "status": "pending"
  }
]
```

**Fields:**
- `id` (string): Unique message identifier
- `from` (string): Sender agent name
- `to` (string): Recipient agent name
- `timestamp` (string): ISO 8601 timestamp
- `body` (string): Message content
- `status` (string): Message status - `pending`, `delivered`, `read`, or `acked`

**Example:**
```bash
curl http://127.0.0.1:8080/api/repos/my-repo/messages
```

**Error Responses:**
- `404 Not Found`: Repository does not exist

---

### 6. Repository Task History

**Endpoint:** `GET /api/repos/{repo}/history`

**Description:** Returns task completion history for a repository (last 50 tasks).

**Path Parameters:**
- `repo` (string): Repository name

**Response Schema:**
```json
[
  {
    "name": "happy-platypus",
    "task": "Add authentication tests",
    "branch": "add-auth-tests",
    "pr_url": "https://github.com/user/repo/pull/42",
    "pr_number": 42,
    "status": "completed",
    "summary": "Added comprehensive authentication tests",
    "failure_reason": "",
    "created_at": "2026-01-22T10:05:00Z",
    "completed_at": "2026-01-22T11:30:00Z"
  }
]
```

**Fields:**
- `name` (string): Agent name
- `task` (string): Task description
- `branch` (string): Git branch name
- `pr_url` (string): Pull request URL (if created)
- `pr_number` (integer): Pull request number (if created)
- `status` (string): Task status - `completed`, `failed`, or `abandoned`
- `summary` (string): Task completion summary
- `failure_reason` (string): Reason for failure (if status is `failed`)
- `created_at` (string): ISO 8601 timestamp of task start
- `completed_at` (string): ISO 8601 timestamp of task completion

**Example:**
```bash
curl http://127.0.0.1:8080/api/repos/my-repo/history
```

**Error Responses:**
- `404 Not Found`: Repository does not exist

---

### 7. Repository Activity Feed

**Endpoint:** `GET /api/repos/{repo}/activity`

**Description:** Returns recent activity feed combining tasks and messages.

**Path Parameters:**
- `repo` (string): Repository name

**Response Schema:**
```json
[
  {
    "type": "task",
    "timestamp": "2026-01-22T11:30:00Z",
    "agent": "happy-platypus",
    "message": "Completed: Add authentication tests",
    "status": "completed"
  },
  {
    "type": "message",
    "timestamp": "2026-01-22T10:15:00Z",
    "agent": "happy-platypus",
    "message": "Message from supervisor: Please review PR #42",
    "status": "pending"
  }
]
```

**Fields:**
- `type` (string): Activity type - `task` or `message`
- `timestamp` (string): ISO 8601 timestamp
- `agent` (string): Associated agent name
- `message` (string): Human-readable activity description
- `status` (string): Status (varies by type)

**Example:**
```bash
curl http://127.0.0.1:8080/api/repos/my-repo/activity
```

**Error Responses:**
- `404 Not Found`: Repository does not exist

---

## Error Handling

### Standard Error Responses

**404 Not Found:**
```
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8

404 page not found
```

**500 Internal Server Error:**
```
HTTP/1.1 500 Internal Server Error
Content-Type: text/plain; charset=utf-8

Internal Server Error
```

### Error Scenarios

1. **Repository Not Found:** Returns `404 Not Found` for `/api/repos/{repo}/*` endpoints
2. **Invalid Endpoint:** Returns `404 Not Found` for unknown paths
3. **JSON Encoding Error:** Returns `500 Internal Server Error` (rare)

---

## Usage Examples

### Using curl

```bash
# Get system status
curl http://127.0.0.1:8080/api/status

# List all repositories
curl http://127.0.0.1:8080/api/repos

# Get agents for a specific repo
curl http://127.0.0.1:8080/api/repos/my-repo/agents

# Get task history
curl http://127.0.0.1:8080/api/repos/my-repo/history

# Pretty-print JSON with jq
curl -s http://127.0.0.1:8080/api/status | jq .
```

### Using JavaScript (fetch)

```javascript
// Get system status
const status = await fetch('http://127.0.0.1:8080/api/status')
  .then(r => r.json());

console.log(`Total agents: ${status.total_agents}`);

// Get all repositories
const repos = await fetch('http://127.0.0.1:8080/api/repos')
  .then(r => r.json());

repos.forEach(repo => {
  console.log(`${repo.name}: ${repo.agent_count} agents`);
});
```

### Using Python (requests)

```python
import requests

# Get system status
response = requests.get('http://127.0.0.1:8080/api/status')
status = response.json()
print(f"Total agents: {status['total_agents']}")

# Get all repositories
response = requests.get('http://127.0.0.1:8080/api/repos')
repos = response.json()
for repo in repos:
    print(f"{repo['name']}: {repo['agent_count']} agents")
```

---

## Implementation Notes

### Logging

All HTTP requests are logged with:
- HTTP method
- URL path
- Request duration

Example log output:
```
GET /api/status 1.234ms
GET /api/repos 2.345ms
GET /api/repos/my-repo/agents 3.456ms
```

### Performance

- All endpoints read from in-memory state (no database queries)
- Response times typically <5ms
- No rate limiting (localhost-only access)
- No caching (state changes frequently)

### Concurrency

- All endpoints are safe for concurrent access
- State reads are protected by mutex locks
- No write operations to worry about

### Content-Type Headers

All JSON responses include:
```
Content-Type: application/json
```

HTML response includes:
```
Content-Type: text/html; charset=utf-8
```

---

## Related Documentation

- [DASHBOARD.md](DASHBOARD.md) - Complete dashboard documentation
- [SPEC.md](../SPEC.md) - MultiClaude architecture
- [AGENTS.md](../AGENTS.md) - Agent types and behavior

