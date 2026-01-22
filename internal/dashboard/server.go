package dashboard

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dlorenc/multiclaude/internal/messages"
	"github.com/dlorenc/multiclaude/internal/state"
)

//go:embed templates/index.html
var indexHTML string

// Server represents the dashboard HTTP server
type Server struct {
	state      *state.State
	msgManager *messages.Manager
	server     *http.Server
}

// NewServer creates a new dashboard server
func NewServer(st *state.State, msgManager *messages.Manager) *Server {
	return &Server{
		state:      st,
		msgManager: msgManager,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Logging middleware
	loggedMux := loggingMiddleware(mux)

	// HTML page
	mux.HandleFunc("/", s.handleIndex)

	// API endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/repos", s.handleRepos)
	mux.HandleFunc("/api/repos/", s.handleRepoEndpoints)

	s.server = &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: loggedMux,
	}

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Dashboard server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return s.Stop()
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// handleIndex serves the HTML dashboard
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// handleStatus returns overall system status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	repos := s.state.GetAllRepos()

	totalAgents := 0
	activeAgents := 0
	idleAgents := 0

	for _, repo := range repos {
		for _, agent := range repo.Agents {
			totalAgents++
			if agent.Task != "" {
				activeAgents++
			} else {
				idleAgents++
			}
		}
	}

	status := map[string]interface{}{
		"total_agents":  totalAgents,
		"active_agents": activeAgents,
		"idle_agents":   idleAgents,
		"repos":         len(repos),
	}

	writeJSON(w, status)
}

// handleRepos returns list of repositories
func (s *Server) handleRepos(w http.ResponseWriter, r *http.Request) {
	repoNames := s.state.ListRepos()
	repos := make([]map[string]interface{}, 0, len(repoNames))

	for _, name := range repoNames {
		repo, exists := s.state.GetRepo(name)
		if !exists {
			continue
		}

		repos = append(repos, map[string]interface{}{
			"name":         name,
			"github_url":   repo.GithubURL,
			"tmux_session": repo.TmuxSession,
			"agent_count":  len(repo.Agents),
		})
	}

	writeJSON(w, repos)
}

// handleRepoEndpoints routes repo-specific endpoints
func (s *Server) handleRepoEndpoints(w http.ResponseWriter, r *http.Request) {
	// Parse path: /api/repos/{repo}/{endpoint}
	path := strings.TrimPrefix(r.URL.Path, "/api/repos/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	repoName := parts[0]
	endpoint := parts[1]

	switch endpoint {
	case "agents":
		s.handleRepoAgents(w, r, repoName)
	case "messages":
		s.handleRepoMessages(w, r, repoName)
	case "history":
		s.handleRepoHistory(w, r, repoName)
	case "activity":
		s.handleRepoActivity(w, r, repoName)
	default:
		http.NotFound(w, r)
	}
}

// handleRepoAgents returns agents for a repository
func (s *Server) handleRepoAgents(w http.ResponseWriter, r *http.Request, repoName string) {
	repo, exists := s.state.GetRepo(repoName)
	if !exists {
		http.NotFound(w, r)
		return
	}

	agents := make([]map[string]interface{}, 0, len(repo.Agents))
	now := time.Now()

	for name, agent := range repo.Agents {
		status := "idle"
		if agent.Task != "" {
			status = "active"
		} else if !agent.LastNudge.IsZero() && now.Sub(agent.LastNudge) > 10*time.Minute {
			status = "stuck"
		}

		agents = append(agents, map[string]interface{}{
			"name":          name,
			"type":          agent.Type,
			"status":        status,
			"task":          agent.Task,
			"worktree_path": agent.WorktreePath,
			"tmux_window":   agent.TmuxWindow,
			"created_at":    agent.CreatedAt,
			"last_nudge":    agent.LastNudge,
		})
	}

	writeJSON(w, agents)
}

// handleRepoMessages returns messages for a repository
func (s *Server) handleRepoMessages(w http.ResponseWriter, r *http.Request, repoName string) {
	repo, exists := s.state.GetRepo(repoName)
	if !exists {
		http.NotFound(w, r)
		return
	}

	allMessages := make([]map[string]interface{}, 0)

	for agentName := range repo.Agents {
		msgs, err := s.msgManager.List(repoName, agentName)
		if err != nil {
			continue
		}

		for _, msg := range msgs {
			allMessages = append(allMessages, map[string]interface{}{
				"id":        msg.ID,
				"from":      msg.From,
				"to":        msg.To,
				"timestamp": msg.Timestamp,
				"body":      msg.Body,
				"status":    msg.Status,
			})
		}
	}

	writeJSON(w, allMessages)
}

// handleRepoHistory returns task history for a repository
func (s *Server) handleRepoHistory(w http.ResponseWriter, r *http.Request, repoName string) {
	history, err := s.state.GetTaskHistory(repoName, 50)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, history)
}

// handleRepoActivity returns recent activity for a repository
func (s *Server) handleRepoActivity(w http.ResponseWriter, r *http.Request, repoName string) {
	repo, exists := s.state.GetRepo(repoName)
	if !exists {
		http.NotFound(w, r)
		return
	}

	activity := make([]map[string]interface{}, 0)

	// Add recent task history
	history, _ := s.state.GetTaskHistory(repoName, 10)
	for _, entry := range history {
		activity = append(activity, map[string]interface{}{
			"type":      "task",
			"timestamp": entry.CompletedAt,
			"agent":     entry.Name,
			"message":   fmt.Sprintf("Completed: %s", entry.Task),
			"status":    entry.Status,
		})
	}

	// Add recent messages
	for agentName := range repo.Agents {
		msgs, err := s.msgManager.List(repoName, agentName)
		if err != nil {
			continue
		}

		// Get last 5 messages per agent
		count := 0
		for i := len(msgs) - 1; i >= 0 && count < 5; i-- {
			msg := msgs[i]
			activity = append(activity, map[string]interface{}{
				"type":      "message",
				"timestamp": msg.Timestamp,
				"agent":     msg.To,
				"message":   fmt.Sprintf("Message from %s: %s", msg.From, msg.Body),
				"status":    msg.Status,
			})
			count++
		}
	}

	writeJSON(w, activity)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


