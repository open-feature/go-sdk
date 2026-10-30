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

// TestMultiProvider_TrackConcurrentWithShutdown verifies race-free concurrent
// tracking and shutdown, and that tracking is skipped after shutdown.
func TestMultiProvider_TrackConcurrentWithShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	tracker := newMockProviderWithEvents(ctrl, "tracker")
	provider, err := NewProvider(StrategyFirstMatch, WithProvider("tracker", tracker))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	require.NoError(t, provider.InitWithContext(ctx, of.EvaluationContext{}))
	t.Cleanup(provider.Shutdown)

	var calls atomic.Int64
	reached := make(chan struct{})
	var reachedOnce sync.Once
	tracker.MockTracker.EXPECT().Track(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(
		func(context.Context, string, of.EvaluationContext, of.TrackingEventDetails) {
			calls.Add(1)
			reachedOnce.Do(func() { close(reached) })
		}).AnyTimes()

	stop := make(chan struct{})
	trackingDone := make(chan struct{})
	go func() {
		defer close(trackingDone)
		for {
			select {
			case <-stop:
				return
			default:
				provider.Track(ctx, "event", of.EvaluationContext{}, of.TrackingEventDetails{})
			}
		}
	}()
	stopTracking := sync.OnceFunc(func() {
		close(stop)
		<-trackingDone
	})
	defer stopTracking()

	select {
	case <-reached:
	case <-ctx.Done():
		t.Fatal("tracking did not reach the child provider")
	}
	require.NoError(t, provider.ShutdownWithContext(ctx))
	stopTracking()

	before := calls.Load()
	provider.Track(ctx, "after-shutdown", of.EvaluationContext{}, of.TrackingEventDetails{})
	require.Equal(t, before, calls.Load(), "tracking after shutdown must be skipped")
}

// TestMultiProvider_EventChannelConcurrentWithReinit reads EventChannel() while Init/Shutdown cycles
// replace the outbound channel, so -race catches an unguarded swap.
func TestMultiProvider_EventChannelConcurrentWithReinit(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := newEventingProvider(ctrl, "provider", nil)
	mp, err := NewProvider(StrategyFirstMatch, WithProvider("provider", provider))
	require.NoError(t, err)

	stop := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			select {
			case <-stop:
				return
			default:
				_ = mp.EventChannel()
			}
		}
	}()
	// Deferred so a failed require below still stops the reader instead of leaving it spinning.
	defer func() {
		close(stop)
		<-readerDone
	}()

	for range 5 {
		require.NoError(t, mp.InitWithContext(t.Context(), of.EvaluationContext{}))
		require.NoError(t, mp.ShutdownWithContext(t.Context()))
	}
}
