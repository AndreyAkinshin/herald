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
		msg := "claude CLI not available"
		if s := strings.TrimSpace(stderr.String()); s != "" {
			msg += ": " + s
		}

		return errors.Environment(msg, err)
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
		msg := "failed to generate notes with Claude"
		if s := strings.TrimSpace(stderr.String()); s != "" {
			msg += ": " + s
		}

		return "", errors.Runtime(msg, err)
	}

	return stripPreamble(stdout.String()), nil
}

// stripPreamble removes unwanted leading content that Claude sometimes adds:
//   - A leading H1 heading (e.g. "# Release Notes for v1.2.0")
//   - Conversational preamble ending with ":" (e.g. "Here are the release notes:")
//     optionally followed by a thematic break ("---")
//
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

	first := lines[start]

	// Strip leading H1 heading
	if headingLevel(first) == 1 {
		return joinAfterSkippingBlanks(lines, start+1)
	}

	// Strip conversational preamble: a non-heading line ending with ":"
	// optionally followed by a "---" thematic break
	if isConversationalPreamble(first) {
		j := skipBlanks(lines, start+1)

		if j < len(lines) && isThematicBreak(lines[j]) {
			j = skipBlanks(lines, j+1)
		}

		if j < len(lines) {
			return strings.Join(lines[j:], "\n")
		}

		return ""
	}

	return output
}

// joinAfterSkippingBlanks skips blank lines starting at index and joins the rest.
func joinAfterSkippingBlanks(lines []string, from int) string {
	j := skipBlanks(lines, from)
	if j < len(lines) {
		return strings.Join(lines[j:], "\n")
	}

	return ""
}

// skipBlanks returns the index of the first non-blank line at or after from.
func skipBlanks(lines []string, from int) int {
	for from < len(lines) && strings.TrimSpace(lines[from]) == "" {
		from++
	}

	return from
}

// isConversationalPreamble returns true if the line looks like LLM preamble
// (e.g. "Here are the release notes:") rather than release notes content.
func isConversationalPreamble(line string) bool {
	trimmed := strings.TrimSpace(line)

	return len(trimmed) > 0 &&
		trimmed[len(trimmed)-1] == ':' &&
		headingLevel(line) == 0
}

// isThematicBreak returns true for markdown thematic breaks (---, ***, ___).
func isThematicBreak(line string) bool {
	trimmed := strings.TrimSpace(line)

	return trimmed == "---" || trimmed == "***" || trimmed == "___"
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
