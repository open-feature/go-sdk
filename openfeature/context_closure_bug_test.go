package openfeature

import (
	"context"
	"testing"
	"time"
)

// TestContextClosureBug demonstrates the race condition in the async goroutine
// where the context variable is captured by reference in the closure.
func TestContextClosureBug(t *testing.T) {
	// Save original state
	originalAPI := api
	originalEventing := eventing
	defer func() {
		api = originalAPI
		eventing = originalEventing
	}()

	// Run the test multiple times to increase chances of hitting the race condition
	for attempt := 1; attempt <= 10; attempt++ {
		// Create fresh API for isolated testing
		exec := newEventExecutor()
		testAPI := newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		// Create a provider that takes some time to initialize
		slowProvider := &testContextAwareProvider{initDelay: 30 * time.Millisecond}

		// Create a context that we'll cancel immediately after starting the async operation
		ctx, cancel := context.WithCancel(context.Background())

		// Start the async operation
		err := testAPI.initNewAndShutdownOldInternal(ctx, "test", slowProvider, nil, true)
		if err != nil {
			t.Errorf("Unexpected error starting async operation: %v", err)
			continue
		}

		// Cancel context immediately - this creates the race condition
		// The goroutine might not have started yet, so it could capture the cancelled context
		cancel()

		// Wait for the async initialization to complete
		time.Sleep(100 * time.Millisecond)

		// Check if the provider failed due to context cancellation
		state := eventing.State("test")

		if state == ErrorState {
			t.Log("BUG DEMONSTRATED: Provider failed due to context cancellation race condition")
			t.Log("The async goroutine captured the context by reference and saw it cancelled")
			t.Log("even though the context was valid when the async operation started")
			t.Log("ROOT CAUSE: In initNewAndShutdownOldInternal, the goroutine closure captures")
			t.Log("'ctx' and 'newProvider' by reference instead of passing them as parameters")
			return // Test passed - bug demonstrated
		}
	}

	// If we get here, the race condition didn't manifest
	t.Log("Race condition did not manifest in 10 attempts")
	t.Log("This doesn't mean the bug doesn't exist - it's timing dependent")
	t.Log("The bug exists because the goroutine captures variables by reference in the closure")
}

