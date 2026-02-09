# herald

Go CLI tool that generates GitHub release notes by analyzing git diffs between releases using Claude.

## Requirements

- Go 1.25+
- [gh CLI](https://cli.github.com/) authenticated
- [claude CLI](https://github.com/anthropics/claude-code) installed

## Installation

```bash
go install github.com/AndreyAkinshin/herald/cmd/herald@latest
```

Or build from source:

```bash
mise run build
```

## Usage

```
herald <tag|last> ["instructions"] [options]

Arguments:
  tag                  Release tag or "last" for latest
  instructions         Custom instructions for Claude (optional)

Options:
  -o, --output <file>  Save notes to file (default: <tmpdir>/herald/<repo>-<tag>.md)
  --no-confirm         Skip confirmation prompt
  --no-footer          Omit herald attribution footer
  --dry-run            Generate notes but don't update release
  -v, --verbose        Detailed output
  --version            Print version and exit
```

## Examples

Generate and preview release notes without updating:

```bash
herald v1.2.0 --dry-run
```

Generate notes for the latest release:

```bash
herald last --dry-run
```

Generate notes with custom instructions:

```bash
herald v1.2.0 "Very detailed api section" --dry-run
```

Generate and update release notes:

```bash
herald v1.2.0
```

Skip confirmation prompt:

```bash
herald v1.2.0 --no-confirm
```

## Using via mise

Herald can be installed as a [mise](https://mise.jdx.dev/) tool via `go:github.com/AndreyAkinshin/herald/cmd/herald`, then wrapped in a mise task for convenient per-project use.

### Local (user-wide)

Install globally so `herald` is available everywhere:

```bash
mise use -g go:github.com/AndreyAkinshin/herald/cmd/herald@latest
```

### Per-project

Add herald as a tool and create a task in your project's `mise.toml`:

```toml
[tools]
"go:github.com/AndreyAkinshin/herald/cmd/herald" = "latest"

[tasks.release-notes]
description = "Generate release notes for a tag"
usage = '''
arg "<tag>"
arg "[instructions]"
'''
run = 'herald "$usage_tag" "$usage_instructions" --dry-run'
```

Then run:

```bash
mise run release-notes v1.2.0
```

## License

MIT
