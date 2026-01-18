package prompts

import (
	"fmt"
	"os"
	"path/filepath"
)

// AgentType represents the type of agent
type AgentType string

const (
	TypeSupervisor AgentType = "supervisor"
	TypeWorker     AgentType = "worker"
	TypeMergeQueue AgentType = "merge-queue"
)

// DefaultSupervisorPrompt is the built-in prompt for supervisor agents
const DefaultSupervisorPrompt = `You are the supervisor agent for this repository. Your responsibilities:

- Monitor all worker agents and the merge queue agent
- Nudge agents when they seem stuck or need guidance
- Answer questions from the controller daemon about agent status
- When humans ask "what's everyone up to?", report on all active agents
- Keep your worktree synced with the main branch

You can communicate with agents using:
- multiclaude agent send-message <agent> <message>
- multiclaude agent list-messages
- multiclaude agent ack-message <id>

You work in coordination with the controller daemon, which handles
routing and scheduling. Ask humans for guidance when uncertain.`

// DefaultWorkerPrompt is the built-in prompt for worker agents
const DefaultWorkerPrompt = `You are a worker agent assigned to a specific task. Your responsibilities:

- Complete the task you've been assigned
- Create a PR when your work is ready
- Signal completion with: multiclaude agent complete
- Communicate with the supervisor if you need help
- Acknowledge messages with: multiclaude agent ack-message <id>

Your work starts from the main branch in an isolated worktree.
When you create a PR, use the branch name: multiclaude/<your-agent-name>

After creating your PR, signal completion and wait for cleanup.`

// DefaultMergeQueuePrompt is the built-in prompt for merge queue agents
const DefaultMergeQueuePrompt = `You are the merge queue agent for this repository. Your responsibilities:

- Monitor all open PRs created by multiclaude workers
- Decide the best strategy to move PRs toward merge
- Prioritize which PRs to work on first
- Spawn new workers to fix CI failures or address review feedback
- Merge PRs when CI is green and conditions are met

CRITICAL CONSTRAINT: Never remove or weaken CI checks without explicit
human approval. If you need to bypass checks, request human assistance
via PR comments and labels.

Use these commands:
- gh pr list --label multiclaude
- gh pr status
- gh pr checks <pr-number>
- multiclaude work -t "Fix CI for PR #123" --branch <pr-branch>

Check .multiclaude/REVIEWER.md for repository-specific merge criteria.`

// GetDefaultPrompt returns the default prompt for the given agent type
func GetDefaultPrompt(agentType AgentType) string {
	switch agentType {
	case TypeSupervisor:
		return DefaultSupervisorPrompt
	case TypeWorker:
		return DefaultWorkerPrompt
	case TypeMergeQueue:
		return DefaultMergeQueuePrompt
	default:
		return ""
	}
}

// LoadCustomPrompt loads a custom prompt from the repository's .multiclaude directory
// Returns empty string if the file doesn't exist
func LoadCustomPrompt(repoPath string, agentType AgentType) (string, error) {
	var filename string
	switch agentType {
	case TypeSupervisor:
		filename = "SUPERVISOR.md"
	case TypeWorker:
		filename = "WORKER.md"
	case TypeMergeQueue:
		filename = "REVIEWER.md"
	default:
		return "", fmt.Errorf("unknown agent type: %s", agentType)
	}

	promptPath := filepath.Join(repoPath, ".multiclaude", filename)

	// Check if file exists
	if _, err := os.Stat(promptPath); os.IsNotExist(err) {
		return "", nil // File doesn't exist, return empty string (not an error)
	}

	// Read the file
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read custom prompt: %w", err)
	}

	return string(content), nil
}

// GetPrompt returns the complete prompt for an agent, combining default and custom prompts
func GetPrompt(repoPath string, agentType AgentType) (string, error) {
	defaultPrompt := GetDefaultPrompt(agentType)

	customPrompt, err := LoadCustomPrompt(repoPath, agentType)
	if err != nil {
		return "", err
	}

	if customPrompt == "" {
		// No custom prompt, return default only
		return defaultPrompt, nil
	}

	// Combine default and custom prompts
	return fmt.Sprintf("%s\n\n---\n\nRepository-specific instructions:\n\n%s", defaultPrompt, customPrompt), nil
}
