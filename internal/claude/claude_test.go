package claude

import "testing"

func TestStripPreamble_no_heading(t *testing.T) {
	input := "Just some plain text\nwith lines"
	got := stripPreamble(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestStripPreamble_text_before_h1(t *testing.T) {
	input := "Here is some preamble text\n# Release v1.0\n\nActual content"
	got := stripPreamble(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestStripPreamble_h1_with_blank_lines(t *testing.T) {
	input := "# Release v1.0\n\n\nContent after blanks"
	want := "Content after blanks"
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_h1_only(t *testing.T) {
	input := "# Just a heading"
	want := ""
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_h2_start(t *testing.T) {
	input := "## Subsection\nSome content"
	want := "## Subsection\nSome content"
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_plain_text(t *testing.T) {
	input := "no headings at all\njust text"
	got := stripPreamble(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestStripPreamble_empty(t *testing.T) {
	got := stripPreamble("")

	if got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestStripPreamble_issue_ref_not_heading(t *testing.T) {
	input := "#123 is a bug\n## Changes\nSome content"
	got := stripPreamble(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestStripPreamble_intro_before_h2(t *testing.T) {
	input := "Introduction paragraph.\n\n## Features\n- item"
	got := stripPreamble(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestStripPreamble_conversational_with_separator(t *testing.T) {
	input := "Here are the release notes:\n\n---\n\nThis release adds new features."
	want := "This release adds new features."
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_conversational_without_separator(t *testing.T) {
	input := "Here are the release notes:\n\nThis release adds new features."
	want := "This release adds new features."
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_conversational_only(t *testing.T) {
	input := "Here are the release notes:"
	want := ""
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_conversational_with_separator_only(t *testing.T) {
	input := "Here are the release notes:\n\n---"
	want := ""
	got := stripPreamble(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPreamble_thematic_break_variants(t *testing.T) {
	for _, sep := range []string{"---", "***", "___"} {
		input := "Sure, here you go:\n\n" + sep + "\n\nContent."
		want := "Content."
		got := stripPreamble(input)

		if got != want {
			t.Errorf("separator %q: got %q, want %q", sep, got, want)
		}
	}
}

func TestIsConversationalPreamble(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"Here are the release notes:", true},
		{"Sure, here you go:", true},
		{"Release notes for v1.2.0:", true},
		{"This is a normal sentence.", false},
		{"## Features:", false},
		{"# Title:", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isConversationalPreamble(tt.line)
		if got != tt.want {
			t.Errorf("isConversationalPreamble(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestIsThematicBreak(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"---", true},
		{"***", true},
		{"___", true},
		{"  ---  ", true},
		{"----", false},
		{"- - -", false},
		{"text", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isThematicBreak(tt.line)
		if got != tt.want {
			t.Errorf("isThematicBreak(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestHeadingLevel_h1(t *testing.T) {
	if got := headingLevel("# Title"); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestHeadingLevel_h3(t *testing.T) {
	if got := headingLevel("### Section"); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestHeadingLevel_issue_ref(t *testing.T) {
	if got := headingLevel("#123 bug"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestHeadingLevel_no_space(t *testing.T) {
	if got := headingLevel("##noSpace"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestHeadingLevel_too_deep(t *testing.T) {
	if got := headingLevel("####### Seven"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestHeadingLevel_empty(t *testing.T) {
	if got := headingLevel(""); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestHeadingLevel_hash_only(t *testing.T) {
	if got := headingLevel("#"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}
