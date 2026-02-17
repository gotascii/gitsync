package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addRemoteRepo(t *testing.T, localPath string, remotePath string) {
	localRepo, err := git.PlainOpen(localPath)
	assert.NoError(t, err)

	_, err = localRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remotePath},
	})
	assert.NoError(t, err)
}

func setupRemoteRepo(t *testing.T) string {
	path := t.TempDir()
	_, err := git.PlainInit(path, true)
	assert.NoError(t, err)

	_, err = git.PlainOpen(path)
	assert.NoError(t, err)

	return path
}

func setupLocalRepo(t *testing.T) (string, *git.Repository) {
	path := t.TempDir()
	repo, err := git.PlainInit(path, false)
	assert.NoError(t, err)

	_, err = git.PlainOpen(path)
	assert.NoError(t, err)

	return path, repo
}

func createUncommittedChange(t *testing.T, localPath string, fileName string) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	err := os.WriteFile(filepath.Join(localPath, fileName), b, 0644)
	assert.NoError(t, err)
}

func createCommit(t *testing.T, localPath string, fileName string) plumbing.Hash {
	createUncommittedChange(t, localPath, fileName)

	local, err := git.PlainOpen(localPath)
	assert.NoError(t, err)

	w, err := local.Worktree()
	assert.NoError(t, err)

	_, err = w.Add(fileName)
	assert.NoError(t, err)

	hash, err := w.Commit("commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	return hash
}

// Empty Local Repo
func TestGitSync_EmptyLocalEmptyRemoteNoUncommitted(t *testing.T) {
	localPath, localRepo := setupLocalRepo(t)
	remotePath := setupRemoteRepo(t)
	addRemoteRepo(t, localPath, remotePath)

	// Try to sync (should be no-op)
	err := GitSync(localPath, "test commit")
	assert.NoError(t, err)

	// Verify still no commits
	_, err = localRepo.Head()
	assert.Error(t, err)
}

func TestGitSync_EmptyLocalEmptyRemoteUncommitted(t *testing.T) {
	localPath, localRepo := setupLocalRepo(t)
	remotePath := setupRemoteRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	createUncommittedChange(t, localPath, "test.txt")

	// Verify we have uncommitted changes
	w, err := localRepo.Worktree()
	assert.NoError(t, err)
	status, err := w.Status()
	assert.NoError(t, err)
	assert.False(t, status.IsClean())

	// Sync
	err = GitSync(localPath, "test commit")
	assert.NoError(t, err)

	// Verify the changes were committed
	head, err := localRepo.Head()
	assert.NoError(t, err)

	commit, err := localRepo.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "test commit", commit.Message)

	// Verify the working directory is now clean
	status, err = w.Status()
	assert.NoError(t, err)
	assert.True(t, status.IsClean())

	// TODO: Test that the remote repo is updated
}

func TestGitSync_EmptyLocalNonEmptyRemoteNoUncommitted(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	emptyLocalPath, emptyLocalRepo := setupLocalRepo(t)
	addRemoteRepo(t, emptyLocalPath, remotePath)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)

	remoteHash := createCommit(t, localPath, "test.txt")

	err := localRepo.Push(&git.PushOptions{})
	assert.NoError(t, err)

	err = GitSync(emptyLocalPath, "test commit")
	assert.NoError(t, err)

	head, err := emptyLocalRepo.Head()
	assert.NoError(t, err)
	assert.Equal(t, remoteHash, head.Hash())
}

