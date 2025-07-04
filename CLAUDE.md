# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitSync is a command-line tool written in Go that automates Git repository synchronization. It handles uncommitted changes, fast-forward merges, and case-sensitive file renames across different platforms.

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

# Run with debug pausing (creates .pause_for_debug file)
DEBUG_PAUSE=1 ./gitsync

# Run with experimental go-git merge implementation
DEBUG_MERGE=1 ./gitsync
```

## Architecture

### Core Components

- **main.go**: CLI entry point and main sync orchestration logic
- **repo.go**: Repository state management and sync decision logic
- **ssh.go**: SSH authentication handling with GIT_SSH_COMMAND support
- **rename.go**: Case-sensitive file rename detection and handling
- **log.go**: Logging system with unique sync IDs
- **merge.go**: Debug merge functionality (experimental)
- **pause.go**: Debug utilities for development

### Key Design Decisions

1. **Git CLI vs go-git**: Production uses Git CLI for fast-forward merges due to go-git limitations with uncommitted files. Debug mode has experimental go-git implementation.

2. **Conservative Sync Strategy**: Refuses potentially destructive operations, explicitly handles different sync states:
   - Local and remote synced: No action
   - Local behind remote: Fast-forward merge
   - Local ahead of remote: Push
   - Branches diverged: Error (manual merge required)

3. **SSH Authentication**: Parses GIT_SSH_COMMAND environment variable for custom SSH keys, falls back to system SSH agent.

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

4. **Error Handling**: Explicit error messages for different failure scenarios, with clear guidance on manual intervention needed.

## Dependencies

- `github.com/go-git/go-git/v5` - Git operations
- `github.com/spf13/cobra` - CLI framework
- `github.com/google/uuid` - Unique ID generation
- `github.com/stretchr/testify` - Testing framework
