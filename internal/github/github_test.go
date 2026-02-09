package github

import (
	"testing"
	"time"
)

func TestFindPreviousRelease_basic(t *testing.T) {
	releases := []Release{
		{TagName: "v1.0", PublishedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v2.0", PublishedAt: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v3.0", PublishedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
	}

	got, err := FindPreviousRelease(releases, "v3.0", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.TagName != "v2.0" {
		t.Errorf("got %q, want %q", got.TagName, "v2.0")
	}
}

func TestFindPreviousRelease_first_release(t *testing.T) {
	releases := []Release{
		{TagName: "v1.0", PublishedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	got, err := FindPreviousRelease(releases, "v1.0", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestFindPreviousRelease_tag_not_found(t *testing.T) {
	releases := []Release{
		{TagName: "v1.0", PublishedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	_, err := FindPreviousRelease(releases, "v9.0", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindPreviousRelease_skips_invalid_tags(t *testing.T) {
	releases := []Release{
		{TagName: "v1.0", PublishedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v2.0-gone", PublishedAt: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v3.0", PublishedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
	}

	validator := func(tag string) bool { return tag != "v2.0-gone" }

	got, err := FindPreviousRelease(releases, "v3.0", validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.TagName != "v1.0" {
		t.Errorf("got %q, want %q", got.TagName, "v1.0")
	}
}

func TestFindPreviousRelease_all_previous_invalid(t *testing.T) {
	releases := []Release{
		{TagName: "v1.0-gone", PublishedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v2.0", PublishedAt: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	validator := func(tag string) bool { return tag != "v1.0-gone" }

	got, err := FindPreviousRelease(releases, "v2.0", validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestGetLatestRelease_basic(t *testing.T) {
	releases := []Release{
		{TagName: "v1.0", PublishedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v3.0", PublishedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		{TagName: "v2.0", PublishedAt: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	got, err := GetLatestRelease(releases)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.TagName != "v3.0" {
		t.Errorf("got %q, want %q", got.TagName, "v3.0")
	}
}

func TestGetLatestRelease_empty(t *testing.T) {
	_, err := GetLatestRelease(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		stderr string
		want   bool
	}{
		{"non-200 OK status code: 429 Too Many Requests body: ...", true},
		{"429", true},
		{"not found", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isRateLimited(tt.stderr)
		if got != tt.want {
			t.Errorf("isRateLimited(%q) = %v, want %v", tt.stderr, got, tt.want)
		}
	}
}
