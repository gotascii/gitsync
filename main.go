// my thinking is to change up the flow (quite a bit...) and to approach it like this:
// if there are local uncommitted changes, commit them
// then, store the local head ref (which may be unchanged cuz there were no local uncommited changes)
// then, try to fetch from the remote
// then, store the remote head ref (which could be nil cuz the remote is empty)
// then compare the refs
// if they are the same there's nothing to do, exit
// if the local branch is ahead of the remote, push that up and exit
// otherwise, if the remote is ahead of local, do a fast forward and exit
// otherwise, do a proper merge
// does that makes sense?

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Helper function to generate a sync run ID
func generateSyncID() string {
	return uuid.New().String()[:8]
}

func logWithID(syncID string, format string, a ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, a...)
	fmt.Printf("[%s][%s] %s\n", syncID, timestamp, message)
}

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

func commitChanges(syncID string, localRepo *git.Repository, commitMsg string) error {
	// localWorktree is required to get Status
	localWorktree, err := localRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	// Are there uncommitted changes in localRepo?
	localStatus, err := localWorktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %v", err)
	}

	if localStatus.IsClean() {
		logWithID(syncID, "No uncommitted changes")
		return nil
	}

	// Handle local uncommitted changes
	if !localStatus.IsClean() {
		logWithID(syncID, "Committing uncommitted changes...")

		if err := handleCaseRenames(localWorktree, localStatus, syncID); err != nil {
			return fmt.Errorf("failed to handle case renames: %v", err)
		}

		if err := localWorktree.AddGlob("."); err != nil {
			return fmt.Errorf("failed to AddGlob files: %v", err)
		}

		// Now commit changes
		_, err = localWorktree.Commit(commitMsg, &git.CommitOptions{
			All:               true,
			AllowEmptyCommits: false,
			Author: &object.Signature{
				Name:  "gitsync",
				Email: "gitsync@local",
				When:  time.Now(),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to commit: %v", err)
		}
		logWithID(syncID, "Committed uncommitted changes")
	}
	return nil
}

func pushToRemote(repo *git.Repository) error {
	err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push to remote: %v", err)
	}
	return nil
}

// GitSync performs the actual git sync operation
func GitSync(repoPath string, commitMsg string) error {
	syncID := generateSyncID()
	logWithID(syncID, "GitSyncing")

	localRepo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %v", err)
	}

	logWithID(syncID, "Fetching latest changes...")
	err = localRepo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		Force:      true,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/remotes/origin/*"),
		},
	})
	if err != nil && err != transport.ErrEmptyRemoteRepository && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch error: %v", err)
	}
	logWithID(syncID, "Fetch complete")

	r, err := NewRepo(localRepo)
	if err != nil {
		return fmt.Errorf("failed to create repo: %v", err)
	}

	if r.CommitsSynced() {
		logWithID(syncID, "No unsynced commits")
	} else if r.FastForwardSyncNeeded() {
		// Fast-forward using git CLI because go-git drops uncommited new files!
		logWithID(syncID, "Fast-forwarding to remote changes...")

		// Get the remote branch name
		remoteBranch := "origin/" + r.LocalHeadRefName

		// Attempt fast-forward merge
		cmd := exec.Command("git", "-C", repoPath, "merge", "--ff-only", remoteBranch)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to fast-forward: %s: %v", string(out), err)
		}

		logWithID(syncID, "Fast-forward complete")
	} else if r.MergeSyncNeeded() {
		return fmt.Errorf("merge detected")
	}

	if err := commitChanges(syncID, localRepo, commitMsg); err != nil {
		return fmt.Errorf("failed to commit uncommitted changes: %v", err)
	}

	if err := r.Reload(); err != nil {
		return fmt.Errorf("failed to reload repo: %v", err)
	}

	if r.PushSyncNeeded() {
		logWithID(syncID, "Pushing to remote...")
		if err := pushToRemote(localRepo); err != nil {
			return fmt.Errorf("failed to push: %v", err)
		}
		logWithID(syncID, "Push complete")
	}

	logWithID(syncID, "Sync complete")

	return nil
}

func execute(cmd *cobra.Command, args []string) error {
	path, _ := cmd.Flags().GetString("path")
	msg, _ := cmd.Flags().GetString("msg")
	return GitSync(path, msg)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:          "gitsync",
		Short:        "Sync a git repository",
		RunE:         execute,
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().String("msg", "Syncing", "sync commit message")
	rootCmd.PersistentFlags().String("path", ".", "path to repo")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
