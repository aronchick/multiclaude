package events

import (
	"context"
	"encoding/json"
	"os/exec"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Agent lifecycle events
	EventAgentStarted EventType = "agent_started"
	EventAgentStopped EventType = "agent_stopped"
	EventAgentIdle    EventType = "agent_idle"
	EventAgentFailed  EventType = "agent_failed"

	// PR events
	EventPRCreated EventType = "pr_created"
	EventPRMerged  EventType = "pr_merged"
	EventPRClosed  EventType = "pr_closed"

	// Task events
	EventTaskAssigned EventType = "task_assigned"
	EventTaskComplete EventType = "task_complete"

	// CI events
	EventCIFailed EventType = "ci_failed"
	EventCIPassed EventType = "ci_passed"

	// Message events
	EventMessageSent EventType = "message_sent"

	// Worker events
	EventWorkerStuck EventType = "worker_stuck"
)

// Event represents a lifecycle event in multiclaude
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	RepoName  string                 `json:"repo_name,omitempty"`
	AgentName string                 `json:"agent_name,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// HookConfig represents the configuration for event hooks
type HookConfig struct {
	// OnEvent is called for all events
	OnEvent string `json:"on_event,omitempty"`

	// Specific event hooks (optional, more granular)
	OnPRCreated     string `json:"on_pr_created,omitempty"`
	OnAgentIdle     string `json:"on_agent_idle,omitempty"`
	OnMergeComplete string `json:"on_merge_complete,omitempty"`
	OnAgentStarted  string `json:"on_agent_started,omitempty"`
	OnAgentStopped  string `json:"on_agent_stopped,omitempty"`
	OnTaskAssigned  string `json:"on_task_assigned,omitempty"`
	OnCIFailed      string `json:"on_ci_failed,omitempty"`
	OnWorkerStuck   string `json:"on_worker_stuck,omitempty"`
	OnMessageSent   string `json:"on_message_sent,omitempty"`
}

// Bus is the event bus that emits events to configured hooks
type Bus struct {
	config HookConfig
	mu     sync.RWMutex
}

// NewBus creates a new event bus with the given configuration
func NewBus(config HookConfig) *Bus {
	return &Bus{
		config: config,
	}
}

// UpdateConfig updates the hook configuration
func (b *Bus) UpdateConfig(config HookConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = config
}

// Emit emits an event to configured hooks
// This is fire-and-forget - no retries, no delivery guarantees
func (b *Bus) Emit(event Event) {
	b.mu.RLock()
	config := b.config
	b.mu.RUnlock()

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		// Can't emit event if we can't marshal it
		return
	}

	// Call generic hook if configured
	if config.OnEvent != "" {
		go b.callHook(config.OnEvent, event.Type, eventJSON)
	}

	// Call specific hook if configured
	specificHook := b.getSpecificHook(event.Type, config)
	if specificHook != "" {
		go b.callHook(specificHook, event.Type, eventJSON)
	}
}

// callHook executes a hook script with the event data
func (b *Bus) callHook(hookPath string, eventType EventType, eventJSON []byte) {
	// Create context with timeout to prevent hung hooks
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute hook: hook_script event_type event_json
	cmd := exec.CommandContext(ctx, hookPath, string(eventType), string(eventJSON))

	// Fire and forget - we don't wait for output or check errors
	// Users are responsible for their own error handling in hooks
	_ = cmd.Run()
}

// getSpecificHook returns the specific hook for an event type, if configured
func (b *Bus) getSpecificHook(eventType EventType, config HookConfig) string {
	switch eventType {
	case EventPRCreated:
		return config.OnPRCreated
	case EventAgentIdle:
		return config.OnAgentIdle
	case EventPRMerged:
		return config.OnMergeComplete
	case EventAgentStarted:
		return config.OnAgentStarted
	case EventAgentStopped:
		return config.OnAgentStopped
	case EventTaskAssigned:
		return config.OnTaskAssigned
	case EventCIFailed:
		return config.OnCIFailed
	case EventWorkerStuck:
		return config.OnWorkerStuck
	case EventMessageSent:
		return config.OnMessageSent
	default:
		return ""
	}
}

// Helper functions to create common events

// NewAgentStartedEvent creates an agent_started event
func NewAgentStartedEvent(repoName, agentName, agentType, task string) Event {
	return Event{
		Type:      EventAgentStarted,
		RepoName:  repoName,
		AgentName: agentName,
		Data: map[string]interface{}{
			"agent_type": agentType,
			"task":       task,
		},
	}
}

// NewAgentStoppedEvent creates an agent_stopped event
func NewAgentStoppedEvent(repoName, agentName, reason string) Event {
	return Event{
		Type:      EventAgentStopped,
		RepoName:  repoName,
		AgentName: agentName,
		Data: map[string]interface{}{
			"reason": reason,
		},
	}
}

// NewAgentIdleEvent creates an agent_idle event
func NewAgentIdleEvent(repoName, agentName string, durationSeconds int) Event {
	return Event{
		Type:      EventAgentIdle,
		RepoName:  repoName,
		AgentName: agentName,
		Data: map[string]interface{}{
			"duration_seconds": durationSeconds,
		},
	}
}

// NewPRCreatedEvent creates a pr_created event
func NewPRCreatedEvent(repoName, agentName string, prNumber int, title, url string) Event {
	return Event{
		Type:      EventPRCreated,
		RepoName:  repoName,
		AgentName: agentName,
		Data: map[string]interface{}{
			"pr_number": prNumber,
			"title":     title,
			"url":       url,
		},
	}
}

// NewPRMergedEvent creates a pr_merged event
func NewPRMergedEvent(repoName string, prNumber int, title string) Event {
	return Event{
		Type:     EventPRMerged,
		RepoName: repoName,
		Data: map[string]interface{}{
			"pr_number": prNumber,
			"title":     title,
		},
	}
}

// NewTaskAssignedEvent creates a task_assigned event
func NewTaskAssignedEvent(repoName, agentName, task string) Event {
	return Event{
		Type:      EventTaskAssigned,
		RepoName:  repoName,
		AgentName: agentName,
		Data: map[string]interface{}{
			"task": task,
		},
	}
}

// NewMessageSentEvent creates a message_sent event
func NewMessageSentEvent(repoName, from, to, messageType, body string) Event {
	return Event{
		Type:     EventMessageSent,
		RepoName: repoName,
		Data: map[string]interface{}{
			"from":         from,
			"to":           to,
			"message_type": messageType,
			"body":         body,
		},
	}
}

// NewCIFailedEvent creates a ci_failed event
func NewCIFailedEvent(repoName string, prNumber int, jobName string) Event {
	return Event{
		Type:     EventCIFailed,
		RepoName: repoName,
		Data: map[string]interface{}{
			"pr_number": prNumber,
			"job_name":  jobName,
		},
	}
}

// NewWorkerStuckEvent creates a worker_stuck event
func NewWorkerStuckEvent(repoName, agentName string, durationMinutes int) Event {
	return Event{
		Type:      EventWorkerStuck,
		RepoName:  repoName,
		AgentName: agentName,
		Data: map[string]interface{}{
			"duration_minutes": durationMinutes,
		},
	}
}
