package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Helper to set up a test repository with a remote
func setupTestRepos(t *testing.T) (string, string) {
	// Create temporary directories for local and remote repos
	localPath := t.TempDir()
	remotePath := t.TempDir()

	// Initialize remote repo
	remote, err := git.PlainInit(remotePath, false)
	assert.NoError(t, err)

	// Create initial commit in remote
	remoteW, err := remote.Worktree()
	assert.NoError(t, err)

	// Create a test file and commit it
	err = os.WriteFile(filepath.Join(remotePath, "test.txt"), []byte("initial"), 0644)
	assert.NoError(t, err)

	_, err = remoteW.Add("test.txt")
	assert.NoError(t, err)

	_, err = remoteW.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Clone remote to create local repo
	_, err = git.PlainClone(localPath, false, &git.CloneOptions{
		URL: remotePath,
	})
	assert.NoError(t, err)

	return localPath, remotePath
}

// Helper to create a new commit in a repository
func createCommit(t *testing.T, repoPath, filename, content, message string) {
	r, err := git.PlainOpen(repoPath)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(repoPath, filename), []byte(content), 0644)
	assert.NoError(t, err)

	_, err = w.Add(filename)
	assert.NoError(t, err)

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)
}

func TestGitSync_CleanState(t *testing.T) {
	localPath, _ := setupTestRepos(t)

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("path", localPath, "")
	cmd.PersistentFlags().String("msg", "test sync", "")

	err := execute(cmd, nil)
	assert.NoError(t, err)

	// Verify no new commits were created
	r, err := git.PlainOpen(localPath)
	assert.NoError(t, err)

	head, err := r.Head()
	assert.NoError(t, err)

	commits := 0
	iter, err := r.Log(&git.LogOptions{From: head.Hash()})
	assert.NoError(t, err)

	err = iter.ForEach(func(c *object.Commit) error {
		commits++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, commits, "Should only have initial commit")
}

func TestGitSync_FastForward(t *testing.T) {
	localPath, remotePath := setupTestRepos(t)

	// Create a new commit in remote
	createCommit(t, remotePath, "remote.txt", "remote change", "Remote commit")

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("path", localPath, "")
	cmd.PersistentFlags().String("msg", "test sync", "")

	err := execute(cmd, nil)
	assert.NoError(t, err)

	// Verify local has fast-forwarded
	r, err := git.PlainOpen(localPath)
	assert.NoError(t, err)

	head, err := r.Head()
	assert.NoError(t, err)

	// Verify content of new file exists
	content, err := os.ReadFile(filepath.Join(localPath, "remote.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "remote change", string(content))

	// Verify no merge commits were created
	commits := 0
	iter, err := r.Log(&git.LogOptions{From: head.Hash()})
	assert.NoError(t, err)

	err = iter.ForEach(func(c *object.Commit) error {
		commits++
		if commits > 1 {
			assert.Equal(t, 1, len(c.ParentHashes), "Should not have any merge commits")
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, commits, "Should have initial + remote commit")
}

func TestGitSync_LocalChanges(t *testing.T) {
	localPath, _ := setupTestRepos(t)

	// Create a local change
	err := os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local change"), 0644)
	assert.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("path", localPath, "")
	cmd.PersistentFlags().String("msg", "test sync", "")

	err = execute(cmd, nil)
	assert.NoError(t, err)

	// Verify commit was created and pushed
	r, err := git.PlainOpen(localPath)
	assert.NoError(t, err)

	head, err := r.Head()
	assert.NoError(t, err)

	commit, err := r.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "test sync", commit.Message)

	// Verify file exists in commit
	tree, err := commit.Tree()
	assert.NoError(t, err)
	_, err = tree.File("local.txt")
	assert.NoError(t, err)
}

func TestGitSync_MergeRequired(t *testing.T) {
	localPath, remotePath := setupTestRepos(t)

	// Create a remote change
	createCommit(t, remotePath, "remote.txt", "remote change", "Remote commit")

	// Create a local change
	err := os.WriteFile(filepath.Join(localPath, "local.txt"), []byte("local change"), 0644)
	assert.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("path", localPath, "")
	cmd.PersistentFlags().String("msg", "test sync", "")

	err = execute(cmd, nil)
	assert.NoError(t, err)

	// Verify merge commit was created
	r, err := git.PlainOpen(localPath)
	assert.NoError(t, err)

	head, err := r.Head()
	assert.NoError(t, err)

	commit, err := r.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Contains(t, commit.Message, "Merge")
	assert.Len(t, commit.ParentHashes, 2, "Should be a merge commit with 2 parents")

	// Verify both files exist
	tree, err := commit.Tree()
	assert.NoError(t, err)
	_, err = tree.File("local.txt")
	assert.NoError(t, err)
	_, err = tree.File("remote.txt")
	assert.NoError(t, err)
}
