package main

import (
	"bytes"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rai/interactive-git/git"
)

// isBinaryFile checks if a file contains binary data by looking for null bytes
// and other non-text indicators in the first 8KB of the file
func isBinaryFile(data []byte) bool {
	// Check for null bytes (strong indicator of binary)
	if bytes.Contains(data, []byte{0}) {
		return true
	}

	// Check for excessive non-printable characters (>30% non-text)
	if len(data) > 0 {
		nonTextCount := 0
		sampleSize := len(data)
		if sampleSize > 8192 {
			sampleSize = 8192 // Sample first 8KB
		}

		for i := 0; i < sampleSize; i++ {
			b := data[i]
			// Allow common text characters, whitespace, and UTF-8 markers
			if !bytes.Contains([]byte{'\t', '\n', '\r', ' '}, []byte{b}) &&
				(b < 32 || b > 126) && (b < 128 || b > 191) {
				nonTextCount++
			}
		}

		// If more than 30% of sampled bytes are non-text, likely binary
		if sampleSize > 0 && (nonTextCount*100/sampleSize) > 30 {
			return true
		}
	}

	return false
}

// Message types for commands

type gitStageMsg struct {
	files []string
	err   error
}

type gitUnstageMsg struct {
	files []string
	err   error
}

type gitRefreshMsg struct{}

type processingMsg struct {
	active bool
}

type gitDiffMsg struct {
	file    string
	content string
	err     error
}

type gitCommitMsg struct {
	success bool
	err     error
	message string
}

type gitAmendMsg struct {
	success bool
	err     error
	message string
}

// stageFilesCmd stages the given files
func (m *Model) stageFilesCmd(files []git.FileItem) tea.Cmd {
	return func() tea.Msg {
		var filePaths []string
		for _, f := range files {
			// Only stage unstaged and untracked files
			if f.Status != git.StatusStaged {
				filePaths = append(filePaths, f.Path)
			}
		}

		if len(filePaths) == 0 {
			return statusMsg{msg: "No files to stage"}
		}

		err := m.gitClient.Stage(filePaths...)
		if err != nil {
			return errorMsg{err: fmt.Sprintf("Failed to stage files: %v", err)}
		}

		return gitStageMsg{files: filePaths, err: nil}
	}
}

// unstageFilesCmd unstages the given files
func (m *Model) unstageFilesCmd(files []git.FileItem) tea.Cmd {
	return func() tea.Msg {
		var filePaths []string
		for _, f := range files {
			// Only unstage staged files
			if f.Status == git.StatusStaged {
				filePaths = append(filePaths, f.Path)
			}
		}

		if len(filePaths) == 0 {
			return statusMsg{msg: "No files to unstage"}
		}

		err := m.gitClient.Unstage(filePaths...)
		if err != nil {
			return errorMsg{err: fmt.Sprintf("Failed to unstage files: %v", err)}
		}

		return gitUnstageMsg{files: filePaths, err: nil}
	}
}

// refreshStatusCmd refreshes the git status
func (m *Model) refreshStatusCmd() tea.Cmd {
	return func() tea.Msg {
		status, err := m.gitClient.Status()
		if err != nil {
			return errorMsg{err: fmt.Sprintf("Failed to refresh status: %v", err)}
		}
		return gitStatusMsg{status: status}
	}
}

// toggleSelectionCmd handles toggling selection of files
func (m *Model) toggleSelectionCmd(files []git.FileItem) tea.Cmd {
	return func() tea.Msg {
		// Determine if we're staging or unstaging
		var staged []string
		var unstaged []string

		for _, f := range files {
			if f.Status == git.StatusStaged {
				staged = append(staged, f.Path)
			} else {
				unstaged = append(unstaged, f.Path)
			}
		}

		// If we have more unstaged than staged, stage them
		// Otherwise unstage them
		if len(unstaged) > len(staged) {
			err := m.gitClient.Stage(unstaged...)
			if err != nil {
				return errorMsg{err: fmt.Sprintf("Failed to stage files: %v", err)}
			}
			return statusMsg{msg: fmt.Sprintf("Staged %d file(s)", len(unstaged))}
		}

		// Unstage the staged files
		err := m.gitClient.Unstage(staged...)
		if err != nil {
			return errorMsg{err: fmt.Sprintf("Failed to unstage files: %v", err)}
		}
		return statusMsg{msg: fmt.Sprintf("Unstaged %d file(s)", len(staged))}
	}
}

