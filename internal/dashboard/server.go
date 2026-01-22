package dashboard

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web
var webFS embed.FS

// Server provides the HTTP server for the dashboard
type Server struct {
	handler *APIHandler
	mux     *http.ServeMux
}

// NewServer creates a new dashboard server
func NewServer(reader *StateReader) *Server {
	handler := NewAPIHandler(reader)
	mux := http.NewServeMux()

	server := &Server{
		handler: handler,
		mux:     mux,
	}

	server.registerRoutes()
	return server
}

// registerRoutes sets up all HTTP routes
func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("/api/state", s.handler.HandleState)
	s.mux.HandleFunc("/api/repos", s.handleReposRoute)
	s.mux.HandleFunc("/api/events", s.handler.HandleEvents)

	// Static files - serve embedded web/ directory
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		// Fallback to empty filesystem if embedding fails
		fmt.Printf("Warning: failed to load embedded web files: %v\n", err)
		webRoot = nil
	}

	if webRoot != nil {
		s.mux.Handle("/", http.FileServer(http.FS(webRoot)))
	} else {
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Web UI not available", http.StatusNotFound)
		})
	}
}

// handleReposRoute routes repository-related requests
func (s *Server) handleReposRoute(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/repos" {
		// List all repos
		s.handler.HandleRepos(w, r)
		return
	}

	if strings.HasSuffix(path, "/agents") {
		// Get agents for a repo
		s.handler.HandleAgents(w, r)
		return
	}

	if strings.HasSuffix(path, "/history") {
		// Get history for a repo
		s.handler.HandleHistory(w, r)
		return
	}

	// Get specific repo details
	s.handler.HandleRepo(w, r)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start starts the HTTP server on the given address
func (s *Server) Start(addr string) error {
	fmt.Printf("Starting multiclaude web dashboard on http://%s\n", addr)
	fmt.Printf("  API endpoints:\n")
	fmt.Printf("    GET /api/state          - Full aggregated state\n")
	fmt.Printf("    GET /api/repos          - List all repositories\n")
	fmt.Printf("    GET /api/repos/{name}   - Repository details\n")
	fmt.Printf("    GET /api/repos/{name}/agents  - Repository agents\n")
	fmt.Printf("    GET /api/repos/{name}/history - Task history\n")
	fmt.Printf("    GET /api/events         - Server-Sent Events (live updates)\n")
	fmt.Printf("\n")
	fmt.Printf("  Web UI: http://%s\n", addr)
	fmt.Printf("\n")

	return http.ListenAndServe(addr, s)
}
