package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// CIStatus represents the status of a CI run
type CIStatus string

const (
	// CIStatusSuccess indicates all CI checks passed
	CIStatusSuccess CIStatus = "success"
	// CIStatusFailure indicates at least one CI check failed
	CIStatusFailure CIStatus = "failure"
	// CIStatusPending indicates CI checks are still running
	CIStatusPending CIStatus = "pending"
	// CIStatusUnknown indicates the CI status could not be determined
	CIStatusUnknown CIStatus = "unknown"
)

// CILayerStatus represents the complete CI status information for a branch
type CILayerStatus struct {
	// Status is the overall CI status
	Status CIStatus `json:"status"`
	// LastCheckTime is when this status was last checked
	LastCheckTime time.Time `json:"last_check_time"`
	// FailureInfo contains details about the failure if Status is CIStatusFailure
	FailureInfo string `json:"failure_info,omitempty"`
	// WorkflowName is the name of the workflow that was checked
	WorkflowName string `json:"workflow_name,omitempty"`
	// RunURL is the URL to the workflow run
	RunURL string `json:"run_url,omitempty"`
}

// workflowRun represents a single workflow run from gh CLI output
type workflowRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	URL        string `json:"url"`
	HeadBranch string `json:"headBranch"`
}

// CheckCIStatus checks the CI status for a specific branch using gh CLI
// It uses 'gh run list' to get the latest workflow run status and parses
// the conclusion field to determine if CI is passing or failing.
func CheckCIStatus(ctx context.Context, owner, repo, branch string) (CILayerStatus, error) {
	status := CILayerStatus{
		LastCheckTime: time.Now(),
		Status:        CIStatusUnknown,
	}

	// Use gh run list to get the latest workflow run for the branch
	// Format: JSON with name, status, conclusion, url, and headBranch fields
	cmd := exec.CommandContext(ctx, "gh", "run", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--branch", branch,
		"--limit", "1",
		"--json", "name,status,conclusion,url,headBranch")

	output, err := cmd.Output()
	if err != nil {
		// Check if this is an exit error with stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			return status, fmt.Errorf("gh run list failed: %s: %w", string(exitErr.Stderr), err)
		}
		return status, fmt.Errorf("gh run list failed: %w", err)
	}

	// Parse the JSON output
	var runs []workflowRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return status, fmt.Errorf("failed to parse gh run list output: %w", err)
	}

	// If no runs found, CI status is unknown
	if len(runs) == 0 {
		status.Status = CIStatusUnknown
		status.FailureInfo = fmt.Sprintf("no workflow runs found for branch %s", branch)
		return status, nil
	}

	// Get the most recent run (should be first in the list)
	run := runs[0]
	status.WorkflowName = run.Name
	status.RunURL = run.URL

	// Determine status based on the workflow run status and conclusion
	// Status can be: "queued", "in_progress", "completed"
	// Conclusion can be: "success", "failure", "cancelled", "skipped", "timed_out", etc.
	switch run.Status {
	case "completed":
		// Check the conclusion to determine success or failure
		switch run.Conclusion {
		case "success":
			status.Status = CIStatusSuccess
		case "failure", "timed_out", "cancelled":
			status.Status = CIStatusFailure
			status.FailureInfo = fmt.Sprintf("workflow '%s' %s", run.Name, run.Conclusion)
		case "skipped":
			// Skipped is not a failure, but also not a clear success
			status.Status = CIStatusUnknown
			status.FailureInfo = fmt.Sprintf("workflow '%s' was skipped", run.Name)
		default:
			status.Status = CIStatusUnknown
			status.FailureInfo = fmt.Sprintf("workflow '%s' has unknown conclusion: %s", run.Name, run.Conclusion)
		}
	case "queued", "in_progress":
		status.Status = CIStatusPending
	default:
		status.Status = CIStatusUnknown
		status.FailureInfo = fmt.Sprintf("workflow '%s' has unknown status: %s", run.Name, run.Status)
	}

	return status, nil
}
