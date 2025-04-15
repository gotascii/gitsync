package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func generateSyncID() string {
	return uuid.New().String()[:8]
}

func logWithID(syncID string, format string, a ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, a...)
	fmt.Printf("[%s][%s] %s\n", syncID, timestamp, message)
}
