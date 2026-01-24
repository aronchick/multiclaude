package prompts

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dlorenc/multiclaude/internal/prompts/commands"
	"github.com/dlorenc/multiclaude/internal/state"
)

// AgentType is an alias for state.AgentType.
// Deprecated: Use state.AgentType directly instead.
type AgentType = state.AgentType

// Deprecated: Use state.AgentTypeSupervisor directly.
const TypeSupervisor = state.AgentTypeSupervisor

// Deprecated: Use state.AgentTypeWorker directly.
const TypeWorker = state.AgentTypeWorker

// Deprecated: Use state.AgentTypeMergeQueue directly.
const TypeMergeQueue = state.AgentTypeMergeQueue

// Deprecated: Use state.AgentTypeWorkspace directly.
const TypeWorkspace = state.AgentTypeWorkspace

// Deprecated: Use state.AgentTypeReview directly.
const TypeReview = state.AgentTypeReview

// Deprecated: Use state.AgentTypePRShepherd directly.
const TypePRShepherd = state.AgentTypePRShepherd

// Embedded default prompts
// Only supervisor and workspace are "hardcoded" - other agent types (worker, merge-queue, review)
// should come from configurable agent definitions in agent-templates.
//
//go:embed supervisor.md
var defaultSupervisorPrompt string

//go:embed workspace.md
var defaultWorkspacePrompt string

// GetDefaultPrompt returns the default prompt for the given agent type.
// Only supervisor and workspace have embedded default prompts.
// Worker, merge-queue, and review prompts should come from agent definitions.
func GetDefaultPrompt(agentType state.AgentType) string {
	switch agentType {
	case state.AgentTypeSupervisor:
		return defaultSupervisorPrompt
	case state.AgentTypeWorkspace:
		return defaultWorkspacePrompt
	case state.AgentTypeWorker, state.AgentTypeMergeQueue, state.AgentTypeReview:
		// These agent types should use configurable agent definitions
		// from ~/.multiclaude/repos/<repo>/agents/ or <repo>/.multiclaude/agents/
		return ""
	default:
		return ""
	}
}