func TestGitSync_EmptyLocalNonEmptyRemoteUncommitted(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	emptyLocalPath, _ := setupLocalRepo(t)
	addRemoteRepo(t, emptyLocalPath, remotePath)
	createUncommittedChange(t, emptyLocalPath, "test.txt")

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)

	remoteHash := createCommit(t, localPath, "remote.txt")

	err := localRepo.Push(&git.PushOptions{})
	assert.NoError(t, err)

	err = GitSync(emptyLocalPath, "test commit")
	assert.NoError(t, err)

	// Re-open the repo to get the latest state
	repo, err := git.PlainOpen(emptyLocalPath)
	assert.NoError(t, err)

	// Use 'repo' for all subsequent checks
	head, err := repo.Head()
	assert.NoError(t, err)

	headCommit, err := repo.CommitObject(head.Hash())
	assert.NoError(t, err)

	// Should have 1 parent (the remote commit)
	assert.Equal(t, 1, len(headCommit.ParentHashes))
	assert.Equal(t, remoteHash, headCommit.ParentHashes[0])

	// Verify both files exist
	_, err = os.Stat(filepath.Join(emptyLocalPath, "test.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(emptyLocalPath, "remote.txt"))
	assert.NoError(t, err)

	// Verify working directory is clean
	w, err := repo.Worktree()
	assert.NoError(t, err)
	status, err := w.Status()
	assert.NoError(t, err)
	assert.True(t, status.IsClean())
}

func TestGitSync_NonEmptyLocalEmptyRemoteNoUncommitted(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	initialHash := createCommit(t, localPath, "test.txt")

	err := GitSync(localPath, "test commit")
	assert.NoError(t, err)

	head, err := localRepo.Head()
	assert.NoError(t, err)
	assert.Equal(t, initialHash, head.Hash())
}

func TestGitSync_NonEmptyLocalEmptyRemoteUncommitted(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	initialHash := createCommit(t, localPath, "test.txt")
	createUncommittedChange(t, localPath, "test2.txt")

	err := GitSync(localPath, "test commit")
	assert.NoError(t, err)

	// Get the new HEAD commit
	head, err := localRepo.Head()
	assert.NoError(t, err)

	// Verify the new commit details
	headCommit, err := localRepo.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "test commit", headCommit.Message)

	// Verify the parent is our initial commit
	assert.Equal(t, 1, len(headCommit.ParentHashes))
	assert.Equal(t, initialHash, headCommit.ParentHashes[0])

	// Verify both files exist and are tracked
	w, err := localRepo.Worktree()
	assert.NoError(t, err)

	status, err := w.Status()
	assert.NoError(t, err)

	// Working directory should be clean
	assert.True(t, status.IsClean())

	// Both files should exist on disk
	_, err = os.Stat(filepath.Join(localPath, "test.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(localPath, "test2.txt"))
	assert.NoError(t, err)
}

func TestGitSync_NonEmptyLocalNonEmptyRemoteNoUncommitted(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	hash := createCommit(t, localPath, "test.txt")

	err := localRepo.Push(&git.PushOptions{})
	assert.NoError(t, err)

	// Store the hash before sync
	beforeHash := hash

	// Run GitSync
	err = GitSync(localPath, "test commit")
	assert.NoError(t, err)

	// Verify HEAD is still at the same commit
	head, err := localRepo.Head()
	assert.NoError(t, err)
	assert.Equal(t, beforeHash, head.Hash())
}

func TestGitSync_NonEmptyLocalNonEmptyRemoteUncommitted(t *testing.T) {
	localPath, localRepo := setupLocalRepo(t)
	remotePath := setupRemoteRepo(t)
	addRemoteRepo(t, localPath, remotePath)

	// Create and push initial commit
	initialHash := createCommit(t, localPath, "initial.txt")
	var err error
	err = localRepo.Push(&git.PushOptions{})
	require.NoError(t, err)

	// Create an uncommitted change
	createUncommittedChange(t, localPath, "test.txt")

	// Verify we have uncommitted changes
	w, err := localRepo.Worktree()
	assert.NoError(t, err)
	status, err := w.Status()
	assert.NoError(t, err)
	assert.False(t, status.IsClean())

	// Run GitSync
	err = GitSync(localPath, "test commit")
	assert.NoError(t, err)

	// Get the new HEAD commit
	head, err := localRepo.Head()
	assert.NoError(t, err)

	// Verify the new commit details
	headCommit, err := localRepo.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "test commit", headCommit.Message)

	// Verify the parent is our initial commit
	assert.Equal(t, 1, len(headCommit.ParentHashes))
	assert.Equal(t, initialHash, headCommit.ParentHashes[0])

	// Verify both files exist and are tracked
	status, err = w.Status()
	assert.NoError(t, err)
	assert.True(t, status.IsClean())

	// Both files should exist on disk
	_, err = os.Stat(filepath.Join(localPath, "initial.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(localPath, "test.txt"))
	assert.NoError(t, err)

	// Verify the changes made it to the remote
	remoteRepo, err := git.PlainOpen(remotePath)
	assert.NoError(t, err)
	remoteHead, err := remoteRepo.Reference("refs/heads/master", true)
	assert.NoError(t, err)
	assert.Equal(t, head.Hash(), remoteHead.Hash())
}

func TestGitSync_NonEmptyLocalBehindNonEmptyRemote(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	initialHash := createCommit(t, localPath, "test.txt")

	err := localRepo.Push(&git.PushOptions{})
	assert.NoError(t, err)

	clonedPath := t.TempDir()
	cloned, err := git.PlainClone(clonedPath, false, &git.CloneOptions{
		URL: remotePath,
	})
	assert.NoError(t, err)

	remoteHash := createCommit(t, clonedPath, "remote.txt")
	err = cloned.Push(&git.PushOptions{})
	assert.NoError(t, err)

	// Run GitSync
	err = GitSync(localPath, "test commit")
	assert.NoError(t, err)

	// Verify local HEAD is now at remote's commit
	head, err := localRepo.Head()
	assert.NoError(t, err)
	assert.Equal(t, remoteHash, head.Hash())

	// Verify initial commit is an ancestor of our new HEAD
	initialCommit, err := localRepo.CommitObject(initialHash)
	assert.NoError(t, err)
	headCommit, err := localRepo.CommitObject(head.Hash())
	assert.NoError(t, err)
	isAncestor, err := initialCommit.IsAncestor(headCommit)
	assert.NoError(t, err)
	assert.True(t, isAncestor)

	// Verify we have both files
	_, err = os.Stat(filepath.Join(localPath, "test.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(localPath, "remote.txt"))
	assert.NoError(t, err)
}

func TestGitSync_NonEmptyLocalBehindNonEmptyRemoteUncommitted(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	createCommit(t, localPath, "initial.txt")

	err := localRepo.Push(&git.PushOptions{})
	require.NoError(t, err)

	clonedPath := t.TempDir()
	cloned, err := git.PlainClone(clonedPath, false, &git.CloneOptions{URL: remotePath})
	require.NoError(t, err)

	remoteHash := createCommit(t, clonedPath, "remote.txt")
	err = cloned.Push(&git.PushOptions{})
	require.NoError(t, err)

	createUncommittedChange(t, localPath, "uncommitted.txt")

	err = GitSync(localPath, "test commit")
	assert.NoError(t, err)

	head, err := localRepo.Head()
	assert.NoError(t, err)

	headCommit, err := localRepo.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "test commit", headCommit.Message)

	assert.Equal(t, 1, len(headCommit.ParentHashes))
	assert.Equal(t, remoteHash, headCommit.ParentHashes[0])

	_, err = os.Stat(filepath.Join(localPath, "initial.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(localPath, "remote.txt"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(localPath, "uncommitted.txt"))
	assert.NoError(t, err)

	w, err := localRepo.Worktree()
	assert.NoError(t, err)
	status, err := w.Status()
	assert.NoError(t, err)
	assert.True(t, status.IsClean())
}

func TestGitSync_MergeDetected(t *testing.T) {
	remotePath := setupRemoteRepo(t)

	localPath, localRepo := setupLocalRepo(t)
	addRemoteRepo(t, localPath, remotePath)
	createCommit(t, localPath, "initial.txt")

	err := localRepo.Push(&git.PushOptions{})
	require.NoError(t, err)

	clonedPath := t.TempDir()
	cloned, err := git.PlainClone(clonedPath, false, &git.CloneOptions{URL: remotePath})
	require.NoError(t, err)

	createCommit(t, clonedPath, "remote.txt")
	err = cloned.Push(&git.PushOptions{})
	require.NoError(t, err)

	createCommit(t, localPath, "local.txt")

	err = GitSync(localPath, "test commit")
	assert.ErrorContains(t, err, "merge detected")
}

// TestGoGit_FastForwardDropsUncommitted verifies go-git's fast-forward behavior with uncommitted files.
// Prior to v5.16.5, go-git's MergeReset dropped uncommitted files (which is why we use Git CLI in main.go).
// As of v5.16.5, go-git preserves uncommitted files during MergeReset.
func TestGoGit_FastForwardPreservesUncommitted(t *testing.T) {
	// Create local repo
	localPath, localRepo := setupLocalRepo(t)

	// Create remote repo and add it as remote
	remotePath := setupRemoteRepo(t)
	addRemoteRepo(t, localPath, remotePath)

	// Create and push initial commit
	_ = createCommit(t, localPath, "initial.txt")
	var err error
	err = localRepo.Push(&git.PushOptions{})
	require.NoError(t, err)

	// Clone remote repo to simulate another user
	clonedPath := t.TempDir()
	cloned, err := git.PlainClone(clonedPath, false, &git.CloneOptions{
		URL: remotePath,
	})
	require.NoError(t, err)

	// Create a commit in the cloned repo and push it
	_ = createCommit(t, clonedPath, "remote.txt")
	err = cloned.Push(&git.PushOptions{})
	require.NoError(t, err)

	// Create an uncommitted file in local repo
	createUncommittedChange(t, localPath, "uncommitted.txt")

	// Verify uncommitted file exists before fast-forward
	uncommittedPath := filepath.Join(localPath, "uncommitted.txt")
	require.FileExists(t, uncommittedPath)

	// Fetch remote changes
	err = localRepo.Fetch(&git.FetchOptions{})
	require.NoError(t, err)

	// Attempt fast-forward merge using go-git
	w, err := localRepo.Worktree()
	require.NoError(t, err)

	// Get remote master reference
	remoteMaster, err := localRepo.Reference("refs/remotes/origin/master", true)
	require.NoError(t, err)

	err = w.Reset(&git.ResetOptions{
		Commit: remoteMaster.Hash(),
		Mode:   git.MergeReset,
	})
	require.NoError(t, err)

	// As of v5.16.5, go-git preserves uncommitted files during MergeReset
	require.FileExists(t, uncommittedPath)
}
