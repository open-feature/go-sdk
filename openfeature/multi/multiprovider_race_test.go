package multi

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestMultiProvider_TrackConcurrentWithShutdown verifies in-flight tracking and post-shutdown skipping.
func TestMultiProvider_TrackConcurrentWithShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracker := newMockProviderWithEvents(ctrl, "tracker")
	provider, err := NewProvider(StrategyFirstMatch, WithProvider("tracker", tracker))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	require.NoError(t, provider.InitWithContext(ctx, of.EvaluationContext{}))
	t.Cleanup(provider.Shutdown)

	started := make(chan struct{})
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()
	var calls atomic.Int64
	tracker.MockTracker.EXPECT().Track(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(
		func(context.Context, string, of.EvaluationContext, of.TrackingEventDetails) {
			if calls.Add(1) == 1 {
				close(started)
				select {
				case <-release:
				case <-ctx.Done():
				}
			}
		}).AnyTimes()

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		provider.Track(ctx, "event", of.EvaluationContext{}, of.TrackingEventDetails{})
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("tracking callback did not start")
	}

	// Keep reading the lifecycle state concurrently with shutdown as well as
	// holding the first tracking callback in flight.
	stressFinished := make(chan struct{})
	go func() {
		defer close(stressFinished)
		for range 10000 {
			provider.Track(ctx, "event", of.EvaluationContext{}, of.TrackingEventDetails{})
		}
	}()

	require.NoError(t, provider.ShutdownWithContext(ctx))
	unblock()
	for _, done := range []<-chan struct{}{finished, stressFinished} {
		select {
		case <-done:
		case <-ctx.Done():
			t.Fatal("tracking did not finish")
		}
	}
	before := calls.Load()
	require.Positive(t, before)
	provider.Track(ctx, "after-shutdown", of.EvaluationContext{}, of.TrackingEventDetails{})
	require.Equal(t, before, calls.Load(), "tracking after shutdown must be skipped")
}
