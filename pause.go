package main

import (
	"fmt"
	"os"
)

func debugPause(syncID, format string, a ...interface{}) error {
	if os.Getenv("DEBUG_PAUSE") != "" {
		logWithID(syncID, "[DEBUG_PAUSE] "+format, a...)
		logWithID(syncID, "[DEBUG_PAUSE] Press Enter to continue...")
		if _, err := fmt.Scanln(); err != nil {
			logWithID(syncID, "[DEBUG_PAUSE] Error reading input: %v", err)
		}
	}
	return nil
}
