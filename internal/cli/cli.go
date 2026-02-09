// Package cli provides the main CLI orchestration.
package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreyAkinshin/herald/internal/claude"
	"github.com/AndreyAkinshin/herald/internal/errors"
	"github.com/AndreyAkinshin/herald/internal/git"
	"github.com/AndreyAkinshin/herald/internal/github"
	"github.com/AndreyAkinshin/herald/internal/prompt"
	"github.com/AndreyAkinshin/herald/internal/term"
)

var tempDir = filepath.Join(os.TempDir(), "herald")

// Config holds CLI configuration.
type Config struct {
	Tag          string
	Instructions string
	Output       string
	Model        string
	Version      string
	NoConfirm    bool
	NoFooter     bool
	DryRun       bool
	Verbose      bool
}

// ParseArgs parses command-line arguments.
// Returns (nil, nil) when --version is requested (caller should print version and exit).
func ParseArgs(version string, args []string) (*Config, error) {
	cfg := &Config{}

	var showVersion bool

	fs := flag.NewFlagSet("herald", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&cfg.Output, "output", "", "")
	fs.StringVar(&cfg.Output, "o", "", "")
	fs.StringVar(&cfg.Model, "model", "", "")
	fs.StringVar(&cfg.Model, "m", "", "")
	fs.BoolVar(&cfg.NoConfirm, "no-confirm", false, "")
	fs.BoolVar(&cfg.NoFooter, "no-footer", false, "")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "")
	fs.BoolVar(&cfg.Verbose, "v", false, "")
	fs.BoolVar(&showVersion, "version", false, "")

	fs.Usage = printUsage

	// Reorder args to put flags before positional args (allows flags anywhere)
	reordered := reorderArgs(args)

	if err := fs.Parse(reordered); err != nil {
		return nil, errors.Config(err.Error())
	}

	if showVersion {
		return nil, nil
	}

	if fs.NArg() < 1 {
		return nil, errors.Config("missing required argument: tag")
	}

	cfg.Tag = fs.Arg(0)
	if fs.NArg() > 1 {
		cfg.Instructions = fs.Arg(1)
	}

	cfg.Version = version

	return cfg, nil
}

func printUsage() {
	var b strings.Builder

	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s — Generate GitHub release notes using Claude\n", term.BoldCyan("herald"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s\n", term.BoldYellow("USAGE"))
	fmt.Fprintf(&b, "    %s %s %s %s\n",
		term.BoldCyan("herald"),
		term.Yellow("<tag|last>"),
		term.Yellow("[\"instructions\"]"),
		term.Dim("[options]"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s\n", term.BoldYellow("ARGUMENTS"))
	fmt.Fprintf(&b, "    %s                     Release tag or %s for latest\n",
		term.Green("tag"), term.Cyan("last"))
	fmt.Fprintf(&b, "    %s            Custom instructions for Claude %s\n",
		term.Green("instructions"), term.Dim("(optional)"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s\n", term.BoldYellow("OPTIONS"))
	fmt.Fprintf(&b, "    %s %s %s    Claude model alias or full name\n",
		term.Green("-m,"), term.Green("--model"), term.Yellow("<model>"))
	fmt.Fprintf(&b, "                            %s\n",
		term.Dim("(e.g. haiku, sonnet, opus)"))
	fmt.Fprintf(&b, "    %s %s %s     Save notes to file\n",
		term.Green("-o,"), term.Green("--output"), term.Yellow("<file>"))
	fmt.Fprintf(&b, "                            %s\n",
		term.Dim("(default: <tmpdir>/herald/<repo>-<tag>.md)"))
	fmt.Fprintf(&b, "        %s        Skip confirmation prompt\n", term.Green("--no-confirm"))
	fmt.Fprintf(&b, "        %s         Omit herald attribution footer\n", term.Green("--no-footer"))
	fmt.Fprintf(&b, "        %s           Generate notes but don't update release\n", term.Green("--dry-run"))
	fmt.Fprintf(&b, "    %s %s           Detailed output\n",
		term.Green("-v,"), term.Green("--verbose"))
	fmt.Fprintf(&b, "        %s           Print version and exit\n", term.Green("--version"))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  %s\n", term.BoldYellow("EXAMPLES"))
	fmt.Fprintf(&b, "    %s\n", term.Dim("herald v1.2.0 --dry-run"))
	fmt.Fprintf(&b, "    %s\n", term.Dim("herald last \"Very detailed api section\""))
	fmt.Fprintf(&b, "    %s\n", term.Dim("herald v1.2.0 --no-confirm"))
	b.WriteString("\n")

	fmt.Fprint(os.Stderr, b.String())
}

// reorderArgs moves all flag-like arguments before positional arguments.
// The input should not include the program name.
func reorderArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	var flags, positional []string

	skipNext := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if skipNext {
			flags = append(flags, arg)
			skipNext = false

			continue
		}

		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			// Check if this flag expects a value (non-boolean flags)
			if strings.HasPrefix(arg, "-o") || strings.HasPrefix(arg, "--output") ||
				strings.HasPrefix(arg, "-m") || strings.HasPrefix(arg, "--model") {
				if !strings.Contains(arg, "=") {
					skipNext = true
				}
			}
		} else {
			positional = append(positional, arg)
		}
	}

	return append(flags, positional...)
}

