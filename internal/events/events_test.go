package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  string
	}{
		{"agent_started", EventAgentStarted, "agent_started"},
		{"agent_stopped", EventAgentStopped, "agent_stopped"},
		{"agent_idle", EventAgentIdle, "agent_idle"},
		{"pr_created", EventPRCreated, "pr_created"},
		{"pr_merged", EventPRMerged, "pr_merged"},
		{"task_assigned", EventTaskAssigned, "task_assigned"},
		{"ci_failed", EventCIFailed, "ci_failed"},
		{"worker_stuck", EventWorkerStuck, "worker_stuck"},
		{"message_sent", EventMessageSent, "message_sent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("EventType %s = %s, want %s", tt.name, tt.eventType, tt.expected)
			}
		})
	}
}

func TestNewAgentStartedEvent(t *testing.T) {
	event := NewAgentStartedEvent("test-repo", "test-agent", "worker", "Test task")

	if event.Type != EventAgentStarted {
		t.Errorf("Type = %s, want %s", event.Type, EventAgentStarted)
	}
	if event.RepoName != "test-repo" {
		t.Errorf("RepoName = %s, want test-repo", event.RepoName)
	}
	if event.AgentName != "test-agent" {
		t.Errorf("AgentName = %s, want test-agent", event.AgentName)
	}
	if event.Data["agent_type"] != "worker" {
		t.Errorf("Data[agent_type] = %v, want worker", event.Data["agent_type"])
	}
	if event.Data["task"] != "Test task" {
		t.Errorf("Data[task] = %v, want Test task", event.Data["task"])
	}
}

func TestNewAgentStoppedEvent(t *testing.T) {
	event := NewAgentStoppedEvent("test-repo", "test-agent", "completed")

	if event.Type != EventAgentStopped {
		t.Errorf("Type = %s, want %s", event.Type, EventAgentStopped)
	}
	if event.Data["reason"] != "completed" {
		t.Errorf("Data[reason] = %v, want completed", event.Data["reason"])
	}
}

func TestEventJSONMarshaling(t *testing.T) {
	event := NewAgentStartedEvent("test-repo", "test-agent", "worker", "Test task")

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var unmarshaled Event
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if unmarshaled.Type != event.Type {
		t.Errorf("Unmarshaled Type = %s, want %s", unmarshaled.Type, event.Type)
	}
	if unmarshaled.RepoName != event.RepoName {
		t.Errorf("Unmarshaled RepoName = %s, want %s", unmarshaled.RepoName, event.RepoName)
	}
}

func TestBusEmit(t *testing.T) {
	// Create a temporary hook script
	tmpDir := t.TempDir()
	hookScript := filepath.Join(tmpDir, "test-hook.sh")
	outputFile := filepath.Join(tmpDir, "output.json")

	// Write a simple hook that saves the event to a file
	// Hook receives: hookPath eventType eventJSON
	hookContent := `#!/bin/bash
echo "$2" > ` + outputFile + `
`
	if err := os.WriteFile(hookScript, []byte(hookContent), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	// Create bus with hook configuration
	config := HookConfig{
		OnEvent: hookScript,
	}
	bus := NewBus(config)

	// Emit an event
	event := NewAgentStartedEvent("test-repo", "test-agent", "worker", "Test task")
	bus.Emit(event)

	// Wait for hook to execute (fire-and-forget, so we need to wait a bit)
	// Try multiple times with increasing delays
	var data []byte
	var err error
	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)
		data, err = os.ReadFile(outputFile)
		if err == nil && len(data) > 0 {
			break
		}
	}

	if err != nil {
		t.Fatalf("Hook did not create output file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Output file is empty")
	}

	var savedEvent Event
	if err := json.Unmarshal(data, &savedEvent); err != nil {
		t.Fatalf("Failed to unmarshal saved event: %v (data: %s)", err, string(data))
	}

	if savedEvent.Type != event.Type {
		t.Errorf("Saved event Type = %s, want %s", savedEvent.Type, event.Type)
	}
}

func TestBusUpdateConfig(t *testing.T) {
	bus := NewBus(HookConfig{})

	newConfig := HookConfig{
		OnEvent:     "/path/to/hook.sh",
		OnPRCreated: "/path/to/pr-hook.sh",
	}

	bus.UpdateConfig(newConfig)

	// Verify config was updated (we can't directly access it, but we can test behavior)
	// This is a basic test - in practice, we'd emit events and verify the right hooks are called
}

func TestSpecificEventHooks(t *testing.T) {
	tmpDir := t.TempDir()
	prHookScript := filepath.Join(tmpDir, "pr-hook.sh")
	prOutputFile := filepath.Join(tmpDir, "pr-output.json")

	// Write PR-specific hook
	// Hook receives: hookPath eventType eventJSON
	hookContent := `#!/bin/bash
echo "$2" > ` + prOutputFile + `
`
	if err := os.WriteFile(prHookScript, []byte(hookContent), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	// Create bus with PR-specific hook
	config := HookConfig{
		OnPRCreated: prHookScript,
	}
	bus := NewBus(config)

	// Emit a PR created event
	event := NewPRCreatedEvent("test-repo", "test-agent", 42, "Test PR", "https://github.com/test/repo/pull/42")
	bus.Emit(event)

	// Wait for hook execution
	time.Sleep(100 * time.Millisecond)

	// Verify hook was called
	if _, err := os.Stat(prOutputFile); os.IsNotExist(err) {
		t.Error("PR hook did not create output file")
	}
}
