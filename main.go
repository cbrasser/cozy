package main

import (
	"fmt"
	"os"

	"github.com/cbrasser/cozy/config"
	"github.com/cbrasser/cozy/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create TUI model
	model := tui.NewModel(cfg)

	// Start the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
