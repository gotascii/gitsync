# CLAUDE.md

## Project Overview

See [README.md](README.md) for project description, installation, usage, and examples. This file contains development-specific architecture and implementation details.

## Development Commands

### Build and Test

```bash
# Build the project
go build

# Run all tests
go test

# Run tests with verbose output
go test -v

# Run specific test
go test -run TestFunctionName

# Build and install globally
go install

# Check dependencies
go mod tidy
```

### Debug Commands

```bash
# Run with debug logging
DEBUG_LOG=1 ./gitsync

# Debug with Delve at key sync points
dlv debug . --init debug.dlv -- [args]

# Run with experimental go-git merge implementation
DEBUG_MERGE=1 ./gitsync
```

### Advanced Debugging

See [README.md](README.md#debugging) for basic Delve usage. The `debug.dlv` script sets breakpoints at these development-relevant points:

- After committing local changes (`main.go:139`)
- Before creating branches (`merge.go:23`) - experimental merge mode only
- Before/after reset operations (`merge.go:38,58`) - experimental merge mode only
- After staging files (`merge.go:114`) - experimental merge mode only

## Architecture

### Core Components

- **main.go**: CLI entry point and main sync orchestration logic
- **repo.go**: Repository state management and sync decision logic
- **ssh.go**: SSH authentication handling with GIT_SSH_COMMAND support
- **rename.go**: Case-sensitive file rename detection and handling
- **log.go**: Logging system with unique sync IDs
- **merge.go**: Debug merge functionality (experimental)

### Key Design Decisions

1. **Git CLI vs go-git**: Production uses Git CLI for fast-forward merges due to go-git limitations with uncommitted files. Debug mode has experimental go-git implementation.

2. **Conservative Sync Strategy**: See [README.md](README.md#how-it-works) for user-facing workflow. Refuses potentially destructive operations.

3. **SSH Authentication**: See [README.md](README.md#authentication) for usage. Parses GIT_SSH_COMMAND environment variable for custom SSH keys, falls back to system SSH agent.

### Repository State Analysis

The `Repo` struct in `repo.go` manages sync state detection:

- `CommitsSynced()`: Checks if local/remote are identical
- `FastForwardSyncNeeded()`: Determines if local is behind remote
- `PushSyncNeeded()`: Determines if local is ahead of remote
- `MergeSyncNeeded()`: Detects when branches have diverged

### Testing Strategy

Tests in `main_test.go` cover comprehensive scenarios:

- Empty repository combinations
- Uncommitted changes handling
- Fast-forward merge scenarios
- Push scenarios
- Merge conflict detection
- Go-git behavior documentation

## Important Implementation Details

1. **Uncommitted Changes**: The tool automatically commits local changes before syncing using the provided commit message.

2. **Case-sensitive Renames**: Special handling in `rename.go` for file renames that only change case (important for macOS/Windows compatibility).

3. **Logging**: Each sync operation gets a unique 8-character ID for tracking. Enable with `DEBUG_LOG=1`.

4. **Error Handling**: See [README.md](README.md#error-handling) for user-facing error scenarios. Implementation uses explicit error messages with clear guidance on manual intervention needed.

## Dependencies

- `github.com/go-git/go-git/v5` - Git operations
- `github.com/spf13/cobra` - CLI framework
- `github.com/google/uuid` - Unique ID generation
- `github.com/stretchr/testify` - Testing framework
