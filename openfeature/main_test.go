package openfeature

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/goleak"
)

// TestMain provides setup and teardown for the entire test suite. After the
// tests run it shuts down the global event executor and then verifies that no
// goroutines were leaked, guarding against regressions of the leak described in
// https://github.com/open-feature/go-sdk/issues/471.
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
//
// It shuts down the current executor and reinitializes a fresh one so the test
// starts from a known-clean goroutine baseline regardless of what previous
// tests left behind. It also registers a t.Cleanup that reinitializes the
// singleton once more, restoring a working executor for subsequent tests after
// the test's deferred goleak verification has run (deferred calls execute
// before t.Cleanup callbacks).
//
// Leak tests should pair this with a deferred shutdown of the global executor
// (registered after the goleak verification so it runs first), e.g.:
//
//	startLeakTest(t)
//	defer goleak.VerifyNone(t)
//	defer shutdownEventing()
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
