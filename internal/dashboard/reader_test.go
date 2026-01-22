package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dlorenc/multiclaude/internal/state"
)

func TestNewStateReader(t *testing.T) {
	// Create temp directory for test state files
	tmpDir, err := os.MkdirTemp("", "dashboard-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test state file
	statePath := filepath.Join(tmpDir, "state.json")
	testState := state.New(statePath)
	testState.AddRepo("test-repo", &state.Repository{
		GithubURL:   "https://github.com/test/repo",
		TmuxSession: "mc-test-repo",
		Agents:      make(map[string]state.Agent),
	})
	if err := testState.Save(); err != nil {
		t.Fatalf("failed to save test state: %v", err)
	}

	// Create state reader
	reader, err := NewStateReader([]string{statePath})
	if err != nil {
		t.Fatalf("NewStateReader failed: %v", err)
	}
	defer reader.Close()

	// Verify state was loaded
	loadedState, ok := reader.GetState(statePath)
	if !ok {
		t.Fatalf("state not found for path: %s", statePath)
	}

	if len(loadedState.Repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(loadedState.Repos))
	}

	if _, ok := loadedState.Repos["test-repo"]; !ok {
		t.Errorf("expected repo 'test-repo' not found")
	}
}

func TestGetAggregatedState(t *testing.T) {
	// Create temp directory for test state files
	tmpDir, err := os.MkdirTemp("", "dashboard-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test state file
	statePath := filepath.Join(tmpDir, "state.json")
	testState := state.New(statePath)

	// Add test repository with agents
	testRepo := &state.Repository{
		GithubURL:   "https://github.com/test/repo",
		TmuxSession: "mc-test-repo",
		Agents: map[string]state.Agent{
			"supervisor": {
				Type:         state.AgentTypeSupervisor,
				WorktreePath: "/path/to/worktree",
				TmuxWindow:   "0",
				CreatedAt:    time.Now(),
			},
			"worker-1": {
				Type:         state.AgentTypeWorker,
				WorktreePath: "/path/to/worktree",
				TmuxWindow:   "1",
				Task:         "Test task",
				CreatedAt:    time.Now(),
			},
		},
	}

	testState.AddRepo("test-repo", testRepo)
	if err := testState.Save(); err != nil {
		t.Fatalf("failed to save test state: %v", err)
	}

	// Create state reader
	reader, err := NewStateReader([]string{statePath})
	if err != nil {
		t.Fatalf("NewStateReader failed: %v", err)
	}
	defer reader.Close()

	// Get aggregated state
	agg := reader.GetAggregatedState()

	if len(agg.Machines) != 1 {
		t.Errorf("expected 1 machine, got %d", len(agg.Machines))
	}

	// Verify the aggregated state contains our test repo
	found := false
	for _, machine := range agg.Machines {
		if repo, ok := machine.Repos["test-repo"]; ok {
			found = true
			if len(repo.Agents) != 2 {
				t.Errorf("expected 2 agents, got %d", len(repo.Agents))
			}
		}
	}

	if !found {
		t.Errorf("test-repo not found in aggregated state")
	}
}

func TestAggregatedStateJSON(t *testing.T) {
	agg := &AggregatedState{
		Machines: map[string]*MachineState{
			"test-machine": {
				Path:  "/path/to/state.json",
				Repos: make(map[string]*state.Repository),
			},
		},
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Marshal to JSON
	data, err := json.Marshal(agg)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	// Verify timestamp is formatted as RFC3339
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	timestamp, ok := result["timestamp"].(string)
	if !ok {
		t.Fatalf("timestamp is not a string")
	}

	if timestamp != "2024-01-01T12:00:00Z" {
		t.Errorf("expected timestamp '2024-01-01T12:00:00Z', got '%s'", timestamp)
	}
}

func TestStateReaderWithNonexistentFile(t *testing.T) {
	// Test with a file that doesn't exist
	reader, err := NewStateReader([]string{"/nonexistent/state.json"})
	if err != nil {
		t.Fatalf("NewStateReader should not fail for nonexistent files: %v", err)
	}
	defer reader.Close()

	// Should have machine entry but with empty repos (state.Load creates empty state for nonexistent files)
	agg := reader.GetAggregatedState()
	if len(agg.Machines) != 1 {
		t.Errorf("expected 1 machine entry (with empty repos), got %d", len(agg.Machines))
	}

	// Verify the machine has no repos
	for _, machine := range agg.Machines {
		if len(machine.Repos) != 0 {
			t.Errorf("expected 0 repos for nonexistent file, got %d", len(machine.Repos))
		}
	}
}
