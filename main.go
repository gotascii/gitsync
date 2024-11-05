package main

import (
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// Helper function to find the case-only renamed file
func findCaseOnlyRename(deletedFile string, status git.Status) bool {
	for otherFile, otherStat := range status {
		// if the file has the same case-insensitive name as the deleted one
		// && it isn't the same file
		// && the file we're looking at is new
		// then it's a case-only rename
		if strings.EqualFold(deletedFile, otherFile) && deletedFile != otherFile && otherStat.Worktree == git.Untracked {
			return true
		}
	}
	return false
}

func execute(cmd *cobra.Command, args []string) error {
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

	if status.IsClean() {
		return nil
	}

	for file, stat := range status {
		if stat.Worktree == git.Deleted {
			if findCaseOnlyRename(file, status) {
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

	err = r.Push(&git.PushOptions{})
	if err != nil {
		return err
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
