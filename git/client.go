package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Client wraps git command execution
type Client struct {
	workDir string
	timeout time.Duration
}

// NewClient creates a new git client for the given directory
func NewClient(dir string) (*Client, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = absDir
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", absDir)
	}

	return &Client{
		workDir: absDir,
		timeout: 10 * time.Second,
	}, nil
}

// execGit executes a git command and returns its output
func (c *Client) execGit(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\n%s", args[0], err, string(output))
	}

	return string(output), nil
}

// CurrentBranch returns the name of the current branch
func (c *Client) CurrentBranch() (string, error) {
	output, err := c.execGit("branch", "--show-current")
	if err != nil {
		return "", err
	}
	// Remove trailing newline
	if len(output) > 0 && output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}
	return output, nil
}

// WorkDir returns the working directory of the git repository
func (c *Client) WorkDir() string {
	return c.workDir
}

// IsRepo checks if a directory is a git repository
func IsRepo(dir string) bool {
	_, err := NewClient(dir)
	return err == nil
}

// GetCurrentWorkingDir returns the current working directory
func GetCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
