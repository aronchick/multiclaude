package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/dlorenc/multiclaude/internal/state"
)

// APIHandler provides HTTP handlers for the dashboard API
type APIHandler struct {
	reader  *StateReader
	mu      sync.RWMutex
	clients map[chan []byte]bool
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(reader *StateReader) *APIHandler {
	handler := &APIHandler{
		reader:  reader,
		clients: make(map[chan []byte]bool),
	}

	// Register for state changes to broadcast to SSE clients
	reader.Watch(func() {
		handler.broadcastStateChange()
	})

	return handler
}

// HandleState returns the full aggregated state
func (h *APIHandler) HandleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agg := h.reader.GetAggregatedState()
	h.writeJSON(w, agg)
}

// HandleRepos returns a list of all repositories across all machines
func (h *APIHandler) HandleRepos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agg := h.reader.GetAggregatedState()

	type RepoInfo struct {
		Name       string `json:"name"`
		Machine    string `json:"machine"`
		GithubURL  string `json:"github_url"`
		AgentCount int    `json:"agent_count"`
	}

	repos := []RepoInfo{}
	for machineName, machine := range agg.Machines {
		for repoName, repo := range machine.Repos {
			repos = append(repos, RepoInfo{
				Name:       repoName,
				Machine:    machineName,
				GithubURL:  repo.GithubURL,
				AgentCount: len(repo.Agents),
			})
		}
	}

	h.writeJSON(w, repos)
}

// HandleRepo returns details for a specific repository
func (h *APIHandler) HandleRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract repo name from URL path
	// Expected: /api/repos/{repoName}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	repoName := parts[3]

	agg := h.reader.GetAggregatedState()

	// Find the repo in any machine
	for _, machine := range agg.Machines {
		if repo, ok := machine.Repos[repoName]; ok {
			h.writeJSON(w, repo)
			return
		}
	}

	http.Error(w, "Repository not found", http.StatusNotFound)
}

// HandleAgents returns agents for a specific repository
func (h *APIHandler) HandleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract repo name from URL path
	// Expected: /api/repos/{repoName}/agents
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	repoName := parts[3]

	agg := h.reader.GetAggregatedState()

	// Find the repo in any machine
	for _, machine := range agg.Machines {
		if repo, ok := machine.Repos[repoName]; ok {
			h.writeJSON(w, repo.Agents)
			return
		}
	}

	http.Error(w, "Repository not found", http.StatusNotFound)
}

// HandleHistory returns task history for a specific repository
func (h *APIHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract repo name from URL path
	// Expected: /api/repos/{repoName}/history
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	repoName := parts[3]

	// Parse optional limit parameter
	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	agg := h.reader.GetAggregatedState()

	// Find the repo in any machine
	for _, machine := range agg.Machines {
		if repo, ok := machine.Repos[repoName]; ok {
			history := repo.TaskHistory
			if history == nil {
				history = []state.TaskHistoryEntry{}
			}

			// Reverse to get most recent first
			result := make([]state.TaskHistoryEntry, len(history))
			for i, entry := range history {
				result[len(history)-1-i] = entry
			}

			// Apply limit if specified
			if limit > 0 && len(result) > limit {
				result = result[:limit]
			}

			h.writeJSON(w, result)
			return
		}
	}

	http.Error(w, "Repository not found", http.StatusNotFound)
}

// HandleEvents provides Server-Sent Events for live updates
func (h *APIHandler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan []byte, 10)

	h.mu.Lock()
	h.clients[clientChan] = true
	h.mu.Unlock()

	// Remove client when done
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientChan)
		close(clientChan)
		h.mu.Unlock()
	}()

	// Send initial state
	agg := h.reader.GetAggregatedState()
	data, err := json.Marshal(agg)
	if err == nil {
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	// Listen for updates or client disconnect
	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

// broadcastStateChange sends state updates to all SSE clients
func (h *APIHandler) broadcastStateChange() {
	agg := h.reader.GetAggregatedState()
	data, err := json.Marshal(agg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client <- data:
		default:
			// Client buffer full, skip this update
		}
	}
}

// writeJSON writes a JSON response
func (h *APIHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}
