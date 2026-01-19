package cli

import (
	"fmt"
	"os"

	"github.com/dlorenc/multiclaude/internal/errors"
	"github.com/dlorenc/multiclaude/internal/socket"
	"github.com/manifoldco/promptui"
	"github.com/mattn/go-isatty"
)

// isInteractive returns true if stdin is a terminal (TTY)
func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// selectRepo prompts the user to select a repository from the available list.
// If only one repo exists, it returns that repo without prompting.
// If no repos exist, it returns an error with guidance.
// If stdin is not a TTY, it falls back to the standard MultipleRepos error.
func (c *CLI) selectRepo() (string, error) {
	// Get list of repos from daemon
	client := socket.NewClient(c.paths.DaemonSock)
	resp, err := client.Send(socket.Request{
		Command: "list_repos",
	})
	if err != nil {
		return "", errors.DaemonCommunicationFailed("listing repositories", err)
	}

	if !resp.Success {
		return "", errors.Wrap(errors.CategoryRuntime, "failed to list repos", fmt.Errorf("%s", resp.Error))
	}

	repos, ok := resp.Data.([]interface{})
	if !ok {
		return "", errors.New(errors.CategoryRuntime, "unexpected response format from daemon")
	}

	// No repos tracked
	if len(repos) == 0 {
		return "", errors.NotInRepo()
	}

	// Extract repo names
	var repoNames []string
	for _, repo := range repos {
		if repoMap, ok := repo.(map[string]interface{}); ok {
			if name, ok := repoMap["name"].(string); ok {
				repoNames = append(repoNames, name)
			}
		}
	}

	// If only one repo, return it directly
	if len(repoNames) == 1 {
		return repoNames[0], nil
	}

	// Check if we can prompt interactively
	if !isInteractive() {
		return "", errors.MultipleRepos()
	}

	// Prompt user to select
	prompt := promptui.Select{
		Label: "Select repository",
		Items: repoNames,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "\U00002705 {{ . | green }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
			return "", errors.New(errors.CategoryUsage, "selection cancelled")
		}
		return "", errors.Wrap(errors.CategoryRuntime, "selection failed", err)
	}

	return result, nil
}

// selectWorkspace prompts the user to select a workspace from the given repository.
// If only one workspace exists, it returns that workspace without prompting.
// If no workspaces exist, it returns an error with guidance.
// If stdin is not a TTY, it falls back to an InvalidUsage error.
func (c *CLI) selectWorkspace(repoName string) (string, error) {
	// Get list of agents from daemon
	client := socket.NewClient(c.paths.DaemonSock)
	resp, err := client.Send(socket.Request{
		Command: "list_agents",
		Args: map[string]interface{}{
			"repo": repoName,
		},
	})
	if err != nil {
		return "", errors.DaemonCommunicationFailed("listing workspaces", err)
	}

	if !resp.Success {
		return "", errors.Wrap(errors.CategoryRuntime, "failed to list workspaces", fmt.Errorf("%s", resp.Error))
	}

	agents, ok := resp.Data.([]interface{})
	if !ok {
		return "", errors.New(errors.CategoryRuntime, "unexpected response format from daemon")
	}

	// Filter for workspaces
	var workspaceNames []string
	for _, agent := range agents {
		if agentMap, ok := agent.(map[string]interface{}); ok {
			agentType, _ := agentMap["type"].(string)
			if agentType == "workspace" {
				if name, ok := agentMap["name"].(string); ok {
					workspaceNames = append(workspaceNames, name)
				}
			}
		}
	}

	// No workspaces
	if len(workspaceNames) == 0 {
		return "", &errors.CLIError{
			Category:   errors.CategoryNotFound,
			Message:    fmt.Sprintf("no workspaces in repository '%s'", repoName),
			Suggestion: "multiclaude workspace add <name>",
		}
	}

	// If only one workspace, return it directly
	if len(workspaceNames) == 1 {
		return workspaceNames[0], nil
	}

	// Check if we can prompt interactively
	if !isInteractive() {
		return "", errors.InvalidUsage("usage: multiclaude workspace connect <name>")
	}

	// Prompt user to select
	prompt := promptui.Select{
		Label: fmt.Sprintf("Select workspace (repo: %s)", repoName),
		Items: workspaceNames,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "\U00002705 {{ . | green }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
			return "", errors.New(errors.CategoryUsage, "selection cancelled")
		}
		return "", errors.Wrap(errors.CategoryRuntime, "selection failed", err)
	}

	return result, nil
}

// resolveRepo resolves the repository name from flags, CWD inference, or interactive selection.
// This is the main entry point for repo resolution that should replace the pattern:
//
//	if r, ok := flags["repo"]; ok {
//	    repoName = r
//	} else {
//	    if inferred, err := c.inferRepoFromCwd(); err == nil {
//	        repoName = inferred
//	    } else {
//	        return errors.MultipleRepos()
//	    }
//	}
func (c *CLI) resolveRepo(flags map[string]string) (string, error) {
	// Check for explicit --repo flag
	if r, ok := flags["repo"]; ok {
		return r, nil
	}

	// Try to infer from CWD
	if inferred, err := c.inferRepoFromCwd(); err == nil {
		return inferred, nil
	}

	// Fall back to interactive selection
	return c.selectRepo()
}
