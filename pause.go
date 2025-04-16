package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func debugPause(syncID, format string, a ...interface{}) error {
	if os.Getenv("DEBUG_PAUSE") != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %v", err)
		}
		pauseFile := filepath.Join(cwd, ".pause_for_debug")
		err = os.WriteFile(pauseFile, []byte("Delete this file to continue..."), 0644)
		if err != nil {
			return fmt.Errorf("failed to create pause file: %v", err)
		}
		logWithID(syncID, "[DEBUG_PAUSE] "+format, a...)
		logWithID(syncID, "[DEBUG_PAUSE] delete %s to continue...", pauseFile)
		for {
			if _, err := os.Stat(pauseFile); os.IsNotExist(err) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
