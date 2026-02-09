// Package github provides GitHub CLI (gh) wrapper functions.
package github

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"slices"
	"sort"
	"time"

	"github.com/AndreyAkinshin/herald/internal/errors"
)

// Release represents a GitHub release.
type Release struct {
	TagName      string    `json:"tagName"`
	PublishedAt  time.Time `json:"publishedAt"`
	IsDraft      bool      `json:"isDraft"`
	IsPrerelease bool      `json:"isPrerelease"`
}

// CheckGHAvailable verifies the gh CLI is installed and authenticated.
func CheckGHAvailable() error {
	cmd := exec.Command("gh", "auth", "status")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.Environment("gh CLI not available or not authenticated", err)
	}

	return nil
}

// ListReleases returns all releases for the current repository.
func ListReleases() ([]Release, error) {
	cmd := exec.Command("gh", "release", "list", "--json", "tagName,publishedAt,isDraft,isPrerelease", "--limit", "999")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.Runtime("failed to list releases", err)
	}

	var releases []Release
	if err := json.Unmarshal(stdout.Bytes(), &releases); err != nil {
		return nil, errors.Runtime("failed to parse releases", err)
	}

	return releases, nil
}

// GetRelease fetches a specific release by tag.
func GetRelease(tag string) (*Release, error) {
	cmd := exec.Command("gh", "release", "view", tag, "--json", "tagName,publishedAt,isDraft,isPrerelease")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.Runtime("failed to get release "+tag, err)
	}

	var release Release
	if err := json.Unmarshal(stdout.Bytes(), &release); err != nil {
		return nil, errors.Runtime("failed to parse release", err)
	}

	return &release, nil
}

// FindPreviousRelease finds the release published immediately before the given tag.
// The tagValidator function is used to filter releases to only those with valid git tags.
// Returns nil (without error) if no valid previous release is found.
func FindPreviousRelease(releases []Release, tag string, tagValidator func(string) bool) (*Release, error) {
	// Sort a copy to avoid mutating the caller's slice
	sorted := slices.Clone(releases)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PublishedAt.After(sorted[j].PublishedAt)
	})

	// Find target release index
	targetIdx := -1
	for i, r := range sorted {
		if r.TagName == tag {
			targetIdx = i

			break
		}
	}

	if targetIdx == -1 {
		return nil, errors.Runtime("release "+tag+" not found", nil)
	}

	// Find the next release with a valid git tag
	for i := targetIdx + 1; i < len(sorted); i++ {
		if tagValidator == nil || tagValidator(sorted[i].TagName) {
			return &sorted[i], nil
		}
	}

	// No valid previous release found
	return nil, nil
}

// GetLatestRelease returns the most recently published release.
func GetLatestRelease(releases []Release) (*Release, error) {
	if len(releases) == 0 {
		return nil, errors.Runtime("no releases found", nil)
	}

	sorted := slices.Clone(releases)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PublishedAt.After(sorted[j].PublishedAt)
	})

	return &sorted[0], nil
}

// UpdateReleaseBody updates the release notes for a given tag.
func UpdateReleaseBody(tag, notesFile string) error {
	cmd := exec.Command("gh", "release", "edit", tag, "--notes-file", notesFile)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.Runtime("failed to update release "+tag, err)
	}

	return nil
}

// RepoInfo holds the repository metadata from gh repo view.
type RepoInfo struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
}

// GetRepoInfo returns the repository name and owner/name in a single gh call.
func GetRepoInfo() (*RepoInfo, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "name,nameWithOwner")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.Runtime("failed to get repository info", err)
	}

	var info RepoInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, errors.Runtime("failed to parse repository info", err)
	}

	return &info, nil
}