// fetchDiffCmd fetches the diff for a file
func (m *Model) fetchDiffCmd(file git.FileItem) tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		if content, ok := m.diffCache[file.Path]; ok {
			return gitDiffMsg{file: file.Path, content: content, err: nil}
		}

		// Fetch diff based on file status
		var content string
		var err error

		switch file.Status {
		case git.StatusStaged:
			// Show staged diff
			content, err = m.gitClient.Diff(file.Path, true)
		case git.StatusUnstaged:
			// Show unstaged diff
			content, err = m.gitClient.Diff(file.Path, false)
		case git.StatusUntracked:
			// Show file contents for untracked files
			contentBytes, readErr := os.ReadFile(file.Path)
			if readErr != nil {
				return gitDiffMsg{file: file.Path, content: fmt.Sprintf("Error reading file: %v", readErr), err: nil}
			}
			// Check if file is binary
			if isBinaryFile(contentBytes) {
				content = "[BINARY] File cannot be previewed"
			} else {
				content = string(contentBytes)
			}
		}

		if err != nil {
			return gitDiffMsg{file: file.Path, content: fmt.Sprintf("Error loading diff: %v", err), err: nil}
		}

		// If no diff content (no changes), show the actual file content instead
		if content == "" && file.Status != git.StatusUntracked {
			// Try to read the file content instead
			contentBytes, readErr := os.ReadFile(file.Path)
			if readErr == nil {
				// Check if file is binary
				if isBinaryFile(contentBytes) {
					content = "[BINARY] File cannot be previewed"
				} else {
					content = string(contentBytes)
				}
			} else {
				content = fmt.Sprintf("(File has no changes)\n\nCould not read file: %v", readErr)
			}
		}

		// Cache the result
		m.diffCache[file.Path] = content

		return gitDiffMsg{file: file.Path, content: content, err: nil}
	}
}

// commitCmd creates a commit with the given message and optional date
func (m *Model) commitCmd(message, date string) tea.Cmd {
	return func() tea.Msg {
		// Validate date if provided
		var validatedDate string
		if date != "" {
			var err error
			validatedDate, err = git.ValidateCommitDate(date)
			if err != nil {
				return gitCommitMsg{success: false, err: err, message: ""}
			}
		}

		// Create the commit
		err := m.gitClient.Commit(message, validatedDate)
		if err != nil {
			return gitCommitMsg{success: false, err: err, message: ""}
		}

		return gitCommitMsg{success: true, err: nil, message: "[OK] Commit created successfully"}
	}
}

// amendMessageCmd amends the HEAD commit message
func (m *Model) amendMessageCmd(message string) tea.Cmd {
	return func() tea.Msg {
		if message == "" {
			return gitAmendMsg{success: false, err: fmt.Errorf("commit message cannot be empty"), message: ""}
		}

		err := m.gitClient.AmendMessage(message)
		if err != nil {
			return gitAmendMsg{success: false, err: err, message: ""}
		}

		return gitAmendMsg{success: true, err: nil, message: "[OK] Commit message amended successfully"}
	}
}

// softResetHeadCmd performs a soft reset of HEAD
func (m *Model) softResetHeadCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.gitClient.SoftResetHead()
		if err != nil {
			return gitAmendMsg{success: false, err: err, message: ""}
		}

		return gitAmendMsg{success: true, err: nil, message: "[OK] HEAD soft reset successfully. Changes staged."}
	}
}
