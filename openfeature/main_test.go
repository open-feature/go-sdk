package openfeature

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/goleak"
)

// TestMain provides setup and teardown for the entire test suite. After the
// tests run it shuts down the global event executor and then verifies that no
// goroutines were leaked.
func TestMain(m *testing.M) {
	code := m.Run()

	if code == 0 {
		// Shut down the global event executor so its background goroutine is
		// stopped before we verify that nothing was leaked.
		if eventing != nil {
			eventing.(*eventExecutor).shutdown()
		}

		if err := goleak.Find(); err != nil {
			fmt.Fprintf(os.Stderr, "goroutine leak detected: %v\n", err)
			code = 1
		}
	}

	os.Exit(code)
}

// startLeakTest prepares the global event executor for a goroutine-leak test.
func startLeakTest(t *testing.T) {
	t.Helper()
	shutdownEventing()
	initSingleton()
	t.Cleanup(initSingleton)
}

// shutdownEventing shuts down the global event executor if one is set.
func shutdownEventing() {
	if eventing != nil {
		eventing.(*eventExecutor).shutdown()
	}
}
