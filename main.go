package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rai/interactive-git/git"
)

func main() {
	// Check if we're in a git repository
	if !git.IsRepo(".") {
		fmt.Fprintln(os.Stderr, "Error: Not in a git repository")
		os.Exit(1)
	}

	// Create the initial model
	m := NewModel()

	// Create a Bubble Tea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
