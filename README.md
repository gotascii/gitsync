# GitSync

GitSync is a command-line tool that automates Git repository synchronization, handling common scenarios like uncommitted changes, fast-forward merges, and case-sensitive file renames. It's designed to make Git synchronization simpler and more reliable.

## Features

- 🔄 Automatic handling of uncommitted changes
- ⚡ Fast-forward merges when possible
- 📝 Case-sensitive file rename detection
- 🔍 Detailed logging with unique sync IDs
- 🚫 Merge conflict detection (prevents automatic merges that would create conflicts)
- 🔁 Push/pull synchronization based on commit history

## Installation

Requires Go 1.23.2 or later.

```bash
go install github.com/gotascii/gitsync@latest
```

## Usage

Basic usage:

```bash
gitsync [flags]
```

Available flags:

- `--path` - Path to the Git repository (default: ".")
- `--msg` - Commit message for any uncommitted changes (default: "Syncing")

## How It Works

GitSync follows this workflow:

1. Commits any local uncommitted changes
2. Fetches latest changes from the remote
3. Analyzes the relationship between local and remote branches
4. Takes appropriate action based on the analysis:
   - If branches are in sync: No action needed
   - If local is ahead: Pushes changes to remote
   - If remote is ahead: Performs fast-forward merge
   - If branches have diverged: Reports merge needed (no automatic merge)

## Examples

Sync current directory with default commit message:

```bash
gitsync
```

Sync specific repository with custom commit message:

```bash
gitsync --path /path/to/repo --msg "Syncing project files"
```

## Error Handling

GitSync will exit with status code 1 and display an error message when:

- A merge would be required (branches have diverged)
- Remote repository is not accessible
- Git operations fail
- Repository path is invalid

## License

[License Type] - See LICENSE file for details
