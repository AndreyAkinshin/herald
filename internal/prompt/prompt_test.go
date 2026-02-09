package prompt

import (
	"strings"
	"testing"
)

func TestGenerate_with_prev_tag(t *testing.T) {
	got := Generate("v2.0", "v1.0", "commit details here", "")

	if !strings.Contains(got, "version v2.0") {
		t.Error("missing target tag in header")
	}

	if !strings.Contains(got, "git diff v1.0^..v2.0") {
		t.Error("missing diff-between-releases row")
	}

	if !strings.Contains(got, "commit details here") {
		t.Error("missing commit details")
	}
}

func TestGenerate_without_prev_tag(t *testing.T) {
	got := Generate("v1.0", "", "commit details here", "")

	if !strings.Contains(got, "version v1.0") {
		t.Error("missing target tag in header")
	}

	if strings.Contains(got, "Show diff for file between releases") {
		t.Error("should not contain diff-between-releases row when prevTag is empty")
	}

	if !strings.Contains(got, "commit details here") {
		t.Error("missing commit details")
	}
}

func TestGenerate_with_instructions(t *testing.T) {
	got := Generate("v1.0", "", "commits", "Very detailed api section")

	if !strings.Contains(got, "Custom Instructions") {
		t.Error("missing custom instructions section")
	}

	if !strings.Contains(got, "Very detailed api section") {
		t.Error("missing custom instructions text")
	}
}

func TestGenerate_without_instructions(t *testing.T) {
	got := Generate("v1.0", "", "commits", "")

	if strings.Contains(got, "Custom Instructions") {
		t.Error("should not contain custom instructions section when empty")
	}
}