// LoadCustomPrompt loads a custom prompt from the repository's .multiclaude directory.
// Returns empty string if the file doesn't exist.
//
// Deprecated: This function is deprecated. Use the configurable agent system instead:
// - Agent definitions: <repo>/.multiclaude/agents/<agent-name>.md
// - Local overrides: ~/.multiclaude/repos/<repo>/agents/<agent-name>.md
// See internal/agents package for the new system.
func LoadCustomPrompt(repoPath string, agentType state.AgentType) (string, error) {
	var filename string
	switch agentType {
	case state.AgentTypeSupervisor:
		filename = "SUPERVISOR.md"
	case state.AgentTypeWorker:
		filename = "WORKER.md"
	case state.AgentTypeMergeQueue:
		filename = "MERGE-QUEUE.md"
	case state.AgentTypePRShepherd:
		filename = "PR-SHEPHERD.md"
	case state.AgentTypeWorkspace:
		filename = "WORKSPACE.md"
	case state.AgentTypeReview:
		filename = "REVIEW.md"
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

// GetPrompt returns the complete prompt for an agent, combining default, custom prompts, CLI docs, and slash commands
func GetPrompt(repoPath string, agentType state.AgentType, cliDocs string) (string, error) {
	defaultPrompt := GetDefaultPrompt(agentType)

	customPrompt, err := LoadCustomPrompt(repoPath, agentType)
	if err != nil {
		return "", err
	}

	// Build the complete prompt
	var result string
	result = defaultPrompt

	// Add fork workflow guidance if in a fork
	forkInfo, err := DetectFork(repoPath)
	if err == nil && forkInfo.IsFork {
		forkPrompt := GenerateForkWorkflowPrompt(forkInfo)
		if forkPrompt != "" {
			result += fmt.Sprintf("\n\n---\n\n%s", forkPrompt)
		}
	}

	// Add CLI documentation
	if cliDocs != "" {
		result += fmt.Sprintf("\n\n---\n\n%s", cliDocs)
	}

	// Add slash commands section
	slashCommands := GetSlashCommandsPrompt()
	if slashCommands != "" {
		result += fmt.Sprintf("\n\n---\n\n%s", slashCommands)
	}

	// Add custom prompt if it exists
	if customPrompt != "" {
		result += fmt.Sprintf("\n\n---\n\nRepository-specific instructions:\n\n%s", customPrompt)
	}

	return result, nil
}

// GenerateTrackingModePrompt generates prompt text explaining which PRs to track
// based on the tracking mode. The trackMode parameter should be "all", "author", or "assigned".
func GenerateTrackingModePrompt(trackMode string) string {
	switch trackMode {
	case "author":
		return `## PR Tracking Mode: Author Only

**IMPORTANT**: This repository is configured to track only PRs where you (or the multiclaude system) are the author.

When listing and monitoring PRs, use:
` + "```bash" + `
gh pr list --author @me --label multiclaude
` + "```" + `

Do NOT process or attempt to merge PRs authored by others. Focus only on PRs created by multiclaude workers.`

	case "assigned":
		return `## PR Tracking Mode: Assigned Only

**IMPORTANT**: This repository is configured to track only PRs where you (or the multiclaude system) are assigned.

When listing and monitoring PRs, use:
` + "```bash" + `
gh pr list --assignee @me --label multiclaude
` + "```" + `

Do NOT process or attempt to merge PRs unless they are assigned to you. Focus only on PRs explicitly assigned to multiclaude.`

	default: // "all"
		return `## PR Tracking Mode: All PRs

This repository is configured to track all PRs with the multiclaude label.

When listing and monitoring PRs, use:
` + "```bash" + `
gh pr list --label multiclaude
` + "```" + `

Monitor and process all multiclaude-labeled PRs regardless of author or assignee.`
	}
}

// GenerateForkWorkflowPrompt generates prompt text explaining fork-based workflow.
// This is injected into all agent prompts when working in a fork.
func GenerateForkWorkflowPrompt(upstreamOwner, upstreamRepo, forkOwner string) string {
	return fmt.Sprintf(`## Fork Workflow (Auto-detected)

You are working in a fork of **%s/%s**.

**Key differences from upstream workflow:**

### Git Remotes
- **origin**: Your fork (%s/%s) - push branches here
- **upstream**: Original repo (%s/%s) - PRs target this repo

### Creating PRs
PRs should target the upstream repository, not your fork:
`+"```bash"+`
# Create a PR targeting upstream
gh pr create --repo %s/%s --head %s:<branch-name>

# View your PRs on upstream
gh pr list --repo %s/%s --author @me
`+"```"+`

### Keeping Synced
Keep your fork updated with upstream:
`+"```bash"+`
# Fetch upstream changes
git fetch upstream main

# Rebase your work
git rebase upstream/main

# Update your fork's main
git checkout main && git merge --ff-only upstream/main && git push origin main
`+"```"+`

### Important Notes
- **You cannot merge PRs** - upstream maintainers do that
- Create branches on your fork (origin), target upstream for PRs
- Keep rebasing onto upstream/main to avoid conflicts
- The pr-shepherd agent handles getting PRs ready for review
`, upstreamOwner, upstreamRepo,
		forkOwner, upstreamRepo,
		upstreamOwner, upstreamRepo,
		upstreamOwner, upstreamRepo, forkOwner,
		upstreamOwner, upstreamRepo)
}

// GetSlashCommandsPrompt returns a formatted prompt section containing all available
// slash commands. This can be included in agent prompts to document the available
// commands.
func GetSlashCommandsPrompt() string {
	var builder strings.Builder

	builder.WriteString("## Slash Commands\n\n")
	builder.WriteString("The following slash commands are available for use:\n\n")

	for _, cmd := range commands.AvailableCommands {
		content, err := commands.GetCommand(cmd.Name)
		if err != nil {
			continue
		}
		builder.WriteString(content)
		builder.WriteString("\n---\n\n")
	}

	return builder.String()
}

// ForkInfo contains information about a repository's fork status
type ForkInfo struct {
	IsFork         bool
	UpstreamRemote string
	UpstreamOwner  string
	UpstreamRepo   string
	ForkOwner      string
	ForkRepo       string
}

// DetectFork determines if the repository is a fork by checking git remotes
// Returns ForkInfo with details about the fork relationship
func DetectFork(repoPath string) (*ForkInfo, error) {
	info := &ForkInfo{
		IsFork: false,
	}

	// Check for upstream remote
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		// If git command fails, assume not a fork
		return info, nil
	}

	remotes := string(output)

	// Check if "upstream" remote exists
	if strings.Contains(remotes, "upstream") {
		info.IsFork = true
		info.UpstreamRemote = "upstream"

		// Extract upstream URL
		for _, line := range strings.Split(remotes, "\n") {
			if strings.HasPrefix(line, "upstream") && strings.Contains(line, "fetch") {
				// Parse URL: upstream	https://github.com/owner/repo.git (fetch)
				// or: upstream	git@github.com:owner/repo.git (fetch)
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					url := parts[1]
					owner, repo := parseGitHubURL(url)
					info.UpstreamOwner = owner
					info.UpstreamRepo = repo
				}
				break
			}
		}
	}

	// Get origin information (the fork)
	for _, line := range strings.Split(remotes, "\n") {
		if strings.HasPrefix(line, "origin") && strings.Contains(line, "fetch") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				url := parts[1]
				owner, repo := parseGitHubURL(url)
				info.ForkOwner = owner
				info.ForkRepo = repo
			}
			break
		}
	}

	return info, nil
}

