package main

import (
	"os"
	"testing"
)

func TestDebugLogWithID(t *testing.T) {
	t.Run("no output when DEBUG_LOG unset", func(t *testing.T) {
		os.Unsetenv("DEBUG_LOG")
		debugLogWithID("abc123", "should not panic: %s", "test")
	})

	t.Run("outputs when DEBUG_LOG set", func(t *testing.T) {
		os.Setenv("DEBUG_LOG", "1")
		defer os.Unsetenv("DEBUG_LOG")
		debugLogWithID("abc123", "debug message: %s", "test")
	})
}
