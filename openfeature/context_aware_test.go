package openfeature

import (
	"context"
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

func (p *testContextAwareProvider) Shutdown() {}

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
		if err != context.DeadlineExceeded {
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
		if err != context.DeadlineExceeded {
			t.Errorf("Expected deadline exceeded, got: %v", err)
		}
		if event.EventType != ProviderError {
			t.Errorf("Expected ProviderError event, got: %v", event.EventType)
		}
	})
}
