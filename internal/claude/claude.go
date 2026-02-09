// Package claude provides Claude CLI wrapper functions.
package claude

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/AndreyAkinshin/herald/internal/errors"
)

// CheckClaudeAvailable verifies the claude CLI is installed.
func CheckClaudeAvailable() error {
	cmd := exec.Command("claude", "--version")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.Environment("claude CLI not available", err)
	}

	return nil
}

// GenerateNotes invokes Claude with the given prompt and returns the generated notes.
// If model is non-empty, it is passed via --model to the claude CLI.
func GenerateNotes(prompt string, model string) (string, error) {
	args := []string{"-p"}
	if model != "" {
		args = append(args, "--model", model)
	}

	cmd := exec.Command("claude", args...)
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.Runtime("failed to generate notes with Claude", err)
	}

	return stripPreamble(stdout.String()), nil
}

// stripPreamble removes any content before the first markdown heading,
// and also strips the heading line itself along with any following blank lines.
func stripPreamble(output string) string {
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		level := headingLevel(line)
		if level == 0 {
			continue
		}

		if level == 1 {
			// Skip the heading line and any following blank lines
			j := i + 1
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j < len(lines) {
				return strings.Join(lines[j:], "\n")
			}
			return ""
		}

		// H2+: keep from here
		return strings.Join(lines[i:], "\n")
	}

	return output
}

// headingLevel returns the markdown heading level (1-6) or 0 if not a heading.
func headingLevel(line string) int {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}

	if i > 0 && i <= 6 && i < len(line) && line[i] == ' ' {
		return i
	}

	return 0
}
