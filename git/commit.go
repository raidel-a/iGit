package git

import (
	"fmt"
	"strings"
	"time"
)

// Commit creates a new commit with the given message and optional date
func (c *Client) Commit(message, date string) error {
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	args := []string{"commit", "-m", message}

	// Add date if provided
	if date != "" {
		args = append(args, "--date", date)
	}

	_, err := c.execGit(args...)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// AmendMessage amends the HEAD commit message
func (c *Client) AmendMessage(message string) error {
	if message == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	_, err := c.execGit("commit", "--amend", "-m", message)
	if err != nil {
		return fmt.Errorf("failed to amend commit: %w", err)
	}

	return nil
}

// GetHeadCommitInfo returns information about the HEAD commit
func (c *Client) GetHeadCommitInfo() (*CommitInfo, error) {
	// Get short hash
	shortHash, err := c.execGit("rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD hash: %w", err)
	}
	shortHash = strings.TrimSpace(shortHash)

	// Get full hash
	fullHash, err := c.execGit("rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD full hash: %w", err)
	}
	fullHash = strings.TrimSpace(fullHash)

	// Get commit message (subject line only)
	message, err := c.execGit("log", "-1", "--pretty=format:%s", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit message: %w", err)
	}
	message = strings.TrimSpace(message)

	// Get author
	author, err := c.execGit("log", "-1", "--pretty=format:%an", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	author = strings.TrimSpace(author)

	// Get date
	date, err := c.execGit("log", "-1", "--pretty=format:%ar", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get date: %w", err)
	}
	date = strings.TrimSpace(date)

	// Check if pushed
	branch, err := c.CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	isPushed := false
	remoteBranch := fmt.Sprintf("origin/%s", branch)
	output, err := c.execGit("branch", "-r", "--contains", "HEAD")
	if err == nil && strings.Contains(output, remoteBranch) {
		isPushed = true
	}

	return &CommitInfo{
		Hash:      fullHash,
		ShortHash: shortHash,
		Message:   message,
		Author:    author,
		Date:      date,
		IsPushed:  isPushed,
	}, nil
}

// ValidateCommitDate validates and formats a commit date
func ValidateCommitDate(dateStr string) (string, error) {
	if dateStr == "" || strings.ToLower(dateStr) == "now" {
		return "", nil // Use current time
	}

	// Try parsing various formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Format("2006-01-02 15:04:05"), nil
		}
	}

	return "", fmt.Errorf("invalid date format: %s (use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", dateStr)
}

// SoftResetHead resets HEAD to HEAD~1 but keeps changes staged
func (c *Client) SoftResetHead() error {
	_, err := c.execGit("reset", "--soft", "HEAD~1")
	if err != nil {
		return fmt.Errorf("failed to soft reset HEAD: %w", err)
	}
	return nil
}

// ShowCommit shows the full commit details
func (c *Client) ShowCommit(ref string) (string, error) {
	output, err := c.execGit("show", "--color=always", ref)
	if err != nil {
		return "", fmt.Errorf("failed to show commit: %w", err)
	}
	return output, nil
}