// parseGitHubURL extracts owner and repo from a GitHub URL
// Handles both HTTPS and SSH formats
func parseGitHubURL(url string) (owner, repo string) {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle HTTPS: https://github.com/owner/repo
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) == 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				return pathParts[0], pathParts[1]
			}
		}
	}

	// Handle SSH: git@github.com:owner/repo
	if strings.Contains(url, "git@github.com:") {
		parts := strings.Split(url, "git@github.com:")
		if len(parts) == 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				return pathParts[0], pathParts[1]
			}
		}
	}

	return "", ""
}

// GenerateForkWorkflowPrompt generates prompt text explaining fork workflows
// based on whether the repository is a fork
func GenerateForkWorkflowPrompt(forkInfo *ForkInfo) string {
	if !forkInfo.IsFork {
		return "" // No additional guidance needed for upstream repos
	}

	upstream := fmt.Sprintf("%s/%s", forkInfo.UpstreamOwner, forkInfo.UpstreamRepo)
	fork := fmt.Sprintf("%s/%s", forkInfo.ForkOwner, forkInfo.ForkRepo)

	return `## Fork Workflow

**IMPORTANT**: This repository is a fork of ` + upstream + `.

You are working in the fork: ` + fork + `

### Fork→Upstream Contribution Strategy

When working in a fork, follow this workflow to contribute changes upstream:

1. **Work on feature branches** - Create branches for each feature/fix:
   ` + "```bash" + `
   git checkout -b feature/my-feature main
   ` + "```" + `

2. **Create PRs to fork main** - When your work is complete and tested:
   ` + "```bash" + `
   git push origin feature/my-feature
   gh pr create --base main  # Targets fork main (` + fork + `)
   ` + "```" + `

3. **Merge to fork main** - Once CI passes and the PR is approved, merge to your fork's main branch
   - This allows you to integrate and test changes in your fork
   - Fork main can contain experimental or in-progress work

4. **Create upstream PRs** - When a feature is complete and ready for upstream:
   ` + "```bash" + `
   # Create PR from your feature branch to upstream main
   gh pr create --repo ` + upstream + ` --head ` + forkInfo.ForkOwner + `:feature/my-feature --base main
   ` + "```" + `

   **CRITICAL**: Create the upstream PR from your **feature branch**, NOT from your fork's main branch.
   - CORRECT: --head ` + forkInfo.ForkOwner + `:feature/my-feature
   - WRONG: --head ` + forkInfo.ForkOwner + `:main

   This ensures the upstream PR contains only the specific feature changes, not all commits from your fork's main.

5. **Keep fork synced** - Regularly sync your fork with upstream:
   ` + "```bash" + `
   git fetch upstream
   git checkout main
   git merge upstream/main
   git push origin main
   ` + "```" + `

### Branch Naming

- Feature branches: ` + "`feature/<description>`" + `
- Bug fixes: ` + "`fix/<description>`" + `
- Multiclaude worker branches: ` + "`multiclaude/<worker-name>`" + `

### PR Target Guidelines

- **Fork PRs** (` + fork + `): Can be more experimental, WIP, or bundled changes
- **Upstream PRs** (` + upstream + `): Should be focused, complete, well-tested features
  - Each upstream PR should be a minimal, self-contained change
  - Review the CONTRIBUTING.md for upstream PR guidelines
  - Upstream maintainers expect high-quality, focused PRs

### Why This Workflow?

This approach allows you to:
- Iterate quickly in your fork without upstream approval
- Bundle related changes together in your fork
- Send focused, minimal PRs upstream
- Maintain a clean upstream contribution history
- Avoid overwhelming upstream maintainers with large PRs`
}
