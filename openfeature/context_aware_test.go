package openfeature

import (
	"context"
	"errors"
	"testing"
	"time"
)

// testContextAwareProvider is a test provider that implements ContextAwareStateHandler
type testContextAwareProvider struct {
	FeatureProvider
	stateContextAwareHandlerForTests
}

func newTestContextAwareProvider(t testing.TB, initDelay time.Duration, customShutdownDelay ...time.Duration) *testContextAwareProvider {
	t.Helper()
	shutdownDelay := initDelay
	if len(customShutdownDelay) == 1 {
		shutdownDelay = customShutdownDelay[0]
	}
	return &testContextAwareProvider{
		FeatureProvider: NoopProvider{},
		stateContextAwareHandlerForTests: stateContextAwareHandlerForTests{
			initF: func(ctx context.Context, _ EvaluationContext) error {
				select {
				case <-time.After(initDelay):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			shutdownF: func(ctx context.Context) error {
				select {
				case <-time.After(shutdownDelay):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			shutdownErrCh: make(chan error, 2),
		},
	}
}

func TestContextAwareInitialization(t *testing.T) {
	installIsolatedAPI(t)

	t.Run("fast provider succeeds within timeout", func(t *testing.T) {
		fastProvider := newTestContextAwareProvider(t, 50*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, fastProvider)
		if err != nil {
			t.Errorf("Expected fast provider to succeed, got error: %v", err)
		}
	})

	t.Run("slow provider times out", func(t *testing.T) {
		slowProvider := newTestContextAwareProvider(t, 800*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, slowProvider)
		if err == nil {
			t.Error("Expected timeout error but got success")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context deadline exceeded, got: %v", err)
		}
	})

	t.Run("async initialization returns immediately", func(t *testing.T) {
		asyncProvider := newTestContextAwareProvider(t, 200*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		start := time.Now()
		err := SetProviderWithContext(ctx, asyncProvider)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Async setup should not fail: %v", err)
		}
		if elapsed > 100*time.Millisecond {
			t.Errorf("Async setup took too long: %v", elapsed)
		}
	})

	t.Run("cancelled context does not fail async setup", func(t *testing.T) {
		// The non-blocking variants hand the context to the background
		// initialization; they do not return an error when it is cancelled.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		if err := SetProviderWithContext(ctx, newTestContextAwareProvider(t, 0)); err != nil {
			t.Errorf("SetProviderWithContext with cancelled context: got %v, want nil", err)
		}
		if err := SetNamedProviderWithContext(ctx, "cancelled-domain", newTestContextAwareProvider(t, 0)); err != nil {
			t.Errorf("SetNamedProviderWithContext with cancelled context: got %v, want nil", err)
		}
	})

	t.Run("named provider with context works", func(t *testing.T) {
		namedProvider := newTestContextAwareProvider(t, 50*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		err := SetNamedProviderWithContextAndWait(ctx, "test-domain", namedProvider)
		if err != nil {
			t.Errorf("Named provider should succeed: %v", err)
		}
	})

	t.Run("backward compatibility with regular provider", func(t *testing.T) {
		legacyProvider := &NoopProvider{}

		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, legacyProvider)
		if err != nil {
			t.Errorf("Legacy provider should work: %v", err)
		}
	})
}

func TestContextAwareStateHandlerDetection(t *testing.T) {
	// Test that the initializerWithContext function correctly detects ContextAwareStateHandler
	evalCtx := EvaluationContext{}

	t.Run("detects ContextAwareStateHandler", func(t *testing.T) {
		provider := newTestContextAwareProvider(t, 50*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()

		event, err := initializerWithContext(ctx, provider, evalCtx)
		if err != nil {
			t.Errorf("Context-aware provider should initialize successfully: %v", err)
		}
		if event.EventType != ProviderReady {
			t.Errorf("Expected ProviderReady event, got: %v", event.EventType)
		}
	})

	t.Run("falls back to regular StateHandler", func(t *testing.T) {
		provider := &NoopProvider{}

		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()

		event, err := initializerWithContext(ctx, provider, evalCtx)
		if err != nil {
			t.Errorf("Regular provider should initialize successfully: %v", err)
		}
		if event.EventType != ProviderReady {
			t.Errorf("Expected ProviderReady event, got: %v", event.EventType)
		}
	})

	t.Run("handles timeout in context-aware provider", func(t *testing.T) {
		provider := newTestContextAwareProvider(t, 500*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		event, err := initializerWithContext(ctx, provider, evalCtx)
		if err == nil {
			t.Error("Expected timeout error")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected deadline exceeded, got: %v", err)
		}
		if event.EventType != ProviderError {
			t.Errorf("Expected ProviderError event, got: %v", event.EventType)
		}
	})
}

func TestContextAwareShutdown(t *testing.T) {
	installIsolatedAPI(t)

	t.Run("context-aware shutdown with timeout", func(t *testing.T) {
		provider := newTestContextAwareProvider(t, 50*time.Millisecond)

		// Set the provider first
		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Now replace it to trigger shutdown
		newProvider := newTestContextAwareProvider(t, 10*time.Millisecond)
		err = SetProviderWithContextAndWait(ctx, newProvider)
		if err != nil {
			t.Errorf("Provider replacement should succeed: %v", err)
		}
	})

	t.Run("shutdown timeout handling", func(t *testing.T) {
		// Create a provider with long shutdown delay that would timeout during shutdown (not init)
		slowShutdownProvider := newTestContextAwareProvider(t, 10*time.Millisecond, 5*time.Second)

		// Set the provider first with generous timeout
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, slowShutdownProvider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Replace with new provider - shutdown happens in background, so this should succeed
		// even if the old provider takes a long time to shut down
		fastProvider := newTestContextAwareProvider(t, 10*time.Millisecond)
		err = SetProviderWithContextAndWait(ctx, fastProvider)
		if err != nil {
			t.Errorf("Provider replacement should succeed even with slow shutdown: %v", err)
		}

		// Wait a bit to let any background shutdown complete
		time.Sleep(100 * time.Millisecond)
	})
}

func TestGlobalContextAwareShutdown(t *testing.T) {
	t.Run("shutdown with context affects all providers", func(t *testing.T) {
		installIsolatedAPI(t)

		// Set up multiple providers
		defaultProvider := newTestContextAwareProvider(t, 50*time.Millisecond)
		namedProvider := newTestContextAwareProvider(t, 50*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		// Set default provider
		err := SetProviderWithContextAndWait(ctx, defaultProvider)
		if err != nil {
			t.Errorf("Default provider setup should succeed: %v", err)
		}

		// Set named provider
		err = SetNamedProviderWithContextAndWait(ctx, "test-service", namedProvider)
		if err != nil {
			t.Errorf("Named provider setup should succeed: %v", err)
		}

		// Shutdown all providers with context
		shutdownCtx, shutdownCancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer shutdownCancel()

		err = ShutdownWithContext(shutdownCtx)
		if err != nil {
			t.Errorf("Global shutdown should succeed: %v", err)
		}
	})

	t.Run("shutdown timeout handling", func(t *testing.T) {
		installIsolatedAPI(t)

		// Set up a provider with fast init but simulates long shutdown delay
		slowShutdownProvider := newTestContextAwareProvider(t, 50*time.Millisecond, 5*time.Second)

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		// Set the provider (this should succeed quickly)
		err := SetProviderWithContextAndWait(ctx, slowShutdownProvider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Try to shutdown with short timeout - this should timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer shutdownCancel()

		err = ShutdownWithContext(shutdownCtx)
		if err == nil {
			t.Error("Expected shutdown timeout error")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context deadline exceeded, got: %v", err)
		}
	})

	t.Run("backward compatibility with regular providers", func(t *testing.T) {
		installIsolatedAPI(t)

		// Set up regular (non-context-aware) providers
		defaultProvider := &NoopProvider{}
		namedProvider := &NoopProvider{}

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		// Set providers
		err := SetProviderWithContextAndWait(ctx, defaultProvider)
		if err != nil {
			t.Errorf("Default provider setup should succeed: %v", err)
		}

		err = SetNamedProviderWithContextAndWait(ctx, "test-service", namedProvider)
		if err != nil {
			t.Errorf("Named provider setup should succeed: %v", err)
		}

		// Shutdown should work even with non-context-aware providers
		shutdownCtx, shutdownCancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer shutdownCancel()

		err = ShutdownWithContext(shutdownCtx)
		if err != nil {
			t.Errorf("Global shutdown should succeed with regular providers: %v", err)
		}
	})

	t.Run("named provider shutdown error is returned", func(t *testing.T) {
		installIsolatedAPI(t)

		shutdownErr := errors.New("named provider shutdown failed")
		provider := newTestContextAwareProvider(t, 10*time.Millisecond)
		provider.shutdownF = func(context.Context) error { return shutdownErr }

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := SetNamedProviderWithContextAndWait(ctx, "failing-provider", provider)
		if err != nil {
			t.Errorf("Named provider setup should succeed: %v", err)
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer shutdownCancel()

		err = ShutdownWithContext(shutdownCtx)
		if err == nil {
			t.Fatal("Expected shutdown error")
		}
		if !errors.Is(err, shutdownErr) {
			t.Errorf("Expected shutdown error to wrap %v, got: %v", shutdownErr, err)
		}
	})
}

func TestContextPropagationFixes(t *testing.T) {
	installIsolatedAPI(t)

	t.Run("shutdown uses passed context timeout", func(t *testing.T) {
		// Create provider with fast init but slow shutdown
		provider := newTestContextAwareProvider(
			t,
			10*time.Millisecond,  // Fast init
			500*time.Millisecond, // Slow shutdown
		)

		// Set provider with long timeout - should succeed
		initCtx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := SetProviderWithContextAndWait(initCtx, provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Replace provider with short timeout - shutdown should respect the timeout
		newProvider := newTestContextAwareProvider(t, 10*time.Millisecond)

		// Use a short timeout that's shorter than the shutdown delay
		replaceCtx, replaceCancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer replaceCancel()

		start := time.Now()
		err = SetProviderWithContextAndWait(replaceCtx, newProvider)
		elapsed := time.Since(start)

		// The init should succeed quickly, shutdown happens async
		if err != nil {
			t.Errorf("Provider replacement should succeed: %v", err)
		}

		// Should complete quickly since init is fast and shutdown is async
		if elapsed > 100*time.Millisecond {
			t.Errorf("Provider replacement took too long: %v (expected < 100ms)", elapsed)
		}

		// Wait for shutdown to complete and capture its error
		select {
		case shutdownErr := <-provider.shutdownErrCh:
			if !errors.Is(shutdownErr, context.DeadlineExceeded) {
				t.Errorf("Expected shutdown to return DeadlineExceeded, got: %v", shutdownErr)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for shutdown to complete")
		}
	})

	t.Run("shutdown respects context cancellation", func(t *testing.T) {
		installIsolatedAPI(t)

		provider := newTestContextAwareProvider(
			t,
			10*time.Millisecond,
			5*time.Second, // Very slow shutdown
		)

		// Set up provider
		err := SetProviderWithContextAndWait(t.Context(), provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Create a context that we'll cancel quickly
		replaceCtx, cancel := context.WithCancel(t.Context())

		// Start provider replacement
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel() // Cancel context during operation
		}()

		newProvider := newTestContextAwareProvider(t, 10*time.Millisecond)
		err = SetProviderWithContextAndWait(replaceCtx, newProvider)
		// Should succeed because init is fast, shutdown is async
		if err != nil {
			t.Errorf("Provider replacement should succeed even with cancellation: %v", err)
		}

		// Wait for shutdown to complete and capture its error
		select {
		case shutdownErr := <-provider.shutdownErrCh:
			if !errors.Is(shutdownErr, context.Canceled) {
				t.Errorf("Expected shutdown to return Canceled, got: %v", shutdownErr)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for shutdown to complete")
		}
	})
}

func TestSimplifiedErrorHandling(t *testing.T) {
	evalCtx := EvaluationContext{}

	t.Run("context cancellation error message", func(t *testing.T) {
		provider := newTestContextAwareProvider(t, 200*time.Millisecond)

		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel immediately

		event, err := initializerWithContext(ctx, provider, evalCtx)
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
		if event.EventType != ProviderError {
			t.Errorf("Expected ProviderError event, got: %v", event.EventType)
		}
		if event.Message != "Provider initialization cancelled" {
			t.Errorf("Expected cancellation message, got: %q", event.Message)
		}
	})

	t.Run("context timeout error message", func(t *testing.T) {
		provider := newTestContextAwareProvider(t, 200*time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		event, err := initializerWithContext(ctx, provider, evalCtx)
		if err == nil {
			t.Error("Expected timeout error")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
		}
		if event.EventType != ProviderError {
			t.Errorf("Expected ProviderError event, got: %v", event.EventType)
		}
		if event.Message != "Provider initialization timed out" {
			t.Errorf("Expected timeout message, got: %q", event.Message)
		}
	})

	t.Run("provider init error takes precedence", func(t *testing.T) {
		// Create a provider that returns a ProviderInitError even with context issues
		provider := &testContextAwareProvider{
			FeatureProvider: NoopProvider{},
			stateContextAwareHandlerForTests: stateContextAwareHandlerForTests{
				initF: func(ctx context.Context, _ EvaluationContext) error {
					initError := &ProviderInitError{
						ErrorCode: ProviderFatalCode,
						Message:   "Custom provider error",
					}
					select {
					case <-time.After(50 * time.Millisecond):
						return initError
					case <-ctx.Done():
						// Still return the provider error even if context is cancelled
						return initError
					}
				},
			},
		}

		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond) // Longer than init
		defer cancel()

		event, err := initializerWithContext(ctx, provider, evalCtx)
		if err == nil {
			t.Error("Expected provider init error")
		}

		// Should get the custom provider error, not a context error
		if event.EventType != ProviderError {
			t.Errorf("Expected ProviderError event, got: %v", event.EventType)
		}
		if event.ErrorCode != ProviderFatalCode {
			t.Errorf("Expected ProviderFatalCode, got: %v", event.ErrorCode)
		}
		if event.Message != "Custom provider error" {
			t.Errorf("Expected custom error message, got: %q", event.Message)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("rapid provider switching", func(t *testing.T) {
		installIsolatedAPI(t)

		providers := []*testContextAwareProvider{
			newTestContextAwareProvider(t, 10*time.Millisecond),
			newTestContextAwareProvider(t, 15*time.Millisecond),
			newTestContextAwareProvider(t, 5*time.Millisecond),
		}

		// Rapidly switch providers
		for i, provider := range providers {
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			err := SetProviderWithContextAndWait(ctx, provider)
			cancel()

			if err != nil {
				t.Errorf("Provider %d setup should succeed: %v", i, err)
			}
		}

		// Let any pending shutdowns complete
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("concurrent operations with different contexts", func(t *testing.T) {
		installIsolatedAPI(t)

		// Use channels to coordinate goroutines
		done := make(chan error, 2)

		// Start two concurrent provider operations
		go func() {
			ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
			defer cancel()

			provider := newTestContextAwareProvider(t, 50*time.Millisecond)
			err := SetProviderWithContextAndWait(ctx, provider)
			done <- err
		}()

		go func() {
			time.Sleep(25 * time.Millisecond) // Start slightly later
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			provider := newTestContextAwareProvider(t, 30*time.Millisecond)
			err := SetNamedProviderWithContextAndWait(ctx, "concurrent-test", provider)
			done <- err
		}()

		// Wait for both to complete
		for i := range 2 {
			if err := <-done; err != nil {
				t.Errorf("Concurrent operation %d failed: %v", i, err)
			}
		}
	})
}
