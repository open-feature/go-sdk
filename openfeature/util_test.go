package openfeature

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test Utils

// event handlers
var (
	h1 func(details EventDetails)
	h2 func(details EventDetails)
	h3 func(details EventDetails)
	h4 func(details EventDetails)
)

func init() {
	h1 = func(details EventDetails) {
		// noop
	}

	h2 = func(details EventDetails) {
		// noop
	}

	h3 = func(details EventDetails) {
		// noop
	}

	h4 = func(details EventDetails) {
		// noop
	}
}

var _ StateHandler = (*stateHandlerForTests)(nil)

// stateHandlerForTests is a StateHandler with callbacks
type stateHandlerForTests struct {
	initF     func(e EvaluationContext) error
	shutdownF func()
}

func (s *stateHandlerForTests) Init(e EvaluationContext) error {
	if s.initF != nil {
		return s.initF(e)
	}
	return nil
}

func (s *stateHandlerForTests) Shutdown() {
	if s.shutdownF != nil {
		s.shutdownF()
	}
}

var _ ContextAwareStateHandler = (*stateContextAwareHandlerForTests)(nil)

// stateContextAwareHandlerForTests is a StateHandler with callbacks
type stateContextAwareHandlerForTests struct {
	initF     func(context.Context, EvaluationContext) error
	shutdownF func(context.Context) error
}

func (s *stateContextAwareHandlerForTests) Init(e EvaluationContext) error {
	return s.InitWithContext(context.Background(), e)
}

func (s *stateContextAwareHandlerForTests) InitWithContext(ctx context.Context, e EvaluationContext) error {
	if s.initF != nil {
		return s.initF(ctx, e)
	}
	return nil
}

func (s *stateContextAwareHandlerForTests) Shutdown() {
	_ = s.ShutdownWithContext(context.Background())
}

func (s *stateContextAwareHandlerForTests) ShutdownWithContext(ctx context.Context) error {
	if s.shutdownF != nil {
		return s.shutdownF(ctx)
	}
	return nil
}

// ProviderEventing is an eventing implementation with invoke capability
type ProviderEventing struct {
	c chan Event
}

func (s ProviderEventing) Invoke(e Event) {
	s.c <- e
}

func (s ProviderEventing) EventChannel() <-chan Event {
	return s.c
}

func (s ProviderEventing) Close() {
	close(s.c)
}

func eventually(t *testing.T, condition func() bool, timeout, interval time.Duration, errMsg string) {
	t.Helper()
	require.Eventually(t, condition, timeout, interval, errMsg)
}
