// Package prompt provides prompt template generation for Claude.
package prompt

import (
	"bytes"
	_ "embed"
	"text/template"
)

type templateData struct {
	TargetTag     string
	PrevTag       string
	CommitDetails string
	Instructions  string
}

//go:embed prompt.tmpl
var promptText string

var promptTemplate = template.Must(template.New("prompt").Parse(promptText))

// Generate creates a prompt for Claude to generate release notes.
func Generate(targetTag, prevTag, commitDetails, instructions string) string {
	var buf bytes.Buffer

	data := templateData{
		TargetTag:     targetTag,
		PrevTag:       prevTag,
		CommitDetails: commitDetails,
		Instructions:  instructions,
	}

	// Template is validated at init via template.Must; execution only fails
	// on write errors to an in-memory buffer, which cannot happen in practice.
	_ = promptTemplate.Execute(&buf, data)

	return buf.String()
}
