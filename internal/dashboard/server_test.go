package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dlorenc/multiclaude/internal/messages"
	"github.com/dlorenc/multiclaude/internal/state"
)

func setupTestServer(t *testing.T) (*Server, *state.State, *messages.Manager) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")
	messagesDir := filepath.Join(tmpDir, "messages")

	st := state.New(statePath)
	msgManager := messages.NewManager(messagesDir)

	// Add test repository
	repo := &state.Repository{
		GithubURL:   "https://github.com/test/repo",
		TmuxSession: "multiclaude-test",
		Agents:      make(map[string]state.Agent),
	}
	if err := st.AddRepo("test-repo", repo); err != nil {
		t.Fatalf("Failed to add repo: %v", err)
	}

	// Add test agents
	supervisor := state.Agent{
		Type:         state.AgentTypeSupervisor,
		WorktreePath: "/path/to/repo",
		TmuxWindow:   "supervisor",
		SessionID:    "test-session-1",
		PID:          12345,
		CreatedAt:    time.Now(),
	}
	if err := st.AddAgent("test-repo", "supervisor", supervisor); err != nil {
		t.Fatalf("Failed to add supervisor: %v", err)
	}

	worker := state.Agent{
		Type:         state.AgentTypeWorker,
		WorktreePath: "/path/to/worktree",
		TmuxWindow:   "worker1",
		SessionID:    "test-session-2",
		PID:          12346,
		Task:         "Test task",
		CreatedAt:    time.Now(),
	}
	if err := st.AddAgent("test-repo", "worker1", worker); err != nil {
		t.Fatalf("Failed to add worker: %v", err)
	}

	server := NewServer(st, msgManager)
	return server, st, msgManager
}

func TestHandleStatus(t *testing.T) {
	server, _, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["total_agents"] != float64(2) {
		t.Errorf("total_agents = %v, want 2", result["total_agents"])
	}

	if result["active_agents"] != float64(1) {
		t.Errorf("active_agents = %v, want 1", result["active_agents"])
	}

	if result["idle_agents"] != float64(1) {
		t.Errorf("idle_agents = %v, want 1", result["idle_agents"])
	}

	if result["repos"] != float64(1) {
		t.Errorf("repos = %v, want 1", result["repos"])
	}
}

func TestHandleRepos(t *testing.T) {
	server, _, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/repos", nil)
	w := httptest.NewRecorder()

	server.handleRepos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 repo, got %d", len(result))
	}

	repo := result[0]
	if repo["name"] != "test-repo" {
		t.Errorf("name = %v, want test-repo", repo["name"])
	}

	if repo["github_url"] != "https://github.com/test/repo" {
		t.Errorf("github_url = %v, want https://github.com/test/repo", repo["github_url"])
	}

	if repo["agent_count"] != float64(2) {
		t.Errorf("agent_count = %v, want 2", repo["agent_count"])
	}
}

func TestHandleRepoAgents(t *testing.T) {
	server, _, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/repos/test-repo/agents", nil)
	w := httptest.NewRecorder()

	server.handleRepoEndpoints(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 agents, got %d", len(result))
	}

	// Check that we have both supervisor and worker
	agentNames := make(map[string]bool)
	for _, agent := range result {
		name, ok := agent["name"].(string)
		if !ok {
			t.Errorf("Agent name is not a string: %v", agent["name"])
			continue
		}
		agentNames[name] = true

		// Check status
		status, ok := agent["status"].(string)
		if !ok {
			t.Errorf("Agent status is not a string: %v", agent["status"])
			continue
		}

		if name == "worker1" && status != "active" {
			t.Errorf("worker1 status = %s, want active", status)
		}
		if name == "supervisor" && status != "idle" {
			t.Errorf("supervisor status = %s, want idle", status)
		}
	}

	if !agentNames["supervisor"] {
		t.Error("supervisor not found in agents")
	}
	if !agentNames["worker1"] {
		t.Error("worker1 not found in agents")
	}
}

func TestHandleIndex(t *testing.T) {
	server, _, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %s, want text/html; charset=utf-8", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}

	// Check for key HTML elements
	if !strings.Contains(body, "MultiClaude Dashboard") {
		t.Error("Response does not contain 'MultiClaude Dashboard'")
	}

	if !strings.Contains(body, "escapeHtml") {
		t.Error("Response does not contain XSS protection function")
	}
}

func TestHandle404(t *testing.T) {
	server, _, _ := setupTestServer(t)

	tests := []struct {
		name string
		path string
	}{
		{"invalid repo", "/api/repos/nonexistent/agents"},
		{"invalid endpoint", "/api/repos/test-repo/invalid"},
		{"invalid path", "/api/invalid"},
		{"root 404", "/nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			// Route to appropriate handler
			if tt.path == "/" || tt.path == "/nonexistent" {
				server.handleIndex(w, req)
			} else if tt.path == "/api/repos/nonexistent/agents" || tt.path == "/api/repos/test-repo/invalid" {
				server.handleRepoEndpoints(w, req)
			} else {
				http.NotFound(w, req)
			}

			if w.Code != http.StatusNotFound {
				t.Errorf("Status code = %d, want %d for path %s", w.Code, http.StatusNotFound, tt.path)
			}
		})
	}
}


