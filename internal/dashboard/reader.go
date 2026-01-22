// Package dashboard provides web dashboard functionality for multiclaude.
// This is a FORK-ONLY feature that upstream explicitly rejects.
package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dlorenc/multiclaude/internal/state"
	"github.com/fsnotify/fsnotify"
)

// StateReader reads and watches multiclaude state files
type StateReader struct {
	paths    []string
	watcher  *fsnotify.Watcher
	mu       sync.RWMutex
	states   map[string]*state.State
	onChange func()
}

// NewStateReader creates a new state reader for the given paths
func NewStateReader(paths []string) (*StateReader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	reader := &StateReader{
		paths:   paths,
		watcher: watcher,
		states:  make(map[string]*state.State),
	}

	// Initial read of all state files
	for _, path := range paths {
		if err := reader.readState(path); err != nil {
			// Log error but continue with other paths
			fmt.Fprintf(os.Stderr, "Warning: failed to read state from %s: %v\n", path, err)
		}

		// Watch the state file for changes
		if err := watcher.Add(path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to watch %s: %v\n", path, err)
		}
	}

	return reader, nil
}

// readState reads a state file from the given path
func (r *StateReader) readState(path string) error {
	// Expand home directory if present
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	s, err := state.Load(path)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.states[path] = s
	r.mu.Unlock()

	return nil
}

// Watch starts watching for state file changes and calls onChange when they occur
func (r *StateReader) Watch(onChange func()) {
	r.onChange = onChange
	go r.watchLoop()
}

// watchLoop handles file system events
func (r *StateReader) watchLoop() {
	for {
		select {
		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// Only care about write events
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Re-read the state file
				if err := r.readState(event.Name); err != nil {
					fmt.Fprintf(os.Stderr, "Error reading state after change: %v\n", err)
				}

				// Notify onChange callback
				if r.onChange != nil {
					r.onChange()
				}
			}

		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		}
	}
}

// GetAggregatedState returns a combined view of all state files
func (r *StateReader) GetAggregatedState() *AggregatedState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agg := &AggregatedState{
		Machines:  make(map[string]*MachineState),
		Timestamp: time.Now(),
	}

	for path, s := range r.states {
		// Use the state file path as the machine identifier
		machineName := filepath.Dir(path)

		machineState := &MachineState{
			Path:        path,
			Repos:       s.Repos,
			CurrentRepo: s.CurrentRepo,
		}

		agg.Machines[machineName] = machineState
	}

	return agg
}

// GetState returns the state for a specific path
func (r *StateReader) GetState(path string) (*state.State, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.states[path]
	return s, ok
}

// Close stops watching and cleans up resources
func (r *StateReader) Close() error {
	return r.watcher.Close()
}

// AggregatedState represents the combined state from multiple machines
type AggregatedState struct {
	Machines  map[string]*MachineState `json:"machines"`
	Timestamp time.Time                `json:"timestamp"`
}

// MachineState represents the state from a single machine
type MachineState struct {
	Path        string                      `json:"path"`
	Repos       map[string]*state.Repository `json:"repos"`
	CurrentRepo string                      `json:"current_repo,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for better API output
func (a *AggregatedState) MarshalJSON() ([]byte, error) {
	type Alias AggregatedState
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(a),
		Timestamp: a.Timestamp.Format(time.RFC3339),
	})
}
