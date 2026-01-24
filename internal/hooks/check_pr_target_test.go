package hooks

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckPRTargetScript tests the check-pr-target.sh script that prevents
// PRs from being created against upstream (dlorenc/multiclaude).
func TestCheckPRTargetScript(t *testing.T) {
	// Find the script path relative to the repo root
	// We need to go up from internal/hooks to find .multiclaude/scripts
	scriptPath := filepath.Join("..", "..", ".multiclaude", "scripts", "check-pr-target.sh")

	// Verify script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skipf("Script not found at %s - skipping test", scriptPath)
	}

	tests := []struct {
		name       string
		input      string
		wantCode   int
		wantStderr string // substring to check in stderr
	}{
		{
			name:     "non-gh command allowed",
			input:    `{"tool_input":{"command":"git status"}}`,
			wantCode: 0,
		},
		{
			name:     "gh pr create without repo flag allowed",
			input:    `{"tool_input":{"command":"gh pr create --title \"test\""}}`,
			wantCode: 0,
		},
		{
			name:     "gh pr create to fork allowed",
			input:    `{"tool_input":{"command":"gh pr create --repo aronchick/multiclaude --title \"test\""}}`,
			wantCode: 0,
		},
		{
			name:     "gh pr create to other repo allowed",
			input:    `{"tool_input":{"command":"gh pr create --repo someuser/somerepo --title \"test\""}}`,
			wantCode: 0,
		},
		{
			name:       "gh pr create to upstream blocked",
			input:      `{"tool_input":{"command":"gh pr create --repo dlorenc/multiclaude --title \"test\""}}`,
			wantCode:   2,
			wantStderr: "dlorenc/multiclaude",
		},
		{
			name:       "gh pr create to upstream with -R flag blocked",
			input:      `{"tool_input":{"command":"gh pr create -R dlorenc/multiclaude"}}`,
			wantCode:   2,
			wantStderr: "dlorenc/multiclaude",
		},
		{
			name:       "case insensitive blocking",
			input:      `{"tool_input":{"command":"gh pr create --repo DLORENC/MULTICLAUDE"}}`,
			wantCode:   2,
			wantStderr: "dlorenc/multiclaude",
		},
		{
			name:     "empty command allowed",
			input:    `{"tool_input":{}}`,
			wantCode: 0,
		},
		{
			name:     "empty input allowed",
			input:    `{}`,
			wantCode: 0,
		},
		{
			name:     "upstream mentioned in body text is allowed",
			input:    `{"tool_input":{"command":"gh pr create --body \"fix for dlorenc/multiclaude issue\""}}`,
			wantCode: 0,
			// Body text references to upstream (like issue numbers) should NOT block
			// Only explicit --repo targeting upstream should be blocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("bash", scriptPath)
			cmd.Stdin = strings.NewReader(tt.input)

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := 0
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else if err != nil {
				t.Fatalf("Failed to run script: %v", err)
			}

			if exitCode != tt.wantCode {
				t.Errorf("Exit code = %d, want %d. Stderr: %s", exitCode, tt.wantCode, stderr.String())
			}

			// Check stderr contains expected message if blocking
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("Stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

// TestHooksJSONValid verifies the hooks.json file is valid JSON with expected structure.
func TestHooksJSONValid(t *testing.T) {
	hooksPath := filepath.Join("..", "..", ".multiclaude", "hooks.json")

	data, err := os.ReadFile(hooksPath)
	if os.IsNotExist(err) {
		t.Skipf("hooks.json not found at %s - skipping test", hooksPath)
	}
	if err != nil {
		t.Fatalf("Failed to read hooks.json: %v", err)
	}

	// Check it contains expected keys
	content := string(data)
	expectedKeys := []string{
		`"hooks"`,
		`"PreToolUse"`,
		`"Bash"`,
		`"matcher"`,
		`"command"`,
		`check-pr-target.sh`,
	}

	for _, key := range expectedKeys {
		if !strings.Contains(content, key) {
			t.Errorf("hooks.json should contain %q", key)
		}
	}
}
