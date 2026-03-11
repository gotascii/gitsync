package main

import (
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func TestFindCaseOnlyRename(t *testing.T) {
	tests := []struct {
		name        string
		deletedFile string
		status      git.Status
		want        bool
	}{
		{
			name:        "matching case-only rename exists",
			deletedFile: "readme.md",
			status: git.Status{
				"readme.md": &git.FileStatus{Worktree: git.Deleted},
				"README.md": &git.FileStatus{Worktree: git.Untracked},
			},
			want: true,
		},
		{
			name:        "no rename - different name",
			deletedFile: "readme.md",
			status: git.Status{
				"readme.md": &git.FileStatus{Worktree: git.Deleted},
				"other.md":  &git.FileStatus{Worktree: git.Untracked},
			},
			want: false,
		},
		{
			name:        "same case - not a rename",
			deletedFile: "readme.md",
			status: git.Status{
				"readme.md": &git.FileStatus{Worktree: git.Deleted},
			},
			want: false,
		},
		{
			name:        "case match but not untracked",
			deletedFile: "readme.md",
			status: git.Status{
				"readme.md": &git.FileStatus{Worktree: git.Deleted},
				"README.md": &git.FileStatus{Worktree: git.Modified},
			},
			want: false,
		},
		{
			name:        "empty status",
			deletedFile: "readme.md",
			status:      git.Status{},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCaseOnlyRename(tt.deletedFile, tt.status)
			assert.Equal(t, tt.want, result)
		})
	}
}
