// Package git provides git repository operations.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AndreyAkinshin/herald/internal/errors"
)

const commitDelimiter = "---HERALD-COMMIT---"

// FindRepoRoot walks up from the current directory to find the git repository root.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Environment("failed to get current directory", err)
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.Environment("not a git repository (or any parent)", nil)
		}

		dir = parent
	}
}

// TagExists checks if a tag exists in the repository.
func TagExists(tag string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", tag)

	return cmd.Run() == nil
}

// GetCommitDetails returns detailed commit information between two refs.
// Each commit includes: full hash, full message (header + body), and list of changed files.
func GetCommitDetails(from, to string) (string, error) {
	return getCommitDetails(from + ".." + to)
}

// GetCommitDetailsFromRoot returns detailed commit information from root to the given ref.
func GetCommitDetailsFromRoot(to string) (string, error) {
	return getCommitDetails(to)
}

func getCommitDetails(revRange string) (string, error) {
	format := fmt.Sprintf("%s%%n%%H%%n%%B%%n%s-STAT", commitDelimiter, commitDelimiter)
	cmd := exec.Command("git", "log", "--stat", "--format="+format, revRange)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Runtime("failed to get commit details", err)
	}

	return parseCommitDetails(stdout.String(), commitDelimiter)
}

func parseCommitDetails(output, delim string) (string, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "(no commits)", nil
	}

	// Split on the commit delimiter to get individual commit blocks
	startMarker := delim
	blocks := strings.Split(output, startMarker)

	var result strings.Builder
	first := true

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" || block == "-STAT" {
			continue
		}

		// Each block has format:
		// <hash>
		// <message body>
		//
		// <delim>-STAT
		// <stat lines>

		// Split on the stat marker
		statMarker := delim + "-STAT"
		parts := strings.SplitN(block, statMarker, 2)

		headerPart := strings.TrimSpace(parts[0])
		if headerPart == "" {
			continue
		}

		// First line is the hash, rest is the message
		lines := strings.SplitN(headerPart, "\n", 2)
		hash := strings.TrimSpace(lines[0])

		if hash == "" {
			continue
		}

		var message string
		if len(lines) > 1 {
			message = strings.TrimSpace(lines[1])
		}

		var stat string
		if len(parts) > 1 {
			stat = strings.TrimSpace(parts[1])
		}

		if !first {
			result.WriteString("\n---\n\n")
		}

		first = false

		fmt.Fprintf(&result, "Commit: %s\n\n", hash)
		result.WriteString(message)
		result.WriteString("\n\nChanged files:\n")

		if stat != "" {
			result.WriteString(stat)
			result.WriteString("\n")
		}
	}

	if first {
		return "(no commits)", nil
	}

	return result.String(), nil
}
