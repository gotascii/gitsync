package main

import (
	"fmt"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/spf13/cobra"
)

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

func pushToRemote(repo *git.Repository, auth transport.AuthMethod) error {
	err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
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

	// Get SSH auth method if available
	auth, err := getAuthMethod()
	if err != nil {
		logWithID(syncID, "Warning: Failed to setup SSH auth: %v", err)
		// Continue without auth, will use default SSH agent
	}

	localRepo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %v", err)
	}

	logWithID(syncID, "Fetching latest changes...")
	err = localRepo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		Force:      true,
		Auth:       auth,
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
		logWithID(syncID, "Fast-forwarding to remote changes...")

		w, err := localRepo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %v", err)
		}
		if r.LocalHeadRef == nil {
			branchRef := plumbing.NewBranchReferenceName(r.LocalHeadRefName)
			err = w.Checkout(&git.CheckoutOptions{
				Branch: branchRef,
				Hash:   r.RemoteHeadRef.Hash(),
				Create: true,
			})
		} else {
			err = w.Reset(&git.ResetOptions{Commit: r.RemoteHeadRef.Hash(), Mode: git.MergeReset})
		}
		if err != nil {
			return fmt.Errorf("failed to fast-forward: %v", err)
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
		if err := pushToRemote(localRepo, auth); err != nil {
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
