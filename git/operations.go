package git

import (
	"fmt"
	"strings"
)

// Stage stages files for commit
func (c *Client) Stage(files ...string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"add", "--"}, files...)
	_, err := c.execGit(args...)
	if err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	return nil
}

// Unstage unstages files
func (c *Client) Unstage(files ...string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"reset", "HEAD", "--"}, files...)
	_, err := c.execGit(args...)
	if err != nil {
		return fmt.Errorf("failed to unstage files: %w", err)
	}

	return nil
}

// Diff returns the diff for a file
func (c *Client) Diff(file string, staged bool) (string, error) {
	args := []string{"diff", "--color=always"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", file)

	output, err := c.execGit(args...)
	if err != nil {
		// git diff returns exit code 1 if there are differences, which is not an error
		if strings.Contains(err.Error(), "exit status 1") {
			return output, nil
		}
		return "", err
	}

	return output, nil
}

// StageAll stages all unstaged and untracked files
func (c *Client) StageAll() error {
	_, err := c.execGit("add", ".")
	if err != nil {
		return fmt.Errorf("failed to stage all files: %w", err)
	}
	return nil
}

// UnstageAll unstages all staged files
func (c *Client) UnstageAll() error {
	_, err := c.execGit("reset", "HEAD")
	if err != nil {
		return fmt.Errorf("failed to unstage all files: %w", err)
	}
	return nil
}
