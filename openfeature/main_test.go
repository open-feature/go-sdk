package openfeature

import (
	"os"
	"testing"
)

// TestMain provides setup and teardown for the entire test suite
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Final cleanup: shut down the global event executor
	if eventing != nil {
		eventing.(*eventExecutor).shutdown()
	}

	os.Exit(code)
}
