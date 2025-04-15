package main

import (
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v5"
)

// Helper function to find the case-only renamed file
func findCaseOnlyRename(deletedFile string, status git.Status) bool {
	for otherFile, otherStat := range status {
		if strings.EqualFold(deletedFile, otherFile) && deletedFile != otherFile && otherStat.Worktree == git.Untracked {
			return true
		}
	}
	return false
}

// handleCaseRenames processes case-sensitive file renames
func handleCaseRenames(w *git.Worktree, status git.Status, syncID string) error {
	for file, stat := range status {
		if stat.Worktree == git.Deleted {
			if findCaseOnlyRename(file, status) {
				logWithID(syncID, "Handling case-sensitive rename for %s", file)
				_, err := w.Remove(file)
				if err != nil {
					return fmt.Errorf("failed to remove file: %v", err)
				}
			}
		}
	}
	return nil
}
