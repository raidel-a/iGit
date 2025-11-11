package git

import (
	"strings"
)

// Status returns the current git status
func (c *Client) Status() (GitStatus, error) {
	output, err := c.execGit("status", "--porcelain", "-u")
	if err != nil {
		return GitStatus{}, err
	}

	status := parseStatusOutput(output)

	// Get current branch
	branch, _ := c.CurrentBranch()
	status.Branch = branch

	// Check if clean
	status.IsClean = len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0

	return status, nil
}

// parseStatusOutput parses the output of `git status --porcelain`
// Format: XY PATH where X is index status, Y is work tree status
func parseStatusOutput(output string) GitStatus {
	var status GitStatus

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) < 4 {
			continue // Invalid line
		}

		x := line[0] // Index status
		y := line[1] // Work tree status
		filepath := line[3:]

		// Remove quotes if present
		filepath = strings.Trim(filepath, "\"")

		// Categorize based on status codes
		switch {
		case x != ' ' && x != '?':
			// Index has changes (staged)
			status.Staged = append(status.Staged, filepath)
		case y != ' ':
			// Work tree has changes (unstaged)
			status.Unstaged = append(status.Unstaged, filepath)
		case x == '?' && y == '?':
			// Untracked
			status.Untracked = append(status.Untracked, filepath)
		}
	}

	return status
}

// StagedCount returns the number of staged files
func (s GitStatus) StagedCount() int {
	return len(s.Staged)
}

// UnstagedCount returns the number of unstaged files
func (s GitStatus) UnstagedCount() int {
	return len(s.Unstaged)
}

// UntrackedCount returns the number of untracked files
func (s GitStatus) UntrackedCount() int {
	return len(s.Untracked)
}

// AllFiles returns all files organized by status
func (s GitStatus) AllFiles() []FileItem {
	var items []FileItem

	// Add unstaged files (marked with -)
	for _, f := range s.Unstaged {
		items = append(items, NewFileItem(f, StatusUnstaged))
	}

	// Add staged files (marked with +)
	for _, f := range s.Staged {
		items = append(items, NewFileItem(f, StatusStaged))
	}

	// Add untracked files (marked with ?)
	for _, f := range s.Untracked {
		items = append(items, NewFileItem(f, StatusUntracked))
	}

	return items
}

// FileItem represents a file in the git status
type FileItem struct {
	Path         string
	Status       FileStatus
	StatusSymbol string
	Selected     bool
}

// NewFileItem creates a new FileItem
func NewFileItem(path string, status FileStatus) FileItem {
	item := FileItem{
		Path:   path,
		Status: status,
	}

	switch status {
	case StatusStaged:
		item.StatusSymbol = "+"
	case StatusUnstaged:
		item.StatusSymbol = "-"
	case StatusUntracked:
		item.StatusSymbol = "?"
	}

	return item
}

// FilterValue implements list.Item interface for filtering
func (f FileItem) FilterValue() string {
	return f.Path
}

// Title implements list.Item interface
func (f FileItem) Title() string {
	return f.Path
}

// Description implements list.Item interface
func (f FileItem) Description() string {
	return ""
}
