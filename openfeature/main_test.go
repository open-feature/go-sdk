package openfeature

import (
	"context"
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
	resetSingleton()
	t.Cleanup(resetSingleton)
}

// installIsolatedAPI replaces the global evaluation API and event executor with
// fresh, isolated instances for the duration of the test (or subtest) and
// returns the new API for further configuration. The previous globals are
// restored and the executor is shut down via t.Cleanup.
func installIsolatedAPI(t *testing.T) *EvaluationAPI {
	t.Helper()

	testAPI := newAPI()
	originalAPI := apiInstance.Swap(testAPI)

	t.Cleanup(func() {
		// ShutdownWithContext (and similar) can reinitialize the global event
		// executor via resetSingleton; shut that replacement down too so it
		// doesn't leak.
		_ = testAPI.Shutdown(context.Background()) //nolint:usetesting
		if current := apiInstance.Swap(originalAPI); current != testAPI {
			_ = current.Shutdown(context.Background()) //nolint:usetesting
		}
	})

	return testAPI
}
