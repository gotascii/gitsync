package main

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repo struct {
	repo                    *git.Repository
	LocalHeadRef            *plumbing.Reference
	RemoteHeadRef           *plumbing.Reference
	LocalHeadRefName        string
	localCommit             *object.Commit
	remoteCommit            *object.Commit
	isLocalAncestorOfRemote bool
	isRemoteAncestorOfLocal bool
}

func NewRepo(repo *git.Repository) (*Repo, error) {
	r := &Repo{repo: repo}
	if err := r.Reload(); err != nil {
		return nil, fmt.Errorf("error creating repo: %v", err)
	}

	return r, nil
}

func (r *Repo) populateLocalHeadRef() error {
	ref, err := r.repo.Reference(plumbing.HEAD, true)
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return fmt.Errorf("resolving local HEAD ref: %v", err)
	}
	r.LocalHeadRef = ref
	return nil
}

func (r *Repo) populateLocalHeadRefName() error {
	if r.LocalHeadRef == nil {
		unresolvedRef, err := r.repo.Reference(plumbing.HEAD, false)
		if err != nil {
			return fmt.Errorf("getting unresolved local HEAD ref: %v", err)
		}
		r.LocalHeadRefName = unresolvedRef.Target().Short()
	} else {
		r.LocalHeadRefName = r.LocalHeadRef.Name().Short()
	}

	return nil
}

func (r *Repo) populateRemoteHeadRef() error {
	refString := "refs/remotes/origin/" + r.LocalHeadRefName
	refName := plumbing.ReferenceName(refString)
	ref, err := r.repo.Reference(refName, true)
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return fmt.Errorf("resolving remote HEAD ref: %v", err)
	}
	r.RemoteHeadRef = ref
	return nil
}

func (r *Repo) notEmpty() bool {
	if r.RemoteHeadRef != nil && r.LocalHeadRef != nil {
		return true
	}

	return false
}

func (r *Repo) populateCommits() error {
	if r.notEmpty() {
		localCommit, err := r.repo.CommitObject(r.LocalHeadRef.Hash())
		if err != nil {
			return fmt.Errorf("getting local commit: %v", err)
		}
		r.localCommit = localCommit

		remoteCommit, err := r.repo.CommitObject(r.RemoteHeadRef.Hash())
		if err != nil {
			return fmt.Errorf("getting remote commit: %v", err)
		}
		r.remoteCommit = remoteCommit
	}

	return nil
}

func (r *Repo) populateAncestry() error {
	if r.localCommit == nil || r.remoteCommit == nil {
		return nil
	}
	isLocalAncestorOfRemote, err := r.localCommit.IsAncestor(r.remoteCommit)
	if err != nil {
		return fmt.Errorf("failed to check ancestry: %v", err)
	}
	r.isLocalAncestorOfRemote = isLocalAncestorOfRemote

	isRemoteAncestorOfLocal, err := r.remoteCommit.IsAncestor(r.localCommit)
	if err != nil {
		return fmt.Errorf("failed to check ancestry: %v", err)
	}
	r.isRemoteAncestorOfLocal = isRemoteAncestorOfLocal

	return nil
}

func (r *Repo) Reload() error {
	if err := r.populateLocalHeadRef(); err != nil {
		return fmt.Errorf("populating local HEAD ref: %v", err)
	}
	if err := r.populateLocalHeadRefName(); err != nil {
		return fmt.Errorf("populating local HEAD ref name: %v", err)
	}
	if err := r.populateRemoteHeadRef(); err != nil {
		return fmt.Errorf("populating remote HEAD ref: %v", err)
	}
	if err := r.populateCommits(); err != nil {
		return fmt.Errorf("populating commits: %v", err)
	}
	if err := r.populateAncestry(); err != nil {
		return fmt.Errorf("populating ancestry: %v", err)
	}

	return nil
}

func (r *Repo) CommitsSynced() bool {
	if r.RemoteHeadRef == nil && r.LocalHeadRef == nil {
		return true
	}

	if r.notEmpty() && r.LocalHeadRef.Hash() == r.RemoteHeadRef.Hash() {
		return true
	}

	return false
}

func (r *Repo) FastForwardSyncNeeded() bool {
	if r.RemoteHeadRef != nil && r.LocalHeadRef == nil {
		return true
	}

	if r.notEmpty() && r.isLocalAncestorOfRemote && !r.isRemoteAncestorOfLocal {
		return true
	}

	return false
}

func (r *Repo) PushSyncNeeded() bool {
	if r.RemoteHeadRef == nil && r.LocalHeadRef != nil {
		return true
	}

	if r.notEmpty() && r.isRemoteAncestorOfLocal && !r.isLocalAncestorOfRemote {
		return true
	}

	return false
}

func (r *Repo) MergeSyncNeeded() bool {
	if r.notEmpty() && !r.isLocalAncestorOfRemote && !r.isRemoteAncestorOfLocal {
		return true
	}

	return false
}
