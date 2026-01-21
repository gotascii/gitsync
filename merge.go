package main

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func debugMerge(syncID string, localRepo *git.Repository, r *Repo) error {
	localWorktree, err := localRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	debugLogWithID(syncID, "LocalHeadRefName: %s", r.LocalHeadRefName)
	debugLogWithID(syncID, "LocalHeadRef: %v", r.LocalHeadRef)
	if r.RemoteHeadRef == nil {
		return fmt.Errorf("RemoteHeadRef is nil during fast-forward")
	}
	// 1. If no local branch, create it and set HEAD
	if r.LocalHeadRef == nil {
		branchRef := plumbing.NewBranchReferenceName(r.LocalHeadRefName)
		err := localRepo.Storer.SetReference(plumbing.NewHashReference(branchRef, r.RemoteHeadRef.Hash()))
		if err != nil {
			return fmt.Errorf("failed to create local branch %s: %v", r.LocalHeadRefName, err)
		}
		err = localRepo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, branchRef))
		if err != nil {
			return fmt.Errorf("failed to set HEAD to new branch %s: %v", r.LocalHeadRefName, err)
		}
		debugLogWithID(syncID, "Created local branch and set HEAD: %s", branchRef)
	}

	// 2. Some sort of ff-merge-like event
	debugLogWithID(syncID, "Doing a ff merge")

	err = localRepo.Merge(*r.RemoteHeadRef, git.MergeOptions{})
	if err != nil {
		return fmt.Errorf("failed to merge: %v", err)
	}

	// err = localWorktree.Reset(&git.ResetOptions{
	// 	Commit: r.RemoteHeadRef.Hash(),
	// 	Mode:   git.MergeReset,
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to soft reset: %v", err)
	// }

	// THIS IS A TERRIBLE IDEA
	// 3. Manually check out all files from the remote commit
	// remoteCommit, err := localRepo.CommitObject(r.RemoteHeadRef.Hash())
	// if err != nil {
	// 	return fmt.Errorf("failed to get remote commit: %v", err)
	// }
	// debugLogWithID(syncID, "Remote HEAD hash: %v, Local HEAD hash: %v", r.RemoteHeadRef.Hash(), r.LocalHeadRef.Hash())
	// tree, err := remoteCommit.Tree()
	// if err != nil {
	// 	return fmt.Errorf("failed to get remote tree: %v", err)
	// }
	// err = tree.Files().ForEach(func(f *object.File) error {
	// 	path := filepath.Join(repoPath, f.Name)
	// 	if _, err := os.Stat(path); err == nil {
	// 		// File exists locally - check if it's in the working tree
	// 		status, err := localWorktree.Status()
	// 		if err != nil {
	// 			return fmt.Errorf("failed to get status: %v", err)
	// 		}
	// 		fileStatus := status.File(f.Name)
	// 		debugLogWithID(syncID, "Status for %s - Worktree: '%c', Staging: '%c'", f.Name, fileStatus.Worktree, fileStatus.Staging)
	// 		if fileStatus.Worktree == git.Unmodified {
	// 			// File exists but is unmodified in working tree - copy it over
	// 			debugLogWithID(syncID, "Updating unchanged file from remote: %s", f.Name)
	// 		} else {
	// 			// File exists and is modified - possible conflict
	// 			return fmt.Errorf("conflict: file %s exists locally and in remote commit", f.Name)
	// 		}
	// 	}
	// 	reader, err := f.Reader()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer reader.Close()
	// 	out, err := os.Create(path)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer out.Close()
	// 	_, err = io.Copy(out, reader)
	// 	return err
	// })
	// if err != nil {
	// 	return fmt.Errorf("sync aborted due to conflict: %v", err)
	// }
	// debugLogWithID(syncID, "Checked out all files from remote commit")

	// 4. Stage all files
	if err := localWorktree.AddGlob("."); err != nil {
		return fmt.Errorf("failed to AddGlob files after checkout: %v", err)
	}

	debugLogWithID(syncID, "Files staged after fast-forward")

	return nil
}
