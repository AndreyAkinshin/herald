# Development Guide

Technical internals for developers and LLM agents.

## Project Overview

CLI tool that generates GitHub release notes by:
1. Finding the previous release via `gh release list`
2. Getting commit details (messages + stat) between releases
3. Generating a prompt for Claude
4. Invoking Claude CLI to generate release notes
5. Optionally updating the release via `gh release edit`

## Build System

Uses mise for task orchestration:

| Task | Description |
|------|-------------|
| `mise run build` | Build binary |
| `mise run test` | Run tests |
| `mise run check` | Run linting |
| `mise run check:fix` | Auto-fix lint issues |
| `mise run clean` | Remove artifacts |
| `mise run ci` | Full CI pipeline |

## Package Structure

```
cmd/herald/main.go              # Entry point, error handling
internal/
├── cli/cli.go                  # Orchestration, arg parsing, workflow
├── errors/errors.go            # Error types, exit codes
├── git/git.go                  # Git operations (commit details, repo detection)
├── github/github.go            # gh CLI wrapper (releases)
├── claude/claude.go            # claude CLI wrapper
├── prompt/prompt.go            # Prompt template generation
└── term/term.go                # ANSI terminal colors
```

## Key Interfaces

### cli.Config

```go
type Config struct {
    Tag          string // Target release tag (or "last")
    Instructions string // Custom instructions for Claude
    Output       string // Output file path
    NoConfirm    bool   // Skip confirmation
    NoFooter     bool   // Omit attribution footer
    DryRun       bool   // Don't update release
    Verbose      bool   // Detailed output
}
```

### github.Release

```go
type Release struct {
    TagName      string
    PublishedAt  time.Time
    IsDraft      bool
    IsPrerelease bool
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error |
| 2 | Config error |
| 3 | Environment error (not in repo, CLI unavailable) |
| 4 | User abort |

## Workflow

1. `ParseArgs()` — Parse CLI flags and positional arguments (returns nil config for `--version`)
2. `verifyEnvironment()` — Check git repo, gh CLI, claude CLI
3. `github.GetRepoInfo()` — Fetch repo name and owner/name in one call
4. `github.ListReleases()` — Fetch all releases once
5. `github.GetLatestRelease()` — Resolve "last" to latest tag (if needed)
6. `github.GetRelease()` — Verify target release exists
7. `github.FindPreviousRelease()` — Find release before target by date
8. `git.GetCommitDetails()` — Get commit messages and stat between tags
9. `prompt.Generate()` — Build Claude prompt
10. `claude.GenerateNotes()` — Invoke Claude CLI
11. Save notes to file
12. Display preview, prompt for confirmation
13. `github.UpdateReleaseBody()` — Update release via gh CLI

## Testing

Run tests with:

```bash
mise run test
```

Tests use standard Go testing patterns. Unit tests cover pure functions (same-package tests for unexported access). Integration tests require git, gh, and claude CLIs.

## Adding Features

1. CLI flags: Add to `cli.ParseArgs()` and `cli.Config`
2. Git operations: Add to `internal/git/git.go`
3. GitHub operations: Add to `internal/github/github.go`
4. Prompt changes: Modify template in `internal/prompt/prompt.go`
