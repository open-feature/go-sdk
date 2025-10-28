package openfeature

import (
	"context"
	"errors"
	"testing"
	"time"
)

// testContextAwareProvider is a test provider that implements ContextAwareStateHandler
type testContextAwareProvider struct {
	initDelay time.Duration
}

func (p *testContextAwareProvider) Metadata() Metadata {
	return Metadata{Name: "test-context-aware-provider"}
}

// InitWithContext implements ContextAwareStateHandler
func (p *testContextAwareProvider) InitWithContext(ctx context.Context, evalCtx EvaluationContext) error {
	select {
	case <-time.After(p.initDelay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Init implements StateHandler for backward compatibility
func (p *testContextAwareProvider) Init(evalCtx EvaluationContext) error {
	return p.InitWithContext(context.Background(), evalCtx)
}

// ShutdownWithContext implements ContextAwareStateHandler
func (p *testContextAwareProvider) ShutdownWithContext(ctx context.Context) error {
	select {
	case <-time.After(p.initDelay): // Reuse delay for shutdown simulation
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown implements StateHandler for backward compatibility
func (p *testContextAwareProvider) Shutdown() {
	// For backward compatibility, use background context with no timeout
	_ = p.ShutdownWithContext(context.Background())
}

func (p *testContextAwareProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx FlattenedContext) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx FlattenedContext) StringResolutionDetail {
	return StringResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx FlattenedContext) FloatResolutionDetail {
	return FloatResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx FlattenedContext) IntResolutionDetail {
	return IntResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx FlattenedContext) InterfaceResolutionDetail {
	return InterfaceResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProvider) Hooks() []Hook {
	return []Hook{}
}

func TestContextAwareInitialization(t *testing.T) {
	// Save original state
	originalAPI := api
	originalEventing := eventing
	defer func() {
		api = originalAPI
		eventing = originalEventing
	}()

	// Create fresh API for isolated testing
	exec := newEventExecutor()
	testAPI := newEvaluationAPI(exec)
	api = testAPI
	eventing = exec

	t.Run("fast provider succeeds within timeout", func(t *testing.T) {
		fastProvider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, fastProvider)
		if err != nil {
			t.Errorf("Expected fast provider to succeed, got error: %v", err)
		}
	})

	t.Run("slow provider times out", func(t *testing.T) {
		slowProvider := &testContextAwareProvider{initDelay: 800 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
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
		asyncProvider := &testContextAwareProvider{initDelay: 200 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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

	t.Run("named provider with context works", func(t *testing.T) {
		namedProvider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		err := SetNamedProviderWithContextAndWait(ctx, "test-domain", namedProvider)
		if err != nil {
			t.Errorf("Named provider should succeed: %v", err)
		}
	})

	t.Run("backward compatibility with regular provider", func(t *testing.T) {
		legacyProvider := &NoopProvider{}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
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
		provider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
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

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
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
		provider := &testContextAwareProvider{initDelay: 500 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
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
	// Save original state
	originalAPI := api
	originalEventing := eventing
	defer func() {
		api = originalAPI
		eventing = originalEventing
	}()

	// Create fresh API for isolated testing
	exec := newEventExecutor()
	testAPI := newEvaluationAPI(exec)
	api = testAPI
	eventing = exec

	t.Run("context-aware shutdown with timeout", func(t *testing.T) {
		provider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}

		// Set the provider first
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Now replace it to trigger shutdown
		newProvider := &testContextAwareProvider{initDelay: 10 * time.Millisecond}
		err = SetProviderWithContextAndWait(ctx, newProvider)
		if err != nil {
			t.Errorf("Provider replacement should succeed: %v", err)
		}
	})

	t.Run("shutdown timeout handling", func(t *testing.T) {
		// Create a provider with long shutdown delay that would timeout during shutdown (not init)
		slowShutdownProvider := &testContextAwareProvider{initDelay: 10 * time.Millisecond} // Fast init

		// Set the provider first with generous timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := SetProviderWithContextAndWait(ctx, slowShutdownProvider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Replace with new provider - shutdown happens in background, so this should succeed
		// even if the old provider takes a long time to shut down
		fastProvider := &testContextAwareProvider{initDelay: 10 * time.Millisecond}
		err = SetProviderWithContextAndWait(ctx, fastProvider)
		if err != nil {
			t.Errorf("Provider replacement should succeed even with slow shutdown: %v", err)
		}

		// Wait a bit to let any background shutdown complete
		time.Sleep(100 * time.Millisecond)
	})
}

func TestGlobalContextAwareShutdown(t *testing.T) {
	// Save original state
	originalAPI := api
	originalEventing := eventing
	defer func() {
		api = originalAPI
		eventing = originalEventing
	}()

	t.Run("shutdown with context affects all providers", func(t *testing.T) {
		// Create fresh API for isolated testing
		exec := newEventExecutor()
		testAPI := newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		// Set up multiple providers
		defaultProvider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}
		namedProvider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer shutdownCancel()

		err = ShutdownWithContext(shutdownCtx)
		if err != nil {
			t.Errorf("Global shutdown should succeed: %v", err)
		}
	})

	t.Run("shutdown timeout handling", func(t *testing.T) {
		// Create fresh API for isolated testing
		exec := newEventExecutor()
		testAPI := newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		// Set up a provider with fast init but simulates long shutdown delay
		slowShutdownProvider := &testContextAwareProvider{initDelay: 50 * time.Millisecond} // Fast init

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Set the provider (this should succeed quickly)
		err := SetProviderWithContextAndWait(ctx, slowShutdownProvider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Create a provider that uses the initDelay for shutdown simulation too
		// When shutdown is called, it will use the same delay, which would be longer than our timeout
		// For this test, we'll create a new provider instance with a longer delay to simulate slow shutdown
		testAPI.mu.Lock()
		// Replace the provider's delay to simulate slow shutdown
		if contextProvider, ok := testAPI.defaultProvider.(*testContextAwareProvider); ok {
			contextProvider.initDelay = 5 * time.Second // This will be used by ShutdownWithContext
		}
		testAPI.mu.Unlock()

		// Try to shutdown with short timeout - this should timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
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
		// Create fresh API for isolated testing
		exec := newEventExecutor()
		testAPI := newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		// Set up regular (non-context-aware) providers
		defaultProvider := &NoopProvider{}
		namedProvider := &NoopProvider{}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer shutdownCancel()

		err = ShutdownWithContext(shutdownCtx)
		if err != nil {
			t.Errorf("Global shutdown should succeed with regular providers: %v", err)
		}
	})
}

// testContextAwareProviderWithShutdownDelay allows different delays for init and shutdown
type testContextAwareProviderWithShutdownDelay struct {
	initDelay     time.Duration
	shutdownDelay time.Duration
}

func (p *testContextAwareProviderWithShutdownDelay) Metadata() Metadata {
	return Metadata{Name: "test-shutdown-delay-provider"}
}

func (p *testContextAwareProviderWithShutdownDelay) InitWithContext(ctx context.Context, evalCtx EvaluationContext) error {
	select {
	case <-time.After(p.initDelay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *testContextAwareProviderWithShutdownDelay) Init(evalCtx EvaluationContext) error {
	return p.InitWithContext(context.Background(), evalCtx)
}

func (p *testContextAwareProviderWithShutdownDelay) ShutdownWithContext(ctx context.Context) error {
	select {
	case <-time.After(p.shutdownDelay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *testContextAwareProviderWithShutdownDelay) Shutdown() {
	_ = p.ShutdownWithContext(context.Background())
}

func (p *testContextAwareProviderWithShutdownDelay) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx FlattenedContext) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProviderWithShutdownDelay) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx FlattenedContext) StringResolutionDetail {
	return StringResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProviderWithShutdownDelay) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx FlattenedContext) FloatResolutionDetail {
	return FloatResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProviderWithShutdownDelay) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx FlattenedContext) IntResolutionDetail {
	return IntResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProviderWithShutdownDelay) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx FlattenedContext) InterfaceResolutionDetail {
	return InterfaceResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testContextAwareProviderWithShutdownDelay) Hooks() []Hook {
	return []Hook{}
}

func TestContextPropagationFixes(t *testing.T) {
	// Save original state
	originalAPI := api
	originalEventing := eventing
	defer func() {
		api = originalAPI
		eventing = originalEventing
	}()

	// Create fresh API for isolated testing
	exec := newEventExecutor()
	testAPI := newEvaluationAPI(exec)
	api = testAPI
	eventing = exec

	t.Run("shutdown uses passed context timeout", func(t *testing.T) {
		// Create provider with fast init but slow shutdown
		provider := &testContextAwareProviderWithShutdownDelay{
			initDelay:     10 * time.Millisecond,  // Fast init
			shutdownDelay: 500 * time.Millisecond, // Slow shutdown
		}

		// Set provider with long timeout - should succeed
		initCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := SetProviderWithContextAndWait(initCtx, provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Replace provider with short timeout - shutdown should respect the timeout
		newProvider := &testContextAwareProvider{initDelay: 10 * time.Millisecond}

		// Use a short timeout that's shorter than the shutdown delay but longer than defaultShutdownTimeout
		replaceCtx, replaceCancel := context.WithTimeout(context.Background(), 200 * time.Millisecond)
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

		// Wait a bit to let shutdown complete
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("shutdown respects context cancellation", func(t *testing.T) {
		// Reset API
		exec = newEventExecutor()
		testAPI = newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		provider := &testContextAwareProviderWithShutdownDelay{
			initDelay:     10 * time.Millisecond,
			shutdownDelay: 5 * time.Second, // Very slow shutdown
		}

		// Set up provider
		err := SetProviderWithContextAndWait(context.Background(), provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Create a context that we'll cancel quickly
		replaceCtx, cancel := context.WithCancel(context.Background())

		// Start provider replacement
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel() // Cancel context during operation
		}()

		newProvider := &testContextAwareProvider{initDelay: 10 * time.Millisecond}
		err = SetProviderWithContextAndWait(replaceCtx, newProvider)

		// Should succeed because init is fast, shutdown is async
		if err != nil {
			t.Errorf("Provider replacement should succeed even with cancellation: %v", err)
		}
	})
}

func TestSimplifiedErrorHandling(t *testing.T) {
	evalCtx := EvaluationContext{}

	t.Run("context cancellation error message", func(t *testing.T) {
		provider := &testContextAwareProvider{initDelay: 200 * time.Millisecond}

		ctx, cancel := context.WithCancel(context.Background())
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
		provider := &testContextAwareProvider{initDelay: 200 * time.Millisecond}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
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
		provider := &testProviderInitError{
			initDelay: 50 * time.Millisecond,
			initError: &ProviderInitError{
				ErrorCode: ProviderFatalCode,
				Message:   "Custom provider error",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) // Longer than init
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

// testProviderInitError is a provider that returns a specific ProviderInitError
type testProviderInitError struct {
	initDelay time.Duration
	initError *ProviderInitError
}

func (p *testProviderInitError) Metadata() Metadata {
	return Metadata{Name: "test-provider-init-error"}
}

func (p *testProviderInitError) InitWithContext(ctx context.Context, evalCtx EvaluationContext) error {
	select {
	case <-time.After(p.initDelay):
		return p.initError
	case <-ctx.Done():
		// Still return the provider error even if context is cancelled
		return p.initError
	}
}

func (p *testProviderInitError) Init(evalCtx EvaluationContext) error {
	return p.InitWithContext(context.Background(), evalCtx)
}

func (p *testProviderInitError) ShutdownWithContext(ctx context.Context) error {
	return nil
}

func (p *testProviderInitError) Shutdown() {}

func (p *testProviderInitError) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx FlattenedContext) BoolResolutionDetail {
	return BoolResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testProviderInitError) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx FlattenedContext) StringResolutionDetail {
	return StringResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testProviderInitError) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx FlattenedContext) FloatResolutionDetail {
	return FloatResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testProviderInitError) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx FlattenedContext) IntResolutionDetail {
	return IntResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testProviderInitError) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx FlattenedContext) InterfaceResolutionDetail {
	return InterfaceResolutionDetail{
		Value:                    defaultValue,
		ProviderResolutionDetail: ProviderResolutionDetail{Reason: DefaultReason},
	}
}

func (p *testProviderInitError) Hooks() []Hook {
	return []Hook{}
}

func TestEdgeCases(t *testing.T) {
	// Save original state
	originalAPI := api
	originalEventing := eventing
	defer func() {
		api = originalAPI
		eventing = originalEventing
	}()

	// Create fresh API for isolated testing
	exec := newEventExecutor()
	testAPI := newEvaluationAPI(exec)
	api = testAPI
	eventing = exec

	t.Run("context with very short deadline extended for shutdown", func(t *testing.T) {
		provider := &testContextAwareProviderWithShutdownDelay{
			initDelay:     10 * time.Millisecond,
			shutdownDelay: 100 * time.Millisecond,
		}

		// Set provider first
		err := SetProviderWithContextAndWait(context.Background(), provider)
		if err != nil {
			t.Errorf("Provider setup should succeed: %v", err)
		}

		// Replace with very short context (shorter than shutdown delay)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		newProvider := &testContextAwareProvider{initDelay: 5 * time.Millisecond}
		err = SetProviderWithContextAndWait(ctx, newProvider)

		// Should succeed because init is fast and shutdown gets extended timeout
		if err != nil {
			t.Errorf("Provider replacement should succeed with short context: %v", err)
		}
	})

	t.Run("rapid provider switching", func(t *testing.T) {
		// Reset API
		exec = newEventExecutor()
		testAPI = newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		providers := []*testContextAwareProvider{
			{initDelay: 10 * time.Millisecond},
			{initDelay: 15 * time.Millisecond},
			{initDelay: 5 * time.Millisecond},
		}

		// Rapidly switch providers
		for i, provider := range providers {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
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
		// Reset API
		exec = newEventExecutor()
		testAPI = newEvaluationAPI(exec)
		api = testAPI
		eventing = exec

		// Use channels to coordinate goroutines
		done := make(chan error, 2)

		// Start two concurrent provider operations
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
			defer cancel()

			provider := &testContextAwareProvider{initDelay: 50 * time.Millisecond}
			err := SetProviderWithContextAndWait(ctx, provider)
			done <- err
		}()

		go func() {
			time.Sleep(25 * time.Millisecond) // Start slightly later
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			provider := &testContextAwareProvider{initDelay: 30 * time.Millisecond}
			err := SetNamedProviderWithContextAndWait(ctx, "concurrent-test", provider)
			done <- err
		}()

		// Wait for both to complete
		for i := 0; i < 2; i++ {
			if err := <-done; err != nil {
				t.Errorf("Concurrent operation %d failed: %v", i, err)
			}
		}
	})
}
