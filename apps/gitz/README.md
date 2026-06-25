# gitz

A CLI tool to list the status of your local git repositories. Instead of navigating into each folder manually, `gitz` quickly scans a directory and tells you exactly which repositories have uncommitted changes, pending pushes, or stashes.

## Installation

```bash
go install github.com/ckinan/lab/apps/gitz@latest
```

## Usage

By default, running `gitz` without arguments will scan the current directory:

```bash
gitz
```

Or pass a specific folder path:

```bash
gitz ~/projects
```

You can also sort the output by `name` (default) or by the time of the `lastCommit` (newest first):

```bash
gitz --sort lastCommit
```

### Configuration (Optional)

If you regularly work across multiple directories or specific repositories, you can create a configuration file. `gitz` looks for configuration in two places (in this order):

1. `~/.gitz.yaml` (Simplest, classic dotfile in your home directory)
2. `~/.config/gitz/config.yaml` (Linux/Mac) or `%AppData%\gitz\config.yaml` (Windows)

Example structure:

```yaml
# Folders that CONTAIN multiple repositories
scan_dirs:
  - ~/projects
  - ~/Documents

# Exact paths to isolated repositories
exact_repos:
  - ~/.dotfiles

# Default sort order ('name' or 'lastCommit')
sort: lastCommit
```

If the configuration file exists, running `gitz` with no arguments will automatically scan all defined locations instead of the current directory.

### Output Example

```text
REPOSITORY    BRANCH    UNCOMMITTED  UNPUSHED  STASHED  LAST COMMIT
frontend      main      Yes          -         Yes      2026-06-24 13:02:02 -0500 (a1b2c3d)
backend       main      -            Yes       -        2026-06-23 11:15:00 -0500 (e4f5g6h)
scripts       master    -            -         -        2026-06-22 09:30:00 -0500 (i7j8k9l)
```
