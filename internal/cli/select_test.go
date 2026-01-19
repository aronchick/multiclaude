package cli

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/dlorenc/multiclaude/internal/socket"
	"github.com/dlorenc/multiclaude/pkg/config"
)

func TestIsInteractive(t *testing.T) {
	// This test verifies that isInteractive can be called without panicking.
	// The actual result depends on the test environment (TTY vs not).
	result := isInteractive()
	// Just verify it returns a boolean without error
	_ = result
}

func TestResolveRepo_WithFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	paths := &config.Paths{
		Root:         tmpDir,
		ReposDir:     filepath.Join(tmpDir, "repos"),
		WorktreesDir: filepath.Join(tmpDir, "wts"),
		DaemonSock:   filepath.Join(tmpDir, "daemon.sock"),
	}
	cli := NewWithPaths(paths, "claude")

	// When --repo flag is provided, it should be used directly
	flags := map[string]string{"repo": "test-repo"}
	repo, err := cli.resolveRepo(flags)
	if err != nil {
		t.Errorf("resolveRepo with flag should not error: %v", err)
	}
	if repo != "test-repo" {
		t.Errorf("resolveRepo = %q, want %q", repo, "test-repo")
	}
}

func TestResolveRepo_InferFromCwd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	paths := &config.Paths{
		Root:         tmpDir,
		ReposDir:     filepath.Join(tmpDir, "repos"),
		WorktreesDir: filepath.Join(tmpDir, "wts"),
		DaemonSock:   filepath.Join(tmpDir, "daemon.sock"),
	}

	// Create repo directory structure
	repoWorktreeDir := filepath.Join(paths.WorktreesDir, "my-repo", "my-agent")
	if err := os.MkdirAll(repoWorktreeDir, 0755); err != nil {
		t.Fatalf("Failed to create repo worktree dir: %v", err)
	}

	// Change to the worktree directory
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	if err := os.Chdir(repoWorktreeDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cli := NewWithPaths(paths, "claude")

	// With empty flags, it should infer from CWD
	flags := map[string]string{}
	repo, err := cli.resolveRepo(flags)
	if err != nil {
		t.Errorf("resolveRepo should infer from CWD: %v", err)
	}
	if repo != "my-repo" {
		t.Errorf("resolveRepo = %q, want %q", repo, "my-repo")
	}
}

func TestSelectRepo_SingleRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")

	paths := &config.Paths{
		Root:         tmpDir,
		ReposDir:     filepath.Join(tmpDir, "repos"),
		WorktreesDir: filepath.Join(tmpDir, "wts"),
		DaemonSock:   sockPath,
	}
	cli := NewWithPaths(paths, "claude")

	// Start mock server that returns a single repo
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create test socket: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read request
		decoder := json.NewDecoder(conn)
		var req socket.Request
		decoder.Decode(&req)

		// Send response with single repo
		resp := socket.Response{
			Success: true,
			Data: []interface{}{
				map[string]interface{}{"name": "only-repo"},
			},
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(resp)
	}()

	// With single repo, selectRepo should return it without prompting
	repo, err := cli.selectRepo()
	if err != nil {
		t.Errorf("selectRepo with single repo should not error: %v", err)
	}
	if repo != "only-repo" {
		t.Errorf("selectRepo = %q, want %q", repo, "only-repo")
	}
}

func TestSelectWorkspace_SingleWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")

	paths := &config.Paths{
		Root:         tmpDir,
		ReposDir:     filepath.Join(tmpDir, "repos"),
		WorktreesDir: filepath.Join(tmpDir, "wts"),
		DaemonSock:   sockPath,
	}
	cli := NewWithPaths(paths, "claude")

	// Start mock server that returns a single workspace
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create test socket: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read request
		decoder := json.NewDecoder(conn)
		var req socket.Request
		decoder.Decode(&req)

		// Send response with single workspace
		resp := socket.Response{
			Success: true,
			Data: []interface{}{
				map[string]interface{}{"name": "default", "type": "workspace"},
			},
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(resp)
	}()

	// With single workspace, selectWorkspace should return it without prompting
	ws, err := cli.selectWorkspace("test-repo")
	if err != nil {
		t.Errorf("selectWorkspace with single workspace should not error: %v", err)
	}
	if ws != "default" {
		t.Errorf("selectWorkspace = %q, want %q", ws, "default")
	}
}

func TestSelectRepo_NoRepos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")

	paths := &config.Paths{
		Root:         tmpDir,
		ReposDir:     filepath.Join(tmpDir, "repos"),
		WorktreesDir: filepath.Join(tmpDir, "wts"),
		DaemonSock:   sockPath,
	}
	cli := NewWithPaths(paths, "claude")

	// Start mock server that returns no repos
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create test socket: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		decoder := json.NewDecoder(conn)
		var req socket.Request
		decoder.Decode(&req)

		resp := socket.Response{
			Success: true,
			Data:    []interface{}{},
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(resp)
	}()

	// With no repos, selectRepo should error
	_, err = cli.selectRepo()
	if err == nil {
		t.Error("selectRepo with no repos should error")
	}
}

func TestSelectWorkspace_NoWorkspaces(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sockPath := filepath.Join(tmpDir, "test.sock")

	paths := &config.Paths{
		Root:         tmpDir,
		ReposDir:     filepath.Join(tmpDir, "repos"),
		WorktreesDir: filepath.Join(tmpDir, "wts"),
		DaemonSock:   sockPath,
	}
	cli := NewWithPaths(paths, "claude")

	// Start mock server that returns no workspaces
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create test socket: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		decoder := json.NewDecoder(conn)
		var req socket.Request
		decoder.Decode(&req)

		resp := socket.Response{
			Success: true,
			Data:    []interface{}{},
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(resp)
	}()

	// With no workspaces, selectWorkspace should error
	_, err = cli.selectWorkspace("test-repo")
	if err == nil {
		t.Error("selectWorkspace with no workspaces should error")
	}
}
