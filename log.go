package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func generateSyncID() string {
	return uuid.New().String()[:8]
}

func debugLogWithID(syncID, format string, a ...interface{}) {
	if os.Getenv("DEBUG_LOG") != "" {
		format = "[DEBUG_LOG] " + format
		logWithID(syncID, format, a...)
	}
}

func logWithID(syncID, format string, a ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, a...)
	fmt.Printf("[%s][%s] %s\n", syncID, timestamp, message)
}
