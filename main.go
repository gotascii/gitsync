package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Helper function to generate a sync run ID
func generateSyncID() string {
	return uuid.New().String()[:8]
}

// Helper function to log with timestamp and sync ID
func logWithID(syncID string, format string, a ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, a...)
	fmt.Printf("[%s][%s] %s\n", syncID, timestamp, message)
}

// Helper function to find the case-only renamed file
func findCaseOnlyRename(deletedFile string, status git.Status) bool {
	for otherFile, otherStat := range status {
		// if the file has the same case-insensitive name as the deleted one
		// && it isn't the same file
		// && the file we're looking at is new <- this might be overly defensive...
		// then it's a case-only rename
		if strings.EqualFold(deletedFile, otherFile) && deletedFile != otherFile && otherStat.Worktree == git.Untracked {
			return true
		}
	}
	return false
}

func execute(cmd *cobra.Command, args []string) error {
	syncID := generateSyncID()
	logWithID(syncID, "Executing gitsync")

	path, _ := cmd.Flags().GetString("path")

	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, _ := r.Worktree()

	// Open the index directly
	index, err := r.Storer.Index()
	if err != nil {
		return err
	}

	// Detect case-only renames
	status, err := w.Status()
	if err != nil {
		return err
	}

	// First fetch the latest changes
	logWithID(syncID, "Fetching latest changes...")
	err = r.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch failed: %v", err)
	}

	// Get references
	head, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %v", err)
	}

	refName := plumbing.ReferenceName("refs/remotes/origin/" + head.Name().Short())
	remoteBranch, err := r.Reference(refName, true)
	if err != nil {
		return fmt.Errorf("failed to get remote branch: %v", err)
	}

	// Check if we need to merge
	needsMerge := head.Hash() != remoteBranch.Hash()
	hasChanges := !status.IsClean()

	if !needsMerge && !hasChanges {
		logWithID(syncID, "Everything up to date, nothing to do")
		return nil
	}

	// If we have upstream changes but no local changes, just fast-forward
	if needsMerge && !hasChanges {
		logWithID(syncID, "Fast-forwarding to upstream changes...")
		// Update the reference to point to the remote commit
		newRef := plumbing.NewHashReference(head.Name(), remoteBranch.Hash())
		err = r.Storer.SetReference(newRef)
		if err != nil {
			return fmt.Errorf("failed to update reference: %v", err)
		}
		logWithID(syncID, "Fast-forward complete")
	} else if needsMerge {
		// Only create a merge commit if we have both upstream and local changes
		logWithID(syncID, "Merging changes...")
		_, err = w.Commit("Merge remote-tracking branch 'origin/main'", &git.CommitOptions{
			All:     true,
			Parents: []plumbing.Hash{head.Hash(), remoteBranch.Hash()},
			Author: &object.Signature{
				Name:  "gitsync",
				Email: "gitsync@local",
				When:  time.Now(),
			},
		})
		if err != nil {
			return fmt.Errorf("merge commit failed: %v", err)
		}
	}

	// Only handle case-sensitive renames and create a sync commit if we have actual changes
	if hasChanges {
		for file, stat := range status {
			if stat.Worktree == git.Deleted {
				if findCaseOnlyRename(file, status) {
					logWithID(syncID, "Handling case-sensitive rename for %s", file)
					// Remove the old file from the index without affecting the working tree
					_, err := index.Remove(file)
					if err != nil {
						return err
					}
				}
			}
		}

		// Save changes to the index
		if err := r.Storer.SetIndex(index); err != nil {
			return err
		}

		err = w.AddGlob(".")
		if err != nil {
			return err
		}

		msg, _ := cmd.Flags().GetString("msg")
		_, err = w.Commit(msg, &git.CommitOptions{
			All:               true,
			AllowEmptyCommits: false,
		})
		if err != nil {
			return err
		}

		logWithID(syncID, "Pushing changes to remote...")
		err = r.Push(&git.PushOptions{
			RemoteName: "origin",
			Progress:   os.Stdout,
		})
		if err != nil {
			logWithID(syncID, "Push error: %v", err)
			return err
		}

		logWithID(syncID, "Successfully pushed changes")
	}

	return nil
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
