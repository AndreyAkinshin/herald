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

// stripPreamble removes a leading H1 heading (e.g. "# Release Notes for v1.2.0")
// that Claude sometimes adds despite being told not to.
// Preserves everything else including introductory paragraphs before ## sections.
func stripPreamble(output string) string {
	lines := strings.Split(output, "\n")

	// Find the first non-blank line
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	if start >= len(lines) {
		return output
	}

	// Only strip if the first non-blank line is an H1
	if headingLevel(lines[start]) != 1 {
		return output
	}

	// Skip the H1 and any following blank lines
	j := start + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}

	if j < len(lines) {
		return strings.Join(lines[j:], "\n")
	}

	return ""
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
