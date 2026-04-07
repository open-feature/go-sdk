package openfeature

import (
	"context"
	"testing"
	"time"
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

// stateHandlerForTests is a StateHandler with callbacks
type stateHandlerForTests struct {
	initF     func(context.Context) error
	shutdownF func(context.Context) error
}

func (s *stateHandlerForTests) Init(ctx context.Context) error {
	if s.initF != nil {
		return s.initF(ctx)
	}
	return nil
}

func (s *stateHandlerForTests) Shutdown(ctx context.Context) error {
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

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}

	t.Fatalf("condition not met: %s", errMsg)
}
