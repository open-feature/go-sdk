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
			eventing.shutdown()
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
	resetSingleton()
	t.Cleanup(resetSingleton)
}

// shutdownEventing shuts down the global event executor if one is set.
func shutdownEventing() {
	if eventing != nil {
		eventing.shutdown()
	}
}

// installIsolatedAPI replaces the global evaluation API and event executor with
// fresh, isolated instances for the duration of the test (or subtest) and
// returns the new API for further configuration. The previous globals are
// restored and the executor is shut down via t.Cleanup.
func installIsolatedAPI(t *testing.T) *EvaluationAPI {
	t.Helper()

	originalAPI := api
	originalEventing := eventing

	exec := newEventExecutor()
	testAPI := newEvaluationAPI(exec)
	api = testAPI
	eventing = exec

	t.Cleanup(func() {
		exec.shutdown()
		// ShutdownWithContext (and similar) can reinitialize the global event
		// executor via resetSingleton; shut that replacement down too so it
		// doesn't leak.
		if eventing != nil && eventing != exec {
			eventing.shutdown()
		}
		api = originalAPI
		eventing = originalEventing
	})

	return testAPI
}
