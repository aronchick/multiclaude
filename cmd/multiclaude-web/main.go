// multiclaude-web provides a read-only web dashboard for multiclaude.
// This is a FORK-ONLY feature that upstream explicitly rejects.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dlorenc/multiclaude/internal/dashboard"
)

const (
	defaultPort      = "8080"
	defaultBind      = "127.0.0.1"
	defaultStatePath = "~/.multiclaude/state.json"
)

func main() {
	var (
		port      = flag.String("port", defaultPort, "Port to listen on")
		bind      = flag.String("bind", defaultBind, "Address to bind to (use 0.0.0.0 for all interfaces)")
		statePath = flag.String("state", defaultStatePath, "Path to multiclaude state.json file")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "multiclaude-web - Read-only web dashboard for multiclaude\n\n")
		fmt.Fprintf(os.Stderr, "This is a FORK-ONLY feature. Upstream multiclaude explicitly rejects\n")
		fmt.Fprintf(os.Stderr, "web interfaces and dashboards. Use this only in the aronchick/multiclaude fork.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Start with defaults (localhost:8080)\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Custom port\n")
		fmt.Fprintf(os.Stderr, "  %s --port 3000\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Listen on all interfaces\n")
		fmt.Fprintf(os.Stderr, "  %s --bind 0.0.0.0\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Custom state file location\n")
		fmt.Fprintf(os.Stderr, "  %s --state /path/to/state.json\n\n", os.Args[0])
	}

	flag.Parse()

	// Expand home directory in state path
	expandedStatePath := *statePath
	if len(*statePath) > 0 && (*statePath)[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		expandedStatePath = filepath.Join(home, (*statePath)[1:])
	}

	// Verify state file exists
	if _, err := os.Stat(expandedStatePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: state file not found: %s\n", expandedStatePath)
		fmt.Fprintf(os.Stderr, "\nIs multiclaude initialized? Try:\n")
		fmt.Fprintf(os.Stderr, "  multiclaude init <github-url>\n")
		os.Exit(1)
	}

	// Create state reader
	reader, err := dashboard.NewStateReader([]string{expandedStatePath})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create state reader: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	// Create and start server
	server := dashboard.NewServer(reader)
	addr := fmt.Sprintf("%s:%s", *bind, *port)

	if err := server.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: server failed: %v\n", err)
		os.Exit(1)
	}
}
