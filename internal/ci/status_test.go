package ci

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestCIStatus_Constants(t *testing.T) {
	// Verify that the constants are defined correctly
	if CIStatusSuccess != "success" {
		t.Errorf("CIStatusSuccess = %q, want %q", CIStatusSuccess, "success")
	}
	if CIStatusFailure != "failure" {
		t.Errorf("CIStatusFailure = %q, want %q", CIStatusFailure, "failure")
	}
	if CIStatusPending != "pending" {
		t.Errorf("CIStatusPending = %q, want %q", CIStatusPending, "pending")
	}
	if CIStatusUnknown != "unknown" {
		t.Errorf("CIStatusUnknown = %q, want %q", CIStatusUnknown, "unknown")
	}
}

func TestCILayerStatus_JSONMarshaling(t *testing.T) {
	now := time.Now().Round(time.Second) // Round to avoid precision issues

	tests := []struct {
		name   string
		status CILayerStatus
	}{
		{
			name: "success status",
			status: CILayerStatus{
				Status:        CIStatusSuccess,
				LastCheckTime: now,
				WorkflowName:  "CI",
				RunURL:        "https://github.com/owner/repo/actions/runs/123",
			},
		},
		{
			name: "failure status with info",
			status: CILayerStatus{
				Status:        CIStatusFailure,
				LastCheckTime: now,
				FailureInfo:   "workflow 'CI' failure",
				WorkflowName:  "CI",
				RunURL:        "https://github.com/owner/repo/actions/runs/124",
			},
		},
		{
			name: "pending status",
			status: CILayerStatus{
				Status:        CIStatusPending,
				LastCheckTime: now,
				WorkflowName:  "CI",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			// Unmarshal back
			var got CILayerStatus
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			// Compare fields
			if got.Status != tt.status.Status {
				t.Errorf("Status = %q, want %q", got.Status, tt.status.Status)
			}
			if !got.LastCheckTime.Equal(tt.status.LastCheckTime) {
				t.Errorf("LastCheckTime = %v, want %v", got.LastCheckTime, tt.status.LastCheckTime)
			}
			if got.FailureInfo != tt.status.FailureInfo {
				t.Errorf("FailureInfo = %q, want %q", got.FailureInfo, tt.status.FailureInfo)
			}
			if got.WorkflowName != tt.status.WorkflowName {
				t.Errorf("WorkflowName = %q, want %q", got.WorkflowName, tt.status.WorkflowName)
			}
			if got.RunURL != tt.status.RunURL {
				t.Errorf("RunURL = %q, want %q", got.RunURL, tt.status.RunURL)
			}
		})
	}
}

func TestWorkflowRun_JSONUnmarshaling(t *testing.T) {
	// Test that we can unmarshal the expected gh run list JSON output
	tests := []struct {
		name     string
		jsonData string
		want     workflowRun
		wantErr  bool
	}{
		{
			name: "successful run",
			jsonData: `{
				"name": "CI",
				"status": "completed",
				"conclusion": "success",
				"url": "https://github.com/owner/repo/actions/runs/123",
				"headBranch": "main"
			}`,
			want: workflowRun{
				Name:       "CI",
				Status:     "completed",
				Conclusion: "success",
				URL:        "https://github.com/owner/repo/actions/runs/123",
				HeadBranch: "main",
			},
			wantErr: false,
		},
		{
			name: "failed run",
			jsonData: `{
				"name": "Tests",
				"status": "completed",
				"conclusion": "failure",
				"url": "https://github.com/owner/repo/actions/runs/124",
				"headBranch": "feature-branch"
			}`,
			want: workflowRun{
				Name:       "Tests",
				Status:     "completed",
				Conclusion: "failure",
				URL:        "https://github.com/owner/repo/actions/runs/124",
				HeadBranch: "feature-branch",
			},
			wantErr: false,
		},
		{
			name: "pending run",
			jsonData: `{
				"name": "Build",
				"status": "in_progress",
				"conclusion": "",
				"url": "https://github.com/owner/repo/actions/runs/125",
				"headBranch": "main"
			}`,
			want: workflowRun{
				Name:       "Build",
				Status:     "in_progress",
				Conclusion: "",
				URL:        "https://github.com/owner/repo/actions/runs/125",
				HeadBranch: "main",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got workflowRun
			err := json.Unmarshal([]byte(tt.jsonData), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Name != tt.want.Name {
					t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
				}
				if got.Status != tt.want.Status {
					t.Errorf("Status = %q, want %q", got.Status, tt.want.Status)
				}
				if got.Conclusion != tt.want.Conclusion {
					t.Errorf("Conclusion = %q, want %q", got.Conclusion, tt.want.Conclusion)
				}
				if got.URL != tt.want.URL {
					t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
				}
				if got.HeadBranch != tt.want.HeadBranch {
					t.Errorf("HeadBranch = %q, want %q", got.HeadBranch, tt.want.HeadBranch)
				}
			}
		})
	}
}

func TestCheckCIStatus_EmptyRuns(t *testing.T) {
	// This test would require mocking the gh CLI, which is beyond the scope
	// of this unit test. In a real implementation, you might use dependency
	// injection to provide a testable interface for executing commands.
	//
	// For now, we document that CheckCIStatus with no runs should return
	// CIStatusUnknown with appropriate failure info.
	t.Skip("Integration test - requires gh CLI and repository access")
}

func TestCheckCIStatus_ContextCancellation(t *testing.T) {
	// Test that context cancellation is respected
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := CheckCIStatus(ctx, "owner", "repo", "main")
	if err == nil {
		t.Error("CheckCIStatus() with cancelled context should return error")
	}
}

// Example test documentation for manual testing
func ExampleCheckCIStatus() {
	ctx := context.Background()
	status, err := CheckCIStatus(ctx, "anthropics", "claude-code", "main")
	if err != nil {
		// Handle error
		return
	}

	switch status.Status {
	case CIStatusSuccess:
		// CI is passing
	case CIStatusFailure:
		// CI is failing - check status.FailureInfo for details
	case CIStatusPending:
		// CI is still running
	case CIStatusUnknown:
		// Status could not be determined
	}
}
