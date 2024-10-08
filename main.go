package main

import (
	"fmt"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

func execute(cmd *cobra.Command, args []string) error {
	path, _ := cmd.Flags().GetString("path")
	msg, _ := cmd.Flags().GetString("msg")

	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, _ := r.Worktree()
	w.AddGlob(".")

	_, err = w.Commit(msg, &git.CommitOptions{})
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
		Use:           "gitsync [path]",
		Short:         "Sync a git repository",
		RunE:          execute,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().String("msg", "Syncing", "sync commit message")
	rootCmd.PersistentFlags().String("path", ".", "path to repo")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
