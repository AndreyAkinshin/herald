package git

import (
	"strings"
	"testing"
)

func TestParseCommitDetails_single(t *testing.T) {
	delim := "---DELIM---"
	input := delim + "\nabc123\nfeat: add feature\n\n" + delim + "-STAT\n file.go | 10 ++++\n 1 file changed"

	got, err := parseCommitDetails(input, delim)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Commit: abc123") {
		t.Error("missing commit hash")
	}

	if !strings.Contains(got, "feat: add feature") {
		t.Error("missing commit message")
	}

	if !strings.Contains(got, "Changed files:") {
		t.Error("missing changed files section")
	}

	if !strings.Contains(got, "file.go") {
		t.Error("missing stat output")
	}
}

func TestParseCommitDetails_multiple(t *testing.T) {
	delim := "---DELIM---"
	input := delim + "\naaa111\nfirst commit\n\n" + delim + "-STAT\n a.go | 1 +\n" +
		delim + "\nbbb222\nsecond commit\n\n" + delim + "-STAT\n b.go | 2 ++\n"

	got, err := parseCommitDetails(input, delim)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Commit: aaa111") {
		t.Error("missing first commit hash")
	}

	if !strings.Contains(got, "Commit: bbb222") {
		t.Error("missing second commit hash")
	}

	if !strings.Contains(got, "---\n") {
		t.Error("missing separator between commits")
	}
}

func TestParseCommitDetails_empty(t *testing.T) {
	got, err := parseCommitDetails("", "---DELIM---")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "(no commits)" {
		t.Errorf("got %q, want %q", got, "(no commits)")
	}
}