// Run executes the main workflow.
func Run(cfg *Config) error {
	// Verify environment
	if err := verifyEnvironment(cfg); err != nil {
		return err
	}

	// Fetch remote tags so CI-created tags are available locally
	logVerbose(cfg, "Fetching tags...")

	if err := git.FetchTags(); err != nil {
		return err
	}

	// Fetch all releases once
	logVerbose(cfg, "Fetching releases...")

	releases, err := github.ListReleases()
	if err != nil {
		return err
	}

	// Resolve "last" to the latest release tag
	if cfg.Tag == "last" {
		logVerbose(cfg, "Resolving latest release...")

		latest, err := github.GetLatestRelease(releases)
		if err != nil {
			return err
		}

		cfg.Tag = latest.TagName
		fmt.Printf("Latest release: %s\n", term.Cyan(cfg.Tag))
	}

	// Fetch repo info (used for default output path and changelog link)
	logVerbose(cfg, "Fetching repository info...")

	repoInfo, err := github.GetRepoInfo()
	if err != nil {
		return err
	}

	// Set default output path if not specified
	if cfg.Output == "" {
		cfg.Output = filepath.Join(tempDir, repoInfo.Name+"-"+cfg.Tag+".md")
	}

	// Get release information
	logVerbose(cfg, "Fetching release %s...", cfg.Tag)

	_, err = github.GetRelease(cfg.Tag)
	if err != nil {
		return err
	}

	logVerbose(cfg, "Finding previous release...")

	prevRelease, err := github.FindPreviousRelease(releases, cfg.Tag, git.TagExists)
	if err != nil {
		return err
	}

	var prevTag string
	var commitDetails string

	if prevRelease != nil {
		logVerbose(cfg, "Previous release: %s", prevRelease.TagName)
		prevTag = prevRelease.TagName

		logVerbose(cfg, "Getting commit details...")

		commitDetails, err = git.GetCommitDetails(prevTag, cfg.Tag)
		if err != nil {
			return err
		}
	} else {
		logVerbose(cfg, "No previous release found, using full history")
		prevTag = ""

		logVerbose(cfg, "Getting commit details from root...")

		commitDetails, err = git.GetCommitDetailsFromRoot(cfg.Tag)
		if err != nil {
			return err
		}
	}

	// Generate prompt and invoke Claude
	promptText := prompt.Generate(cfg.Tag, prevTag, commitDetails, cfg.Instructions)

	// Save prompt to file
	promptPath := strings.TrimSuffix(cfg.Output, ".md") + "-prompt.md"

	outputDir := filepath.Dir(promptPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return errors.Runtime("failed to create output directory", err)
	}

	if err := os.WriteFile(promptPath, []byte(promptText), 0o644); err != nil {
		return errors.Runtime("failed to write prompt file", err)
	}

	fmt.Printf("Prompt saved to %s\n", term.Cyan(promptPath))

	if cfg.Verbose {
		fmt.Println(term.Dim("\n--- Prompt ---"))
		fmt.Println(promptText)
		fmt.Println(term.Dim("--- End Prompt ---"))
		fmt.Println()
	}

	// Remove previous release notes file so a stale result is never mistaken for a fresh one
	_ = os.Remove(cfg.Output)

	fmt.Println("Generating release notes with Claude...")

	notes, err := claude.GenerateNotes(promptText, cfg.Model)
	if err != nil {
		return err
	}

	// Append "Full Changelog" link if there's a previous release
	if prevRelease != nil {
		notes = appendFullChangelog(notes, repoInfo.NameWithOwner, prevRelease.TagName, cfg.Tag)
	}

	// Append herald attribution footer
	if !cfg.NoFooter {
		notes = appendFooter(notes, cfg.Version)
	}

	// Save to file
	if err := os.WriteFile(cfg.Output, []byte(notes), 0o644); err != nil {
		return errors.Runtime("failed to write output file", err)
	}

	fmt.Printf("Release notes saved to %s\n", term.Cyan(cfg.Output))

	// Display preview
	fmt.Println(term.Dim("\n--- Preview ---"))
	fmt.Println(notes)
	fmt.Println(term.Dim("--- End Preview ---"))

	// Handle dry-run
	if cfg.DryRun {
		fmt.Println(term.Yellow("\nDry run: release not updated"))

		return nil
	}

	// Confirm and update
	if !cfg.NoConfirm {
		if !confirm("Update release " + cfg.Tag + " with these notes?") {
			return errors.UserAbort()
		}
	}

	fmt.Println("Updating release...")

	if err := github.UpdateReleaseBody(cfg.Tag, cfg.Output); err != nil {
		return err
	}

	fmt.Println(term.Green("Release " + cfg.Tag + " updated successfully"))

	return nil
}

