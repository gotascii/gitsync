package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// parseSSHKeyPath extracts the identity file path from GIT_SSH_COMMAND
func parseSSHKeyPath(sshCommand string) (string, error) {
	if sshCommand == "" {
		return "", fmt.Errorf("GIT_SSH_COMMAND is empty")
	}

	// GIT_SSH_COMMAND typically looks like: "ssh -i /path/to/key -o ..."
	parts := strings.Fields(sshCommand)
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "-i" {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("no identity file (-i) found in GIT_SSH_COMMAND")
}

// getAuthMethod returns an auth method based on SSH key if available
func getAuthMethod() (transport.AuthMethod, error) {
	sshCommand := os.Getenv("GIT_SSH_COMMAND")
	if sshCommand == "" {
		return nil, nil // No auth method, will use default SSH agent
	}

	keyPath, err := parseSSHKeyPath(sshCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key path: %v", err)
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create public keys: %v", err)
	}

	return publicKeys, nil
}
