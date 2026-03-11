package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSSHKeyPath(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantPath   string
		wantErrMsg string
	}{
		{
			name:     "standard ssh command with key",
			input:    "ssh -i /home/user/.ssh/id_rsa",
			wantPath: "/home/user/.ssh/id_rsa",
		},
		{
			name:     "ssh command with extra flags before key",
			input:    "ssh -o StrictHostKeyChecking=no -i /path/to/key",
			wantPath: "/path/to/key",
		},
		{
			name:     "ssh command with extra flags after key",
			input:    "ssh -i /path/to/key -o StrictHostKeyChecking=no",
			wantPath: "/path/to/key",
		},
		{
			name:       "empty string",
			input:      "",
			wantErrMsg: "GIT_SSH_COMMAND is empty",
		},
		{
			name:       "no -i flag",
			input:      "ssh -o StrictHostKeyChecking=no",
			wantErrMsg: "no identity file (-i) found in GIT_SSH_COMMAND",
		},
		{
			name:       "-i as last token",
			input:      "ssh -i",
			wantErrMsg: "no identity file (-i) found in GIT_SSH_COMMAND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := parseSSHKeyPath(tt.input)
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPath, path)
			}
		})
	}
}