func verifyEnvironment(cfg *Config) error {
	logVerbose(cfg, "Verifying git repository...")

	if _, err := git.FindRepoRoot(); err != nil {
		return err
	}

	logVerbose(cfg, "Verifying gh CLI...")

	if err := github.CheckGHAvailable(); err != nil {
		return err
	}

	logVerbose(cfg, "Verifying claude CLI...")

	if err := claude.CheckClaudeAvailable(); err != nil {
		return err
	}

	return nil
}

func logVerbose(cfg *Config, format string, args ...any) {
	if cfg.Verbose {
		msg := fmt.Sprintf(format, args...)
		fmt.Println(term.Dim(msg))
	}
}

func confirm(message string) bool {
	fmt.Printf("%s %s ", term.Bold(message), term.Dim("[y/N]:"))

	reader := bufio.NewReader(os.Stdin)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

// appendFooter appends a herald attribution footer to the release notes.
func appendFooter(notes, version string) string {
	footer := fmt.Sprintf("*Release notes generated by [herald v%s](https://github.com/AndreyAkinshin/herald)*", version)
	trimmed := strings.TrimRight(notes, "\n")

	return trimmed + "\n\n" + footer + "\n"
}

// appendFullChangelog appends a "Full Changelog" link to the release notes.
func appendFullChangelog(notes, repoNameWithOwner, prevTag, currentTag string) string {
	link := fmt.Sprintf("**Full Changelog**: https://github.com/%s/compare/%s...%s",
		repoNameWithOwner, prevTag, currentTag)

	// Ensure proper spacing before the link
	trimmed := strings.TrimRight(notes, "\n")

	return trimmed + "\n\n" + link + "\n"
}
